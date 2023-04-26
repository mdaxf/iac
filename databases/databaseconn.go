package dbconn

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB
var (
	/*
		mysql,sqlserver, goracle
	*/
	DatabaseType = "sqlserver"

	/*
		user:password@tcp(localhost:3306)/mydb
		server=%s;port=%d;user id=%s;password=%s;database=%s
	*/
	DatabaseConnection = "server=delsrv000039;user id=flxadmin;password=DS@3ds23;database=LPM23"

	once sync.Once
	err  error
)

func ConnectDB() error {
	/*once.Do(func() {
		DB, err = sql.Open(DatabaseType, DatabaseConnection)
		if err != nil {
			panic(err.Error())
		}

		//	defer DB.Close()
		DB.SetMaxIdleConns(1000)
		DB.SetMaxOpenConns(10000)
	}) */
	DB, err = sql.Open(DatabaseType, DatabaseConnection)
	if err != nil {
		panic(err.Error())
	}

	DB.SetMaxIdleConns(5)
	DB.SetMaxOpenConns(10)

	return nil
}

func DBPing() error {

	return DB.Ping()

}

func Query(querystr string, args ...interface{}) (*sql.Rows, error) {
	fmt.Println(string(querystr))
	//fmt.Println(string(args))
	stmt, err := DB.Prepare(querystr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	jsonString, err := json.Marshal(stmt)
	if err != nil {
		fmt.Println("Error marshaling json:", err)

	}
	fmt.Println(string(jsonString))

	rows, err := stmt.Query(args...)

	if err != nil {
		return nil, err
	}

	jsonString, err = json.Marshal(rows)
	if err != nil {
		fmt.Println("Error marshaling json:", err)

	}
	fmt.Println(string(jsonString))

	return rows, nil
}
