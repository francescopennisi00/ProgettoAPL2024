import confluent_kafka
import json
import grpc
import notifier_um_pb2
import notifier_um_pb2_grpc
from email.message import EmailMessage
import ssl
import smtplib
import mysql.connector
import os
import time
import sys
import logging


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

BOOTSTRAP_SERVER_KAFKA = 'kafka-service:9092'
GROUP_ID = 'group1'
TOPIC_NAME = "event_to_be_notified"
TIMEOUT_POLL_REQUEST = 5.0
ATTEMPTS = 5
UM_ADDRESS = 'um-service:50051'
SMTP_SERVER_NAME = 'smtp.gmail.com'
SMTP_SERVER_PORT = 465


def commit_completed(er, partitions):
    if er:
        logger.error(str(er))
    else:
        logger.info("Commit done!\n")
        logger.info("Committed partition offsets: " + str(partitions) + "\n")
        logger.info("Notification fetched and stored in DB in order to be sent!\n")


class EmailSender:
    def __init__(self, service_address, email_address, email_password):
        self._gRPC_server_address = service_address
        self._email_address = email_address
        self._email_password = email_password

    # communication with user manager in order to get user's email
    def fetch_email_address(self, userid):
        try:
            with grpc.insecure_channel(self._gRPC_server_address) as channel:
                stub = notifier_um_pb2_grpc.NotifierUmStub(channel)
                response = stub.RequestEmail(notifier_um_pb2.Request(user_id=userid))
                logger.info("Fetched email: " + response.email + "\n")
                email_to_return = response.email
        except grpc.RpcError as error:
            logger.error("gRPC error! -> " + str(error) + "\n")
            email_to_return = "null"
        return email_to_return

    @staticmethod
    def __insert_rule_in_mail_text(rule, value, name_loc, country, state):
        if rule == "max_temp" or rule == "min_temp":
            return f"The temperature in {name_loc} ({country}, {state}) is {str(value)} °C!\n"
        elif rule == "max_humidity" or rule == "min_humidity":
            return f"The humidity in {name_loc} ({country}, {state}) is {str(value)} %!\n"
        elif rule == "max_pressure" or rule == "min_pressure":
            return f"The pressure in {name_loc} ({country}, {state}) is {str(value)} hPa!\n"
        elif rule == "max_cloud" or rule == "min_cloud":
            return f"The the percentage of sky covered by clouds in {name_loc} ({country}, {state}) is {str(value)} %\n"
        elif rule == "max_wind_speed" or rule == "min_wind_speed":
            return f"The wind speed in {name_loc} ({country}, {state}) is {str(value)} m/s!\n"
        elif rule == "wind_direction":
            return f"The wind direction in {name_loc} ({country}, {state}) is {value}!\n"
        elif rule == "rain":
            return f"Warning! In {name_loc} ({country}, {state}) is raining! Arm yourself with an umbrella!\n"
        elif rule == "snow":
            return f"Warning! In {name_loc} ({country}, {state}) is snowing! Be careful and enjoy the snow!\n"

    # send notification by email
    def send_email(self, recipient_email, violated_rules, name_location, country, state):
        subject = "Weather Alert Notification! "
        body = "Warning! Some weather parameters that you specified have been violated!\n\n"
        rules_list = violated_rules.get("violated_rules")  # extracting list of key-value pairs of violated_rules
        for element in rules_list:
            if element:
                rule = next(iter(element))
                body += self.__insert_rule_in_mail_text(rule, element[rule], name_location, country, state)
        em = EmailMessage()
        em['From'] = self._email_address
        em['To'] = recipient_email
        em['Subject'] = subject
        em.set_content(body)
        context = ssl.create_default_context()
        try:
            with smtplib.SMTP_SSL(SMTP_SERVER_NAME, SMTP_SERVER_PORT, context=context) as smtp:
                smtp.login(self._email_address, self._email_password)
                smtp.sendmail(self._email_address, recipient_email, em.as_string())
                boolean_to_return = True
        except smtplib.SMTPException as exception:
            logger.error("SMTP protocol error! -> " + str(exception) + "\n")
            boolean_to_return = False
        return boolean_to_return


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
        self.__init_secret('PASSWORD')
        self.__init_secret('APP_PASSWORD')
        self.__init_secret('EMAIL')

    @staticmethod
    def __init_secret(env_var_name):
        secret_path = os.environ.get(env_var_name)
        with open(secret_path, 'r') as file:
            secret_value = file.read()
        os.environ[env_var_name] = secret_value
        logger.info(f"Initialized {env_var_name}.\n")


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
        service_address=UM_ADDRESS,
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
            for attempt in range(ATTEMPTS):
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
        bootstrap_servers=BOOTSTRAP_SERVER_KAFKA,
        group_id=GROUP_ID,
        commit_completed_callback=commit_completed
    )

    # start Kafka subscription in order to retrieve messages written by Worker on broker
    kafka_consumer.start_subscription(TOPIC_NAME)

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
                if msg.error().code() == confluent_kafka.KafkaError.UNKNOWN_TOPIC_OR_PART:
                    raise SystemExit
            else:

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
                        raise SystemExit
                if not db.commit_update():  # to make changes effective after inserting ALL the violated_rules
                    raise SystemExit("Error in commit update in DB")
                db.close()

                # make commit to Kafka broker after Kafka msg has been stored in DB
                # we give some attempts to retrying commit in order to avoid replication of the same email
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
