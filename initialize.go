package main

import (
	dbconn "github.com/mdaxf/iac/databases"
)

func initialize() {

	initializeDatabase()
}

func initializeDatabase() {

	err := dbconn.ConnectDB()
	if err != nil {
		panic(err.Error())
	}

	//log.Printf("database connection initialized: %d", dbconn.DB.Stats().OpenConnections)
	/*	err := dbconn.DBPing()
		if err != nil {
			panic(err.Error())
		}

		rows, err := dbconn.Query("SELECT ID,Name,FamilyName FROM EMPLOYEE WHERE LoginName = '?'", "w4g")
		if err != nil {
			panic(err.Error())
		}
		defer rows.Close()
		log.Println(fmt.Printf("rows: %v\n", rows)) */
}
