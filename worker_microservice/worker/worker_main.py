import time
import json
import os
import sys
from classes.database_connector import DatabaseConnector
from classes.rest_querier import RESTQuerier
from classes.secrets_initializer import SecretInitializer
from classes.kafka_producer import KafkaProducer
from classes.kafka_consumer import KafkaConsumer
from utils.logger import logger
import utils.constants as constants


# compare values obtained from OpenWeather API call with those that have been placed into the DB
# for recoverability from faults that occur before to possibly publish violated rules
# returns the violated rules to be sent in the form of a dictionary that contains many other
# dictionary with key = user_id and value = the list of (violated rule-current value) pairs
# there is another key-value pair in the outer dictionary with key = "location" and value = array
# that contains information about the location in common for all the entries to be entered into the DB
def check_rules(api_response, db_connect):
    def check_max(key, actual_value, target_value, temp_dict):
        target_value = float(target_value)
        if actual_value > target_value:
            temp_dict[key] = actual_value

    def check_min(key, actual_value, target_value, temp_dict):
        target_value = float(target_value)
        if actual_value < target_value:
            temp_dict[key] = actual_value

    def check_rain(key, actual_value, target_value, temp_dict):
        if actual_value is True:
            temp_dict[key] = actual_value

    def check_snow(key, actual_value, target_value, temp_dict):
        if actual_value is True:
            temp_dict[key] = actual_value

    def check_wind_direction(key, actual_value, target_value, temp_dict):
        if target_value == actual_value:
            temp_dict[key] = actual_value

    # Function's dict
    check_functions = {
        "max": check_max,
        "min": check_min,
        "rain": check_rain,
        "snow": check_snow,
        "wind_direction": check_wind_direction
    }

    rules_list = db_connect.execute_query("SELECT rules FROM current_work")
    if not db_connect.close():
        logger.error("Error in DB connection close!\n")
        return False
    if rules_list is False:
        logger.error("Error in DB query execution!\n")
        return False
    if rules_list:
        event_dict = dict()
        check_functions_keys = check_functions.keys()
        at_least_one_notification = False  # if no user is interested in any constraint, this remains False
        for rules in rules_list:  # every rules entry is referred to one different user
            user_violated_rules_list = list()
            rules_json = json.loads(rules[0])
            keys_set_target = set(rules_json.keys())
            one_rule_different_from_null = False  # if this variable remains False, user isn't interested in any rule
            for key in keys_set_target:
                temp_dict = dict()
                if key != "user_id" and key != "location":
                    if rules_json.get(key) != "null":
                        one_rule_different_from_null = True
                        for prefix in check_functions_keys:
                            if prefix in key:
                                check_functions[prefix](key, api_response.get(key), rules_json.get(key), temp_dict)
                                user_violated_rules_list.append(temp_dict)
                                break
            # we insert entry for the user only if he is interested in at least one constraint
            if one_rule_different_from_null:
                event_dict[rules_json.get("user_id")] = user_violated_rules_list
                at_least_one_notification = True  # since this is set to True, Worker will publish message to Notifier
        json_location = rules_list[0][0]  # all entries in rules_list have the same location
        dict_location = json.loads(json_location)
        event_dict['location'] = dict_location.get('location')
        if at_least_one_notification:
            return json.dumps(event_dict)  # we return message to send to Notifier
        else:
            return "{}"  # we return this in order to avoid message sending to Notifier


# function for recovering unchecked rules when worker goes down before publishing notification event
def find_current_work():
    db_connection = DatabaseConnector(
        hostname=os.environ.get("HOSTNAME"),
        port=os.environ.get("PORT"),
        user=os.environ.get("USER"),
        password=os.environ.get("PASSWORD"),
        database=os.environ.get("DATABASE")
    )
    result = db_connection.execute_query("SELECT rules FROM current_work")
    if result:
        # we are interested in any of the entries because we are interested in location and
        # all the entries in current_works are related to the same location
        result = result[0]
        dict_row = json.loads(result[0])  # extract json of rules to be checked and convert it into dict
        location_info = dict_row.get('location')
        # make OpenWeather API call
        querier = RESTQuerier()
        actual_weather_values = querier.make_query(apikey=os.environ.get('APIKEY'), location=location_info)
        events_to_be_sent = check_rules(actual_weather_values, db_connection)
    elif result is False:
        return result
    else:
        events_to_be_sent = "{}"
    # if no user is interested in any constraint, events_to_be_sent is "{}" because in this case check_rules return "{}"
    return events_to_be_sent


if __name__ == "__main__":

    logger.info("Start notifier main")
    secret_initializer = SecretInitializer()
    secret_initializer.init_secrets()
    logger.info("ENV variables initialization done")

    # create table current_work if not exists.
    # This table will contain many entries but all relating to the same message from the WMS
    # and therefore all with the same location
    # in particular, each entry includes a rules field that contains many key-value pairs in a JSON
    # in the form of { "user_id": value, "location": [name,lat,long,country,state],
    # "rule_name" : actual value, "other_rule_name" : actual value }
    # in the JSON there are all the rules in which the user is interested plus those in which
    # the user is not interested but for which at least one other user is interested in.
    # In this second case, the actual value of the rule is "null"
    db_connector = DatabaseConnector(
        hostname=os.environ.get("HOSTNAME"),
        port=os.environ.get("PORT"),
        user=os.environ.get("USER"),
        password=os.environ.get("PASSWORD"),
        database=os.environ.get("DATABASE")
    )
    outcome = db_connector.execute_query(
        query=
        "CREATE TABLE IF NOT EXISTS current_work (id INTEGER PRIMARY KEY AUTO_INCREMENT, rules JSON NOT NULL, time_stamp TIMESTAMP NOT NULL)",
        commit=True,
        select=False
    )
    if not outcome:
        sys.exit("Error in creating 'current_work' DB table")

    if not db_connector.close():
        sys.exit("Error in closing DB connection")

    # instantiating Kafka producer instance
    kafka_producer = KafkaProducer(
        bootstrap_servers=constants.BOOTSTRAP_SERVER_KAFKA,
        acks=constants.ACKS_KAFKA_PRODUCER_PARAMETER
    )

    # creation of the topic on which to publish (if not yet exists)
    KafkaProducer.create_topic(constants.BOOTSTRAP_SERVER_KAFKA, constants.PUBLICATION_TOPIC_NAME)

    # instantiating Kafka consumer instance
    kafka_consumer = KafkaConsumer(
        bootstrap_servers=constants.BOOTSTRAP_SERVER_KAFKA,
        group_id=constants.GROUP_ID
    )

    # start Kafka subscription in order to retrieve messages written by Worker on broker
    kafka_consumer.start_subscription(constants.SUBSCRIPTION_TOPIC_NAME)

    logger.info("Starting while true\n")

    try:
        while True:

            logger.info("New iteration!\n")

            # call to find_current_work and publish them in topic "event_to_be_sent"
            current_work = find_current_work()
            if current_work != '{}' and current_work is not False:  # {} is the JSON representation of empty dictionary.
                while not kafka_producer.produce_kafka_message(
                        constants.PUBLICATION_TOPIC_NAME, constants.BOOTSTRAP_SERVER_KAFKA, current_work):
                    pass
            else:
                logger.info("There is no backlog of work\n")

            # polling messages in Kafka topic "event_update"
            msg = kafka_consumer.poll_message()

            if msg is None:
                # No message available within timeout.
                # Initial message consumption may take up to
                # `session.timeout.ms` for the consumer group to
                # rebalance and start consuming
                logger.info("Waiting for message or event/error in poll()\n")
                continue

            elif msg.error():
                logger.info('error: {}\n'.format(msg.error()))
                if KafkaConsumer.topic_not_found(msg):
                    continue

            else:

                # consume Kafka message
                if KafkaConsumer.consume_kafka_message(msg) is False:
                    raise SystemExit("Error in commit updates in DB")

                # make commit to Kafka broker after Kafka msg has been stored in DB
                # we give some attempts to retrying commit in order to avoid potential
                # replication of the same email
                for t in range(constants.ATTEMPTS):
                    if kafka_consumer.commit_async():
                        break
                    time.sleep(1)
                else:
                    logger.error("Exhausted attempts to commit to Kafka broker!\n")
                    raise SystemExit

    except (KeyboardInterrupt, SystemExit):  # to terminate correctly with either CTRL+C or docker stop
        pass
    finally:
        # Leave group and commit final offsets
        kafka_consumer.close_consumer()
