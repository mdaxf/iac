package notifications

import (
	"fmt"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/controllers/common"
	notif "github.com/mdaxf/iac/notifications"
)

type NotificationController struct {
}

// CreateNotification handles the creation of a new notification.
// It retrieves the request body and user information from the context,
// creates a new notification using the provided data, and returns the created notification.
// If any error occurs during the process, it returns an error response.

func (n *NotificationController) CreateNotification(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.CreateNotification", elapsed)
	}()

	/*	defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error: %v", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	requestobj, clientid, user, err := common.GetRequestBodyandUserbyJson(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get request information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user
	ndata := make(map[string]interface{})
	ndata = requestobj["data"].(map[string]interface{})
	err = notif.CreateNewNotification(ndata, user)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to create the notification: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Create notification to respository with data: %s", logger.ConvertJson(ndata)))

	ctx.JSON(http.StatusOK, gin.H{"data": ndata})
}

// GetNotificationsbyUser retrieves the notifications for a specific user.
// It takes a gin.Context as input and returns the notifications as JSON data.
// The function logs performance metrics and any errors encountered during execution.
func (n *NotificationController) GetNotificationsbyUser(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.GetNotificationsbyUser", elapsed)
	}()
	/*
		defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error: %v", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	items, err := notif.GetNotificationsbyUser(user)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the list from notification: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get notification list from respository with data: %s", logger.ConvertJson(items)))

	ctx.JSON(http.StatusOK, gin.H{"data": items})
}

// ResponseNotification handles the HTTP request for updating a notification.
// It retrieves the request body and user information from the context, updates the notification,
// and returns the updated notification data in the response.

func (n *NotificationController) ResponseNotification(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notifications.ResponseNotification", elapsed)
	}()
	/*
		defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error: %v", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	requestobj, clientid, user, err := common.GetRequestBodyandUserbyJson(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get request information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user
	ndata := make(map[string]interface{})
	ndata = requestobj["data"].(map[string]interface{})
	comments := requestobj["comments"].(string)
	status := int(requestobj["status"].(float64))
	err = notif.UpdateNotification(ndata, user, comments, status)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to update the notification: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Update notification to respository with data: %s", logger.ConvertJson(ndata)))

	ctx.JSON(http.StatusOK, gin.H{"data": ndata})

}
