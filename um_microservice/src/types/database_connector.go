package types

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

type DatabaseConnector struct {
	dbConn      *sql.DB
	transaction *sql.Tx
}

func (database *DatabaseConnector) StartDBConnection(dataSource string) (*sql.DB, error) {
	var err error
	database.dbConn, err = sql.Open("mysql", dataSource)
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("Error connecting to the database: %v\n", err)
		return database.dbConn, err
	}
	return database.dbConn, nil
}

func (database *DatabaseConnector) CloseConnection() error {
	if database.dbConn != nil {
		err := database.dbConn.Close()
		if err != nil {
			log.SetPrefix("[ERROR] ")
			log.Printf("DB connection closing error: %v\n", err)
			return err
		} else {
			database.dbConn = nil
			return nil
		}
	}
	return nil
}

func (database *DatabaseConnector) ExecuteQuery(query string, fetchOne ...bool) (outcome sql.Result, results []string, resErr error) {

	//default value of fetchOne is false
	if len(fetchOne) == 0 {
		fetchOne = append(fetchOne, false)
	}

	if database.dbConn != nil {
		results = make([]string, 0)
		var result string
		if strings.HasPrefix(query, "SELECT ") {
			if fetchOne[0] == true {
				err := database.dbConn.QueryRow(query).Scan(&result)
				log.SetPrefix("[INFO] ")
				log.Printf("RESULT: %s\n", result)
				results = append(results, result)
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
					errorVar := res.Scan(&result)
					log.SetPrefix("[INFO] ")
					log.Printf("RESULT[%d]: %s\n", i, result)
					results = append(results, result)
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
	log.SetPrefix("[ERROR] ")
	log.Println("DB connection already closed!")
	er := sql.ErrConnDone
	return nil, results, er
}

func (database *DatabaseConnector) BeginTransaction() (*sql.Tx, error) {
	var err error
	database.transaction, err = database.dbConn.Begin()
	if err != nil {
		log.SetPrefix("[ERROR] ")
		log.Printf("DB error in starting transaction: %v\n", err)
		return nil, err
	}
	return database.transaction, nil
}

func (database *DatabaseConnector) ExecIntoTransaction(query string, fetchOne ...bool) (outcome sql.Result, results []string, resErr error) {

	//default value of fetchOne is false
	if len(fetchOne) == 0 {
		fetchOne = append(fetchOne, false)
	}

	if database.transaction != nil {
		results = make([]string, 0)
		var result string
		if strings.HasPrefix(query, "SELECT ") {
			if fetchOne[0] == true {
				err := database.transaction.QueryRow(query).Scan(&result)
				log.SetPrefix("[INFO] ")
				log.Printf("RESULT: %s\n", result)
				results = append(results, result)
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
					errorVar := res.Scan(&result)
					log.SetPrefix("[INFO] ")
					log.Printf("RESULT[%d]: %s\n", i, result)
					results = append(results, result)
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
	log.SetPrefix("[ERROR] ")
	log.Println("DB transaction already terminated!")
	er := sql.ErrTxDone
	return nil, results, er
}

func (database *DatabaseConnector) RollbackTransaction() error {
	if database.transaction != nil {
		var err error
		err = database.transaction.Rollback()
		if err != nil {
			if rollbackErr := database.transaction.Rollback(); rollbackErr != nil {
				log.Printf("Unable to rollback: %v\n", rollbackErr)
				return rollbackErr
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
	log.SetPrefix("[ERROR] ")
	log.Println("DB transaction already terminated!")
	er := sql.ErrTxDone
	return er
}

func (database *DatabaseConnector) CommitTransaction() error {
	if database.transaction != nil {
		err := database.transaction.Commit()
		if err != nil {
			log.SetPrefix("[ERROR] ")
			log.Printf("DB transaction commit error: %v\n", err)
			return err
		} else {
			database.transaction = nil
			database.CloseTransaction()
			return nil
		}
	}
	return nil
}

func (database *DatabaseConnector) CloseTransaction() {
	if database.transaction != nil {
		database.transaction = nil
	}
}
