package auth

import (
	"fmt"
	"net/http"
	"strings"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"

	jwt "github.com/dgrijalva/jwt-go"
)

var jwtsecretKey = "IACFramework"

func Generate_authentication_token(userID string, loginName string, ClientID string) (string, string, string, error) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Authorization"}
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			log.PerformanceWithDuration("auth.Generate_authentication_token", elapsed)
		}()
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("Error in auth.Generate_authentication_token: %s", r))
				return
			}
		}() */

	log.Debug("Authorization function is called.")

	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = userID
	atClaims["login_name"] = loginName
	atClaims["client_id"] = ClientID
	fmt.Println(config.SessionCacheTimeout, time.Duration(config.SessionCacheTimeout)*time.Second)
	createdt := time.Now().Format("2006-01-02 15:04:05")
	expiredt := time.Now().Add(time.Duration(config.SessionCacheTimeout) * time.Second)
	expires := expiredt.Unix()
	atClaims["exp"] = expires

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(jwtsecretKey))
	if err != nil {
		log.Error(fmt.Sprintf("Authorization Error:%s", err.Error()))
		return "", "", "", err
	}
	log.Debug(fmt.Sprintf("Authorization Token:%s", token))
	return token, createdt, string(expiredt.Format("2006-01-02 15:04:05")), nil
}
func GetUserInformation(c *gin.Context) (string, string, string, error) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Authorization"}
	//	log.Debug(fmt.Sprintf("Authorization validation function is called for tocken: %s ", tokenString))
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			log.PerformanceWithDuration("auth.GetUserInformation", elapsed)
		}()
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("Error in auth.GetUserInformation: %s", r))
				return
			}
		}()  */
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Error(fmt.Sprintf("Missing Authorization header"))
		return "", "", "", nil
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
		log.Error(fmt.Sprintf("Invalid token format: %s", authHeader))
		return "", "", "", nil
	}
	tokenString := bearerToken[1]

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Provide the secret key used during token generation
		secretKey := []byte(jwtsecretKey) // Replace with your actual secret key
		return secretKey, nil
	})
	if err != nil {

		log.Error(fmt.Sprintf("Failed to parse token:%s", err.Error()))
		return "", "", "", err
	}
	if token.Valid {
		// Check if the token has expired
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {

			log.Error(fmt.Sprintf("Invalid token claims: %s", tokenString))
			return "", "", "", err
		}
		UserID := claims["user_id"].(string)
		LoginName := claims["login_name"].(string)
		ClientID := claims["client_id"].(string)
		return UserID, LoginName, ClientID, nil
	}
	return "", "", "", err
}
func ValidateToken(tokenString string) (bool, error) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Authorization"}
	//	log.Debug(fmt.Sprintf("Authorization validation function is called for tocken: %s ", tokenString))
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			log.PerformanceWithDuration("auth.ValidateToken", elapsed)
		}()
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("Error in auth.ValidateToken: %s", r))
				return
			}
		}()  */
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Provide the secret key used during token generation
		secretKey := []byte(jwtsecretKey) // Replace with your actual secret key
		return secretKey, nil
	})
	if err != nil {

		log.Error(fmt.Sprintf("Failed to parse token:%s", err.Error()))
		return false, err
	}

	// Check if the token is valid
	if token.Valid {
		// Check if the token has expired
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {

			log.Error(fmt.Sprintf("Invalid token claims: %s", tokenString))
			return false, err
		}

		expirationTime := claims["exp"].(float64)
		expiration := time.Unix(int64(expirationTime), 0)

		if expiration.Before(time.Now()) {

			log.Error(fmt.Sprintf("Token has expired: %s", tokenString))
			return false, err
		} else {
			log.Debug(fmt.Sprintf("Token is valid: %s", tokenString))

		}
	} else {
		log.Error(fmt.Sprintf("Token is invalid: %s", tokenString))

		return false, err
	}
	return true, nil
}

func Extendexptime(tokenString string) (string, string, string, error) {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Authorization"}

	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			log.PerformanceWithDuration("auth.Extendexptime", elapsed)
		}()
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("Error in auth.Extendexptime: %s", r))
				return
			}
		}()
	*/
	log.Debug(fmt.Sprintf("Extend the token function is called for tocken: %s ", tokenString))

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Provide the secret key used during token generation
		secretKey := []byte(jwtsecretKey) // Replace with your actual secret key
		return secretKey, nil
	})

	if err != nil {

		log.Error(fmt.Sprintf("Failed to parse token:%s", err.Error()))
		return "", "", "", err
	}

	// Check if the token is valid and has not expired
	if !token.Valid {
		log.Error(fmt.Sprintf("Token is invalid: %s", tokenString))
		return "", "", "", err
	}

	// Get the current time
	now := time.Now()

	// Calculate the new expiration time (e.g., extend it by 1 hour from the current time)
	expirationTime := now.Add(time.Duration(config.SessionCacheTimeout) * time.Second)

	// Update the "exp" (expiration) claim in the token with the new expiration time
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = expirationTime.Unix()

	// Create a new token with the updated claims
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the new token with the secret key to get the final token string
	newTokenString, err := newToken.SignedString([]byte(jwtsecretKey))
	if err != nil {
		log.Error(fmt.Sprintf("Create new token Error:%s", err.Error()))
		return "", "", "", err
	}
	log.Debug(fmt.Sprintf("New Token:%s, createdon %s, exp: %s", newTokenString, now.Format("2006-01-02 15:04:05"), expirationTime.Format("2006-01-02 15:04:05")))
	return newTokenString, now.Format("2006-01-02 15:04:05"), expirationTime.Format("2006-01-02 15:04:05"), nil
}

func AuthMiddleware() gin.HandlerFunc {
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Authorization"}
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			log.PerformanceWithDuration("auth.AuthMiddleware", elapsed)
		}()
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("Error in auth.AuthMiddleware: %s", r))
				return
			}
		}()
	*/
	log.Debug(fmt.Sprintf("Authorization for the API call"))

	return func(c *gin.Context) {
		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		//	log.Debug(fmt.Sprintf("Authorization Header:%s %s", authHeader, c.Request.URL.Path))

		if c.Request.URL.Path == "/favicon.ico" || c.Request.URL.Path == "/user/login" || c.Request.URL.Path == "/user/changepwd" || strings.Contains(c.Request.URL.Path, "/user/image") || strings.Contains(c.Request.URL.Path, "/portal") {
			//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		} else if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		//	log.Debug(fmt.Sprintf("Authorization Header:%s", bearerToken))
		if len(bearerToken) != 2 || (strings.ToLower(bearerToken[0]) != "bearer" && strings.ToLower(bearerToken[0]) != "apikey") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}
		authtype := strings.ToLower(bearerToken[0])
		tokenString := bearerToken[1]

		if authtype == "apikey" {
			if tokenString != config.ApiKey {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid APIKey"})
				return
			}
		} else {
			ok, err := ValidateToken(tokenString)

			if err != nil || !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				return
			}
		}
		/*
			// Parse the JWT token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate the signing method and return the secret key
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				return
			}

			// If the token is valid, set the token claims in the Gin context for further use
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				c.Set("userID", claims["userID"])
			} */

		c.Next()
	}
}

func protectedHandler(c *gin.Context) {
	// This is the protected REST API endpoint.
	// You can access the userID from the context.
	log := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Authorization"}
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			log.PerformanceWithDuration("auth.protectedHandler", elapsed)
		}()
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("Error in auth.protectedHandler: %s", r))
				return
			}
		}() */
	log.Debug(fmt.Sprintf("Authorization the user"))

	userID, _ := c.Get("userID")

	// You can implement your logic here to process the request for the authenticated user.
	c.JSON(http.StatusOK, gin.H{"message": "Hello, user " + userID.(string)})

}
