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


def commit_completed(er, partitions):
    if er:
        logger.error(str(er))
    else:
        logger.info("Commit done!\n")
        logger.info("Committed partition offsets: " + str(partitions) + "\n")
        logger.info("Rules fetched and stored in DB in order to save current work!\n")


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
                    self.commit_update()  # to make changes effective
                    return True
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
        except mysql.connector.Error as error:
            logger.error("MySQL Exception raised! -> " + str(error) + "\n")
            try:
                self._connection.rollback()
            except Exception as exe:
                logger.error(f" MySQL Exception raised in rollback: {exe}\n")
        raise SystemExit

    def close(self):
        try:
            self._connection.close()
        except mysql.connector.Error as error:
            logger.error("MySQL Exception raised! -> " + str(error) + "\n")
        raise SystemExit


def make_query(query):
    try:
        resp = requests.get(url=query)
        resp.raise_for_status()
        response = resp.json()
        if response.get('cod') != 200:
            raise Exception('Query failed: ' + response.get('message'))
        logger.info(json.dumps(response) + "\n")
        return response
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


# compare values obtained from OpenWeather API call with those that have been placed into the DB
# for recoverability from faults that occur before to possibly publish violated rules
# returns the violated rules to be sent in the form of a dictionary that contains many other
# dictionary with key = user_id and value = the list of (violated rule-current value) pairs
# there is another key-value pair in the outer dictionary with key = "location" and value = array
# that contains information about the location in common for all the entries to be entered into the DB
def check_rules(db_cursor, api_response):
    db_cursor.execute("SELECT rules FROM current_work WHERE worker_id = %s", (str(worker_id),))
    rules_list = db_cursor.fetchall()
    event_dict = dict()
    for rules in rules_list:
        user_violated_rules_list = list()
        rules_json = json.loads(rules[0])
        keys_set_target = set(rules_json.keys())
        for key in keys_set_target:
            temp_dict = dict()
            if "max" in key and rules_json.get(key) != "null":
                if api_response.get(key) > rules_json.get(key):
                    temp_dict[key] = api_response.get(key)
            elif "min" in key and rules_json.get(key) != "null":
                if api_response.get(key) < rules_json.get(key):
                    temp_dict[key] = api_response.get(key)
            elif key == "rain" and rules_json.get(key) != "null" and api_response.get("rain") == True:
                temp_dict[key] = api_response.get(key)
            elif key == "snow" and rules_json.get(key) != "null" and api_response.get("snow") == True:
                temp_dict[key] = api_response.get(key)
            elif key == "wind_direction" and rules_json.get(key) != "null" and rules_json.get(key) == api_response.get(key):
                temp_dict[key] = api_response.get(key)
            user_violated_rules_list.append(temp_dict)
        event_dict[rules_json.get("user_id")] = user_violated_rules_list
    json_location = rules_list[0][0]  # all entries in rules_list have the same location
    dict_location = json.loads(json_location)
    event_dict['location'] = dict_location.get('location')
    return json.dumps(event_dict)


# function to formatting data returned by OpenWeather API according to our business logic
def format_data(data):
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


# function for recovering unchecked rules when worker goes down before publishing notification event
def find_current_work():
    try:
        with (mysql.connector.connect(host=os.environ.get('HOSTNAME'), port=os.environ.get('PORT'),
                                      user=os.environ.get('USER'), password=os.environ.get('PASSWORD'),
                                      database=os.environ.get('DATABASE')) as db_conn):
            # without a buffered cursor, the results are "lazily" loaded, meaning that "fetchone"
            # actually only fetches one row from the full result set of the query.
            # When you will use the same cursor again, it will complain that you still have
            # n-1 results (where n is the result set amount) waiting to be fetched.
            # However, when you use a buffered cursor the connector fetches ALL rows behind the scenes,
            # and you just take one from the connector so the mysql db won't complain.
            # buffered=True is needed because we next will use db_cursor as first parameter of check_rules
            db_cursor = db_conn.cursor(buffered=True)
            db_cursor.execute("SELECT rules FROM current_work WHERE worker_id = %s", (str(worker_id),))
            result = db_cursor.fetchone()
            if result:
                dict_row = json.loads(result[0])
                # all entries in current_works are related to the same location
                location_info = dict_row.get('location')
                # make OpenWeather API call
                apikey = os.environ.get('APIKEY')
                rest_call = f"https://api.openweathermap.org/data/2.5/weather?lat={location_info[1]}&lon={location_info[2]}&units=metric&appid={apikey}"
                data = make_query(rest_call)
                formatted_data = format_data(data)
                events_to_be_sent = check_rules(db_cursor, formatted_data)
            else:
                events_to_be_sent = "{}"
    except mysql.connector.Error as error:
        logger.error("Exception raised! -> " + str(error) + "\n")
        try:
            db_conn.rollback()
        except Exception as ex:
            logger.error(f"Exception raised in rollback: {ex}\n")
        return False
    return events_to_be_sent


class KafkaProducer:

    def __init__(self, bootstrap_servers, group_id, acks):
        self._producer = confluent_kafka.Producer({
            'bootstrap.servers': bootstrap_servers,
            'group.id': group_id,
            'acks': acks
        })

    @staticmethod
    def _create_topic(broker, topic_name):
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
                found = True
        if not found:
            new_topic = NewTopic(topic_name, 1, 1)  # Number-of-partitions = 1, Number-of-replicas = 1
            kadmin.create_topics([new_topic,])

    def produce_kafka_message(self, topic_name, kafka_broker, message):
        # Publish on the specific topic
        try:
            self._producer.produce(topic_name, value=message, callback=self._delivery_callback)
        except BufferError:
            logger.error(
                '%% Local producer queue is full (%d messages awaiting delivery): try again\n' % len(kafka_broker))
            return False
        # Wait until the message have been delivered
        logger.error("Waiting for message to be delivered\n")
        self._producer.flush()
        return True

    @staticmethod
    # Optional per-message delivery callback (triggered by poll() or flush())
    # when a message has been successfully delivered or permanently
    # failed delivery (after retries).
    def _delivery_callback(err, msg):
        if err:
            logger.error('%% Message failed delivery: %s\n' % err)
            raise SystemExit("Exiting after error in delivery message to Kafka broker\n")
        else:
            logger.info('%% Message delivered to %s, partition[%d] @ %d\n' %
                        (msg.topic(), msg.partition(), msg.offset()))
            db_connect = DatabaseConnector(
                hostname=os.environ.get("HOSTNAME"),
                port=os.environ.get("PORT"),
                user=os.environ.get("USER"),
                password=os.environ.get("PASSWORD"),
                database=os.environ.get("DATABASE")
            )
            if not db_connect.execute_query(query="DELETE FROM current_work", commit=True, select=False):
                raise SystemExit
            db_connect.close()


class KafkaConsumer:

    def __init__(self, bootstrap_servers, group_id, commit_completed_callback):
        self._consumer = confluent_kafka.Consumer({
            'bootstrap.servers': bootstrap_servers,
            'group.id': group_id,
            'enable.auto.commit': False,
            'auto.offset.reset': 'latest',
            'on_commit': commit_completed_callback
        })

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
        if not hasattr(cls, '_instance'):
            cls._instance = super().__new__(cls)
        return cls._instance

    # setting env variables for secrets
    def init_secrets(self):
        self._init_secret('PASSWORD')
        self._init_secret('APIKEY')

    @staticmethod
    def _init_secret(env_var_name):
        secret_path = os.environ.get(env_var_name)
        with open(secret_path, 'r') as file:
            secret_value = file.read()
        os.environ[env_var_name] = secret_value
        logger.info(f"Initialized {env_var_name}.\n")


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
        sys.exit()

    db_connector.close()

    # instantiating Kafka producer instance
    kafka_producer = KafkaProducer(
        bootstrap_servers=BOOTSTRAP_SERVER_KAFKA,
        group_id=GROUP_ID,
        acks=ACKS_KAFKA_PRODUCER_PARAMETER
    )

    # instantiating Kafka consumer instance
    kafka_consumer = KafkaConsumer(
        bootstrap_servers=BOOTSTRAP_SERVER_KAFKA,
        group_id=GROUP_ID,
        commit_completed_callback=commit_completed
    )

    # start Kafka subscription in order to retrieve messages written by Worker on broker
    kafka_consumer.start_subscription(SUBSCRIPTION_TOPIC_NAME)

    logger.info("Starting while true\n")

    try:
        while True:

            logger.info("New iteration!\n")

            # call to find_current_work and publish them in topic "event_to_be_sent"
            current_work = find_current_work()
            if current_work != '{}':  # JSON representation of an empty dictionary.
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
                        if key != "location" and key != "rows_id":  # TODO: rows_id? Maybe to be removed
                            temp_dict[key] = data.get(key)[i]
                    temp_dict['location'] = loc
                    json_to_insert = json.dumps(temp_dict)
                    if not db.execute_query(
                        query="INSERT INTO current_work (rules, time_stamp) VALUES (%s, CURRENT_TIMESTAMP())",
                        params= (json_to_insert, ),
                        commit=False,
                        select=True
                    ):
                        raise SystemExit
                db.commit_update()  # to make changes effective after inserting ALL the violated_rules
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
