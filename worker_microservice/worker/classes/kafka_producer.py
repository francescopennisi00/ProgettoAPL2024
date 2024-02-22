import os
import time

import confluent_kafka
from confluent_kafka.admin import AdminClient, NewTopic
from worker.classes.database_connector import DatabaseConnector
from worker.utils.logger import logger


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
            try:
                kadmin.create_topics([new_topic,])
                logger.info("TOPIC CREATION STARTED!\n")
                time.sleep(3)  # wait to topic creation completion
                list_topics_metadata = kadmin.list_topics()
                topics = list_topics_metadata.topics  # Returns a dict()
                logger.info(f"LIST_TOPICS: {list_topics_metadata}")
                logger.info(f"TOPICS: {topics}")
                topic_names = set(topics.keys())
                logger.info(f"TOPIC_NAMES: {topic_names}")
            except confluent_kafka.KafkaException as err:
                logger.error(f"Error in creating topic:{topic_name}: {err}")

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
