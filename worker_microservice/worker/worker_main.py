import time
import confluent_kafka
from confluent_kafka.admin import AdminClient, NewTopic
import json
import mysql.connector
import os
import sys
import requests
import logging


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

TIMEOUT_POLL_REQUEST = 5.0
BOOTSTRAP_SERVER_KAFKA = 'kafka-service:9092'
GROUP_ID = 'group2'
SUBSCRIPTION_TOPIC_NAME = "event_update"
PUBLICATION_TOPIC_NAME = "event_to_be_notified"
ACKS_KAFKA_PRODUCER_PARAMETER = 1
ATTEMPTS = 5
ENDPOINT_REST_OPEN_WEATHER = f"https://api.openweathermap.org/data/2.5/weather?lat=LATITUDE&lon=LONGITUDE&units=metric&appid=APIKEY"


class DatabaseConnector:
    def __init__(self, hostname, port, user, password, database):
        try:
            self._connection = mysql.connector.connect(
                host=hostname,
                port=port,
                user=user,
                password=password,
                database=database
            )
        except mysql.connector.Error as err:
            logger.error("MySQL Exception raised! -> " + str(err) + "\n")
            raise SystemExit

    def execute_query(self, query, params=None, select=True, commit=False):
        try:
            cursor = self._connection.cursor()
            cursor.execute(query, params)
            if not select:
                if commit:
                    cursor.close()
                    if self.commit_update():  # to make changes effective
                        return True
                    else:
                        return False
                else:
                    # in this case insert, delete or update query was executed but commit will be done later
                    cursor.close()
                    return True
            else:
                results = cursor.fetchall()
                cursor.close()
                return results
        except mysql.connector.Error as err:
            logger.error("MySQL Exception raised! -> " + str(err) + "\n")
            return False

    def commit_update(self):
        try:
            self._connection.commit()
            return True
        except mysql.connector.Error as error:
            logger.error("MySQL Exception raised! -> " + str(error) + "\n")
            try:
                self._connection.rollback()
            except Exception as exe:
                logger.error(f" MySQL Exception raised in rollback: {exe}\n")
            return False

    def close(self):
        try:
            self._connection.close()
            return True
        except mysql.connector.Error as error:
            logger.error("MySQL Exception raised! -> " + str(error) + "\n")
        return False


class KafkaProducer:

    def __init__(self, bootstrap_servers, group_id, acks):
        self._producer = confluent_kafka.Producer({
            'bootstrap.servers': bootstrap_servers,
            'group.id': group_id,
            'acks': acks
        })

    @staticmethod
    def create_topic(broker, topic_name):
        admin_conf = {'bootstrap.servers': broker}
        kadmin = AdminClient(admin_conf)
        # Create topic if not exists
        list_topics_metadata = kadmin.list_topics()
        topics = list_topics_metadata.topics  # Returns a dict()
        logger.info(f"LIST_TOPICS: {list_topics_metadata}")
        logger.info(f"TOPICS: {topics}")
        topic_names = set(topics.keys())
        logger.info(f"TOPIC_NAMES: {topic_names}")
        found = False
        for name in topic_names:
            if name == topic_name:
                logger.info(f"TOPIC {topic_name} FOUND\n")
                found = True
        if not found:
            logger.info(f"TOPIC {topic_name} NOT FOUND: CREATING IT\n")
            new_topic = NewTopic(topic_name, 1, 1)  # Number-of-partitions = 1, Number-of-replicas = 1
            kadmin.create_topics([new_topic,])
            logger.info("TOPIC CREATION STARTED!\n")

    @staticmethod
    # Optional per-message delivery callback (triggered by poll() or flush())
    # when a message has been successfully delivered or permanently
    # failed delivery (after retries).
    def __delivery_callback(err, message):
        if err:
            logger.error('%% Message failed delivery: %s\n' % err)
            raise SystemExit("Exiting after error in delivery message to Kafka broker\n")
        else:
            logger.info('%% Message delivered to %s, partition[%d] @ %d\n' %
                        (message.topic(), message.partition(), message.offset()))
            db_connect = DatabaseConnector(
                hostname=os.environ.get("HOSTNAME"),
                port=os.environ.get("PORT"),
                user=os.environ.get("USER"),
                password=os.environ.get("PASSWORD"),
                database=os.environ.get("DATABASE")
            )
            if not db_connect.execute_query(query="DELETE FROM current_work", commit=True, select=False):
                raise SystemExit
            if not db_connect.close():
                raise SystemExit("Error in closing DB connection")

    def produce_kafka_message(self, topic_name, kafka_broker, message):
        # Publish on the specific topic
        try:
            self._producer.produce(topic_name, value=message, callback=KafkaProducer.__delivery_callback)
        except BufferError:
            logger.error(
                '%% Local producer queue is full (%d messages awaiting delivery): try again\n' % len(kafka_broker))
            return False
        # Wait until the message have been delivered
        logger.error("Waiting for message to be delivered\n")
        self._producer.flush()
        return True


class KafkaConsumer:

    def __init__(self, bootstrap_servers, group_id):
        self._consumer = confluent_kafka.Consumer({
            'bootstrap.servers': bootstrap_servers,
            'group.id': group_id,
            'enable.auto.commit': False,
            'auto.offset.reset': 'latest',
            'on_commit': KafkaConsumer.__commit_completed
        })

    @staticmethod
    def __commit_completed(er, partitions):
        if er:
            logger.error(str(er))
        else:
            logger.info("Commit done!\n")
            logger.info("Committed partition offsets: " + str(partitions) + "\n")
            logger.info("Rules fetched and stored in DB in order to save current work!\n")

    def start_subscription(self, topic):
        try:
            self._consumer.subscribe([topic])
        except confluent_kafka.KafkaException as ke:
            logger.error("Kafka exception raised! -> " + str(ke) + "\n")
            self._consumer.close()
            sys.exit("Terminate after Exception raised in Kafka topic subscribe\n")
        except Exception as ke:
            logger.error("Kafka exception raised! -> " + str(ke) + "\n")
            self._consumer.close()
            sys.exit("Terminate after general exception raised in Kafka subscription\n")

    def poll_message(self):
        return self._consumer.poll(timeout=TIMEOUT_POLL_REQUEST)

    def commit_async(self):
        try:
            self._consumer.commit(asynchronous=True)
        except Exception as e:
            logger.error("Error in commit to Kafka broker! -> " + str(e) + "\n")
            return False
        return True

    def close_consumer(self):
        self._consumer.close()


# implementing pattern Singleton for SecretInitializer class
class SecretInitializer:

    _instance = None  # starter value: no object initially instantiated

    # if SecretInitializer object already instantiated, then return it without re-instantiating
    def __new__(cls):
        if not cls._instance:
            cls._instance = super().__new__(cls)
        return cls._instance

    # setting env variables for secrets
    def init_secrets(self):
        self.__init_secret('PASSWORD')
        self.__init_secret('APIKEY')

    @staticmethod
    def __init_secret(env_var_name):
        secret_path = os.environ.get(env_var_name)
        with open(secret_path, 'r') as file:
            secret_value = file.read()
        os.environ[env_var_name] = secret_value
        logger.info(f"Initialized {env_var_name}.\n")


# implementing pattern Singleton for SecretInitializer class
class RESTQuerier:

    _querier_instance = None  # starter value: no object initially instantiated

    # if RESTQuerier object is already instantiated, then return it without re-instantiating
    def __new__(cls):
        if not cls._querier_instance:
            cls._instance = super().__new__(cls)
        return cls._querier_instance

    def make_query(self, apikey, location):
        rest_call = ENDPOINT_REST_OPEN_WEATHER.replace("LATITUDE", location[1])
        rest_call = rest_call.replace("LONGITUDE", location[2])
        rest_call = rest_call.replace("APIKEY", apikey)
        logger.info(f"ENDPOINT OPEN WEATHER: {rest_call}\n")
        try:
            resp = requests.get(url=rest_call)
            resp.raise_for_status()
            response = resp.json()
            if response.get('cod') != 200:
                raise Exception('Query failed: ' + response.get('message'))
            logger.info("OPEN WEATHER RESPONSE: " + json.dumps(response) + "\n")
            return self.__format_data(response)
        except requests.JSONDecodeError as er:
            logger.error(f'JSON Decode error: {er}\n')
            raise SystemExit
        except requests.HTTPError as er:
            logger.error(f'HTTP Error: {er}\n')
            raise SystemExit
        except requests.exceptions.RequestException as er:
            logger.error(f'Request failed: {er}\n')
            raise SystemExit
        except Exception as er:
            logger.error(f'Error: {er}\n')
            raise SystemExit

    @staticmethod
    # function to formatting data returned by OpenWeather API according to our business logic
    def __format_data(data):
        output_json_dict = dict()
        output_json_dict['max_temp'] = data["main"]["temp"]
        output_json_dict['min_temp'] = data["main"]["temp"]
        output_json_dict['max_humidity'] = data["main"]["humidity"]
        output_json_dict['min_humidity'] = data["main"]["humidity"]
        output_json_dict['max_pressure'] = data["main"]["pressure"]
        output_json_dict['min_pressure'] = data["main"]["pressure"]
        output_json_dict["max_wind_speed"] = data["wind"]["speed"]
        output_json_dict["min_wind_speed"] = data["wind"]["speed"]
        output_json_dict["max_cloud"] = data["clouds"]["all"]
        output_json_dict["min_cloud"] = data["clouds"]["all"]
        direction_list = ["N", "NE", "E", " SE", "S", "SO", "O", "NO"]
        j = 0
        for i in range(0, 316, 45):
            if (i - 22.5) <= data["wind"]["deg"] <= (i + 22.5):
                output_json_dict["wind_direction"] = direction_list[j]
                break
            j = j + 1
        if data["weather"][0]["main"] == "Rain":
            output_json_dict["rain"] = True
        else:
            output_json_dict["rain"] = False
        if data["weather"][0]["main"] == "Snow":
            output_json_dict["snow"] = True
        else:
            output_json_dict["snow"] = False
        return output_json_dict


# compare values obtained from OpenWeather API call with those that have been placed into the DB
# for recoverability from faults that occur before to possibly publish violated rules
# returns the violated rules to be sent in the form of a dictionary that contains many other
# dictionary with key = user_id and value = the list of (violated rule-current value) pairs
# there is another key-value pair in the outer dictionary with key = "location" and value = array
# that contains information about the location in common for all the entries to be entered into the DB
def check_rules(api_response, db_connect):
    def check_max(key, actual_value, target_value, temp_dict):
        if actual_value > target_value:
            temp_dict[key] = actual_value

    def check_min(key, actual_value, target_value, temp_dict):
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
    if rules_list:
        event_dict = dict()
        check_functions_keys = check_functions.keys()
        for rules in rules_list:
            user_violated_rules_list = list()
            rules_json = json.loads(rules[0])
            keys_set_target = set(rules_json.keys())
            for key in keys_set_target:
                temp_dict = dict()
                if rules_json.get(key) != "null":
                    for prefix in check_functions_keys:
                        if prefix in key:
                            check_functions[prefix](key, api_response.get(key), rules_json.get(key), temp_dict)
                            break
                user_violated_rules_list.append(temp_dict)
            event_dict[rules_json.get("user_id")] = user_violated_rules_list
        json_location = rules_list[0][0]  # all entries in rules_list have the same location
        dict_location = json.loads(json_location)
        event_dict['location'] = dict_location.get('location')
        return json.dumps(event_dict)


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
        bootstrap_servers=BOOTSTRAP_SERVER_KAFKA,
        group_id=GROUP_ID,
        acks=ACKS_KAFKA_PRODUCER_PARAMETER
    )

    # creation of the topic on which to publish
    KafkaProducer.create_topic(BOOTSTRAP_SERVER_KAFKA, PUBLICATION_TOPIC_NAME)

    # instantiating Kafka consumer instance
    kafka_consumer = KafkaConsumer(
        bootstrap_servers=BOOTSTRAP_SERVER_KAFKA,
        group_id=GROUP_ID
    )

    # start Kafka subscription in order to retrieve messages written by Worker on broker
    kafka_consumer.start_subscription(SUBSCRIPTION_TOPIC_NAME)

    logger.info("Starting while true\n")

    try:
        while True:

            logger.info("New iteration!\n")

            # call to find_current_work and publish them in topic "event_to_be_sent"
            current_work = find_current_work()
            if current_work != '{}' and current_work is not False:  # {} is the JSON representation of an empty dictionary.
                while not kafka_producer.produce_kafka_message(PUBLICATION_TOPIC_NAME, BOOTSTRAP_SERVER_KAFKA, current_work):
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
                if msg.error().code() == confluent_kafka.KafkaError.UNKNOWN_TOPIC_OR_PART:
                    raise SystemExit
            else:

                # Check for Kafka message

                # each Kafka message is related to a single location, in order to reduce as much as
                # possible the number of OpenWeatherAPI query that worker must do, and it is a JSON
                # that contains many key-value pairs such as key = "location" and value = location
                # info in a list, then key = "user_id" and value = list of user_id of the users that are
                # interested in the location, and lastly many other key-value pairs with
                # key = rule_name and value = target value list for all the user according the order of
                # user id in user_id list

                # if a user is not interested in a specific rule for the location, then its rule value
                # corresponding to the user id is set at "null", while if no user is interested in a
                # specific rule for the location, then the key-value pair with key = rule name is not
                # in the Kafka message

                record_key = msg.key()
                logger.info("RECORD_KEY: " + str(record_key))
                record_value = msg.value()
                logger.info("RECORD_VALUE: " + str(record_value))
                data = json.loads(record_value)
                logger.info("DATA: " + str(data))
                # update current_work in DB
                userId_list = data.get("user_id")
                logger.info("USER_ID_LIST: " + str(userId_list))
                loc = data.get('location')
                logger.info("LOCATION: " + str(loc))

                # connection with DB and store event message to be published

                db = DatabaseConnector(
                    hostname=os.environ.get("HOSTNAME"),
                    port=os.environ.get("PORT"),
                    user=os.environ.get("USER"),
                    password=os.environ.get("PASSWORD"),
                    database=os.environ.get("DATABASE")
                )
                for i in range(0, len(userId_list)):
                    temp_dict = dict()
                    for key in set(data.keys()):
                        if key != "location":
                            temp_dict[key] = data.get(key)[i]
                    temp_dict['location'] = loc
                    json_to_insert = json.dumps(temp_dict)
                    if not db.execute_query(
                        query="INSERT INTO current_work (rules, time_stamp) VALUES (%s, CURRENT_TIMESTAMP())",
                        params=(json_to_insert, ),
                        commit=False,
                        select=False
                    ):
                        raise SystemExit
                if not db.commit_update():  # to make changes effective after inserting ALL the violated_rules
                    raise SystemExit("Error in commit updates in DB")
                db.close()

                # make commit to Kafka broker after Kafka msg has been stored in DB
                # we give some attempts to retrying commit in order to avoid potential
                # replication of the same email
                for t in range(ATTEMPTS):
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
