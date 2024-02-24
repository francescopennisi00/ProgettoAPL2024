package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	notifierUm "um_microservice/proto/notifier_um"
	wmsUm "um_microservice/proto/wms_um"
	grpcC "um_microservice/src/communication_grpc"
	httpC "um_microservice/src/http_handlers"
	umTypes "um_microservice/src/types"
	umUtils "um_microservice/src/utils"
)

var wg sync.WaitGroup

var (
	portWMS      = flag.Int("portWMS", umUtils.PortWMS, "The server port for WMS")
	portNotifier = flag.Int("portNotifier", umUtils.PortNotifier, "The server port for Notifier")
)

func serveWMS() {
	defer wg.Done()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *portWMS))
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to listen to requests from WMS: %v\n", err)
	}
	wmsServer := grpc.NewServer()
	wmsUm.RegisterWMSUmServer(wmsServer, &grpcC.UmWmsServer{})
	log.SetPrefix("[INFO] ")
	log.Printf("WMS server listening at %v\n", lis.Addr())
	if err := wmsServer.Serve(lis); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to serve WMS: %v\n", err)
	}
}

func serveNotifier() {
	defer wg.Done()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *portNotifier))
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to listen to requests from Notifier: %v\n", err)
	}
	notifierServer := grpc.NewServer()
	notifierUm.RegisterNotifierUmServer(notifierServer, &grpcC.UmNotifierServer{})
	log.SetPrefix("[INFO] ")
	log.Printf("Notifier server listening at %v\n", lis.Addr())
	if err := notifierServer.Serve(lis); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to serve Notifier: %v\n", err)
	}
}

func serveAPIGateway() {
	defer wg.Done()
	port := umUtils.PortAPIGateway

	hostname, _ := os.Hostname()
	log.SetPrefix("[INFO] ")
	log.Printf("Hostname: %s server starting on port: %s\n", hostname, port)

	router := mux.NewRouter()
	router.HandleFunc("/login", httpC.LoginHandler).Methods("POST")
	router.HandleFunc("/register", httpC.RegisterHandler).Methods("POST")
	router.HandleFunc("/delete_account", httpC.DeleteAccountHandler).Methods("POST")
	log.SetPrefix("[ERROR] ")
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}

func main() {

	siInstance := umTypes.NewSecretInitializer()
	siInstance.InitSecrets()

	log.SetPrefix("[INFO] ")
	log.Println("ENV variables initialization done!")

	// Creating table 'users' if not exits
	var dbConn umTypes.DatabaseConnector
	_, err := dbConn.StartDBConnection(umUtils.DBConnString)
	defer func(database *umTypes.DatabaseConnector) {
		_ = database.CloseConnection()
	}(&dbConn)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Exit after DB connection error! -> %s\n", err)
	}

	query := "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTO_INCREMENT, email VARCHAR(30) UNIQUE NOT NULL, password VARCHAR(64) NOT NULL)"
	_, _, err = dbConn.ExecuteQuery(query)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Exit after DB error in creating 'users' table: %v\n", err)
	}

	log.SetPrefix("[INFO] ")
	log.Println("Starting serving Notifier goroutine!")
	go serveNotifier()
	log.SetPrefix("[INFO] ")
	log.Println("Starting serving API gateway goroutine!")
	go serveAPIGateway()
	log.SetPrefix("[INFO] ")
	log.Println("Starting serving WMS goroutine!")
	go serveWMS()
	wg.Add(3)
	wg.Wait()
	log.SetPrefix("[INFO] ")
	log.Println("All goroutines have finished. Exiting...")

}
