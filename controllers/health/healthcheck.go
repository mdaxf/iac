package health

import (
	//"log"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

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
		iLog.Performance(fmt.Sprintf(" %s elapsed time: %v", "controllers.health.CheckHealth", elapsed))
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Health Check error: %s", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	data, err := health.CheckSystemHealth(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})

}
