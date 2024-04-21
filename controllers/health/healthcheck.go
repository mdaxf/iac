package health

import (
	//"log"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/health"
	"github.com/mdaxf/iac/logger"
)

type HealthController struct {
}

// CheckHealth is a function that handles the health check request.
// It retrieves user information from the request context, logs the health check activity,
// and calls the CheckSystemHealth function to get the system health data.
// If there is an error retrieving user information or checking the system health,
// it returns an error response with the corresponding status code.
// Otherwise, it returns a success response with the system health data.

func (f *HealthController) CheckHealth(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "health"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.health.CheckHealth", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("Health Check error: %s", err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug("Health Check")
	data, err := health.CheckSystemHealth(c)
	iLog.Debug(fmt.Sprintf("Health Check Result: %v", data))

	iLog.Debug(fmt.Sprintf("all node health data: %v", com.NodeHeartBeats))

	nodehealth := make(map[string]interface{})
	nodehealth["Result"] = data
	nodehealth["Node"] = com.IACNode
	nodehealth["timestamp"] = time.Now().UTC()
	//nodehealth["ServiceStatus"] = (com.NodeHeartBeats[com.IACNode["AppID"].(string)].(map[string]interface{}))["ServiceStatus"]
	com.NodeHeartBeats[com.IACNode["AppID"].(string)] = nodehealth

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": com.NodeHeartBeats})

}
