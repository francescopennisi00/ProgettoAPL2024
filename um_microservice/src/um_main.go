package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	pb "um_microservice"
	"um_microservice/src/types"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedNotifierUmServer
}

func (s *server) RequestEmail(ctx context.Context, in *pb.Request) (*pb.Reply, error) {

	var dbConn types.DatabaseConnector
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE"))
	db, err := dbConn.StartDBConnection(dataSource)
	defer func(database *types.DatabaseConnector) {
		err := database.CloseConnection()
		if err != nil {
			os.Exit(-1)
		}
	}(&dbConn)
	if err != nil {
		return &pb.Reply{Email: "null"}, nil
	}

	//TODO: replace following code with call to DatabaseConnector's method
	row := db.QueryRow("SELECT email FROM users WHERE id=?", int(in.UserId))
	var email string
	err = row.Scan(&email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.SetPrefix("[INFO] ")
			log.Printf("Email %s not present anymore: %v\n", email, err)
			return &pb.Reply{Email: "not present anymore"}, nil
		} else {
			log.SetPrefix("[ERROR] ")
			log.Printf("Error scanning row: %v\n", err)
			return &pb.Reply{Email: "null"}, err
		}
	} else {
		log.SetPrefix("[INFO] ")
		log.Printf("Returning email %s\n to Notifier", email)
		return &pb.Reply{Email: email}, nil
	}

}

func serveNotifier() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.SetPrefix("[ERROR]")
		log.Fatalf("Failed to listen to requests from Notifier: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterNotifierUmServer(s, &server{})
	log.SetPrefix("[INFO] ")
	log.Printf("Notifier server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.SetPrefix("[ERROR] ")
		log.Fatalf("Failed to serve Notifier: %v", err)
	}
}

func main() {

	siInstance := types.NewSecretInitializer()
	siInstance.InitSecrets()

	log.Println("ENV variables initialization done!")
}
