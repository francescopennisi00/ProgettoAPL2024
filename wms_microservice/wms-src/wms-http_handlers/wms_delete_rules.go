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

func DeleteRulesHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		wmsUtils.SetResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if contentType := request.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error: the request must be in JSON format")
		return
	}

	var location wmsUtils.LocationType
	err := json.NewDecoder(request.Body).Decode(&location)
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

	if len(location.Location) == 0 {
		// if the client doesn't respect the right JSON format of the request, len(location.Location) == 0
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}

	// extract information about location whose rules have to be deleted for the logged user

	locationName := location.Location[0]
	latitude := location.Location[1]
	latitudeFloat, errConv := strconv.ParseFloat(latitude, 64)
	if errConv != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error during latitude conversion from string to float64: %v", errConv))
	}
	roundedLatitude := wmsUtils.Round(latitudeFloat, 3)
	longitude := location.Location[2]
	longitudeFloat, errParse := strconv.ParseFloat(longitude, 64)
	if errParse != nil {
		wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error during longitude conversion from string to float64: %v", errParse))
	}
	roundedLongitude := wmsUtils.Round(longitudeFloat, 3)
	countryCode := location.Location[3]
	stateCode := location.Location[4]

	log.SetPrefix("[INFO] ")
	log.Printf("LOCATION %s %s %s %s %s\n\n", locationName, latitude, longitude, countryCode, stateCode)

	// verify is there are user constraints for the logged user corresponding to the location indicated
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
	query := fmt.Sprintf("SELECT * FROM locations WHERE ROUND(latitude,3) = %f AND ROUND(longitude,3) = %f AND location_name ='%s'", roundedLatitude, roundedLongitude, locationName)
	_, rowsLocation, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			log.SetPrefix("[INFO] ")
			log.Println("There is no entry with that latitude and longitude")
			wmsUtils.SetResponseMessage(writer, http.StatusBadRequest, "There are no parameters to delete with that location")
			return
		} else {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in DB query select: %v", errorVar))
			return
		}
	} else {
		locationId = rowsLocation[0][0]
		query := fmt.Sprintf("DELETE FROM user_constraints WHERE user_id = %d and location_id = %s", idUser, locationId)
		_, _, err := dbConn.ExecuteQuery(query)
		if err != nil {
			wmsUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in database delete: %v", err))
			return
		}
		wmsUtils.SetResponseMessage(writer, http.StatusOK, fmt.Sprintf("Your constraints for the location %s have been correctly deleted", locationName))
		return
	}

}
