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

// TODO: implement this one
func formatRulesResponse(rules interface{}) string {
	return ""
}

func ShowRulesHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodGet {
		wmsUtils.SetResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Communication with UserManager in order to authenticate the user and retrieve user id
	var idUser int64
	authorizationHeader := request.Header.Get("Authorization")
	if authorizationHeader != "" && strings.HasPrefix(authorizationHeader, "Bearer ") {
		idUser, err := grpcC.AuthenticateAndRetrieveUserId(authorizationHeader)
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
	_, err := dbConn.StartDBConnection(wmsUtils.DBConnString)
	defer func(database *wmsTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", err))
		return
	}
	query := fmt.Sprintf("SELECT location_id, rules, trigger_period FROM user_constraints WHERE user_id = %d", idUser)
	_, rowsLocation, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, "There is no rules that you have indicated! Please insert location, rules and trigger period!")
			return
		} else {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in DB query select: %v", errorVar))
			return
		}
	} else {

		// TODO: forse va rivisto! Per ora così!
		var rulesList []interface{} // List of (location_info, rules, trigger_period key-value pairs)

		for _, rowLocation := range rowsLocation {
			locationId := rowLocation[0]
			rulesJsonString := rowLocation[1]
			triggerPeriod := rowLocation[2]
			// query to DB in order to retrieve information about location by location_id
			query = fmt.Sprintf("SELECT location_name, country_code, state_code FROM locations WHERE id = %s", locationId)
			_, _, err = dbConn.ExecuteQuery(query)
			if err != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in DB query select execution: %v", err))
				return
			}

			//TODO: forse anche questi dati vanno rivisti! Per ora così!
			var tempList []interface{}
			var locationMap map[string][]string
			locationMap["location"] = rowLocation
			tempList = append(tempList, locationMap)
			var rules wmsUtils.Rules
			err := json.Unmarshal([]byte(rulesJsonString), &rules)
			if err != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in unmarshal JSON rules into Rules struct: %v", err))
				return
			}
			tempList = append(tempList, rules)
			var triggerPeriodMap map[string]string
			triggerPeriodMap["trigger_period"] = triggerPeriod
			tempList = append(tempList, triggerPeriodMap)
			rulesList = append(rulesList, tempList)
		}
		rulesReturned := formatRulesResponse(rulesList)
		//TODO: for now we return a string: maybe we can return a JSON
		wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("YOUR RULES: <br><br> %s", rulesReturned))
		return
	}
}
