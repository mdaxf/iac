package documents

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

func InitMongDB(DatabaseConnection string, DatabaseName string) (*DocDB, error) {
	doc := &DocDB{
		DatabaseConnection: DatabaseConnection,
		DatabaseName:       DatabaseName,
		iLog:               logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MongoDB Connection"},
	}

	return doc.ConnectMongoDB()

}

func (doc *DocDB) ConnectMongoDB() (*DocDB, error) {

	doc.iLog.Info(fmt.Sprintf("Connect Database: %s %s", doc.DatabaseType, doc.DatabaseConnection))

	var err error

	doc.MongoDBClient, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
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

func (doc *DocDB) UpdateCollection(collectionname string, filter bson.M, update bson.M, idata interface{}) error {

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

func (doc *DocDB) InsertCollection(collectionname string, idata interface{}) error {

	MongoDBCollection := doc.MongoDBDatabase.Collection(collectionname)

	data, err := doc.convertToBsonM(idata)
	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to update data from collection with error: %s", err))
		return err
	}

	_, err = MongoDBCollection.InsertOne(context.Background(), data)

	if err != nil {
		doc.iLog.Error(fmt.Sprintf("failed to insert data from collection with error: %s", err))
	}

	return err
}

func (doc *DocDB) convertToBsonM(data interface{}) (bson.M, error) {
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
