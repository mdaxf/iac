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

func (n *Notification) GetNotificationsbyUser(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetListofCollectionData"}

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
