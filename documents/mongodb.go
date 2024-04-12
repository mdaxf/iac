package documents

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	//	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/logger"
	//	"github.com/mdaxf/iac/framework/documentdb/mongodb"
)

type DocDB struct {
	MongoDBClient        *mongo.Client
	MongoDBDatabase      *mongo.Database
	MongoDBCollection_TC *mongo.Collection
	/*
	 */
	DatabaseType       string
	DatabaseConnection string
	DatabaseName       string
	iLog               logger.Log
	monitoring         bool
}

/*
var DatabaseType       = "mongodb"
var DatabaseConnection = "mongodb://localhost:27017"
var DatabaseName       = "IAC_CFG"
*/

// InitMongoDB initializes a MongoDB connection and returns a DocDB object.
// It takes two parameters: DatabaseConnection (the MongoDB connection string) and DatabaseName (the name of the database).
// It returns a pointer to a DocDB object and an error if any.
// The function logs the performance duration and recovers from any panics.

func InitMongoDB(DatabaseConnection string, DatabaseName string) (*DocDB, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MongoDB Connection"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("documents.InitMongoDB", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to documents.InitMongoDB with error: %s", err))
			return
		}
	}()

	doc := &DocDB{
		DatabaseConnection: DatabaseConnection,
		DatabaseName:       DatabaseName,
		iLog:               iLog,
		monitoring:         false,
	}

	_, err := doc.ConnectMongoDB()

	if err != nil {
		iLog.Error(fmt.Sprintf("There is error to connect to MongoDB with error: %s", err))
		return doc, err
	}

	if doc.monitoring == false {
		go func() {
			doc.MonitorAndReconnect()
		}()

	}

	return doc, nil
}

// ConnectMongoDB establishes a connection to the MongoDB database.
// It returns a pointer to the DocDB struct and an error if any.
func (doc *DocDB) ConnectMongoDB() (*DocDB, error) {
	// Measure the execution time of the function
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.ConnectMongoDB", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.ConnectMongoDB with error: %s", err))
			return
		}
	}()
	/*
		myMongodb := &mongodb.MyDocDB{
			DatabaseConnection: doc.DatabaseConnection,
			DatabaseName:       doc.DatabaseName,
		}

		err := myMongodb.Connect()

		if err != nil {
			doc.iLog.Error(fmt.Sprintf("There os error to connect to MongoDB %s %s", doc.DatabaseConnection, doc.DatabaseName))
			return nil, err
		}

		doc.MongoDBClient = myMongodb.MongoDBClient
		doc.MongoDBDatabase = myMongodb.MongoDBDatabase

		com.IACDocDBConn = &com.DocDB{
			DatabaseConnection: myMongodb.DatabaseConnection,
			DatabaseName:       myMongodb.DatabaseName,
			MongoDBClient:      myMongodb.MongoDBClient,
			MongoDBDatabase:    myMongodb.MongoDBDatabase}

		return doc, nil  */
	// Log the database connection details
	doc.iLog.Info(fmt.Sprintf("Connect Database: %s %s", doc.DatabaseType, doc.DatabaseConnection))

	var err error

	// Create a new MongoDB client
	doc.MongoDBClient, err = mongo.NewClient(options.Client().ApplyURI(doc.DatabaseConnection))
	if err != nil {
		doc.iLog.Critical(fmt.Sprintf("failed to connect mongodb with error: %s", err))
		return doc, err
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to the MongoDB server
	err = doc.MongoDBClient.Connect(ctx)
	if err != nil {
		doc.iLog.Critical(fmt.Sprintf("failed to connect mongodb with error: %s", err))
		return doc, err
	}

	// Set the MongoDB database
	doc.MongoDBDatabase = doc.MongoDBClient.Database(doc.DatabaseName)

	//	if doc.monitoring == false {
	//		doc.monitorAndReconnect()
	//	}

	err = doc.MongoDBClient.Ping(context.Background(), nil)
	if err != nil {
		doc.iLog.Critical(fmt.Sprintf("failed to connect mongodb with error: %s", err))
		return doc, err
	}
	return doc, err
}

func (doc *DocDB) MonitorAndReconnect() {
	// Function execution logging
	//	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MongoDB.monitorAndReconnect"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("MongoDB.monitorAndReconnect", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("monitorAndReconnect defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	doc.monitoring = true
	for {

		err := doc.MongoDBClient.Ping(context.Background(), nil)
		if err != nil {
			doc.iLog.Error(fmt.Sprintf("MongoDB connection lost with ping error %v, reconnecting...", err))

			_, err := doc.ConnectMongoDB()

			if err != nil {
				doc.iLog.Error(fmt.Sprintf("Failed to reconnect to MongoDB %s with error:%v", doc.DatabaseConnection, err))
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			} else {
				time.Sleep(1 * time.Second)
				doc.iLog.Debug(fmt.Sprintf("MongoDB reconnected successfully"))
				continue
			}
		} else {
			time.Sleep(1 * time.Second) // Check connection every 60 seconds
			continue
		}
	}

}

// QueryCollection queries a MongoDB collection with the specified filter and projection.
// It returns an array of documents that match the filter, along with any error that occurred.
// The function logs the performance duration and recovers from any panics.
// The function uses the MongoDB Go driver to query the collection.
func (doc *DocDB) QueryCollection(collectionname string, filter bson.M, projection bson.M) ([]bson.M, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.QueryCollection", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.QueryCollection with error: %s", err))
			return
		}
	}()

	MongoDBCollection := doc.MongoDBDatabase.Collection(collectionname)

	options := options.Find()
	options.SetProjection(projection)

	cursor, err := MongoDBCollection.Find(context.TODO(), filter, options)

	if err != nil {
		doc.iLog.Critical(fmt.Sprintf("failed to get data from collection with error: %s", err))
	}

	defer cursor.Close(context.Background())

	var results []bson.M

	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			doc.iLog.Error(fmt.Sprintf("failed to decode data from collection with error: %s", err))
			return nil, err
		}
		results = append(results, result)
	}

	if err := cursor.Err(); err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to get data from collection with error: %s", err))
		return nil, err
	}

	return results, nil
}

func (doc *DocDB) GetItembyUUID(collectionname string, uuid string) (bson.M, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.GetDefaultItembyName", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.GetDefaultItembyName with error: %s", err))
			return
		}
	}()

	MongoDBCollection := doc.MongoDBDatabase.Collection(collectionname)

	filter := bson.M{"uuid": uuid}

	doc.iLog.Debug(fmt.Sprintf("GetDefaultItembyName: %s from collection:%s", filter, collectionname))
	//var result bson.Raw
	var result bson.M
	err := MongoDBCollection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to get data from collection with error: %s", err))
	}
	doc.iLog.Debug(fmt.Sprintf("GetDefaultItembyName: %s", result))

	return result, err
}

// GetDefaultItembyName retrieves the default item from the specified collection by name.
// It takes the collection name and the name of the item as input parameters.
// It returns the retrieved item as a bson.M object and an error if any.
// The function logs the performance duration and recovers from any panics.
// The function uses the MongoDB Go driver to query the collection.
func (doc *DocDB) GetDefaultItembyName(collectionname string, name string) (bson.M, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.GetDefaultItembyName", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.GetDefaultItembyName with error: %s", err))
			return
		}
	}()

	MongoDBCollection := doc.MongoDBDatabase.Collection(collectionname)

	filter := bson.M{"name": name, "isdefault": true}

	doc.iLog.Debug(fmt.Sprintf("GetDefaultItembyName: %s from collection:%s", filter, collectionname))
	//var result bson.Raw
	var result bson.M
	err := MongoDBCollection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to get data from collection with error: %s", err))
	}
	doc.iLog.Debug(fmt.Sprintf("GetDefaultItembyName: %s", result))

	return result, err
}

// GetItembyID retrieves an item from the specified collection by its ID.
// It takes the collection name and the ID as parameters.
// It returns the item as a bson.M object and an error if any.
// The function logs the performance duration and recovers from any panics.
// The function uses the MongoDB Go driver to query the collection.

func (doc *DocDB) GetItembyID(collectionname string, id string) (bson.M, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.GetItembyID", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.GetItembyID with error: %s", err))
			return
		}
	}()

	MongoDBCollection := doc.MongoDBDatabase.Collection(collectionname)

	objectid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to convert id to objectid with error: %s", err))
	}

	filter := bson.M{"_id": objectid}

	doc.iLog.Debug(fmt.Sprintf("GetItembyID: %s from collection:%s", filter, collectionname))
	//var result bson.Raw
	var result bson.M
	err = MongoDBCollection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to get data from collection with error: %s", err))
	}
	doc.iLog.Debug(fmt.Sprintf("GetItembyID: %s", result))
	/*
		jsonBytes, err := bson.MarshalExtJSON(result, true, false)
		if err != nil {
			doc.iLog.Error(fmt.Sprintf("failed to convert data to json with error: %s", err))
		}
		jsonString := string(jsonBytes)
		doc.iLog.Debug(fmt.Sprintf("GetItembyID result: %s", jsonString))
	*/
	return result, err
}

// GetItembyName retrieves an item from the specified collection by its name.
// It takes the collection name and the name as parameters.
// It returns the item as a bson.M object and an error if any.
// The function logs the performance duration and recovers from any panics.
// The function uses the MongoDB Go driver to query the collection.
func (doc *DocDB) UpdateCollection(collectionname string, filter bson.M, update bson.M, idata interface{}) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.UpdateCollection", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.UpdateCollection with error: %s", err))
			return
		}
	}()

	MongoDBCollection := doc.MongoDBDatabase.Collection(collectionname)

	if update == nil && idata != nil {

		data, err := doc.convertToBsonM(idata)
		if err != nil {
			doc.iLog.Error(fmt.Sprintf("failed to update data from collection with error: %s", err))
			return err
		}
		_, err = MongoDBCollection.ReplaceOne(context.Background(), filter, data)
		if err != nil {
			doc.iLog.Error(fmt.Sprintf("failed to update data from collection with error: %s", err))
		}
		return err
	} else {
		_, err := MongoDBCollection.UpdateOne(context.Background(), filter, update)

		if err != nil {
			doc.iLog.Critical(fmt.Sprintf("failed to update data from collection with error: %s", err))
		}
		return err
	}

}

// InsertCollection inserts a new item into the specified collection.
// It takes the collection name and the item as parameters.
// It returns the result of the insert operation and an error if any.
// The function logs the performance duration and recovers from any panics.
// The function uses the MongoDB Go driver to insert the item into the collection.

func (doc *DocDB) InsertCollection(collectionname string, idata interface{}) (*mongo.InsertOneResult, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.InsertCollection", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.InsertCollection with error: %s", err))
			return
		}
	}()
	MongoDBCollection := doc.MongoDBDatabase.Collection(collectionname)

	data, err := doc.convertToBsonM(idata)
	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to update data from collection with error: %s", err))
		return nil, err
	}

	insertResult, err := MongoDBCollection.InsertOne(context.Background(), data)

	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to insert data from collection with error: %s", err))
	}

	return insertResult, err
}

// DeleteItemFromCollection deletes an item from the specified collection by its ID.
// It takes the collection name and the ID as parameters.
// It returns an error if any.
// The function logs the performance duration and recovers from any panics.
// The function uses the MongoDB Go driver to delete the item from the collection.
func (doc *DocDB) DeleteItemFromCollection(collectionname string, documentid string) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.DeleteItemFromCollection", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.DeleteItemFromCollection with error: %s", err))
			return
		}
	}()

	doc.iLog.Debug(fmt.Sprintf("Delete the item %s from collection %s", documentid, collectionname))

	MongoDBCollection := doc.MongoDBDatabase.Collection(collectionname)

	objectid, err := primitive.ObjectIDFromHex(documentid)
	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to convert id to objectid with error: %s", err))
		return err
	}

	filter := bson.M{"_id": objectid}

	_, err = MongoDBCollection.DeleteOne(context.Background(), filter)

	if err != nil {
		doc.iLog.Error(fmt.Sprintf("Delete the item %s from collection %s error %s!", documentid, collectionname, err))
		return err
	}
	return nil
}

// DeleteCollection deletes a collection from the MongoDB database.
// It takes the collection name as a parameter.
// It returns an error if any.
// The function logs the performance duration and recovers from any panics.
// The function uses the MongoDB Go driver to delete the collection.

func (doc *DocDB) convertToBsonM(data interface{}) (bson.M, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.convertToBsonM", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				doc.iLog.Error(fmt.Sprintf("There is error to documents.convertToBsonM with error: %s", err))
				return
			}
		}()
	*/dataBytes, err := bson.Marshal(data)
	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to convert data to bson.M with error: %s", err))
		return nil, err
	}
	var dataBsonM bson.M
	err = bson.Unmarshal(dataBytes, &dataBsonM)
	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to convert data to bson.M with error: %s", err))
		return nil, err
	}
	return dataBsonM, nil
}
