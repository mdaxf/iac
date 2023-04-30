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
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	dbconn "github.com/mdaxf/iac/databases"

	"github.com/mdaxf/iac/config"
)

func execLogin(ctx *gin.Context, username string, password string) {

	querystr := fmt.Sprintf(LoginQuery, username)
	log.Println(fmt.Sprintf("query:%s", querystr))

	rows, err := dbconn.DB.Query(querystr)
	defer rows.Close()
	if err != nil {
		panic(err.Error())

		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for rows.Next() {
		var ID int
		var Name string
		var FamilyName string

		err = rows.Scan(&ID, &Name, &FamilyName)
		if err != nil {
			panic(err.Error())
		}
		log.Println(fmt.Sprintf("ID:%d  Name:%s  FamilyName:%s", ID, Name, FamilyName))

		user := User{ID: ID, Username: Name + " " + FamilyName, Email: "", Password: password, SessionID: uuid.New().String()}

		Columns := []string{"LastSignOnDate", "LastUpdateOn", "LastUpdatedBy"}
		Values := []interface{}{time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"), username}
		Wherestr := fmt.Sprintf("ID= %d", ID)

		whereArgs := []interface{}{}

		index, errr := dbconn.TableUpdate(TableName, Columns, Values, Wherestr, whereArgs)
		if errr != nil {
			panic(errr.Error())
		}
		log.Println(fmt.Sprintf("index:%d", index))
		exist, err := config.SessionCache.IsExist(ctx, "USER_"+string(rune(ID)))

		if err != nil && exist {
			config.SessionCache.Delete(ctx, "USER_"+string(rune(ID)))

		}
		config.SessionCache.Put(ctx, "USER_"+string(rune(ID)), user, 10*time.Minute)

		ctx.JSON(http.StatusOK, user)
		return
	}

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
