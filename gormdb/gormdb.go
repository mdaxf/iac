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

	dbconn "github.com/mdaxf/iac/databases"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global GORM database instance
var DB *gorm.DB

// InitGormDB initializes the GORM database connection using the existing SQL connection
func InitGormDB() error {
	if dbconn.DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Get DSN from the existing connection
	// We'll need to reconstruct it or get it from config
	// For now, use a basic initialization

	// Open GORM with MySQL driver using the existing connection
	var err error
	DB, err = gorm.Open(mysql.New(mysql.Config{
		Conn: dbconn.DB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return fmt.Errorf("failed to initialize GORM: %v", err)
	}

	return nil
}
