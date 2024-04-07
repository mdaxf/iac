package bpmcontroller

import (
	"encoding/json"
	"fmt"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mdaxf/iac/controllers/common"

	"github.com/mdaxf/iac/engine/trancode"

	"github.com/mdaxf/iac/logger"
)

type BPMController struct {
	Name    string                 `json:"name"`
	Version string                 `json:"version"`
	Inputs  map[string]interface{} `json:"inputs"`
	Mode    int                    `json:"mode"`
}

func (b *BPMController) ExecuteBPM(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ExecuteBPM"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("BPMController.ExecuteBPM", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var bpm BPMController
	err = json.Unmarshal(body, &bpm)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("BPMController.ExecuteBPM: %v", bpm))

	if bpm.Name == "" {
		iLog.Error("BPM Name is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "BPM Name is required"})
		return
	}

	systemsessions := make(map[string]interface{})
	systemsessions["UserNo"] = user
	systemsessions["ClientID"] = clientid

	outputs, err := trancode.Execute(bpm.Name, bpm.Inputs, systemsessions)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error executing BPM: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("BPMController.ExecuteBPM: %v", outputs))
	ctx.JSON(http.StatusOK, gin.H{"data": outputs})

}
