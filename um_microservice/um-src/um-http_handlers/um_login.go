package um_http_handlers

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
	umTypes "um_microservice/um-src/um-types"
	umUtils "um_microservice/um-src/um-utils"
)

func LoginHandler(writer http.ResponseWriter, request *http.Request) {

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
	log.SetPrefix("[INFO] ")
	log.Println(fmt.Sprintf("EMAIL: %s", email))
	password := cred.Password
	hashPsw := umUtils.CalculateHash(password)
	log.SetPrefix("[INFO] ")
	log.Println(fmt.Sprintf("HASH PSW: %s", hashPsw))
	var dbConn umTypes.DatabaseConnector
	_, err = dbConn.StartDBConnection(umUtils.DBConnString)
	defer func(database *umTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		umUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %v", err))
		return
	}

	query := fmt.Sprintf("SELECT email, password FROM users WHERE email='%s' AND password='%s'", email, hashPsw)
	log.SetPrefix("[INFO] ")
	log.Println(fmt.Sprintf("EXECUTING QUERY: %s", query))
	_, res, errorVar := dbConn.ExecuteQuery(query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			log.SetPrefix("[INFO] ")
			log.Println("No rows founded!")
			umUtils.SetResponseMessage(writer, http.StatusUnauthorized, "Email or password wrong! Retry!")
			return
		} else {
			umUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in select from DB table 'users': %v", errorVar))
			return
		}
	}
	log.SetPrefix("[INFO] ")
	log.Println(fmt.Sprintf("Founded a row: %s, %s", res[0][0], res[0][1]))
	log.Println("Email and password correctly inserted!")

	tokenExpireTime := time.Now().Add(time.Duration(umUtils.HourTokenExpiration) * time.Hour) // 3 days from now

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
		umUtils.SetResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in signing JWT Token: %v\n", errV))
		return
	}

	umUtils.SetResponseMessage(writer, http.StatusOK, fmt.Sprintf("Login successfully made! JWT Token: %s", tokenString))
	return
}
