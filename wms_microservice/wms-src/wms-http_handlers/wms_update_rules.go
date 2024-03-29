package wms_http_handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	grpcC "wms_microservice/wms-src/wms-communication_grpc"
	wmsTypes "wms_microservice/wms-src/wms-types"
	wmsUtils "wms_microservice/wms-src/wms-utils"
)

func UpdateRulesHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		wmsUtils.SetResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if contentType := request.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error: the request must be in JSON format")
		return
	}
	var rules wmsUtils.RulesFromPostRequest
	err := json.NewDecoder(request.Body).Decode(&rules)
	if err != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error in reading data: %v", err))
		return
	}
	// Communication with UserManager in order to authenticate the user and retrieve user id
	var idUser int64
	authorizationHeader := request.Header.Get("Authorization")
	if authorizationHeader != "" && strings.HasPrefix(authorizationHeader, "Bearer ") {
		idUser, err = grpcC.AuthenticateAndRetrieveUserId(authorizationHeader)
		if err != nil {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in communication with authentication server: %v", err))
			return
		}
		switch idUser {
		case -1:
			wmsUtils.SetResponseMessage(writer, http.StatusUnauthorized, "JWT Token expired: login required!")
			return
		case -2:
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, "Error in communication with DB in order to authentication: retry!")
			return
		case -3:
			wmsUtils.SetResponseMessage(writer, http.StatusUnauthorized, "JWT Token is not valid: login required!")
			return
		}
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusUnauthorized, "JWT Token not provided: login required!")
		return
	}
	triggerPeriod := rules.TriggerPeriod
	// inserted only because Decode function don't forbid to send body request with a json tag different from "trigger_period"
	// and, if user do that, the field rules.TriggerPeriod is filled with the zero value of string i.e. empty string ""
	if triggerPeriod == "" {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}
	if len(rules.Location) == 0 {
		// if the client doesn't respect the right JSON format of the request, len(rules.Location) == 0
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}
	locationName := rules.Location[0]
	locationLatitude := rules.Location[1]
	latitudeFloat, errConv := strconv.ParseFloat(locationLatitude, 64)
	if errConv != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error during latitude conversion from string to float64: %v", errConv))
	}
	roundedLatitude := wmsUtils.Round(latitudeFloat, 3)
	locationLongitude := rules.Location[2]
	longitudeFloat, errParse := strconv.ParseFloat(locationLongitude, 64)
	if errParse != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error during longitude conversion from string to float64: %v", errParse))
	}
	roundedLongitude := wmsUtils.Round(longitudeFloat, 3)
	countryCode := rules.Location[3]
	stateCode := rules.Location[4]

	log.SetPrefix("[INFO] ")
	log.Printf("LOCATION %s %s %s %s %s\n\n", locationName, locationLatitude, locationLongitude, countryCode, stateCode)

	var dbConn wmsTypes.DatabaseConnector
	_, err = dbConn.StartDBConnection(wmsUtils.DBConnString)
	defer func(database *wmsTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", err))
		return
	}
	var locationId string
	query := fmt.Sprintf("SELECT * FROM locations WHERE ROUND(latitude,3) = %f and ROUND(longitude,3) = %f and location_name = '%s'", roundedLatitude, roundedLongitude, locationName)
	_, rowsLocation, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			log.SetPrefix("[INFO] ")
			log.Println("There is no entry with that latitude and longitude")
			query = fmt.Sprintf("INSERT INTO locations (location_name, latitude, longitude, country_code, state_code) VALUES ('%s', %f, %f, '%s', '%s')", locationName, roundedLatitude, roundedLongitude, countryCode, stateCode)
			result, _, err := dbConn.ExecuteQuery(query)
			if err != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in database insert: %v", err))
				return
			}
			locationIdInt64, _ := result.LastInsertId()
			locationId = strconv.FormatInt(locationIdInt64, 10)
			log.SetPrefix("[INFO] ")
			log.Println("New location correctly inserted!")
		} else {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in DB query select from 'location' table: %v", errorVar))
			return
		}
	} else {
		locationId = rowsLocation[0][0]
	}

	// insert rule's values into a RulesIntoDB type variable
	var rulesIntoDB wmsUtils.RulesIntoDB

	if rules.Rules.MaxTemp != "" {
		rulesIntoDB.MaxTemp = rules.Rules.MaxTemp
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.MinTemp != "" {
		rulesIntoDB.MinTemp = rules.Rules.MinTemp
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.MaxHumidity != "" {
		rulesIntoDB.MaxHumidity = rules.Rules.MaxHumidity
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.MinHumidity != "" {
		rulesIntoDB.MinHumidity = rules.Rules.MinHumidity
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.MaxPressure != "" {
		rulesIntoDB.MaxPressure = rules.Rules.MaxPressure
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.MinPressure != "" {
		rulesIntoDB.MinPressure = rules.Rules.MinPressure
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.MaxWindSpeed != "" {
		rulesIntoDB.MaxWindSpeed = rules.Rules.MaxWindSpeed
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.MinWindSpeed != "" {
		rulesIntoDB.MinWindSpeed = rules.Rules.MinWindSpeed
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.WindDirection != "" {
		rulesIntoDB.WindDirection = rules.Rules.WindDirection
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.Rain != "" {
		rulesIntoDB.Rain = rules.Rules.Rain
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.Snow != "" {
		rulesIntoDB.Snow = rules.Rules.Snow
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.MaxCloud != "" {
		rulesIntoDB.MaxCloud = rules.Rules.MaxCloud
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	if rules.Rules.MinCloud != "" {
		rulesIntoDB.MinCloud = rules.Rules.MinCloud
	} else {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	jsonRulesBytes, errMar := json.Marshal(rulesIntoDB)
	if errMar != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error into convert rules json into string: %v", errMar))
		return
	}
	jsonRulesString := string(jsonRulesBytes)
	query = fmt.Sprintf("SELECT * FROM user_constraints WHERE user_id = %d and location_id = %s", idUser, locationId)
	_, _, er := dbConn.ExecuteQuery(query)
	if er != nil {
		if errors.Is(er, sql.ErrNoRows) {
			query = fmt.Sprintf("INSERT INTO user_constraints (user_id, location_id, rules, time_stamp, trigger_period) VALUES(%d, %s, '%s', CURRENT_TIMESTAMP, %s)", idUser, locationId, jsonRulesString, triggerPeriod)
			res, _, err := dbConn.ExecuteQuery(query)
			if err != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in database insert: %v", err))
				return
			}
			// idRule is useful for the client in order to load rule in UI
			idRule, errorId := res.LastInsertId()
			if errorId != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in retrieving id of last insert item into 'user_constraints' table: %v", errorId))
				return
			}
			wmsUtils.SetResponseMessage(writer, http.StatusOK, fmt.Sprintf("New user constraints correctly inserted! Id of the new constraints: %d", idRule))
			return
		} else {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in DB query select from 'user_constraints' table: %v", er))
			return
		}
	} else {
		// we found user constraints
		_, err := dbConn.BeginTransaction()
		if err != nil {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in starting DB transaction: %v", err))
			return
		}
		query := fmt.Sprintf("UPDATE user_constraints SET rules = '%s' WHERE user_id = %d AND location_id = %s", jsonRulesString, idUser, locationId)
		_, _, errQuery := dbConn.ExecIntoTransaction(query)
		if errQuery != nil {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in executing query into DB transaction: %v", errQuery))
			return
		}
		query = fmt.Sprintf("UPDATE user_constraints SET trigger_period = %s WHERE user_id = %d and location_id = %s", triggerPeriod, idUser, locationId)
		_, _, errQuery = dbConn.ExecIntoTransaction(query)
		if errQuery != nil {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in executing query into DB transaction: %v", errQuery))
			return
		}
		errorCommit := dbConn.CommitTransaction()
		if errorCommit != nil {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in commit DB transaction: %v", errorCommit))
			return
		}
		wmsUtils.SetResponseMessage(writer, http.StatusOK, "Updated table user_constraints correctly!")
		return
	}

}
