// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gormdb

import (
	"fmt"
	"strings"

	dbconn "github.com/mdaxf/iac/databases"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global GORM database instance
var DB *gorm.DB

// InitGormDB initializes the GORM database connection using the existing SQL connection
func InitGormDB() error {
	if dbconn.DB == nil {
		return fmt.Errorf("database connection not initialized - dbconn.DB is nil. Please check if ConnectDB() was called successfully")
	}

	// Test the connection first
	if err := dbconn.DB.Ping(); err != nil {
		return fmt.Errorf("database connection is not alive: %v. Please check database connectivity", err)
	}

	// Determine the database type and initialize GORM with the appropriate driver
	dbType := strings.ToLower(dbconn.DatabaseType)
	var dialector gorm.Dialector
	var err error

	switch dbType {
	case "mysql":
		dialector = mysql.New(mysql.Config{
			Conn: dbconn.DB,
		})
	case "postgres", "postgresql":
		dialector = postgres.New(postgres.Config{
			Conn: dbconn.DB,
		})
	case "sqlserver", "mssql":
		return fmt.Errorf("GORM does not support SQL Server due to driver conflicts. SQL Server is supported via the legacy database layer")
	default:
		return fmt.Errorf("unsupported database type for GORM: %s. Supported types: mysql, postgres", dbType)
	}

	// Open GORM with the appropriate driver
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return fmt.Errorf("failed to initialize GORM with %s driver: %v", dbType, err)
	}

	// Verify GORM DB is working
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB from GORM: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("GORM database connection test failed: %v", err)
	}

	return nil
}
