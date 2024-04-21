package component

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
)

type IACComponentController struct {
}

type HeartBeat struct {
	Node          map[string]interface{} `json:Node"`
	Result        map[string]interface{} `json:Result"`
	ServiceStatus map[string]interface{} `json:ServiceStatus"`
	time          time.Time              `json:time"`
}

func (f *IACComponentController) ComponentHeartbeat(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Component.heartbeat"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.component.heartbeat", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var data HeartBeat
	err = json.Unmarshal(body, &data)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// update the component status dataset
	//iLog.Debug(fmt.Sprintf("Component.heartbeat: %v", data))

	com.NodeHeartBeats[data.Node["AppID"].(string)] = data

	removeNotResponseComponentNodeHeartBeats(iLog)

	c.JSON(http.StatusOK, gin.H{"Status": "Success"})
}

func (f *IACComponentController) ComponentClose(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Component.close"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.component.close", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var data HeartBeat
	err = json.Unmarshal(body, &data)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// update the component status dataset
	iLog.Debug("Component close")

	if com.NodeHeartBeats[data.Node["AppID"].(string)] != nil {
		data.Node["CloseTime"] = data.time
		data.Node["Status"] = "Closed"
	}

	c.JSON(http.StatusOK, gin.H{"data": "Component closed"})

}

func removeNotResponseComponentNodeHeartBeats(iLog logger.Log) {
	for key, value := range com.NodeHeartBeats {
		node := value.(map[string]interface{})
		if node == nil {
			iLog.Error(fmt.Sprintf("NodeHeartBeats[%s] is nil", key))
			delete(com.NodeHeartBeats, key)
			continue
		}
		nodeTimestamp, ok := node["timestamp"]

		if !ok {
			iLog.Error(fmt.Sprintf("NodeHeartBeats[%s][timestamp] is nil", key))
			delete(com.NodeHeartBeats, key)
			continue
		}

		lasteHeartBeatTime, ok := nodeTimestamp.(time.Time)

		if !ok {
			iLog.Error(fmt.Sprintf("NodeHeartBeats[%s][timestamp] is not time.Time", key))
			delete(com.NodeHeartBeats, key)
			continue
		}

		if lasteHeartBeatTime.IsZero() {
			iLog.Error(fmt.Sprintf("NodeHeartBeats[%s][timestamp] is zero", key))
			delete(com.NodeHeartBeats, key)
			continue
		}

		if lasteHeartBeatTime.Add(time.Minute * 30).Before(time.Now().UTC()) {
			iLog.Debug(fmt.Sprintf("NodeHeartBeats[%s][timestamp] is not response", key))
			delete(com.NodeHeartBeats, key)
		}

		if node["Status"] == "Closed" {
			iLog.Debug(fmt.Sprintf("NodeHeartBeats[%s][Status] is Closed", key))
			delete(com.NodeHeartBeats, key)
		}
	}
}
