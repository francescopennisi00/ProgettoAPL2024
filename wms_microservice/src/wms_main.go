package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"sync"
	"time"
	wmsUm "wms_microservice/proto"
	grpcC "wms_microservice/src/communication_grpc"
	"wms_microservice/src/types"
)

var wg sync.WaitGroup

var (
	portUm = flag.Int("portUm", 50052, "The server port for UM")
)

func timer(duration time.Duration, event chan<- bool) {
	for {
		time.Sleep(duration)
		event <- true
	}
}
func serveUm() {
	defer wg.Done()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *portUm))
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to listen to requests from Um: %v\n", err)
	}
	wmsServer := grpc.NewServer()
	wmsUm.RegisterWMSUmServer(wmsServer, &grpcC.WmsUmServer{})
	log.SetPrefix("[INFO] ")
	log.Printf("UM server listening at %v\n", lis.Addr())
	if err := wmsServer.Serve(lis); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to serve UM: %v\n", err)
	}
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
		log.Fatalf("Exit after DB connection error! -> %s\n", err)
	}

	query := "CREATE TABLE IF NOT EXISTS locations (id INTEGER PRIMARY KEY AUTO_INCREMENT, location_name VARCHAR(100) NOT NULL, latitude FLOAT NOT NULL, longitude FLOAT NOT NULL, country_code VARCHAR(10) NOT NULL, state_code VARCHAR(70) NOT NULL, UNIQUE KEY location_tuple (location_name, latitude, longitude));"
	_, _, err = dbConn.ExecuteQuery(query)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Exit after DB error in creating 'users' table: %v\n", err)
	}
	query = "CREATE TABLE IF NOT EXISTS user_constraints (id INTEGER PRIMARY KEY AUTO_INCREMENT, user_id INTEGER NOT NULL, location_id INTEGER NOT NULL, rules JSON NOT NULL, time_stamp TIMESTAMP NOT NULL, trigger_period INTEGER NOT NULL, checked BOOLEAN NOT NULL DEFAULT FALSE, FOREIGN KEY (location_id) REFERENCES locations(id), UNIQUE KEY user_location_id (user_id, location_id));"
	_, _, err = dbConn.ExecuteQuery(query)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Exit after DB error in creating 'users' table: %v\n", err)
	}

	// TODO Kafka Producer inizialization

	log.SetPrefix("[INFO] ")
	log.Println("Starting serving goroutine!")
	// TODO APIGateway?
	// go serveAPIGateway()
	go serveUm()

	log.Println("Starting timer thread!")
	expiredTimerEvent := make(chan bool)
	go timer(60*time.Second, expiredTimerEvent)

	for {
		// Wait for expired timer event
		<-expiredTimerEvent

		// TODO Check in DB to find events to send

	}
	// Codice inraggiungibile lasciamo perdere wg.wait() oppure usiamo un'altra goroutine per
	// fare check in DB to find events to send?
	wg.Add(1)
	wg.Wait()
	log.SetPrefix("[INFO] ")
	log.Println("All goroutines have finished. Exiting...")

}
