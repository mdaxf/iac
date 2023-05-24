package documents

var DocDBCon *DocDB

func ConnectDB(DatabaseType string, DatabaseConnection string, DatabaseName string) {

	if DatabaseType == "mongodb" {
		var err error
		DocDBCon, err = InitMongDB(DatabaseConnection, DatabaseName)

		if err != nil {
			//ilog.Error(fmt.Sprintf("initialize Database error: %s", err.Error()))
		}
	}

}
