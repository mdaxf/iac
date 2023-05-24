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
	//"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
)

type UserController struct{}

func (c *UserController) Login(ctx *gin.Context) {
	// Retrieve a list of users from the database
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UserController"}
	log.Debug("Login handle function is called.")

	var user LoginUserData
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := user.Username
	password := user.Password

	log.Debug(fmt.Sprintf("Login:%s  %s", username, password))

	execLogin(ctx, username, password)

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

func (c *UserController) Logout(ctx *gin.Context) {
	// Retrieve a list of users from the database

	// Send the list of users in the response
	var user LoginUserData
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := user.ID
	execLogout(ctx, string(userID))
	ctx.JSON(http.StatusOK, "Logoutsessionid")
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
