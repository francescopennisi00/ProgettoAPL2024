package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	notifierUm "um_microservice/proto/notifier_um"
	wmsUm "um_microservice/proto/wms_um"
	"um_microservice/src/types"
)

type credentials struct {
	email    string
	password string
}

type umNotifierServer struct {
	notifierUm.UnimplementedNotifierUmServer
}

type umWmsServer struct {
	wmsUm.UnimplementedWMSUmServer
}

var wg sync.WaitGroup

var (
	portWMS      = flag.Int("portWMS", 50052, "The server port for WMS")
	portNotifier = flag.Int("port", 50051, "The server port for Notifier")
)

func calculateHash(inputString string) string {
	sha256Hash := sha256.New()
	sha256Hash.Write([]byte(inputString))
	hashResult := sha256Hash.Sum(nil)
	return hex.EncodeToString(hashResult)
}

func setResponseMessage(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(w, "%s\n", message)
}

func deleteUserConstraintsByUserId(userId string) error {

	//convert userId from string to int64
	num, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		fmt.Printf("Error during conversion: %v\n", err)
		return err
	}

	// start connection to gRPC server
	conn, errV := grpc.Dial("wms-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)
	if errV != nil {
		return errV
	}
	client := wmsUm.NewWMSUmClient(conn)

	response, errVar := client.RequestDeleteUser_Constraints(context.Background(), &wmsUm.User{UserId: num})
	if errVar != nil {
		return errVar
	}
	if response.Code != 0 {
		errVariable := errors.New("something went wrong in WMS while deleting user constraints")
		return errVariable
	}

	return nil //no error occurred
}

func (s *umNotifierServer) RequestUserIdViaJWTToken(ctx context.Context, in *wmsUm.Request) (*wmsUm.Reply, error) {

	// extracting JWT token from the request
	tokenString := in.GetJwtToken()

	// extracting token information without verifying them: needed in order to retrieve user email
	token, err := jwt.Parse(tokenString, nil)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error in parsing JWT Token without verifying it: %v\n", err)
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error in extracting JWT Token without verifying it with Claims: %v\n", err)
		return nil, err
	}
	email, okBool := claims["email"].(string)
	if !okBool {
		log.SetPrefix("[ERROR] ")
		log.Printf("Impossible to extract email from JWT Token without verifying it: %v\n", err)
		return nil, err
	}

	// Retrieve user's password from DB in oder to verifying JWT Token and authenticate him
	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err = dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("DB connection error! -> %v\n", err)
	}
	query := fmt.Sprintf("SELECT id, password FROM users WHERE email= %s", email)
	_, row, errV := dbConn.ExecuteQuery(query, true)
	if errV != nil {
		if errors.Is(errV, sql.ErrNoRows) {
			return &wmsUm.Reply{UserId: -3}, nil // token is not valid: email not present
		}
		log.SetPrefix("[ERROR] ")
		log.Printf("DB query error! -> %v\n", errV)
		return nil, errV
	}

	//convert user id from string to int64
	var idUser int64
	idUser, err = strconv.ParseInt(row[0], 10, 64)
	if err != nil {
		fmt.Printf("Error: user id returned by DB is not an integer: -> %v\n", err)
		return nil, err
	}

	//password is already a string: no conversion is needed
	password := row[1]

	// verify JWT Token with password as secret
	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(password), nil
	})
	if errors.Is(err, jwt.ErrTokenExpired) {
		return &wmsUm.Reply{UserId: -1}, nil //token is expired
	}
	if !token.Valid {
		return &wmsUm.Reply{UserId: -3}, nil //token is not valid: password incorrect
	}

	// Everything succeeded, return user ID
	return &wmsUm.Reply{UserId: idUser}, nil

}

func (s *umNotifierServer) RequestEmail(ctx context.Context, in *notifierUm.Request) (*notifierUm.Reply, error) {

	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err := dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		return &notifierUm.Reply{Email: "null"}, nil
	}

	userId := in.UserId
	query := fmt.Sprintf("SELECT email FROM users WHERE id=%d", userId)
	_, email, errorVar := dbConn.ExecuteQuery(query, true)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			log.SetPrefix("[INFO] ")
			log.Printf("Email not present anymore")
			return &notifierUm.Reply{Email: "not present anymore"}, nil
		} else {
			log.SetPrefix("[ERROR] ")
			log.Printf("DB Error: %v\n", errorVar)
			return &notifierUm.Reply{Email: "null"}, errorVar
		}
	} else {
		emailString := email[0]
		return &notifierUm.Reply{Email: emailString}, nil
	}
}

func deleteAccountHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		setResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if contentType := request.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		setResponseMessage(writer, http.StatusBadRequest, "Error: the request must be in JSON format")
		return
	}
	var cred credentials
	err := json.NewDecoder(request.Body).Decode(&cred)
	if err != nil {
		setResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error in reading data: %s", err))
		return
	}
	email := cred.email
	password := cred.password
	hashPsw := calculateHash(password)
	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err = dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		setResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %s", err))
		return
	}

	query := fmt.Sprintf("SELECT id, email, password FROM users WHERE email=%s and password=%s", email, hashPsw)
	_, row, errorVar := dbConn.ExecuteQuery(query, true)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			setResponseMessage(writer, http.StatusUnauthorized, "Email or password wrong! Retry!")
			return
		} else {
			setResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in select from DB table 'users': %s", errorVar))
			return
		}
	}

	result := deleteUserConstraintsByUserId(row[0])
	if result == nil {
		query := fmt.Sprintf("DELETE FROM users WHERE email=%s and password=%s", email, hashPsw)
		_, _, errV := dbConn.ExecuteQuery(query)
		if errV != nil {
			setResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in database delete: %s", err))
			return
		} else {
			setResponseMessage(writer, http.StatusOK, "Account deleted with relative user constraints!")
			return
		}
	}
	setResponseMessage(writer, http.StatusInternalServerError, "Error in gRPC communication, account not deleted")
	return
}

func loginHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		setResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if contentType := request.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		setResponseMessage(writer, http.StatusBadRequest, "Error: the request must be in JSON format")
		return
	}
	var cred credentials
	err := json.NewDecoder(request.Body).Decode(&cred)
	if err != nil {
		setResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error in reading data: %s", err))
		return
	}
	email := cred.email
	password := cred.password
	hashPsw := calculateHash(password)
	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err = dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		setResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %s", err))
		return
	}

	query := fmt.Sprintf("SELECT email, password FROM users WHERE email='%s' AND password='%s'", email, hashPsw)
	_, _, errorVar := dbConn.ExecuteQuery(query, true)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			setResponseMessage(writer, http.StatusUnauthorized, "Email or password wrong! Retry!")
			return
		} else {
			setResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in select from DB table 'users': %s", errorVar))
			return
		}
	}

	tokenExpireTime := time.Now().Add(72 * time.Hour) // 3 days from now

	// Create the payload
	payload := jwt.MapClaims{
		"email": email,
		"exp":   tokenExpireTime.Unix(),
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// Sign the token with the hashed password
	tokenString, errV := token.SignedString([]byte(hashPsw))
	if errV != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error in signing JWT Token: %v\n", errV)
		return
	}

	setResponseMessage(writer, http.StatusOK, fmt.Sprintf("Login successfully made! JWT Token: %s", tokenString))
	return
}

func registerHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		setResponseMessage(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if contentType := request.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		setResponseMessage(writer, http.StatusBadRequest, "Error: the request must be in JSON format")
		return
	}
	var cred credentials
	err := json.NewDecoder(request.Body).Decode(&cred)
	if err != nil {
		setResponseMessage(writer, http.StatusBadRequest, fmt.Sprintf("Error in reading data: %s", err))
		return
	}
	email := cred.email
	password := cred.password
	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err = dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		setResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %s", err))
		return
	}

	query := fmt.Sprintf("SELECT email FROM users WHERE email=%s", email)
	_, _, errorVar := dbConn.ExecuteQuery(query, true)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			// if there is no row this means that the user is not yet registered
			hashPsw := calculateHash(password)
			query = fmt.Sprintf("INSERT INTO users (email, password) VALUES (%s, %s)", email, hashPsw)
			_, _, err := dbConn.ExecuteQuery(query)
			if err != nil {
				setResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in database insert: %s", err))
				return
			}
		} else {
			setResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in connecting to database: %s", err))
			return
		}
	} else {
		setResponseMessage(writer, http.StatusBadRequest, "Email already in use! Try to sign in!")
		return
	}

	setResponseMessage(writer, http.StatusOK, "Registration made successfully! Now try to sign in!")
}

func serveWMS() {
	defer wg.Done()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *portWMS))
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to listen to requests from WMS: %v", err)
	}
	wmsServer := grpc.NewServer()
	wmsUm.RegisterWMSUmServer(wmsServer, &umWmsServer{})
	log.SetPrefix("[INFO] ")
	log.Printf("WMS server listening at %v", lis.Addr())
	if err := wmsServer.Serve(lis); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to serve WMS: %v", err)
	}
}

func serveNotifier() {
	defer wg.Done()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *portNotifier))
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to listen to requests from Notifier: %v", err)
	}
	notifierServer := grpc.NewServer()
	notifierUm.RegisterNotifierUmServer(notifierServer, &umNotifierServer{})
	log.SetPrefix("[INFO] ")
	log.Printf("Notifier server listening at %v", lis.Addr())
	if err := notifierServer.Serve(lis); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to serve Notifier: %v", err)
	}
}

func serveAPIGateway() {
	defer wg.Done()
	port := "50053"

	hostname, _ := os.Hostname()
	log.SetPrefix("[INFO] ")
	log.Printf("Hostname: %s server starting on port: %s", hostname, port)

	router := mux.NewRouter()
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/register", registerHandler).Methods("POST")
	router.HandleFunc("/delete_account", deleteAccountHandler).Methods("POST")
	log.SetPrefix("[ERROR] ")
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}

func main() {

	siInstance := types.NewSecretInitializer()
	siInstance.InitSecrets()

	log.SetPrefix("[INFO] ")
	log.Println("ENV variables initialization done!")

	// Creating table 'users' if not exits
	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err := dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Exit after DB connection error! -> %s", err)
	}

	query := "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTO_INCREMENT, email VARCHAR(30) UNIQUE NOT NULL, password VARCHAR(64) NOT NULL)"
	_, _, err = dbConn.ExecuteQuery(query)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Exit after DB error in creating 'users' table: %v", err)
	}

	log.SetPrefix("[INFO] ")
	log.Println("Starting notifier serving goroutine!")

	go serveNotifier()
	go serveAPIGateway()
	go serveWMS()
	wg.Add(3)
	wg.Wait()
	log.SetPrefix("[INFO] ")
	log.Println("All goroutines have finished. Exiting...")

}
