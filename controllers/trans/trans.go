package trans

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/documents"
)

type TranCodeController struct {
}

func (e *TranCodeController) ExecuteTranCode(ctx *gin.Context) {
	/*	jsonString, err := json.Marshal(ctx.Request)
		if err != nil {
			fmt.Println("Error marshaling json:", err)
			return
		}
		log.Println(string(jsonString))  */
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}

	var tcdata TranCodeData
	if err := ctx.BindJSON(&tcdata); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//log.Print(tcdata.TranCode)
	iLog.Info(fmt.Sprintf("Start process transaction code %s's %s ", tcdata.TranCode, "Execute"))
	//tcode, err := e.getTransCode(tcdata.TranCode)
	filter := bson.M{"trancodename": tcdata.TranCode}

	tcode, err := documents.DocDBCon.QueryCollection("Transaction_Code", filter, nil)

	if err != nil {
		iLog.Error(fmt.Sprintf("Get transaction code %s's error", tcdata.TranCode))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := json.Marshal(tcode[0])
	if err != nil {

		iLog.Error(fmt.Sprintf("Error marshaling json:", err.Error()))
		return
	}

	iLog.Debug(fmt.Sprintf("transaction code %s's json: %s", tcdata.TranCode, string(jsonString)))
	/*trancode := types.TranCode{}
	err = json.Unmarshal(jsonString, &trancode)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
		return
	} */
	code, err := trancode.Configtoobj(string(jsonString))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
		return
	}

	tf := trancode.NewTranFlow(code, tcdata.inputs, map[string]interface{}{}, nil)
	outputs, err := tf.Execute()

	if err == nil {
		iLog.Debug(fmt.Sprintf("End process transaction code %s's %s ", tcdata.TranCode, "Execute"))
		ctx.JSON(http.StatusOK, gin.H{"Outputs": outputs})
		return
	} else {
		iLog.Error(fmt.Sprintf("End process transaction code %s's %s with error %s", tcdata.TranCode, "Execute", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"execution failed": err.Error()})
	}
}

func (e *TranCodeController) getTransCode(name string) (types.TranCode, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}
	iLog.Debug(fmt.Sprintf("Get transaction code /%s/%s%s", "trancodes", name, ".json"))

	data, err := ioutil.ReadFile(fmt.Sprintf("./%s/%s%s", "trancodes", name, ".json"))
	if err != nil {

		iLog.Error(fmt.Sprintf("failed to read configuration file: %v", err))
		return types.TranCode{}, fmt.Errorf("failed to read configuration file: %v", err)
	}

	//filter := bson.M{"trancodename": name}
	//iLog.Debug(fmt.Sprintf("Get transaction code detail data from respository with filter: %v", filter))
	//data, err := documents.DocDBCon.QueryCollection("Transaction_Code", filter, nil)

	iLog.Debug(fmt.Sprintf("Get transaction code data: %s%s%s", "trancodes", name, ".json", string(data)))
	return trancode.Bytetoobj(data)
}

func (e *TranCodeController) GetTranCodeListFromRespository(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetTranCodeListFromRespository"}
	iLog.Debug(fmt.Sprintf("Get transaction code list from respository"))

	projection := bson.M{
		"_id":            1,
		"trancodename":   1,
		"version":        1,
		"status":         1,
		"firstfuncgroup": 1,
		"system":         1,
		"description":    1,
	}
	tcitems, err := documents.DocDBCon.QueryCollection("Transaction_Code", nil, projection)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the transaction code list: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, tcitem := range tcitems {
		iLog.Debug(fmt.Sprintf("Get transaction code %s", tcitem["trancodename"]))
	}
	/*
		outputs := map[string]interface{}{
			"trancode": tcitems,
		} */

	ctx.JSON(http.StatusOK, gin.H{"Outputs": tcitems})
}

func (e *TranCodeController) GetTranCodeDetailFromRespository(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetTranCodeDetailFromRespository"}
	iLog.Debug(fmt.Sprintf("Get transaction code detail data from respository: %v", ctx.Params))

	var tcdata TranCodeData
	if err := ctx.BindJSON(&tcdata); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//log.Print(tcdata.TranCode)

	iLog.Debug(fmt.Sprintf("Get transaction code detail data from respository with code: %s", tcdata.TranCode))
	//filedvalue := primitive.ObjectID(param.ID)
	filter := bson.M{"trancodename": tcdata.TranCode}
	iLog.Debug(fmt.Sprintf("Get transaction code detail data from respository with filter: %v", filter))
	tcitems, err := documents.DocDBCon.QueryCollection("Transaction_Code", filter, nil)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the transaction code list: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, tcitem := range tcitems {
		iLog.Debug(fmt.Sprintf("Get transaction code %s", tcitem["trancodename"]))
	}

	ctx.JSON(http.StatusOK, gin.H{"Outputs": tcitems})
}

func (e *TranCodeController) UpdateTranCodeToRespository(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCodeController"}
	iLog.Debug(fmt.Sprintf("Update transaction code to respository!"))

	var tcdata TranCodeData
	if err := ctx.BindJSON(&tcdata); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := tcdata.TranCode
	uuid := tcdata.UUID
	idata := tcdata.Data

	iLog.Debug(fmt.Sprintf("Update transaction code to respository with code: %s", name))
	iLog.Debug(fmt.Sprintf("Update transaction code to respository with uuid: %s", uuid))
	iLog.Debug(fmt.Sprintf("Update transaction code to respository with data: %s", logger.ConvertJson(idata)))

	jsonData, err := json.Marshal(idata)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to Marshal data: %v", err))
	}

	var tData TranCode

	err = json.Unmarshal(jsonData, &tData)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to unmarshal data: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if name == "" && tData.Name != "" {
		name = tData.Name
	}

	if uuid == "" && tData.UUID != "" {
		uuid = tData.UUID
	}

	id := ""
	ok := false
	if id, ok = idata["_id"].(string); ok {
		iLog.Debug(fmt.Sprintf("Update transaction code to respository with _id: %s", id))

	}

	isdefault := tData.IsDefault

	iLog.Debug(fmt.Sprintf("Update transaction code to respository with _id: %s", id))

	if isdefault {
		iLog.Debug(fmt.Sprintf("Update transaction code to in respository to set default to false: %s", name))
		filter := bson.M{"isdefault": true,
			"trancodename": name}
		update := bson.M{"$set": bson.M{"isdefault": false, "system.updatedon": time.Now()}, "system.updatedby": "system"}

		iLog.Debug(fmt.Sprintf("Update transaction code to in respository with filter: %v", filter))
		iLog.Debug(fmt.Sprintf("Update transaction code to in respository with update: %v", update))
		err := documents.DocDBCon.UpdateCollection("Transaction_Code", filter, update, nil)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to update transaction code: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if id == "" && idata != nil {
		utcTime := time.Now().UTC()
		idata["system.updatedon"] = utcTime
		idata["system.updatedby"] = "system"
		idata["system.createdon"] = utcTime
		idata["system.createdby"] = "system"

		iLog.Debug(fmt.Sprintf("Insert transaction code to respository with code: %s", name))
		err := documents.DocDBCon.InsertCollection("Transaction_Code", idata)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to insert transaction code: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else if idata != nil {

		parsedObjectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to parse object id: %v", err))
		}

		iLog.Debug(fmt.Sprintf("Update transaction code to respository with code: %s", name))
		//filedvalue := primitive.ObjectID(param.ID)
		filter := bson.M{"_id": parsedObjectID}
		iLog.Debug(fmt.Sprintf("Update transaction code to respository with filter: %v", filter))

		idata["system.updatedon"] = time.Now().UTC()
		idata["system.updatedby"] = "system"
		delete(idata, "_id")

		iLog.Debug(fmt.Sprintf("Update transaction code to respository with data: %s", logger.ConvertJson(idata)))

		err = documents.DocDBCon.UpdateCollection("Transaction_Code", filter, nil, idata)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to update transaction code: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"Outputs": "Success"})

}

type TranCodeData struct {
	TranCode string                 `json:"code"`
	UUID     string                 `json:"uuid"`
	Data     map[string]interface{} `json:"data"`
	inputs   map[string]interface{} `json:"Inputs"`
}

type TranCode struct {
	ID             string           "json:'_id'"
	UUID           string           "json:'uuid'"
	Name           string           "json:'trancodename'"
	Version        string           "json:'version'"
	IsDefault      bool             "json:'isdefault'"
	Status         int              "json:'status'"
	Firstfuncgroup string           "json:'firstfuncgroup'"
	SystemData     types.SystemData "json:'system'"
	Description    string           "json:'description'"
}
