import confluent_kafka
import sys
import json
import os
from notifier.utils.logger import logger
import notifier.utils.constants as constants
from notifier.classes.database_connector import DatabaseConnector


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
        record_key = msg.key()
        logger.info("RECORD KEY " + str(record_key))
        record_value = msg.value()
        logger.info("RECORD VALUE " + str(record_value))
        data = json.loads(record_value)
        location_name = data.get("location")[0]
        location_country = data.get("location")[3]
        location_state = data.get("location")[4]
        del data["location"]  # now all key-value pairs have user_id value as key
        user_id_set = set(data.keys())

        # connection with DB and store events to be notified

        db = DatabaseConnector(
            hostname=os.environ.get("HOSTNAME"),
            port=os.environ.get("PORT"),
            user=os.environ.get("USER"),
            password=os.environ.get("PASSWORD"),
            database=os.environ.get("DATABASE")
        )
        for user_id in user_id_set:
            temp_dict = dict()
            temp_dict["violated_rules"] = data.get(user_id)
            violated_rules = json.dumps(temp_dict)
            bool_result = db.execute_query(
                query="INSERT INTO events (user_id, location_name, location_country, location_state, rules, time_stamp, sent) VALUES(%s, %s, %s, %s, %s, CURRENT_TIMESTAMP, FALSE)",
                params=(str(user_id), location_name, location_country, location_state, violated_rules),
                commit=False,
                select=False)
            if not bool_result:
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
