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

package user

import (
	"fmt"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	dbconn "github.com/mdaxf/iac/databases"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
)

func execLogin(ctx *gin.Context, username string, password string) {

	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}
	log.Debug("Login execution function is called.")
	querystr := fmt.Sprintf(LoginQuery, username)

	log.Debug(fmt.Sprintf("Query:%s", querystr))

	iDBTx, err := dbconn.DB.Begin()
	defer iDBTx.Rollback()

	if err != nil {
		log.Error(fmt.Sprintf("Begin error:%s", err.Error()))
		ctx.JSON(http.StatusInternalServerError, "Login failed")
	}

	dboperation := dbconn.NewDBOperation(username, iDBTx, "User Login")

	rows, err := dboperation.Query(querystr)
	defer rows.Close()
	if err != nil {

		log.Error(fmt.Sprintf("Query error:%s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for rows.Next() {
		var ID int
		var Name string
		var FamilyName string

		err = rows.Scan(&ID, &Name, &FamilyName)
		if err != nil {
			log.Error(fmt.Sprintf("Row Scan error:%s", err.Error()))
			ctx.JSON(http.StatusInternalServerError, "Login failed")
			return
		}
		log.Debug(fmt.Sprintf("ID:%d  Name:%s  FamilyName:%s", ID, Name, FamilyName))

		user := User{ID: ID, Username: Name + " " + FamilyName, Email: "", Password: password, SessionID: uuid.New().String()}

		log.Debug(fmt.Sprintf("user:%v", user))

		Columns := []string{"LastSignOnDate", "LastUpdateOn", "LastUpdatedBy"}
		Values := []string{time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"), username}
		datatypes := []int{0, 0, 0}
		Wherestr := fmt.Sprintf("ID= %d", ID)

		index, errr := dboperation.TableUpdate(TableName, Columns, Values, datatypes, Wherestr)

		if errr != nil {
			log.Error(fmt.Sprintf("TableUpdate error:%s", errr.Error()))
			ctx.JSON(http.StatusInternalServerError, "Login failed")
			return
		}
		log.Debug(fmt.Sprintf("index:%d", index))

		iDBTx.Commit()

		exist, err := config.SessionCache.IsExist(ctx, "USER_"+string(rune(ID)))

		if err != nil && exist {
			config.SessionCache.Delete(ctx, "USER_"+string(rune(ID)))

		}
		config.SessionCache.Put(ctx, "USER_"+string(rune(ID)), user, 10*time.Minute)
		log.Debug(fmt.Sprintf("User:%s login successful!", user))

		ctx.JSON(http.StatusOK, user)
		return
	}
	log.Error(fmt.Sprintf("Login failed for user:%s", username))

	ctx.JSON(http.StatusNotFound, "Login failed")

}

/*
	func CheckUserLoginSession(ctx *gin.Context, UserID int) (User, error) {
		exist, err := config.SessionCache.IsExist(ctx, "USER_"+string(rune(UserID)))

		if err != nil && exist {
			return User{config.SessionCache.Get("USER_" + string(rune(UserID)))}, err
		}

		return User{}, err

}

	func UpdateUserSession(ctx *gin.Context, UserID int) (User, error) {
		user, err := CheckUserLoginSession(ctx, UserID)

		if err != nil {
			config.SessionCache.Delete(ctx, "USER_"+string(rune(UserID)))
			config.SessionCache.Put(ctx, "USER_"+string(rune(UserID)), user, 10*time.Minute)
		}
		return user, err
	}
*/
func execLogout(ctx *gin.Context, UserID string) (string, error) {

	config.SessionCache.Delete(ctx, "USER_"+UserID)

	return "OK", nil
}
