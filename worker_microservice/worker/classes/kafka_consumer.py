import confluent_kafka
import sys
import json
import os
from worker.utils.logger import logger
import worker.utils.constants as constants
from worker.classes.database_connector import DatabaseConnector


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
    def topic_not_found(msg):
        if msg.error().code() == confluent_kafka.KafkaError.UNKNOWN_TOPIC_OR_PART:
            return True
        else:
            return False

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
        return self._consumer.poll(timeout=constants.TIMEOUT_POLL_REQUEST)

    @staticmethod
    def consume_kafka_message(msg):

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
        userid_list = data.get("user_id")
        logger.info("USER_ID_LIST: " + str(userid_list))
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
        for i in range(0, len(userid_list)):
            temp_dict = dict()
            for key in set(data.keys()):
                if key != "location" and key != "rows_id":
                    temp_dict[key] = data.get(key)[i]
            temp_dict['location'] = loc
            json_to_insert = json.dumps(temp_dict)
            if not db.execute_query(
                    query="INSERT INTO current_work (rules, time_stamp) VALUES (%s, CURRENT_TIMESTAMP())",
                    params=(json_to_insert, ),
                    commit=False,
                    select=False
            ):
                db.close()
                return False
        if not db.commit_update():  # to make changes effective after inserting ALL the violated_rules
            db.close()
            return False
        db.close()
        return True

    def commit_async(self):
        try:
            self._consumer.commit(asynchronous=True)
        except Exception as e:
            logger.error("Error in commit to Kafka broker! -> " + str(e) + "\n")
            return False
        return True

    def close_consumer(self):
        self._consumer.close()


