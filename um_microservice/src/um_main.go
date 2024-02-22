package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"sync"
	pb "um_microservice"
	"um_microservice/src/types"
)

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
			log.Printf("DB Error: %v\n", err)
			return &pb.Reply{Email: "null"}, err
		}
	} else {
		if emailString, ok := email[0].(string); ok {
			return &pb.Reply{Email: emailString}, nil
		} else {
			log.Println("Email is not a string!")
			return &pb.Reply{Email: "null"}, nil
		}
	}
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

	fmt.Println("Starting notifier serving thread!")
	wg.Add(1)
	go serveNotifier()
	wg.Wait()
	log.SetPrefix("[INFO]")
	log.Println("All goroutines have finished. Exiting...")

}
