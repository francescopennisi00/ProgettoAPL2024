import confluent_kafka
import sys
from notifier.utils.logger import logger
import notifier.utils.constants as constants


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

    def commit_async(self):
        try:
            self._consumer.commit(asynchronous=True)
        except Exception as e:
            logger.error("Error in commit to Kafka broker! -> " + str(e) + "\n")
            return False
        return True

    def close_consumer(self):
        self._consumer.close()
