package types

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
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
	if database.dbConn != nil {
		err := database.dbConn.Close()
		if err != nil {
			log.SetPrefix("[ERROR]")
			log.Printf("DB connection closing error: %v", err)
			return err
		} else {
			database.dbConn = nil
			return nil
		}
	}
	return nil
}

func (database *DatabaseConnector) ExecuteQuery(fetchOne bool, query string) (outcome sql.Result, results []interface{}, resErr error) {

	if database.dbConn != nil {
		if strings.HasPrefix(query, "SELECT ") {
			if fetchOne == true {
				row := database.dbConn.QueryRow(query)
				err := row.Scan(results)
				if err != nil {
					return nil, results, err
				} else {
					return nil, results, nil
				}
			} else {
				res, err := database.dbConn.Query(query)
				if err != nil {
					return nil, results, nil
				}
				i := 0
				for res.Next() {
					errorVar := res.Scan(results[i])
					if errorVar != nil {
						return nil, results, errorVar
					}
					i++
				}
				return nil, results, nil
			}
		} else {
			exec, err := database.dbConn.Exec(query)
			if err != nil {
				return nil, results, err
			}
			return exec, results, nil
		}
	}
	log.SetPrefix("[ERROR]")
	log.Println("DB connection already closed!")
	er := sql.ErrConnDone
	return nil, results, er
}
