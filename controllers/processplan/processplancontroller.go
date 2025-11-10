package processplan

import (
	"encoding/json"
	"fmt"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/documents"

	"github.com/mdaxf/iac/controllers/common"
)

type ProcessPlanController struct {
}

type CollectionData struct {
	Data      map[string]interface{} `json:"data"`
	Operation string                 `json:"operation"`
	Keys      []string               `json:"keys"`
}

func (c *ProcessPlanController) buildProjectionFromJSON(jsonData []byte, converttype string) (bson.M, error) {
	// Parse JSON into a Go map
	var jsonMap map[string]interface{}
	err := json.Unmarshal(jsonData, &jsonMap)
	if err != nil {
		return nil, err
	}

	// Build the projection
	projection := bson.M{}
	c.buildProjection(jsonMap, "", projection, converttype)

	return projection, nil
}

// buildProjection recursively builds a projection document based on the provided JSON map.
// The projection document is used to specify which fields to include or exclude in a MongoDB query.
// Parameters:
//   - jsonMap: The JSON map containing the field names and values.
//   - prefix: The prefix to be added to the field names.
//   - projection: The projection document being built.
//   - converttype: The type of conversion being performed (e.g., "filter").

func (c *ProcessPlanController) buildProjection(jsonMap map[string]interface{}, prefix string, projection bson.M, converttype string) {
	for key, value := range jsonMap {
		fieldName := key
		if prefix != "" {
			fieldName = prefix + "." + key
		}

		if key == "_id" && converttype == "filter" {
			parsedObjectID, _ := primitive.ObjectIDFromHex(value.(string))
			projection[fieldName] = parsedObjectID
		} else {
			switch v := value.(type) {
			case bool:
				projection[fieldName] = v

			case map[string]interface{}:
				subProjection := bson.M{}
				c.buildProjection(v, fieldName, subProjection, converttype)
				if len(subProjection) > 0 {
					projection[fieldName] = subProjection
				}
			default:
				projection[fieldName] = value
			}
		}
	}
}

func (c *ProcessPlanController) GetListofProcessPlan(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetListofCollectionData"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("collectionop.GetListofCollectionData", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.ClientID = clientid
	iLog.User = user

	var data CollectionData
	/*if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	*/
	iLog.Debug(fmt.Sprintf("Get collection list from respository with body: %s", body))
	print("body:", string(len(body)), body)
	collectionName := "Process_Plan"

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error umarshal body: %v", err))
		//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		//	return
	}

	operation := data.Operation
	items := data.Data
	/*
		condition := map[string]interface{}{}
		for _, key := range Keys {
			condition[key] = 1
		}
	*/
	jsonData, err := json.Marshal(items)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error marshaling json: %v", err))
	}

	iLog.Debug(fmt.Sprintf("Collection Name: %s, operation: %s data: %s", collectionName, operation, jsonData))

	if collectionName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name"})
		return
	}

	if jsonData == nil {
		jsonData = []byte("{}")
	}

	projection, _ := c.buildProjectionFromJSON(jsonData, "projection")

	collectionitems, err := documents.DocDBCon.QueryCollection(collectionName, nil, projection)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the list from collection: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get collection list from respository with data: %s", logger.ConvertJson(collectionitems)))
	if collectionitems == nil {
		collectionitems = []primitive.M{}
	}
	ctx.JSON(http.StatusOK, gin.H{"data": collectionitems})
}

func (c *ProcessPlanController) GetDetailCProcessPlanbyID(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDetailCollectionData"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("collectionop.GetDetailCollectionDatabyID", elapsed)
	}()

	/*	defer func() {
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

		iLog.Debug(fmt.Sprintf("Get collection detail data from respository"))

		var data CollectionData
		if err := ctx.BindJSON(&data); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	*/
	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var data CollectionData
	/*if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	*/
	iLog.Debug(fmt.Sprintf("GetDetailCollectionDatabyID from respository with body: %s", body))
	/*
		err = json.Unmarshal(body, &data)
		if err != nil {
			// Handle the error
			iLog.Error(fmt.Sprintf("Error umarshal body: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	*/
	id := ""
	ok := false
	if id, ok = ctx.Params.Get("id"); ok {
		iLog.Debug(fmt.Sprintf("Get collection detail data from respository with path id: %s", id))
	} else if id, ok = data.Data["_id"].(string); ok {
		iLog.Debug(fmt.Sprintf("Get collection detail data from respository with _id: %s", id))
	}

	// id from path is not used by this handler; use the _id from the request body instead
	collectionName := "Process_Plan"
	operation := "query"
	//list := data.Data
	value := id

	iLog.Debug(fmt.Sprintf("Collection Name: %s, operation: %s data: %s", collectionName, operation))
	if collectionName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name"})
		return
	}

	if value == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	/*	jsonData, err := json.Marshal(list)

			if err != nil {
				iLog.Error(fmt.Sprintf("Error marshaling json: %v", err))
			}
		//filter, _ := c.buildProjectionFromJSON(jsonData, "projection")
		parsedObjectID, _ := primitive.ObjectIDFromHex(value.(string))
		filter := bson.M{"_id": parsedObjectID}

		iLog.Debug(fmt.Sprintf("Collection Name: %s, operation: %s data: %s", collectionName, operation, filter))
	*/
	collectionitems, err := documents.DocDBCon.GetItembyID(collectionName, id)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the detail data from collection: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": collectionitems})
}

// UpdateCollectionData updates the collection data in the repository.
// It retrieves the user information from the request context and binds the JSON data.
// If the collection name is invalid, it returns an error response.
// If the data contains an "_id" field, it updates the collection with the specified ID.
// If the data does not contain an "_id" field, it inserts a new collection into the repository.
// The updated or inserted collection is then returned as a response.
func (c *ProcessPlanController) CreateProcessPlan(ctx *gin.Context) {
	c.UpdateCollectionData(ctx, "Process_Plan")
}

func (c *ProcessPlanController) UpdateProcessPlan(ctx *gin.Context) {
	c.UpdateCollectionData(ctx, "Process_Plan")
}

func (c *ProcessPlanController) UpdateProcessPlanHistory(ctx *gin.Context) {
	c.UpdateCollectionData(ctx, "Process_Plan_History")
}

func (c *ProcessPlanController) UpdateCollectionData(ctx *gin.Context, collectionName string) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UpdateCollectionData"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("collectionop.UpdateCollectionData", elapsed)
	}()
	/*	defer func() {
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

		iLog.Debug(fmt.Sprintf("update collection data to respository"))

		var data CollectionData
		if err := ctx.BindJSON(&data); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	*/
	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user
	/*
		var data CollectionData

		iLog.Debug(fmt.Sprintf("UpdateCollectionData in respository with body: %s", body))

		err = json.Unmarshal(body, &data)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error umarshal body: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		collectionName := CollectionName
		operation := data.Operation
		list := data.Data
		idata := reflect.ValueOf(list)
		Keys := data.Keys

		iLog.Debug(fmt.Sprintf("Collection Name: %s, operation: %s data: %s", collectionName, operation, list))
		if collectionName == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name"})
			return
		}
	*/

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		iLog.Error(fmt.Sprintf("Error parsing JSON:", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name"})
		return
	}
	/*_, err := json.Marshal(list)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error marshaling json: %v", err))
	} */

	id := ""
	ok := false
	if id, ok = ctx.Params.Get("id"); ok {
		iLog.Debug(fmt.Sprintf("Update collection to respository with path id: %s", id))
	} else if id, ok = data["_id"].(string); ok {
		iLog.Debug(fmt.Sprintf("Update collection to respository with _id: %s", id))
	}

	isdefault := false
	if isdefault, ok = data["isdefault"].(bool); ok {
		iLog.Debug(fmt.Sprintf("Update collection to respository with isdefault: %t", isdefault))
	}
	name := ""
	if name, ok = data["name"].(string); ok {
		iLog.Debug(fmt.Sprintf("Update collection to respository with name: %s", name))
	}

	if isdefault {

		Keys := []string{"name"}
		condition := map[string]interface{}{}
		condition["isdefault"] = true
		for _, key := range Keys {
			condition[key] = data[key]
		}

		iLog.Debug(fmt.Sprintf("Update collection to in respository to set default to false: %s", condition))

		con, err := json.Marshal(condition)

		if err != nil {
			iLog.Error(fmt.Sprintf("Error marshaling json: %v", err))
		}

		filter, _ := c.buildProjectionFromJSON(con, "filter")

		collectionitems, err := documents.DocDBCon.QueryCollection(collectionName, filter, nil)

		if err == nil && collectionitems != nil {
			/*
				update := bson.M{"$set": bson.M{"isdefault": false, "system.updatedon": time.Now()}, "system.updatedby": "system"}

				iLog.Debug(fmt.Sprintf("Update collection to in respository with filter: %v", filter))
				iLog.Debug(fmt.Sprintf("Update collection to in respository with update: %v", update))

				err = documents.DocDBCon.UpdateCollection(collectionName, filter, update, nil)
				if err != nil {
					iLog.Error(fmt.Sprintf("failed to update collection: %v", err))
					ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
			*/
		}

	}
	if id == "" && data != nil {
		utcTime := time.Now().UTC()
		system := map[string]interface{}{
			"modifiedon": utcTime,
			"modifiedby": "system",
			"createdon":  utcTime,
			"createdby":  "system",
		}

		data["system"] = system
		delete(data, "_id")
		iLog.Debug(fmt.Sprintf("Insert collection to respository: %s", collectionName))
		insertResult, err := documents.DocDBCon.InsertCollection(collectionName, data)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to insert collection: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id = insertResult.InsertedID.(primitive.ObjectID).Hex()
		//	list["_id"] = id

	} else if data != nil {

		parsedObjectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to parse object id: %v", err))
		}

		iLog.Debug(fmt.Sprintf("Update transaction code to respository with code: %s", name))
		//filedvalue := primitive.ObjectID(param.ID)
		filter := bson.M{"_id": parsedObjectID}
		iLog.Debug(fmt.Sprintf("Update transaction code to respository with filter: %v", filter))
		system := data["system"].(map[string]interface{})
		system["modifiedon"] = time.Now().UTC()
		system["modifiedby"] = "system"
		data["system"] = system
		//list["system.updatedon"] = time.Now().UTC()
		//list["system.updatedby"] = "system"

		delete(data, "_id")

		iLog.Debug(fmt.Sprintf("Update transaction code to respository with data: %s", logger.ConvertJson(data)))

		err = documents.DocDBCon.UpdateCollection(collectionName, filter, nil, data)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to update collection: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	rdata := make(map[string]interface{})
	rdata["id"] = id

	result := map[string]interface{}{
		"data":   rdata,
		"status": "Success",
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// DeleteCollectionDatabyID deletes a collection data by its ID.
// It takes a gin.Context as input and returns the deleted data as JSON response.
// If there is an error, it returns an error message as JSON response.

func (c *ProcessPlanController) DeleteProcessPlanbyID(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "DeleteCollectionData"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("collectionop.DeleteCollectionDatabyID", elapsed)
	}()
	/*	defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("collectionop.DeleteCollectionDatabyID Error: %v", err))
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

		iLog.Debug(fmt.Sprintf("delete collection data to respository"))
		var data CollectionData
		if err := ctx.BindJSON(&data); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	*/
	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var data CollectionData

	iLog.Debug(fmt.Sprintf("DeleteCollectionDatabyID from respository with body: %s", body))

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error umarshal body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := ""
	ok := false
	if id, ok = ctx.Params.Get("id"); ok {
		iLog.Debug(fmt.Sprintf("Update collection to respository with path id: %s", id))
	} else if id, ok = data.Data["_id"].(string); ok {
		iLog.Debug(fmt.Sprintf("Update collection to respository with _id: %s", id))
	}

	collectionName := "Process_Plan"
	operation := data.Operation
	list := data.Data
	value := id

	iLog.Debug(fmt.Sprintf("Collection Name: %s, operation: %s data: %s", collectionName, operation, list))
	if collectionName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name"})
		return
	}

	if value == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	err = documents.DocDBCon.DeleteItemFromCollection(collectionName, value)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Delete item from collection error!"})
		return
	}

	rdata := make(map[string]interface{})
	rdata["id"] = value
	result := map[string]interface{}{
		"data":   rdata,
		"status": "Success",
	}

	ctx.JSON(http.StatusOK, gin.H{"data": result})
}
