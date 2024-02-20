import grpc
import notifier_um_pb2
import notifier_um_pb2_grpc
from email.message import EmailMessage
import ssl
import smtplib
from notifier.utils.logger import logger
import notifier.utils.constants as constants


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
        match rule:
            case "max_temp" | "min_temp":
                return f"The temperature in {name_loc} ({country}, {state}) is {str(value)} °C!\n"
            case "max_humidity" | "min_humidity":
                return f"The humidity in {name_loc} ({country}, {state}) is {str(value)} %!\n"
            case "max_pressure" | "min_pressure":
                return f"The pressure in {name_loc} ({country}, {state}) is {str(value)} hPa!\n"
            case "max_cloud" | "min_cloud":
                return f"The the percentage of sky covered by clouds in {name_loc} ({country}, {state}) is {str(value)} %\n"
            case "max_wind_speed" | "min_wind_speed":
                return f"The wind speed in {name_loc} ({country}, {state}) is {str(value)} m/s!\n"
            case "wind_direction":
                return f"The wind direction in {name_loc} ({country}, {state}) is {value}!\n"
            case "rain":
                return f"Warning! In {name_loc} ({country}, {state}) is raining! Arm yourself with an umbrella!\n"
            case "snow":
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
            with smtplib.SMTP_SSL(constants.SMTP_SERVER_NAME, constants.SMTP_SERVER_PORT, context=context) as smtp:
                smtp.login(self._email_address, self._email_password)
                smtp.sendmail(self._email_address, recipient_email, em.as_string())
                boolean_to_return = True
        except smtplib.SMTPException as exception:
            logger.error("SMTP protocol error! -> " + str(exception) + "\n")
            boolean_to_return = False
        return boolean_to_return
