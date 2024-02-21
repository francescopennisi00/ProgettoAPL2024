package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strconv"
	"time"
	pb "um_microservice"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedNotifierUmServer
}

func (s *server) RequestEmail(ctx context.Context, in *pb.Request) (*pb.Reply, error) {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("HOSTNAME"), os.Getenv("PORT"), os.Getenv("DATABASE")))
	defer db.Close()
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error connecting to the database: %v", err)
		return &pb.Reply{Email: "null"}, nil
	}

	row := db.QueryRow("SELECT email FROM users WHERE id=?", strconv.Itoa(int(pb.Request.)))

	var email string
	err = row.Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			email = "not present anymore"
		} else {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
	}

	DBendTime := time.Now()
	QUERY_DURATIONS_HISTOGRAM.Observe(float64(DBendTime.Sub(DBstartTime).Nanoseconds()))

	RESPONSE_TO_NOTIFIER.Inc()

	return &Reply{Email: email}, nil
}

/*
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
*/

func main() {

}
