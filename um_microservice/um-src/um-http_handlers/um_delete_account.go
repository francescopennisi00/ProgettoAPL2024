package um_http_handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	grpcC "um_microservice/um-src/um-communication_grpc"
	umTypes "um_microservice/um-src/um-types"
	umUtils "um_microservice/um-src/um-utils"
)

func DeleteAccountHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		umUtils.SetResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if contentType := request.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		umUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error: the request must be in JSON format")
		return
	}
	var cred umUtils.Credentials
	err := json.NewDecoder(request.Body).Decode(&cred)
	if err != nil {
		umUtils.SetResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error in reading data: %v", err))
		return
	}
	email := cred.Email
	// inserted only because Decode function don't forbid to send body request with a json tag different from "email"
	// and, if user do that, the field Email is filled with the zero value of string i.e. empty string ""
	if email == "" {
		umUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}
	password := cred.Password
	// inserted only because Decode function don't forbid to send body request with a json tag different from "password"
	// and, if user do that, the field Password is filled with the zero value of string i.e. empty string ""
	if password == "" {
		umUtils.SetResponseMessage(writer, http.StatusBadRequest, "Error in reading data! The request must be in the correct JSON format")
		return
	}
	hashPsw := umUtils.CalculateHash(password)
	var dbConn umTypes.DatabaseConnector
	_, err = dbConn.StartDBConnection(umUtils.DBConnString)
	defer func(database *umTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		umUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", err))
		return
	}

	query := fmt.Sprintf("SELECT id, email, password FROM users WHERE email='%s' and password='%s'", email, hashPsw)
	_, row, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			umUtils.SetResponseMessage(writer, http.StatusUnauthorized, "Email or password wrong! Retry!")
			return
		} else {
			umUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in select from DB table 'users': %v", errorVar))
			return
		}
	}

	result := grpcC.DeleteUserConstraintsByUserId(row[0][0])
	if result == nil {
		query := fmt.Sprintf("DELETE FROM users WHERE email='%s' and password='%s'", email, hashPsw)
		_, _, errV := dbConn.ExecuteQuery(query)
		if errV != nil {
			umUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in database delete: %v", errV))
			return
		} else {
			umUtils.SetResponseMessage(writer, http.StatusOK, "Account deleted with relative user constraints!")
			return
		}
	}
	umUtils.SetResponseMessage(writer, http.StatusInternalServerError, "Error in gRPC communication, account not deleted")
	return
}
