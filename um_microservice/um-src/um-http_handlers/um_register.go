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
	password := cred.Password
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
