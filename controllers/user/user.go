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
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
)

type UserController struct{}

func (c *UserController) Login(ctx *gin.Context) {
	// Retrieve a list of users from the database
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.login", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("login defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	log.Debug("Login handle function is called.")

	var user LoginUserData
	if err := ctx.BindJSON(&user); err != nil {
		log.Error(fmt.Sprintf("Login error:%s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := user.Username
	password := user.Password
	token := user.Token
	ClientID := user.ClientID
	Renew := user.Renew

	log.Debug(fmt.Sprintf("Login:%s  %s  token: %s  renew:%s", username, password, token, Renew))

	execLogin(ctx, username, password, token, ClientID, Renew)

	/*
		//log.Println(fmt.Sprintf("Database open connection:%d", &dbconn.DB.Stats().OpenConnections))
		querystr := fmt.Sprintf("SELECT ID,Name,FamilyName FROM EMPLOYEE WHERE LoginName='%s'", username)
		log.Println(fmt.Sprintf("query:%s", querystr))
		rows, err := dbconn.DB.Query(querystr)
		if err != nil {
			panic(err.Error())
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
			//panic(err.Error())
		}
		defer rows.Close()

		log.Println(fmt.Printf("rows: %v\n", rows))

		for rows.Next() {
			var ID int
			var Name string
			var FamilyName string

			err = rows.Scan(&ID, &Name, &FamilyName)
			if err != nil {
				panic(err.Error())
			}
			log.Println(fmt.Sprintf("ID:%d  Name:%s  FamilyName:%s", ID, Name, FamilyName))

			user := User{ID: ID, Username: Name + " " + FamilyName, Email: "", Password: password, SessionID: uuid.New()}

			exist, err := config.SessionCache.IsExist(ctx, "USER_"+string(rune(ID)))

			if err != nil && exist {
				config.SessionCache.Delete(ctx, "USER_"+string(rune(ID)))

			}
			config.SessionCache.Put(ctx, "USER_"+string(rune(ID)), user, 10*time.Minute)

			ctx.JSON(http.StatusOK, user)
			return
		}

		ctx.JSON(http.StatusNotFound, "Login failed")
	*/

}

func (c *UserController) Image(ctx *gin.Context) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.Image", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("Image defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	log.User = userno
	log.ClientID = clientid

	log.Debug("Get User Image handle function is called.")

	log.Debug(fmt.Sprintf("Get User Image:%s", ctx.Param("username")))

	username := ctx.Query("username")
	/*var user LoginUserData
	if err := ctx.BindJSON(&user); err != nil {
		log.Error(fmt.Sprintf("Get User Image error:%s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := user.Username */

	log.Debug(fmt.Sprintf("Get User Image:%s", username))

	PictureUrl, err := getUserImage(username, clientid)

	if err != nil {
		log.Error(fmt.Sprintf("Get User Image error:%s", err.Error()))
		PictureUrl = "images/avatardefault.png"
	}

	if PictureUrl == "" {
		PictureUrl = "images/avatardefault.png"
	}

	log.Debug(fmt.Sprintf("Get User Image:%s", PictureUrl))
	ctx.JSON(http.StatusOK, PictureUrl)
}

func (c *UserController) Logout(ctx *gin.Context) {
	// Retrieve a list of users from the database
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.Logout", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("Logout defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	// Send the list of users in the response
	var user LoginUserData
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//userID := user.ID
	token := user.Token
	execLogout(ctx, token)
	ctx.JSON(http.StatusOK, "Logoutsessionid")
}

func (c *UserController) ChangePassword(ctx *gin.Context) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.ChangePassword", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("ChangePassword defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	log.User = userno
	log.ClientID = clientid

	log.Debug("Change Password handle function is called.")

	var user ChangePwdData
	if err := ctx.BindJSON(&user); err != nil {
		log.Error(fmt.Sprintf("Change Password handle function  error:%s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := user.Username
	oldpassword := user.OldPassword
	newpassword := user.NewPassword

	log.Debug(fmt.Sprintf("Change Password:%s  %s  %s", username, oldpassword, newpassword))

	execChangePassword(ctx, username, oldpassword, newpassword, clientid)
}

func (c *UserController) UserMenus(ctx *gin.Context) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("controllers.user.UserMenus", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("UserMenus defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	log.User = userno
	log.ClientID = clientid
	log.Debug("Get User menus handle function is called.")

	log.Debug(fmt.Sprintf("Get User menus:%s", ctx.Param("username")))

	userID := ctx.Query("userid")
	Mobile := ctx.Query("mobile")
	parentID := ctx.Query("parentid")

	log.Debug(fmt.Sprintf("Get User menus:%s, %s", userID, Mobile))

	isMobile := false
	if Mobile == "1" {
		isMobile = true
	}
	num, err := strconv.Atoi(userID)
	if err != nil {
		log.Error(fmt.Sprintf("Get User menus error:%s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	num1, err := strconv.Atoi(parentID)
	if err != nil {
		log.Error(fmt.Sprintf("Get User menus error:%s", err.Error()))
		num1 = -1
	}

	jdata, err := getUserMenus(num, isMobile, num1, userno, clientid)

	if err != nil {
		log.Error(fmt.Sprintf("Get User menus error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Debug(fmt.Sprintf("Get User menus:%s", jdata))
	ctx.JSON(http.StatusOK, jdata)
}

func (c *UserController) List(ctx *gin.Context) {
	// Retrieve a list of users from the database
	users := []User{ /* ... */ }

	// Send the list of users in the response
	ctx.JSON(http.StatusOK, users)
}

func (c *UserController) Create(ctx *gin.Context) {
	// Retrieve user data from the request body
	var user User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	// Save the user data to the database
	if err := SaveUser(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user data"})
		return
	}

	// Send the saved user data in the response
	ctx.JSON(http.StatusOK, user)
}

func SaveUser(user *User) error {
	// Save the user data to the database
	return nil
}
