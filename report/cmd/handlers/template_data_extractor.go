package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
)

func extractData(ctx *gin.Context) (interface{}, error) {
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.handlers.extractData"}

	encodedData := ctx.Query("d")
	if encodedData == "" {
		return map[string]interface{}{}, nil
	}

	strData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to decode base64 data with error: %v", err))
		return nil, err
	}

	var jsonData interface{}
	if err := json.Unmarshal(strData, &jsonData); err != nil {
		iLog.Error(fmt.Sprintf("failed to unmarshal data to json with error: %v", err))
		return nil, err
	}

	return jsonData, nil
}
