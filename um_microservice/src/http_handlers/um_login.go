package http_handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
	"um_microservice/src/types"
	"um_microservice/src/utils"
)

func LoginHandler(writer http.ResponseWriter, request *http.Request) {

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
	_, err = dbConn.StartDBConnection(utils.DBConnString)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		utils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", err))
		return
	}

	query := fmt.Sprintf("SELECT email, password FROM users WHERE email='%s' AND password='%s'", email, hashPsw)
	_, _, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			utils.SetResponseMessage(writer, http.StatusUnauthorized, "Email or password wrong! Retry!")
			return
		} else {
			utils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in select from DB table 'users': %v", errorVar))
			return
		}
	}

	tokenExpireTime := time.Now().Add(time.Duration(utils.HourTokenExpiration) * time.Hour) // 3 days from now

	// create the payload
	payload := jwt.MapClaims{
		"email": email,
		"exp":   tokenExpireTime.Unix(),
	}

	// create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// sign the token with the hashed password
	tokenString, errV := token.SignedString([]byte(hashPsw))
	if errV != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error in signing JWT Token: %v\n", errV)
		return
	}

	utils.SetResponseMessage(writer, http.StatusOK, fmt.Sprintf("Login successfully made! JWT Token: %s", tokenString))
	return
}
