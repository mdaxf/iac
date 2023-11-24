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
	"strconv"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	dbconn "github.com/mdaxf/iac/databases"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/framework/auth"
	"github.com/mdaxf/iac/logger"
	"golang.org/x/crypto/bcrypt"
)

func execLogin(ctx *gin.Context, username string, password string, clienttoken string, ClientID string, Renew bool) {

	log := logger.Log{ModuleName: logger.API, User: username, ClientID: ClientID, ControllerName: "UserController"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.execLogin", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("execLogin defer error: %s", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	log.Debug("Login execution function is called.")
	log.Debug(fmt.Sprintf("login parameters:%s  %s  token: %s  renew:%s", username, password, clienttoken, Renew))
	fmt.Println("Session Timeout:", config.SessionCacheTimeout)
	if Renew {

		user, err := config.SessionCache.Get(ctx, clienttoken)

		log.Debug(fmt.Sprintf("SessionCache user:%v for token:%s", user, clienttoken))

		if err != nil {
			log.Error(fmt.Sprintf("SessionCache error:%s for token", err.Error(), clienttoken))
		} else {
			log.Debug(fmt.Sprintf("SessionCache user:%v for token:%s", user, clienttoken))
			if user != nil {
				var tokenuser User
				if val, ok := user.(User); ok {
					tokenuser = val
				} else {
					//	log.Debug(fmt.Sprintf("SessionCache error for token:%s", clienttoken))
					for key, value := range user.(map[string]interface{}) {
						log.Debug(fmt.Sprintf("key:%s value:%v", key, value))
						if key == "token" {
							tokenuser.Token = value.(string)
						} else if key == "expirateon" {
							tokenuser.ExpirateOn = value.(string)
						} else if key == "id" {
							tokenuser.ID = int(value.(int32))
						} else if key == "username" {
							tokenuser.Username = value.(string)
						} else if key == "language" {
							tokenuser.Language = value.(string)
						} else if key == "timezone" {
							tokenuser.TimeZone = value.(string)
						} else if key == "clientid" {
							tokenuser.ClientID = value.(string)
						}
					}
				}

				if tokenuser.Token == clienttoken {
					layout := "2006-01-02 15:04:05"
					parsedTime, err := time.Parse(layout, tokenuser.ExpirateOn)
					log.Debug(fmt.Sprintf("token %s expirate on:%s expired? %s", tokenuser.Token, parsedTime, parsedTime.Before(time.Now())))
					if err != nil {
						log.Error(fmt.Sprintf("SessionCache error:%s for token:%s", err.Error(), clienttoken))
					} else if parsedTime.Before(time.Now()) {
						log.Debug(fmt.Sprintf("renew the session for user:%s, token: %s", username, tokenuser.Token))

						token, createdt, expdt, err := auth.Extendexptime(tokenuser.Token)
						if err != nil {
							log.Error(fmt.Sprintf("SessionCache error:%s for token:%s", err.Error(), clienttoken))
						} else {
							ID := tokenuser.ID
							Username := tokenuser.Username
							//	hasedPassword := user.(User).Password
							language := tokenuser.Language
							timezone := tokenuser.TimeZone
							user = User{ID: ID, Username: Username, Language: language, TimeZone: timezone, ClientID: ClientID, CreatedOn: createdt, ExpirateOn: expdt, Token: token}
							config.SessionCache.Delete(ctx, clienttoken)
							config.SessionCache.Put(ctx, token, user, time.Duration(config.SessionCacheTimeout)*time.Second)
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

	//hasedPassword, err := hashPassword(password)
	querystr := fmt.Sprintf(LoginQuery, username)

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
		var storedPassword string
		/*
			err = rows.Scan(&ID, &Name, &FamilyName)
			if err != nil {
				log.Error(fmt.Sprintf("Row Scan error:%s", err.Error()))
				ctx.JSON(http.StatusInternalServerError, "Login failed")
				return
			} */

		if len(jdata) == 0 {
			log.Error(fmt.Sprintf("User:%s not found", username))
			ctx.JSON(http.StatusNotFound, "Login failed")
			return
		}

		ID = int(jdata[0]["ID"].(int64))
		Name = jdata[0]["Name"].(string)
		FamilyName = jdata[0]["LastName"].(string)
		storedPassword = jdata[0]["Password"].(string)
		language := jdata[0]["LanguageCode"].(string)
		timezone := jdata[0]["TimeZoneCode"].(string)

		if jdata[0]["Password"] != nil && storedPassword != "" {
			err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
			if err != nil {
				log.Error(fmt.Sprintf("Password compare error:%s", err.Error()))
				ctx.JSON(http.StatusNotFound, "Login failed")
				return
			}
		}

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
		user := User{ID: ID, Username: username, Language: language, TimeZone: timezone, ClientID: ClientID, CreatedOn: createdt, ExpirateOn: expdt, Token: token}

		log.Debug(fmt.Sprintf("user:%v", user))

		config.SessionCache.Put(ctx, sessionid, user, time.Duration(config.SessionCacheTimeout)*time.Second)
		log.Debug(fmt.Sprintf("User:%s login successful!", user))

		ctx.JSON(http.StatusOK, user)
		return
	}

	iDBTx.Rollback()
	log.Error(fmt.Sprintf("Login failed for user:%s", username))

	ctx.JSON(http.StatusNotFound, "Login failed")

}

func getUserImage(username string, clientid string) (string, error) {
	log := logger.Log{ModuleName: logger.API, User: username, ClientID: clientid, ControllerName: "UserController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.getUserImage", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("getUserImage defer error: %s", err))
			//ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

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

func getUserMenus(userID int, isMobile bool, parentID int, username string, clientid string) ([]map[string]interface{}, error) {
	log := logger.Log{ModuleName: logger.API, User: username, ClientID: clientid, ControllerName: "UserController"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.getUserMenus", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("getUserMenus defer error: %s", err))
			//ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	log.Debug("get user menus execution function is called.")

	Mobile := 0
	Desktop := 1

	if isMobile {
		Mobile = 1
		Desktop = 0
	}

	querystr := fmt.Sprintf(GetUserMenusQuery, userID, Mobile, Desktop, parentID, userID)

	log.Debug(fmt.Sprintf("Query:%s", querystr))

	iDBTx, err := dbconn.DB.Begin()

	if err != nil {
		log.Error(fmt.Sprintf("Begin error:%s", err.Error()))
		return nil, err
	}

	defer iDBTx.Rollback()

	dboperation := dbconn.NewDBOperation(strconv.Itoa(userID), iDBTx, "User Menus")

	jdata, err := dboperation.Query_Json(querystr)

	if err != nil {
		log.Error(fmt.Sprintf("Query menu error:%s", err.Error()))
		return jdata, err
	}

	log.Debug(fmt.Sprintf("Query result:%v", jdata))

	return jdata, nil
}

func execChangePassword(ctx *gin.Context, username string, oldpassword string, newpassword string, clientid string) error {
	log := logger.Log{ModuleName: logger.API, User: username, ClientID: clientid, ControllerName: "UserController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.execChangePassword", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("execChangePassword defer error: %s", err))
			//ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	log.Debug("execChangePassword execution function is called.")

	result, jdata, err := validatePassword(username, oldpassword)

	if err != nil {
		log.Error(fmt.Sprintf("validatePassword error:%s", err.Error()))

		ctx.JSON(http.StatusInternalServerError, "Validate old password failed")
		return err
	}

	if result == false {
		log.Error(fmt.Sprintf("validatePassword error:%s", err.Error()))
		ctx.JSON(http.StatusInternalServerError, "Validate old password failed")
		return err
	}

	hashedPassword, err := hashPassword(newpassword)

	if jdata != nil {

		ID := int(jdata[0]["ID"].(int64))

		log.Debug(fmt.Sprintf("user ID:%d  new hashed password: %s ", ID, hashedPassword))

		Columns := []string{"Password", "PasswordLastChangeDate", "UpdatedOn", "UpdatedBy"}
		Values := []string{hashedPassword, time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"), username}
		datatypes := []int{0, 0, 0}
		Wherestr := fmt.Sprintf("ID= %d", ID)

		iDBTx, err := dbconn.DB.Begin()
		defer iDBTx.Rollback()

		if err != nil {
			log.Error(fmt.Sprintf("Begin error:%s", err.Error()))
			ctx.JSON(http.StatusInternalServerError, "Change password failed")
			return err
		}

		dboperation := dbconn.NewDBOperation(username, iDBTx, "User ChangePassword")

		index, errr := dboperation.TableUpdate(TableName, Columns, Values, datatypes, Wherestr)

		if errr != nil {
			log.Error(fmt.Sprintf("TableUpdate error:%s", errr.Error()))
			ctx.JSON(http.StatusInternalServerError, "Change password failed")
			return errr
		}

		if index == 0 {
			log.Error(fmt.Sprintf("TableUpdate error:%s", errr.Error()))
			ctx.JSON(http.StatusInternalServerError, "Change password failed")
			return errr
		}

		iDBTx.Commit()
		ctx.JSON(http.StatusOK, "OK")
		return nil
	}

	ctx.JSON(http.StatusInternalServerError, "Change password failed")
	return nil
}

func execLogout(ctx *gin.Context, token string) (string, error) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.execLogout", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("execLogout defer error: %s", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("Get user information Error: %v", err))

		return "", err
	}
	log.ClientID = clientid
	log.User = user
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

func validatePassword(username string, password string) (bool, []map[string]interface{}, error) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.validatePassword", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("validatePassword defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	log.Debug(fmt.Sprintf("validate password function is called. username: %s ", username))

	//	hashedPassword, err := hashPassword(password)

	jdata := []map[string]interface{}{}
	/*
		if err != nil {
			log.Error(fmt.Sprintf("hashPassword error:%s", err.Error()))
			return false, nil, err
		}
	*/
	querystr := fmt.Sprintf(LoginQuery, username)

	log.Debug(fmt.Sprintf("Query:%s", querystr))

	iDBTx, err := dbconn.DB.Begin()
	defer iDBTx.Rollback()

	if err != nil {
		log.Error(fmt.Sprintf("Begin error:%s", err.Error()))
		return false, nil, err
	}

	dboperation := dbconn.NewDBOperation(username, iDBTx, "User Login")

	jdata, err = dboperation.Query_Json(querystr)

	if err != nil {

		log.Error(fmt.Sprintf("Query error:%s", err.Error()))
		return false, nil, err

	}

	log.Debug(fmt.Sprintf("Query result:%v", jdata))

	if jdata != nil {

		iDBTx.Commit()

		if len(jdata) == 0 {
			return false, nil, err
		}

		storedPassword := jdata[0]["Password"].(string)

		if storedPassword == "" || jdata[0]["Password"] == nil {
			return true, jdata, nil
		}

		err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		if err != nil {
			log.Error(fmt.Sprintf("CompareHashAndPassword error:%s", err.Error()))
			return false, nil, err
		}

		return true, jdata, nil
	}
	return false, nil, nil
}
