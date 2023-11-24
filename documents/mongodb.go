package documents

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mdaxf/iac/logger"
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
}

/*
var DatabaseType       = "mongodb"
var DatabaseConnection = "mongodb://localhost:27017"
var DatabaseName       = "IAC_CFG"
*/

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
	}

	return doc.ConnectMongoDB()

}

func (doc *DocDB) ConnectMongoDB() (*DocDB, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.ConnectMongoDB", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.ConnectMongoDB with error: %s", err))
			return
		}
	}()

	doc.iLog.Info(fmt.Sprintf("Connect Database: %s %s", doc.DatabaseType, doc.DatabaseConnection))

	var err error

	doc.MongoDBClient, err = mongo.NewClient(options.Client().ApplyURI(doc.DatabaseConnection))
	if err != nil {
		doc.iLog.Critical(fmt.Sprintf("failed to connect mongodb with error: %s", err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err = doc.MongoDBClient.Connect(ctx)

	if err != nil {
		doc.iLog.Critical(fmt.Sprintf("failed to connect mongodb with error: %s", err))
	}

	doc.MongoDBDatabase = doc.MongoDBClient.Database(doc.DatabaseName)

	return doc, err
}

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

func (doc *DocDB) convertToBsonM(data interface{}) (bson.M, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		doc.iLog.PerformanceWithDuration("documents.convertToBsonM", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			doc.iLog.Error(fmt.Sprintf("There is error to documents.convertToBsonM with error: %s", err))
			return
		}
	}()
	dataBytes, err := bson.Marshal(data)
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
