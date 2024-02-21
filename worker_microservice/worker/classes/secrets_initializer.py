import os
from worker.utils.logger import logger


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