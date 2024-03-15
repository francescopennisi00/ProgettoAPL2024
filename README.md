# APL Project
Advanced Programming Languages Project - 2024 - UniCT

## Table of Contents
- [Introduction](#introduction)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Introduction

The system in question is built following a microservices architectural pattern.
Microservices are packaged in special containers making use of Docker containerization technology,
which allows the microservices to be collected and isolated in complete runtime environments equipped with all the necessary files for execution,
to ensure portability to any infrastructure (hardware and software) that is Docker-enabled.
The objective of the system is to allow registered users to indicate, for each location of interest,
weather parameters, such as the maximum temperature or the possible presence of rain.
These parameters will be monitored by the system itself, and if the specified weather conditions are violated,
users will be notified by e-mail.
To enable users to use the application more easily, a GUI-equipped client was developed.
For more details we recommend to read documentation. [Documentation.](https://github.com/francescopennisi00/ProgettoAPL2024/blob/main/RelazioneAPLGenovesePennisi2024.pdf)

## Installation

### Prerequisites

- Docker
- Visual Studio with .NET MAUI workload

### Steps

-       git clone https://github.com/francescopennisi00/ProgettoDSBD.git
- Secrets have to be created in /docker directory <br>
  #### How to create Secrets
  In order to run application on Docker, you need to create text files with your secrets and put them in /docker directory.
  You have to name them in this way:
    - apikey.txt
    - app_password.txt
    - email.txt
    - notifier_DB_root_psw.txt
    - um_DB_root_psw.txt
    - wms_DB_root_psw.txt
    - worker_DB_root_psw.txt <br> <br>

  The last four must contain the password of root user to the associated databases. <br>
  To get the weather forecast you need to use the services offered by OpenWeather,
  so we recommend that you register at this link (https://openweathermap.org/api) to free service option and get your
  API key, that you have to put it in the secret apikey.txt.<br>
  The secret email.txt must contain the email that you want to use, you have to use gmail like provider.<br>
  The secret app_password.txt must contain the password that gmail provides to your account in order to be used
  in applications such as this one. (It is not the password with which you access to your gmail account). <br>


- Modify your /etc/hosts file in Linux or \Windows\System32\drivers\etc\hosts file in Windows and insert the following line:

            127.0.0.1 localhost weather.com

- Run application
  #### How to run application
  Open a terminal and go to the project directory and then in ./docker directory. Then execute <br>

            docker compose up

## Usage

### Usage with C# WeatherClient
You can interact with server-side application using C# WeatherClient.
To do this, you can deploy WeatherClient application on your architecture or start it directly inside Visual Studio. <br>
In both case, you have to open project at ./WeatherClient/WeatherClient/WeatherClient.csproj into Visual Studio.
If you simply want to start client application without deploying it, you have to select the target machine in the Visual Studio toolbar and click on "Run without executing debug".
If you want to deploy app in your local machine, you must follow the guide at https://learn.microsoft.com/en-us/dotnet/maui/deployment/?view=net-maui-8.0.

### Usage with POSTMAN
You can interact with server-side application using POSTMAN as client in order to interact with the system by REST API. <br>
For more information on the meaning and usefulness of endpoints we recommend to read documentation. [Documentation.](https://github.com/francescopennisi00/ProgettoAPL2024/blob/main/RelazioneAPLGenovesePennisi2024.pdf)
<br> From POSTMAN, you have to use this base endpoint:

            http://weather.com:8080

### Endpoints
You have to insert the request body in the <i> Body</i> section, select <i>raw</i> and use JSON format. 

- usermanager/register POST

            //In this way we fill the request body 
            {
            "email": "your_email@gmail.com",
            "password":"your_psw"
            }

- usermanager/login POST

            //In this way we fill the request body
            {
            "email": "your_email@gmail.com",
            "password":"your_psw"
            }

- usermanager/delete_account POST

            //In this way we fill the request body
            {
            "email": "your_email@gmail.com",
            "password":"your_psw"
            }

For next endpoints you have to put the JWT Token in the HTTP Header,
with Authorization as key and "Bearer your_JWT_token" as value. 
<br> All numerical values must be entered as strings such as "your_numerical_value". <br>

- wms/update_rules POST

            //In this way we fill the request body 
            {
            "trigger_period": "trigger_period",
            "location": ["location_name", "latitude", "longitude", "country_code", "state_code"],
            "rules" {
                      "max_temp" : "your value" or "null" if you are not interested,
                      "min_temp": your value or "null" if you are not interested,
                      "max_humidity": your value or "null" if you are not interested,
                      "min_humidity": your value or "null" if you are not interested,
                      "max_pressure": your value or "null" if you are not interested,
                      "min_pressure": your value or "null" if you are not interested,
                      "max_wind_speed": your value or "null" if you are not interested,
                      "min_wind_speed": your value or "null" if you are not interested,
                      "wind_direction": your value or "null" if you are not interested,
                      "rain": your value or "null" if you are not interested,
                      "snow": your value or "null" if you are not interested,
                      "max_cloud": your value or "null" if you are not interested,
                      "min_cloud": your value or "null" if you are not interested
                     }
            }
- wms/show_rules GET

- wms/update_rules/delete_user_constraints_by_location POST

            //In this way we fill the request body
            {
            "location": ["location_name", "latitude", "longitude", "country_code", "state_code"]
            }

## Contributing
We welcome contributions from the community! If you'd like to contribute, please follow these steps:

1. Fork, then clone the repo:
2. Create a new branch: `git checkout -b feature-branch`
3. Make your changes
4. Push to your fork and submit a pull request

## License
This project is licensed under the terms of the MIT license. See [LICENSE](LICENSE) for more details