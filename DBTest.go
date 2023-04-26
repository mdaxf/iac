package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/denisenkom/go-mssqldb"
)

var db *sql.DB
var err error

func dbtest() {
	// Set up the connection string
	connString := "server=delsrv000039;user id=flxadmin;password=DS@3ds23;database=LPM23"

	// Open a connection to the database
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error connecting to database: ", err.Error())
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database: ", err.Error())
	}
}

func querydata() {
	// Query the database for all rows in the users table
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database: ", err.Error())
	}
	rows, err := db.Query("SELECT ID, LoginName FROM EMPLOYEE")
	if err != nil {
		log.Fatal("Error executing query: ", err.Error())
	}
	defer rows.Close()

	// Loop through the rows and print the values of each column
	for rows.Next() {
		var username string
		var email string
		err = rows.Scan(&username, &email)
		if err != nil {
			log.Fatal("Error scanning row: ", err.Error())
		}
		fmt.Println(username, email)
	}
}
