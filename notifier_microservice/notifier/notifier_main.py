import json
import os
import time
import sys
from utils.logger import logger
import utils.constants as constants
from classes.database_connector import DatabaseConnector
from classes.email_sender import EmailSender
from classes.secrets_initializer import SecretInitializer
from classes.kafka_consumer import KafkaConsumer


# connection with DB and update the entry of the notification sent
def update_event_sent(event_id):

    database_connector = DatabaseConnector(
        hostname=os.environ.get("HOSTNAME"),
        port=os.environ.get("PORT"),
        user=os.environ.get("USER"),
        password=os.environ.get("PASSWORD"),
        database=os.environ.get("DATABASE")
    )

    outcome = database_connector.execute_query(
        query="UPDATE events SET sent=TRUE, time_stamp=CURRENT_TIMESTAMP WHERE id = %s",
        params=(str(event_id),),
        commit=True,
        select=False
    )
    if not outcome:
        return False

    database_connector.close()

    return True


# find events to send, send them by email and update events in DB
def find_event_not_sent():

    database_connector = DatabaseConnector(
        hostname=os.environ.get("HOSTNAME"),
        port=os.environ.get("PORT"),
        user=os.environ.get("USER"),
        password=os.environ.get("PASSWORD"),
        database=os.environ.get("DATABASE")
    )

    # instantiating EmailSender instance in order to allow for sending emails
    email_sender = EmailSender(
        service_address=constants.UM_ADDRESS,
        email_address=os.environ.get('EMAIL'),
        email_password=os.environ.get('APP_PASSWORD')
    )

    results = database_connector.execute_query("SELECT * FROM events WHERE sent=FALSE")
    if results:
        for x in results:
            email = email_sender.fetch_email_address(x[1])
            if email == "null":
                database_connector.close()
                return False
            if email == "not present anymore":
                boolean = database_connector.execute_query(
                    query="DELETE FROM events WHERE id=%s",
                    params=(x[0], ),
                    commit=True,
                    select=False
                )
                if boolean:
                    continue
                else:
                    raise SystemExit
            loc_name = x[2]
            loc_country = x[3]
            loc_state = x[4]
            rules_violated = json.loads(x[5])
            res = email_sender.send_email(email, rules_violated, loc_name, loc_country, loc_state)
            if not res:
                database_connector.close()
                return "error_in_send_email"
            # we give 5 attempts to try to update the DB in order
            # to avoid resending the email as much as possible
            for attempt in range(constants.ATTEMPTS):
                if update_event_sent(x[0]):
                    database_connector.close()
                    break
                time.sleep(1)
            else:
                raise SystemExit


if __name__ == "__main__":

    logger.info("Start notifier main")
    secret_initializer = SecretInitializer()
    secret_initializer.init_secrets()
    logger.info("ENV variables initialization done")

    # instantiating MySQL DB Connector instance in order to crate table 'events' if not exists
    db_connector = DatabaseConnector(
        hostname=os.environ.get("HOSTNAME"),
        port=os.environ.get("PORT"),
        user=os.environ.get("USER"),
        password=os.environ.get("PASSWORD"),
        database=os.environ.get("DATABASE")
    )

    bool_outcome = db_connector.execute_query(
        query="CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY AUTO_INCREMENT, user_id INTEGER NOT NULL, location_name VARCHAR(70) NOT NULL, location_country VARCHAR(10) NOT NULL, location_state VARCHAR(30) NOT NULL, rules JSON NOT NULL, time_stamp TIMESTAMP NOT NULL, sent BOOLEAN NOT NULL)",
        commit=True,
        select=False
    )
    if not bool_outcome:
        sys.exit("Error in creating table events")

    if not db_connector.close():
        sys.exit("Error in closing DB connection")

    # instantiating Kafka consumer instance
    kafka_consumer = KafkaConsumer(
        bootstrap_servers=constants.BOOTSTRAP_SERVER_KAFKA,
        group_id=constants.GROUP_ID
    )

    # start Kafka subscription in order to retrieve messages written by Worker on broker
    kafka_consumer.start_subscription(constants.TOPIC_NAME)

    logger.info("Starting while true\n")

    try:
        while True:

            logger.info("New iteration!\n")

            # Looking for entries that have sent == False
            result = find_event_not_sent()
            if result == "error_in_send_email":
                continue

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
                    raise SystemExit("Error in commit update in DB")

                # make commit to Kafka broker after Kafka msg has been stored in DB
                # we give some attempts to retrying commit in order to avoid replication of the same email
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
