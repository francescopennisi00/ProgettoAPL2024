package http_handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	grpcC "um_microservice/src/communication_grpc"
	"um_microservice/src/types"
	"um_microservice/src/utils"
)

func DeleteAccountHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		utils.SetResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if contentType := request.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		utils.SetResponseMessage(writer, http.StatusBadRequest, "Error: the request must be in JSON format")
		return
	}
	var cred utils.Credentials
	err := json.NewDecoder(request.Body).Decode(&cred)
	if err != nil {
		utils.SetResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error in reading data: %v", err))
		return
	}
	email := cred.Email
	password := cred.Password
	hashPsw := utils.CalculateHash(password)
	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err = dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		utils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", err))
		return
	}

	query := fmt.Sprintf("SELECT id, email, password FROM users WHERE email='%s' and password='%s'", email, hashPsw)
	_, row, errorVar := dbConn.ExecuteQuery(query, true)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			utils.SetResponseMessage(writer, http.StatusUnauthorized, "Email or password wrong! Retry!")
			return
		} else {
			utils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in select from DB table 'users': %v", errorVar))
			return
		}
	}

	result := grpcC.DeleteUserConstraintsByUserId(row[0])
	if result == nil {
		query := fmt.Sprintf("DELETE FROM users WHERE email='%s' and password='%s'", email, hashPsw)
		_, _, errV := dbConn.ExecuteQuery(query)
		if errV != nil {
			utils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in database delete: %v", err))
			return
		} else {
			utils.SetResponseMessage(writer, http.StatusOK, "Account deleted with relative user constraints!")
			return
		}
	}
	utils.SetResponseMessage(writer, http.StatusInternalServerError, "Error in gRPC communication, account not deleted")
	return
}
