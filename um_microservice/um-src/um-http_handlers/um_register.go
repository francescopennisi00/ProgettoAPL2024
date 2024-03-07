package um_http_handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	umTypes "um_microservice/um-src/um-types"
	umUtils "um_microservice/um-src/um-utils"
)

func RegisterHandler(writer http.ResponseWriter, request *http.Request) {

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
	var dbConn umTypes.DatabaseConnector
	_, err = dbConn.StartDBConnection(umUtils.DBConnString)
	defer func(database *umTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		umUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", err))
		return
	}

	query := fmt.Sprintf("SELECT email FROM users WHERE email='%s'", email)
	_, _, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			// if there is no row this means that the user is not yet registered
			hashPsw := umUtils.CalculateHash(password)
			query = fmt.Sprintf("INSERT INTO users (email, password) VALUES ('%s', '%s')", email, hashPsw)
			_, _, err := dbConn.ExecuteQuery(query)
			if err != nil {
				umUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in database insert: %v", err))
				return
			}
			umUtils.SetResponseMessage(writer, http.StatusOK, "Registration made successfully! Now try to sign in!")
			return
		} else {
			umUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", errorVar))
			return
		}
	} else {
		umUtils.SetResponseMessage(writer, http.StatusBadRequest, "Email already in use! Try to sign in!")
		return
	}
}
