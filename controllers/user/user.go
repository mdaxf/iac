package user

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	dbconn "github.com/mdaxf/iac/databases"
)

type UserController struct{}

func (c *UserController) Login(ctx *gin.Context) {
	// Retrieve a list of users from the database
	var user User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := user.Username
	password := user.Password
	log.Println(fmt.Sprintf("Login:%s  %s", username, password))

	//log.Println(fmt.Sprintf("Database open connection:%d", &dbconn.DB.Stats().OpenConnections))

	rows, err := dbconn.DB.Query("SELECT ID,Name,FamilyName FROM EMPLOYEE")
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

		user := User{ID: string(rune(ID)), Username: Name + " " + FamilyName, Email: "", Password: password, SessionID: "ashadasdasdghashgd"}

		ctx.JSON(http.StatusOK, user)
		return
	}

	ctx.JSON(http.StatusNotFound, "Login failed")

}

func (c *UserController) Logout(ctx *gin.Context) {
	// Retrieve a list of users from the database

	// Send the list of users in the response
	ctx.JSON(http.StatusOK, "Loginsessionid")
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

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	SessionID string `json:"sessionid"`
}
