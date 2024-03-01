package wms_http_handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	grpcC "wms_microservice/wms-src/wms-communication_grpc"
	wmsTypes "wms_microservice/wms-src/wms-types"
	wmsUtils "wms_microservice/wms-src/wms-utils"
)

func ShowRulesHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodGet {
		wmsUtils.SetResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Communication with UserManager in order to authenticate the user and retrieve user id
	var idUser int64
	var err error
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

	var dbConn wmsTypes.DatabaseConnector
	_, err = dbConn.StartDBConnection(wmsUtils.DBConnString)
	defer func(database *wmsTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", err))
		return
	}
	var rulesList []wmsUtils.ShowRulesOutput // List of (location_info, rules, trigger_period key-value pairs)
	query := fmt.Sprintf("SELECT location_id, rules, trigger_period FROM user_constraints WHERE user_id = %d", idUser)
	_, userConstraintsRows, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			wmsUtils.SetJSONResponse(writer, http.StatusOK, rulesList)
			return
		} else {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in DB query select from 'user_constraints' table: %v", errorVar))
			return
		}
	} else {

		for _, userConstraintsRow := range userConstraintsRows {
			locationId := userConstraintsRow[0]
			rulesJsonString := userConstraintsRow[1]
			triggerPeriod := userConstraintsRow[2]

			// query to DB in order to retrieve information about location by location_id
			query = fmt.Sprintf("SELECT location_name, latitude, longitude, state_code FROM locations WHERE id = %s", locationId)
			_, locationRows, err := dbConn.ExecuteQuery(query)
			if err != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in DB query select execution from 'location' table: %v", err))
				return
			}

			var rulesListItem wmsUtils.ShowRulesOutput
			rulesListItem.Location = locationRows[0]
			var rules wmsUtils.RulesIntoDB
			err = json.Unmarshal([]byte(rulesJsonString), &rules)
			if err != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in unmarshal JSON rules into Rules struct: %v", err))
				return
			}
			rulesListItem.Rules = rules
			rulesListItem.TriggerPeriod = triggerPeriod
			rulesList = append(rulesList, rulesListItem)
		}
		wmsUtils.SetJSONResponse(writer, http.StatusOK, rulesList)
		return
	}
}
