package healthcheck

import (
	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mdaxf/iac/health"
)

type HealthController struct {
}

func (f *HealthController) CheckHealth(c *gin.Context) {

	data, err := health.CheckSystemHealth(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})

}
