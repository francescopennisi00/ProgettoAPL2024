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
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	pb "um_microservice"
	"um_microservice/src/types"
)

type credentials struct {
	email    string
	password string
}

var wg sync.WaitGroup

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedNotifierUmServer
}

func (s *server) RequestEmail(ctx context.Context, in *pb.Request) (*pb.Reply, error) {

	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err := dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		return &pb.Reply{Email: "null"}, nil
	}

	userId := in.UserId
	query := fmt.Sprintf("SELECT email FROM users WHERE id=%d", userId)
	_, email, errorVar := dbConn.ExecuteQuery(true, query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			log.SetPrefix("[INFO] ")
			log.Printf("Email not present anymore")
			return &pb.Reply{Email: "not present anymore"}, nil
		} else {
			log.SetPrefix("[ERROR] ")
			log.Printf("DB Error: %v\n", errorVar)
			return &pb.Reply{Email: "null"}, errorVar
		}
	} else {
		emailString := email[0]
		return &pb.Reply{Email: emailString}, nil
	}
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
	_, _, errorVar := dbConn.ExecuteQuery(true, query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			setResponseMessage(writer, http.StatusUnauthorized, "Email or password wrong! Retry!")
			return
		} else {
			setResponseMessage(writer, http.StatusInternalServerError, fmt.Sprintf("Error in SELECT users, error in connecting to database: %s", errorVar))
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
	tokenString, err := token.SignedString([]byte(hashPsw))
	if err != nil {
		fmt.Println("Error in signing JWT Token:", err)
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
	_, _, errorVar := dbConn.ExecuteQuery(true, query)
	if errorVar != nil {
		if errors.Is(errorVar, sql.ErrNoRows) {
			// if there is no row this means that the user is not yet registered
			hashPsw := calculateHash(password)
			query = fmt.Sprintf("INSERT INTO users (email, password) VALUES (%s, %s)", email, hashPsw)
			_, _, err := dbConn.ExecuteQuery(false, query)
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
func serveNotifier() {
	defer wg.Done()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.SetPrefix("[ERROR]")
		log.Fatalf("Failed to listen to requests from Notifier: %v", err)
	}
	notifierServer := grpc.NewServer()
	pb.RegisterNotifierUmServer(notifierServer, &server{})
	log.SetPrefix("[INFO] ")
	log.Printf("Notifier server listening at %v", lis.Addr())
	if err := notifierServer.Serve(lis); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to serve Notifier: %v", err)
	}
}

func serveAPIgateway() {
	defer wg.Done()
	port := "50053"

	hostname, _ := os.Hostname()
	log.SetPrefix("[INFO]")
	log.Printf("Hostname: %s server starting on port: %s", hostname, port)

	router := mux.NewRouter()
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/register", registerHandler).Methods("POST")
	//router.HandleFunc("/delete_account", deleteAccountHandler).Methods("POST")
	log.SetPrefix("[ERROR]")
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}

func main() {

	siInstance := types.NewSecretInitializer()
	siInstance.InitSecrets()

	log.Println("ENV variables initialization done!")

	// Creating table 'users' if not exits
	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	_, err := dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR]")
		log.Fatalf("Exit after da DB connection error! -> %s", err)
	}

	query := "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTO_INCREMENT, email VARCHAR(30) UNIQUE NOT NULL, password VARCHAR(64) NOT NULL)"
	_, _, err = dbConn.ExecuteQuery(false, query)
	if err != nil {
		log.SetPrefix("[ERROR]")
		log.Fatalf("Exit after DB query execution error!")
	}

	fmt.Println("Starting notifier serving goroutine!")

	go serveNotifier()
	go serveAPIgateway()
	wg.Add(2)
	wg.Wait()
	log.SetPrefix("[INFO]")
	log.Println("All goroutines have finished. Exiting...")

}
