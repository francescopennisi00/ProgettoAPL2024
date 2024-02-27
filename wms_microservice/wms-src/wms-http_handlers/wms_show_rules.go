package wms_http_handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	grpcC "wms_microservice/wms-src/wms-communication_grpc"
	wmsTypes "wms_microservice/wms-src/wms-types"
	wmsUtils "wms_microservice/wms-src/wms-utils"
)

// this function produce an output string that contains only information about rules that user is interested in
func reduceRuleMap(rule wmsUtils.RulesIntoDB) (string, error) {

	// Convert the rule struct in a byte slice in order to Unmarshal it into a map
	ruleBytes, err := json.Marshal(rule)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Println("Error during JSON marshaling:", err)
		return "", err
	}

	type MapOfRulesUserIsInterestedIn map[string]string
	var ruleMap MapOfRulesUserIsInterestedIn
	err = json.Unmarshal(ruleBytes, &ruleMap)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Println("Error during JSON unmarshal:", err)
		return "", err
	}

	for key, value := range ruleMap {
		if value == "null" {
			delete(ruleMap, key)
		}
	}

	//execute marshaling of the ruleMap without "null" values into a JSON string
	ruleMapBytes, errMar := json.Marshal(ruleMap)
	if errMar != nil {
		log.SetPrefix("[ERROR] ")
		log.Println("Error during JSON marshal of final string:", err)
		return "", err
	}

	return string(ruleMapBytes), nil

}

// TODO: maybe we have to change it in order to return a JSON
func formatRulesResponse(rules []wmsUtils.ShowRulesOutput) string {
	stringToBeReturned := "No rules inserted!"
	locationRulesString := ""
	counter := 1

	for _, rule := range rules {
		ruleLocationMapString, err := json.Marshal(rule.Location)
		if err != nil {
			return fmt.Sprintf("Error in marshaling rule.Location: %v\n", err)
		}
		triggerPeriodMapString, errM := json.Marshal(rule.TriggerPeriod)
		if errM != nil {
			return fmt.Sprintf("Error in marshaling rule.TriggerPeriod: %v\n", errM)
		}
		ruleMapString, errReduce := reduceRuleMap(rule.Rules)
		if errReduce != nil {
			return fmt.Sprintf("Error in reduceRuleMap: %v\n", errReduce)
		}
		tempString := fmt.Sprintf("LOCATION {%d}<br>%s<br>%s<br>%s<br><br>", counter, ruleLocationMapString, ruleMapString, triggerPeriodMapString)
		locationRulesString = locationRulesString + tempString
		counter++
	}
	if locationRulesString != "" {
		stringToBeReturned = locationRulesString
	}
	return stringToBeReturned
}

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
	query := fmt.Sprintf("SELECT location_id, rules, trigger_period FROM user_constraints WHERE user_id = %d", idUser)
	_, rowsLocation, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			wmsUtils.SetResponseMessage(writer, http.StatusOK, "There is no rules that you have indicated! Please insert location, rules and trigger period!")
			return
		} else {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in DB query select from 'user_constraints' table: %v", errorVar))
			return
		}
	} else {

		var rulesList []wmsUtils.ShowRulesOutput // List of (location_info, rules, trigger_period key-value pairs)

		for _, rowLocation := range rowsLocation {
			locationId := rowLocation[0]
			rulesJsonString := rowLocation[1]
			triggerPeriod := rowLocation[2]
			// query to DB in order to retrieve information about location by location_id
			query = fmt.Sprintf("SELECT location_name, country_code, state_code FROM locations WHERE id = %s", locationId)
			_, _, err = dbConn.ExecuteQuery(query)
			if err != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in DB query select execution from 'location' table: %v", err))
				return
			}

			var rulesListItem wmsUtils.ShowRulesOutput
			rulesListItem.Location = rowLocation
			var rules wmsUtils.RulesIntoDB
			err := json.Unmarshal([]byte(rulesJsonString), &rules)
			if err != nil {
				wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in unmarshal JSON rules into Rules struct: %v", err))
				return
			}
			rulesListItem.Rules = rules
			rulesListItem.TriggerPeriod = triggerPeriod
			rulesList = append(rulesList, rulesListItem)
		}
		stringToReturn := formatRulesResponse(rulesList)
		//TODO: for now we return a string: maybe we can return a JSON
		wmsUtils.SetResponseMessage(writer, http.StatusOK, fmt.Sprintf("YOUR RULES: <br><br> %s", stringToReturn))
		return
	}
}
