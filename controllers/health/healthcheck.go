package health

import (
	//"log"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/health"
	"github.com/mdaxf/iac/logger"
)

type HealthController struct {
}

func (f *HealthController) CheckHealth(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "health"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.health.CheckHealth", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Health Check error: %s", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
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

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})

}
