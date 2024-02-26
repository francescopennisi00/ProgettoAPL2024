package um_types

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
	umUtils "um_microservice/um-src/um-utils"
)

type DatabaseConnector struct {
	dbConn      *sql.DB
	transaction *sql.Tx
}

func (database *DatabaseConnector) StartDBConnection(dataSource string) (*sql.DB, error) {
	var err error
	database.dbConn, err = sql.Open(umUtils.DBDriver, dataSource)
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

func (database *DatabaseConnector) ExecuteQuery(query string) (outcome sql.Result, results [][]string, resErr error) {

	if database.dbConn != nil {

		if strings.HasPrefix(query, "SELECT ") {

			// result is a []string variable for generic row, while named return parameter results is a
			// [][]string and contains all the rows of the query result
			var result []string

			// execute query and put rows into res
			res, err := database.dbConn.Query(query)
			if err != nil {
				return nil, results, err
			}

			// extract column names in order to know the number of columns
			columns, errCol := res.Columns()
			if errCol != nil {
				return nil, results, errCol
			}

			// create a pointers array with the same length of columns number (required for Scan)
			pointers := make([]interface{}, len(columns))
			for j := range pointers {
				pointers[j] = new(string)
			}

			// row index (initially 0)
			i := 0

			// managing rows
			for res.Next() {
				errorVar := res.Scan(pointers...) //scan of the whole row
				if errorVar != nil {
					return nil, results, errorVar
				}
				// extract values stored in pointers and add then to result
				for _, pointer := range pointers {
					result = append(result, *pointer.(*string))
				}

				log.Printf("RESULT[%d]: ", i)
				for _, item := range result {
					log.Printf("%s - ", item)
				}
				// append row to rows array ([][]string) to be returned
				results = append(results, result)
				i++
			}
			return nil, results, nil

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

func (database *DatabaseConnector) ExecIntoTransaction(query string) (outcome sql.Result, results [][]string, resErr error) {

	if database.transaction != nil {
		if strings.HasPrefix(query, "SELECT ") {

			// result is a []string variable for generic row, while named return parameter results is a
			// [][]string and contains all the rows of the query result
			var result []string

			// execute query and put rows into res
			res, err := database.dbConn.Query(query)
			if err != nil {
				return nil, results, err
			}

			// extract column names in order to know the number of columns
			columns, errCol := res.Columns()
			if errCol != nil {
				return nil, results, errCol
			}

			// create a pointers array with the same length of columns number (required for Scan)
			pointers := make([]interface{}, len(columns))
			for j := range pointers {
				pointers[j] = new(string)
			}

			// row index (initially 0)
			i := 0

			// managing rows
			for res.Next() {
				errorVar := res.Scan(pointers...) //scan of the whole row
				if errorVar != nil {
					return nil, results, errorVar
				}
				// extract values stored in pointers and add then to result
				for _, pointer := range pointers {
					result = append(result, *pointer.(*string))
				}

				log.Printf("RESULT[%d]: ", i)
				for _, item := range result {
					log.Printf("%s - ", item)
				}
				// append row to rows array ([][]string) to be returned
				results = append(results, result)
				i++
			}
			return nil, results, nil
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
