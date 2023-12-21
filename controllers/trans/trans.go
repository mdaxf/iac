package trans

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/documents"

	"github.com/mdaxf/iac/controllers/common"
)

type TranCodeController struct {
}

// ExecuteTranCode executes the transaction code based on the request context.
// It retrieves the user information from the request, fetches the transaction code data,
// and executes the transaction flow. The outputs of the transaction are returned in the response.
func (e *TranCodeController) ExecuteTranCode(ctx *gin.Context) {
	/*	jsonString, err := json.Marshal(ctx.Request)
		if err != nil {
			fmt.Println("Error marshaling json:", err)
			return
		}
		log.Println(string(jsonString))  */
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.ExecuteTranCode", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("ExecuteTranCode defer error: %s", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	iLog.User = userno
	iLog.ClientID = clientid

	//var tcdata TranCodeData
	tcdata, err := getDataFromRequest(ctx)

	iLog.Info(fmt.Sprintf("Start process transaction code %s's %s: %s", tcdata.TranCode, "Execute", tcdata.Inputs))

	//tcode, err := e.getTransCode(tcdata.TranCode)
	filter := bson.M{"trancodename": tcdata.TranCode, "isdefault": true}

	tcode, err := documents.DocDBCon.QueryCollection("Transaction_Code", filter, nil)

	if err != nil {
		iLog.Error(fmt.Sprintf("Get transaction code %s's error", tcdata.TranCode))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("transaction code %s's data: %s", tcdata.TranCode, tcode))
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

	/*var inputs_json map[string]interface{}
	err = json.Unmarshal([]byte(tcdata.inputs), &inputs_json)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
		inputs_json = nil
	}
	iLog.Debug(fmt.Sprintf("transaction code %s's json inputs: %s", tcdata.TranCode, inputs_json))
	*/
	systemsessions := make(map[string]interface{})
	systemsessions["UserNo"] = userno
	systemsessions["ClientID"] = clientid
	tf := trancode.NewTranFlow(code, tcdata.Inputs, systemsessions, nil, nil, nil)
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

// UnitTest is a handler function for performing unit tests on transaction codes.
// It retrieves the user information from the request context, gets the transaction code data,
// and executes the unit test using the transaction code and system sessions.
// The outputs of the unit test are returned in the response.
func (e *TranCodeController) UnitTest(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.UnitTest", elapsed)
	}()

	/*	defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("UnitTest defer error: %s", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	iLog.User = userno
	iLog.ClientID = clientid
	systemsessions := make(map[string]interface{})
	systemsessions["UserNo"] = userno
	systemsessions["ClientID"] = clientid

	//var tcdata TranCodeData
	tcdata, err := getDataFromRequest(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get transaction code %s's error", tcdata.TranCode))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Start process transaction code %s's %s: %s", tcdata.TranCode, "Unit Test", tcdata.Inputs))

	outputs, err := trancode.ExecuteUnitTest(tcdata.TranCode, systemsessions)

	if err == nil {
		iLog.Debug(fmt.Sprintf("End process transaction code %s's %s ", tcdata.TranCode, "Unit Test"))
		ctx.JSON(http.StatusOK, gin.H{"Outputs": outputs})
		return
	} else {
		iLog.Error(fmt.Sprintf("End process transaction code %s's %s with error %s", tcdata.TranCode, "Unit Test", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"execution failed": err.Error()})
	}

}

// TestbyTestData is a function that handles the testing of transaction codes using test data.
// It receives a gin.Context object as a parameter.
// The function retrieves the user information from the request context and logs it.
// It then retrieves the transaction code data from the request and performs unit testing using the provided test data.
// The function returns the outputs of the unit testing or an error if the execution fails.

func (e *TranCodeController) TestbyTestData(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.TestbyTestData", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("TestbyTestData defer error: %s", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	iLog.User = userno
	iLog.ClientID = clientid
	//var
	//var tcdata TranCodeData
	tcdata, err := getDataFromRequest(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get transaction code %s's error", tcdata.TranCode))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	systemsessions := make(map[string]interface{})
	systemsessions["UserNo"] = userno
	systemsessions["ClientID"] = clientid
	iLog.Info(fmt.Sprintf("Start process transaction code %s's %s: %s", tcdata.TranCode, "Unit Test", tcdata.Inputs))

	outputs, err := trancode.ExecuteUnitTestWithTestData(tcdata.TranCode, tcdata.Inputs, systemsessions)

	if err == nil {
		iLog.Debug(fmt.Sprintf("End process transaction code %s's %s ", tcdata.TranCode, "Unit Test"))
		ctx.JSON(http.StatusOK, gin.H{"Outputs": outputs})
		return
	} else {
		iLog.Error(fmt.Sprintf("End process transaction code %s's %s with error %s", tcdata.TranCode, "Unit Test", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"execution failed": err.Error()})
	}
}

// Execute executes the transaction code with the given inputs, user, and client ID.
// It returns the outputs of the transaction code and any error that occurred during execution.

func (e *TranCodeController) Execute(Code string, externalinputs map[string]interface{}, user string, clientid string) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: user, ClientID: clientid, ControllerName: "TranCode"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.Execute", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("Execute defer error: %s", err))
				//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	iLog.Info(fmt.Sprintf("Start process transaction code %s with inputs: %s ", Code, externalinputs))

	iLog.Info(fmt.Sprintf("Start process transaction code %s's %s ", Code, "Execute"))
	/*
		filter := bson.M{"trancodename": Code, "isdefault": true}

		tcode, err := documents.DocDBCon.QueryCollection("Transaction_Code", filter, nil)

		if err != nil {
			iLog.Error(fmt.Sprintf("Get transaction code %s's error", Code))

			return nil, err
		}
		iLog.Debug(fmt.Sprintf("transaction code %s's data: %s", Code, tcode))
		jsonString, err := json.Marshal(tcode[0])
		if err != nil {

			iLog.Error(fmt.Sprintf("Error marshaling json:", err.Error()))
			return nil, err
		}

		trancodeobj, err := trancode.Configtoobj(string(jsonString))
		if err != nil {
			iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
			return nil, err
		}

		iLog.Debug(fmt.Sprintf("transaction code %s's json: %s", trancodeobj, string(jsonString)))

		if err != nil {
			iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
			return nil, err
		}
	*/
	trancodeobj, err := trancode.GetTranCodeDatabyCode(Code)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
		return nil, err
	}
	iLog.Debug(fmt.Sprintf("transaction code %s's json: %s", Code, trancodeobj))
	systemsessions := make(map[string]interface{})
	systemsessions["UserNo"] = user
	systemsessions["ClientID"] = clientid
	tf := trancode.NewTranFlow(trancodeobj, externalinputs, systemsessions, nil, nil, nil)
	outputs, err := tf.Execute()

	if err == nil {
		iLog.Debug(fmt.Sprintf("End process transaction code %s's %s ", Code, "Execute"))
		return outputs, nil

	} else {
		iLog.Error(fmt.Sprintf("End process transaction code %s's %s with error %s", Code, "Execute", err.Error()))
		return nil, err
	}
}

// getTransCode retrieves the transaction code for a given name, user, and client ID.
// It reads the transaction code data from a JSON file and converts it into a TranCode object.
// The function logs debug, error, and performance information using the logger package.
// Parameters:
// - name: the name of the transaction code
// - user: the user associated with the transaction code
// - clientid: the client ID associated with the transaction code
// Returns:
// - types.TranCode: the retrieved transaction code
// - error: any error that occurred during the retrieval process

func (e *TranCodeController) getTransCode(name string, user string, clientid string) (types.TranCode, error) {
	iLog := logger.Log{ModuleName: logger.API, User: user, ClientID: clientid, ControllerName: "TranCode"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.getTransCode", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("getTransCode defer error: %s", err))
				//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
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

// GetTranCodeListFromRespository retrieves the transaction code list from the repository.
// It requires a valid gin.Context as input.
// It returns the transaction code list as JSON data in the gin.Context response.

func (e *TranCodeController) GetTranCodeListFromRespository(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetTranCodeListFromRespository"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.GetTranCodeListFromRespository", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("GetTranCodeListFromRespository defer error: %s", err))
				//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	iLog.User = userno
	iLog.ClientID = clientid
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

	ctx.JSON(http.StatusOK, gin.H{"data": tcitems})
}

// GetTranCodeDetailFromRespository retrieves transaction code details from the repository.
// It expects a JSON payload containing the transaction code data.
// The function first extracts the user information from the request context.
// Then, it binds the JSON payload to the TranCodeData struct.
// Next, it constructs a filter based on the TranCodeData.TranCode field.
// Finally, it queries the "Transaction_Code" collection in the database using the filter and returns the results as JSON.
// If any error occurs during the process, it returns an error response.

func (e *TranCodeController) GetTranCodeDetailFromRespository(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetTranCodeDetailFromRespository"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.GetTranCodeDetailFromRespository", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("GetTranCodeDetailFromRespository defer error: %s", err))
				//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	iLog.User = userno
	iLog.ClientID = clientid
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

// UpdateTranCodeToRespository updates the transaction code in the repository.
// It receives a gin.Context object as a parameter and returns no values.
// The function first logs the start time and defer logs the elapsed time.
// It then retrieves the user information from the request using common.GetRequestUser.
// If there is an error retrieving the user information, it logs the error and returns a JSON response with the error message.
// The function then binds the JSON data from the request to a TranCodeData struct.
// It logs the transaction code, UUID, and data.
// The function marshals the data to JSON and unmarshals it into a TranCode struct.
// If the name or UUID is empty in the request data but not in the TranCode struct, it updates the name or UUID accordingly.
// If the isdefault flag is true, it updates the existing transaction code with the same name to set isdefault to false.
// If there is no ID in the request data, it inserts a new transaction code with the provided data.
// If there is an ID in the request data, it updates the existing transaction code with the provided data.
// Finally, it returns a JSON response with the ID and status.

func (e *TranCodeController) UpdateTranCodeToRespository(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCodeController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.UpdateTranCodeToRespository", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("UpdateTranCodeToRespository defer error: %s", err))
				//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	iLog.User = userno
	iLog.ClientID = clientid

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

	if name == "" && tData.TranCodeName != "" {
		name = tData.TranCodeName
	}

	if uuid == "" && tData.UUID != "" {
		uuid = tData.UUID
	}
	iLog.Debug(fmt.Sprintf("Update transaction code to respository with code: %s", name))
	iLog.Debug(fmt.Sprintf("Update transaction code to respository with uuid: %s", uuid))
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
		update := bson.M{"$set": bson.M{"isdefault": false, "system.updatedon": time.Now().UTC(), "system.updatedby": "system"}}

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
		insertResult, err := documents.DocDBCon.InsertCollection("Transaction_Code", idata)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to insert transaction code: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id = insertResult.InsertedID.(primitive.ObjectID).Hex()
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

	result := map[string]interface{}{
		"id":     id,
		"status": "Success",
	}
	ctx.JSON(http.StatusOK, gin.H{"Outputs": result})

}

// TranCodeRevision is a function that handles the revision of transaction codes.
// It receives a gin.Context object as a parameter and returns no values.
// The function performs the following steps:
// 1. Logs the start time of the function execution.
// 2. Defer a function to log the performance duration of the function.
// 3. Retrieves the user information from the request context.
// 4. Sets the user and client ID in the logger.
// 5. Retrieves the request body from the JSON request.
// 6. Validates and extracts the necessary fields from the request body.
// 7. Checks if the transaction code already exists.
// 8. Queries the transaction code from the database.
// 9. Updates the existing transaction code if the isdefault flag is set.
// 10. Prepares the new transaction code object with the revised information.
// 11. Inserts the new transaction code into the database.
// 12. Returns the ID and status of the inserted transaction code as a JSON response.

func (e *TranCodeController) TranCodeRevision(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCodeController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.TranCodeRevision", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("TranCodeRevision defer error: %s", err))
				//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	iLog.User = userno
	iLog.ClientID = clientid

	iLog.Debug(fmt.Sprintf("Revision transaction code to respository!"))

	request, err := common.GetRequestBodybyJson(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get request body"})
		return
	}

	id, ok := request["_id"].(string)
	if !ok {
		iLog.Error(fmt.Sprintf("Failed to get _id from request"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get _id from request"})
		return
	}
	iLog.Debug(fmt.Sprintf("Revision transaction code to respository with _id: %s", id))

	parsedObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse object id: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse object id"})
		return
	}
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to parse object id: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"_id": parsedObjectID}
	iLog.Debug(fmt.Sprintf("Revision transaction code to respository with filter: %v", filter))

	newvision := request["version"].(string)
	iLog.Debug(fmt.Sprintf("Revision transaction code to respository with version: %s", newvision))
	newname := request["trancodename"].(string)
	iLog.Debug(fmt.Sprintf("Revision transaction code to respository with trancodename: %s", newname))
	isdefault := request["isdefault"].(bool)
	iLog.Debug(fmt.Sprintf("Revision transaction code to respository with isdefault: %s", isdefault))

	vfilter := bson.M{"trancodename": newname, "version": newvision}

	existedobj, err := ValidateIfObjectExist(vfilter, userno, clientid)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to query transaction code: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if existedobj {
		iLog.Error(fmt.Sprintf("The trancode: %s with version: %s already exist!", newname, newvision))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "the trancode alreay exist!"})
		return
	}

	tcitems, err := documents.DocDBCon.QueryCollection("Transaction_Code", filter, nil)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to query transaction code: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(tcitems) == 0 {
		iLog.Error(fmt.Sprintf("Failed to query transaction code: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to query transaction code"})
		return
	}

	tcitem := tcitems[0]
	iLog.Debug(fmt.Sprintf("Revision transaction code to respository with tcitem: %v", tcitem))

	trancodename := tcitem["trancodename"].(string)
	iLog.Debug(fmt.Sprintf("Revision transaction code to respository with trancodename: %s", trancodename))

	if isdefault {
		iLog.Debug(fmt.Sprintf("Revision transaction code to in respository to set default to false: %s", trancodename))
		filter := bson.M{"isdefault": true,
			"trancodename": newname}
		update := bson.M{"$set": bson.M{"isdefault": false, "system.updatedon": time.Now().UTC(), "system.updatedby": "system"}}

		iLog.Debug(fmt.Sprintf("Revision transaction code to in respository with filter: %v", filter))
		iLog.Debug(fmt.Sprintf("Revision transaction code to in respository with update: %v", update))
		err := documents.DocDBCon.UpdateCollection("Transaction_Code", filter, update, nil)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to update transaction code: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	utcTime := time.Now().UTC()
	tcitem["system.updatedon"] = utcTime
	tcitem["system.updatedby"] = "system"
	tcitem["system.createdon"] = utcTime
	tcitem["system.createdby"] = "system"
	tcitem["trancodename"] = newname
	tcitem["version"] = newvision
	tcitem["isdefault"] = isdefault
	tcitem["status"] = 1
	tcitem["uuid"] = uuid.New().String()
	tcitem["description"] = utcTime.String() + ": Revision transaction code " + trancodename + " to " + newname + " with version " + newvision + " \n " + tcitem["description"].(string)
	delete(tcitem, "_id")

	iLog.Debug(fmt.Sprintf("Revision transaction code to respository with tcitem: %v", tcitem))

	insertResult, err := documents.DocDBCon.InsertCollection("Transaction_Code", tcitem)

	if err != nil {
		iLog.Error(fmt.Sprintf("failed to insert transaction code: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id = insertResult.InsertedID.(primitive.ObjectID).Hex()

	result := map[string]interface{}{
		"id":     id,
		"status": "Success",
	}
	ctx.JSON(http.StatusOK, gin.H{"Outputs": result})

}

// ValidateIfObjectExist checks if an object exists in the collection based on the provided filter.
// It takes the filter, user number, and client ID as parameters.
// It returns a boolean indicating whether the object exists and an error if any.

func ValidateIfObjectExist(filter bson.M, userno string, clientid string) (bool, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ValidateIfObjectExist"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.ValidateIfObjectExist", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("ValidateIfObjectExist defer error: %s", err))
				//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	iLog.User = userno
	iLog.ClientID = clientid
	iLog.Debug(fmt.Sprintf("Validate if object exist in collection"))

	collectionitems, err := documents.DocDBCon.QueryCollection("Transaction_Code", filter, nil)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to query collection: %v", err))
		return false, err
	}
	if len(collectionitems) == 0 {
		return false, nil
	}
	return true, nil
}

// getDataFromRequest is a function that retrieves data from the request context and returns it as a TranCodeData struct.
// It also logs performance metrics and any errors encountered during the process.
// The function takes a gin.Context parameter and returns a TranCodeData struct and an error.

func getDataFromRequest(ctx *gin.Context) (TranCodeData, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDataFromRequest"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.trans.getDataFromRequest", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("getDataFromRequest defer error: %s", err))
				//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()  */
	_, userno, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetRequestUser error: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return TranCodeData{}, err
	}

	iLog.User = userno
	iLog.ClientID = clientid

	iLog.Debug(fmt.Sprintf("GetDataFromRequest"))

	var data TranCodeData
	body, err := ioutil.ReadAll(ctx.Request.Body)
	iLog.Debug(fmt.Sprintf("GetDataFromRequest body: %s", body))
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
		return data, err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest Unmarshal error: %s", err.Error()))
		return data, err
	}
	iLog.Debug(fmt.Sprintf("GetDataFromRequest data: %s", data))
	return data, nil
}

type TranCodeData struct {
	TranCode string                 `json:"code"`
	Version  string                 `json:"version"`
	UUID     string                 `json:"uuid"`
	Data     map[string]interface{} `json:"data"`
	Inputs   map[string]interface{} `json:"inputs"`
}

type TranCode struct {
	ID             string           "json:'_id'"
	UUID           string           "json:'uuid'"
	TranCodeName   string           "json:'trancodename'"
	Version        string           "json:'version'"
	IsDefault      bool             "json:'isdefault'"
	Status         int              "json:'status'"
	Firstfuncgroup string           "json:'firstfuncgroup'"
	SystemData     types.SystemData "json:'system'"
	Description    string           "json:'description'"
}
