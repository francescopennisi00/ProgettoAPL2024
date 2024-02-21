package types

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
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
			return nil
		}
	}
	database.dbConn = nil
	return nil
}

func (database *DatabaseConnector) ExecuteQuery(query string, fetchOne bool, selectQ bool, params ...any) (booleanOutcome bool, result sql.Result, row *sql.Row, rows *sql.Rows, resErr error) {

	if database.dbConn != nil {
		if selectQ == true {
			if fetchOne == true {
				row := database.dbConn.QueryRow(query, params)
				var scanVariable sql.Row
				err := row.Scan(scanVariable)
				if err != nil {
					return false, result, &scanVariable, rows, err
				}
			} else {
				res, err := database.dbConn.Query(query, params)
				if err != nil {
					return false, result, row, rows, err
				}
				for res.Next() {
					var scanVariable sql.Row
					err := res.Scan(&scanVariable)
					if err != nil {
						return false, result, &scanVariable, rows, err
					}
				}
				return true, result, row, res, err
			}
		} else {
			exec, err := database.dbConn.Exec(query, params)
			if err != nil {
				return false, result, row, rows, err
			}
			return false, exec, row, rows, err
		}
	}
	return false, result, row, rows, nil
}
