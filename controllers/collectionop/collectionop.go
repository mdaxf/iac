package collectionop

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/documents"
)

type CollectionController struct {
}

type CollectionData struct {
	CollectionName string                 `json:"collectionname"`
	Data           map[string]interface{} `json:"data"`
	Operation      string                 `json:"operation"`
	Keys           []string               `json:"keys"`
}

func (c *CollectionController) GetListofCollectionData(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetListofCollectionData"}
	iLog.Debug(fmt.Sprintf("Get collection list from respository"))

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer ctx.Request.Body.Close()

	iLog.Debug(fmt.Sprintf("Get collection list from respository with body: %s", body))

	var data CollectionData
	/*if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	*/
	err = json.Unmarshal(body, &data)
	if err != nil {
		// Handle the error
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collectionName := data.CollectionName
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

	projection, _ := c.buildProjectionFromJSON(jsonData, "projection")

	collectionitems, err := documents.DocDBCon.QueryCollection(collectionName, nil, projection)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the list from collection: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get collection list from respository with data: %s", logger.ConvertJson(collectionitems)))

	ctx.JSON(http.StatusOK, gin.H{"data": collectionitems})
}
func (c *CollectionController) GetDetailCollectionData(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDetailCollectionData"}
	iLog.Debug(fmt.Sprintf("Get collection detail data from respository"))

	var data CollectionData
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collectionName := data.CollectionName
	operation := data.Operation
	list := data.Data

	iLog.Debug(fmt.Sprintf("Collection Name: %s, operation: %s data: %s", collectionName, operation, list))
	if collectionName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name"})
		return
	}
	filter := bson.M(list)

	iLog.Debug(fmt.Sprintf("Collection Name: %s, operation: %s data: %s", collectionName, operation, filter))
	collectionitems, err := documents.DocDBCon.QueryCollection(collectionName, filter, nil)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the detail data from collection: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": collectionitems})
}
func (c *CollectionController) GetDetailCollectionDatabyID(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDetailCollectionData"}
	iLog.Debug(fmt.Sprintf("Get collection detail data from respository"))

	var data CollectionData
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collectionName := data.CollectionName
	operation := data.Operation
	list := data.Data
	value := list["_id"]

	iLog.Debug(fmt.Sprintf("Collection Name: %s, operation: %s data: %s", collectionName, operation, list))
	if collectionName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name"})
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
	collectionitems, err := documents.DocDBCon.GetItembyID(collectionName, value.(string))

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the detail data from collection: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": collectionitems})
}
func (c *CollectionController) GetDetailCollectionDatabyName(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDetailCollectionDatabyName"}
	iLog.Debug(fmt.Sprintf("Get default collection detail data from respository"))

	var data CollectionData
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collectionName := data.CollectionName

	list := data.Data
	value := list["name"]

	iLog.Debug(fmt.Sprintf("Collection Name: %s, data: %s", collectionName, list))
	if collectionName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name"})
		return
	}

	collectionitems, err := documents.DocDBCon.GetDefaultItembyName(collectionName, value.(string))

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the detail data from collection: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": collectionitems})
}
func (c *CollectionController) UpdateCollectionData(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UpdateCollectionData"}
	iLog.Debug(fmt.Sprintf("update collection data to respository"))

	var data CollectionData
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collectionName := data.CollectionName
	operation := data.Operation
	list := data.Data
	idata := reflect.ValueOf(list)
	Keys := data.Keys

	iLog.Debug(fmt.Sprintf("Collection Name: %s, operation: %s data: %s", collectionName, operation, list))
	if collectionName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name"})
		return
	}

	/*_, err := json.Marshal(list)

	if err != nil {
		iLog.Error(fmt.Sprintf("Error marshaling json: %v", err))
	} */

	id := ""
	ok := false
	if id, ok = list["_id"].(string); ok {
		iLog.Debug(fmt.Sprintf("Update collection to respository with _id: %s", id))
	}

	isdefault := false
	if isdefault, ok = list["isdefault"].(bool); ok {
		iLog.Debug(fmt.Sprintf("Update collection to respository with isdefault: %t", isdefault))
	}
	name := ""
	if name, ok = list["name"].(string); ok {
		iLog.Debug(fmt.Sprintf("Update collection to respository with name: %s", name))
	}

	if isdefault {

		condition := map[string]interface{}{}
		condition["isdefault"] = true
		for _, key := range Keys {
			condition[key] = list[key]
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
	if id == "" && list != nil {
		utcTime := time.Now().UTC()
		system := map[string]interface{}{
			"updatedon": utcTime,
			"updatedby": "system",
			"createdon": utcTime,
			"createdby": "system",
		}

		list["system"] = system

		iLog.Debug(fmt.Sprintf("Insert collection to respository: %s", collectionName))
		insertResult, err := documents.DocDBCon.InsertCollection(collectionName, list)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to insert collection: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id = insertResult.InsertedID.(primitive.ObjectID).Hex()

	} else if list != nil {

		parsedObjectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to parse object id: %v", err))
		}

		iLog.Debug(fmt.Sprintf("Update transaction code to respository with code: %s", name))
		//filedvalue := primitive.ObjectID(param.ID)
		filter := bson.M{"_id": parsedObjectID}
		iLog.Debug(fmt.Sprintf("Update transaction code to respository with filter: %v", filter))
		system := list["system"].(map[string]interface{})
		system["updatedon"] = time.Now().UTC()
		system["updatedby"] = "system"
		list["system"] = system
		//list["system.updatedon"] = time.Now().UTC()
		//list["system.updatedby"] = "system"

		delete(list, "_id")

		iLog.Debug(fmt.Sprintf("Update transaction code to respository with data: %s", logger.ConvertJson(idata)))

		err = documents.DocDBCon.UpdateCollection(collectionName, filter, nil, list)
		if err != nil {
			iLog.Error(fmt.Sprintf("failed to update collection: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	result := map[string]interface{}{
		"data":   list,
		"status": "Success",
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}
func (c *CollectionController) buildProjectionFromJSON(jsonData []byte, converttype string) (bson.M, error) {
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

func (c *CollectionController) buildProjection(jsonMap map[string]interface{}, prefix string, projection bson.M, converttype string) {
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
