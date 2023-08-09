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
	dbconn "github.com/mdaxf/iac/databases"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/framework/auth"
	"github.com/mdaxf/iac/logger"

	"golang.org/x/crypto/bcrypt"
)

func execLogin(ctx *gin.Context, username string, password string, clienttoken string, ClientID string, Renew bool) {

	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}
	log.Debug("Login execution function is called.")
	log.Debug(fmt.Sprintf("login parameters:%s  %s  token: %s  renew:%s", username, password, clienttoken, Renew))

	if Renew {

		user, err := config.SessionCache.Get(ctx, clienttoken)

		log.Debug(fmt.Sprintf("SessionCache user:%v for token:%s", user, clienttoken))

		if err != nil {
			log.Error(fmt.Sprintf("SessionCache error:%s for token", err.Error(), clienttoken))
		} else {
			log.Debug(fmt.Sprintf("SessionCache user:%v for token:%s", user, clienttoken))
			if user != nil {
				if user.(User).Token == clienttoken {
					layout := "2006-01-02 15:04:05"
					parsedTime, err := time.Parse(layout, user.(User).ExpirateOn)
					log.Debug(fmt.Sprintf("token %s expirate on:%s expired? %s", user.(User).Token, parsedTime, parsedTime.Before(time.Now())))
					if err != nil {
						log.Error(fmt.Sprintf("SessionCache error:%s for token:%s", err.Error(), clienttoken))
					} else if parsedTime.Before(time.Now()) {
						log.Debug(fmt.Sprintf("renew the session for user:%s, token: %s", username, user.(User).Token))

						token, createdt, expdt, err := auth.Extendexptime(user.(User).Token)
						if err != nil {
							log.Error(fmt.Sprintf("SessionCache error:%s for token:%s", err.Error(), clienttoken))
						} else {
							ID := user.(User).ID
							Username := user.(User).Username
							hasedPassword := user.(User).Password
							user = User{ID: ID, Username: Username, Email: "", Password: hasedPassword, ClientID: ClientID, CreatedOn: createdt, ExpirateOn: expdt, Token: token}
							config.SessionCache.Delete(ctx, clienttoken)
							config.SessionCache.Put(ctx, token, user, 15*time.Minute)
							ctx.JSON(http.StatusOK, user)
							return
						}
					}
					log.Debug(fmt.Sprintf("token %s already expried", clienttoken))
				}
			}

		}
		log.Error(fmt.Sprintf("Renew session failed for user:%s", username))

		ctx.JSON(http.StatusNotFound, "Renew failed")

		return
	}

	hasedPassword, err := hashPassword(password)
	querystr := fmt.Sprintf(LoginQuery, username, hasedPassword)

	log.Debug(fmt.Sprintf("Query:%s", querystr))

	iDBTx, err := dbconn.DB.Begin()
	defer iDBTx.Rollback()

	if err != nil {
		log.Error(fmt.Sprintf("Begin error:%s", err.Error()))
		ctx.JSON(http.StatusInternalServerError, "Login failed")
	}

	dboperation := dbconn.NewDBOperation(username, iDBTx, "User Login")
	/*
		rows, err := dboperation.Query(querystr)
		defer rows.Close()
		if err != nil {

			log.Error(fmt.Sprintf("Query error:%s", err.Error()))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Debug(fmt.Sprintf("Query result:%v", rows))

	*/
	jdata, err := dboperation.Query_Json(querystr)

	if err != nil {

		log.Error(fmt.Sprintf("Query error:%s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Debug(fmt.Sprintf("Query result:%v", jdata))

	if jdata != nil {
		var ID int
		var Name string
		var FamilyName string
		/*
			err = rows.Scan(&ID, &Name, &FamilyName)
			if err != nil {
				log.Error(fmt.Sprintf("Row Scan error:%s", err.Error()))
				ctx.JSON(http.StatusInternalServerError, "Login failed")
				return
			} */
		ID = int(jdata[0]["ID"].(int64))
		Name = jdata[0]["Name"].(string)
		FamilyName = jdata[0]["LastName"].(string)

		log.Debug(fmt.Sprintf("ID:%d  Name:%s  FamilyName:%s", ID, Name, FamilyName))

		Columns := []string{"LastSignOnDate", "UpdatedOn", "UpdatedBy"}
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

		token, createdt, expdt, err := auth.Generate_authentication_token(string(rune(ID)), username, ClientID)

		sessionid := token
		exist, err := config.SessionCache.IsExist(ctx, sessionid)

		if err != nil && exist {
			config.SessionCache.Delete(ctx, sessionid)

		}
		user := User{ID: ID, Username: username, Email: "", Password: hasedPassword, ClientID: ClientID, CreatedOn: createdt, ExpirateOn: expdt, Token: token}

		log.Debug(fmt.Sprintf("user:%v", user))

		config.SessionCache.Put(ctx, sessionid, user, 15*time.Minute)
		log.Debug(fmt.Sprintf("User:%s login successful!", user))

		ctx.JSON(http.StatusOK, user)
		return
	}

	iDBTx.Rollback()
	log.Error(fmt.Sprintf("Login failed for user:%s", username))

	ctx.JSON(http.StatusNotFound, "Login failed")

}

func getUserImage(username string) (string, error) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}
	log.Debug("Get User Image execution function is called.")

	querystr := fmt.Sprintf(GetUserImageQuery, username)

	log.Debug(fmt.Sprintf("Get User image Query:%s", querystr))

	iDBTx, err := dbconn.DB.Begin()
	defer iDBTx.Rollback()

	if err != nil {
		log.Error(fmt.Sprintf("Begin error:%s", err.Error()))
		return "", err
	}

	dboperation := dbconn.NewDBOperation(username, iDBTx, "User Login")

	jdata, err := dboperation.Query_Json(querystr)

	if err != nil {

		log.Error(fmt.Sprintf("Query error:%s", err.Error()))
		return "", err

	}

	log.Debug(fmt.Sprintf("Query result:%v", jdata))

	if jdata != nil {
		var PictureUrl string

		if len(jdata) == 0 {
			return "", nil
		}

		if jdata[0]["PictureUrl"] == nil {
			return "", nil
		}

		PictureUrl = jdata[0]["PictureUrl"].(string)

		iDBTx.Commit()

		log.Debug(fmt.Sprintf("PictureUrl:%s", PictureUrl))

		return PictureUrl, nil

	}
	return "", nil
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
func execLogout(ctx *gin.Context, token string) (string, error) {

	config.SessionCache.Delete(ctx, token)

	return "OK", nil
}

func hashPassword(password string) (string, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hashedPassword := string(hashedPasswordBytes)
	return hashedPassword, nil
}
