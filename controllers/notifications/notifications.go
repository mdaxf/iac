package notifications

import (
	"fmt"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/controllers/common"
	notif "github.com/mdaxf/iac/notification"
)

type Notification struct {
}

func (n *Notification) CreateNotification(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.CreateNotification", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

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

func (n *Notification) GetNotificationsbyUser(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.GetNotificationsbyUser", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

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

func (n *Notification) ResponseNotification(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notifications.ResponseNotification", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

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
	comments := ndata["comments"].(string)

	err = notif.UpdateNotification(ndata, user, comments)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to update the notification: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Update notification to respository with data: %s", logger.ConvertJson(ndata)))

	ctx.JSON(http.StatusOK, gin.H{"data": ndata})

}
