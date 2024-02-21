package types

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

type DatabaseConnector struct {
	dbConn *sql.DB
}

func (database *DatabaseConnector) StartDBConnection(dataSource string) (*sql.DB, error) {
	var err error
	database.dbConn, err = sql.Open("mysql", dataSource)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error connecting to the database: %v", err)
		return database.dbConn, err
	}
	return database.dbConn, nil
}

func (database *DatabaseConnector) CloseConnection() error {
	err := database.dbConn.Close()
	if err != nil {
		log.SetPrefix("[ERROR]")
		log.Printf("DB connection closing error: %v", err)
		return err
	}
	return nil
}

func (database *DatabaseConnector) ExecuteQuery(envVarName string) {

	//TODO: replace following code with implementation
	secretPath := os.Getenv(envVarName)
	secretValue, err := os.ReadFile(secretPath)
	if err != nil {
		log.Fatalf("Error reading secret file for %s: %v", envVarName, err)
	}
	err = os.Setenv(envVarName, string(secretValue))
	if err != nil {
		log.Printf("Initialized %s.\n", envVarName)
		return
	}
}
