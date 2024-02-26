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

// UpdateRulesHandler TODO: da rivedere nella gestione del json nel body della request
func UpdateRulesHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		wmsUtils.SetResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if contentType := request.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error: the request must be in JSON format")
		return
	}
	var rules wmsUtils.Rules
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
	locationName := rules.LocationInfo[0]
	locationLatitude := rules.LocationInfo[1]
	latitudeFloat, errConv := strconv.ParseFloat(locationLatitude, 64)
	if errConv != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error during latitude conversion from string to float64: %v", errConv))
	}
	roundedLatitude := wmsUtils.Round(latitudeFloat, 3)
	locationLongitude := rules.LocationInfo[2]
	longitudeFloat, errParse := strconv.ParseFloat(locationLatitude, 64)
	if errParse != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error during latitude conversion from string to float64: %v", errParse))
	}
	roundedLongitude := wmsUtils.Round(longitudeFloat, 3)
	countryCode := rules.LocationInfo[3]
	stateCode := rules.LocationInfo[4]
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
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", errorVar))
			return
		}
	} else {
		locationId = rowsLocation[0][0]
	}
	jsonRulesBytes, errMar := json.Marshal(rules)
	if errMar != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error into convert rules json into string: %v", errMar))
		return
	}
	jsonRulesString := string(jsonRulesBytes)
	query = fmt.Sprintf("SELECT * FROM user_constraints WHERE user_id = %d and location_id = %s", idUser, locationId)
	_, _, er := dbConn.ExecuteQuery(query)
	if er != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			query = fmt.Sprintf("INSERT INTO user_constraints (user_id, location_id, rules, time_stamp, trigger_period) VALUES(%d, %s, '%s', CURRENT_TIMESTAMP, %s)", idUser, locationId, jsonRulesString, triggerPeriod)
			_, _, err = dbConn.ExecuteQuery(query)
			if err != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in database insert: %v", err))
				return
			}
			wmsUtils.SetResponseMessage(writer, http.StatusOK, "New user_constraints correctly inserted!")
			return
		} else {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", er))
			return
		}
	} else {
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
