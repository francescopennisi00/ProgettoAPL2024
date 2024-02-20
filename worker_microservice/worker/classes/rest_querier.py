import json
import requests
import worker.utils.constants as constants
from worker.utils.logger import logger


# implementing pattern Singleton for SecretInitializer class
class RESTQuerier:

    _querier_instance = None  # starter value: no object initially instantiated

    # if RESTQuerier object is already instantiated, then return it without re-instantiating
    def __new__(cls):
        if not cls._querier_instance:
            cls._instance = super().__new__(cls)
        return cls._querier_instance

    def make_query(self, apikey, location):
        rest_call = constants.ENDPOINT_REST_OPEN_WEATHER.replace("LATITUDE", location[1])
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
