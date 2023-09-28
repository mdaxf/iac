// The package is migrated from beego, you can get from following link:
// import(
//   "github.com/beego/beego/v2/client/cache"
// )
// Copyright 2023. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mdaxf/iac/framework/berror"

	"github.com/mdaxf/iac/logger"
)

// Cache DocumentDB Cache adapter.
type DocumentDBCache struct {
	MongoDBClient        *mongo.Client
	MongoDBDatabase      *mongo.Database
	MongoDBCollection_TC *mongo.Collection
	/*
	 */
	CollectionName string
	iLog           logger.Log
}
type Item struct {
	Key        string      `bson:"key"`
	Value      interface{} `bson:"value"`
	Expiration int32       `bson:"expiration"`
}

// NewMemCache creates a new memcache adapter.
func NewDocumentDBCache() Cache {
	return &DocumentDBCache{}
}

// Get get value from memcache.
func (doc *DocumentDBCache) Get(ctx context.Context, key string) (interface{}, error) {
	var item Item

	err := doc.MongoDBDatabase.Collection(doc.CollectionName).FindOne(ctx, bson.M{"key": key}).Decode(&item)
	if err == nil {
		return item.Value, nil
	} else {
		return nil, berror.Wrapf(err, MemCacheCurdFailed,
			"could not read data from memcache, please check your key, network and connection. Root cause: %s",
			err.Error())
	}
}

// GetMulti gets a value from a key in memcache.
func (doc *DocumentDBCache) GetMulti(ctx context.Context, keys []string) ([]interface{}, error) {
	rv := make([]interface{}, len(keys))
	var item Item
	filter := bson.M{"key": bson.M{"$in": keys}}
	cur, err := doc.MongoDBDatabase.Collection(doc.CollectionName).Find(ctx, filter)
	if err != nil {
		return rv, berror.Wrapf(err, MemCacheCurdFailed,
			"could not read multiple key-values from memcache, "+
				"please check your keys, network and connection. Root cause: %s",
			err.Error())
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		err := cur.Decode(&item)
		if err != nil {
			return rv, berror.Wrapf(err, MemCacheCurdFailed,
				"could not read multiple key-values from memcache, "+
					"please check your keys, network and connection. Root cause: %s",
				err.Error())
		}
		rv = append(rv, item.Value)
	}
	if err := cur.Err(); err != nil {
		return rv, berror.Wrapf(err, MemCacheCurdFailed,
			"could not read multiple key-values from memcache, "+
				"please check your keys, network and connection. Root cause: %s",
			err.Error())
	}
	return rv, nil
}

// Put puts a value into memcache.
func (doc *DocumentDBCache) Put(ctx context.Context, key string, val interface{}, timeout time.Duration) error {
	item := Item{Key: key, Expiration: int32(timeout / time.Second), Value: val}
	_, err := doc.MongoDBDatabase.Collection(doc.CollectionName).InsertOne(ctx, item)
	if err != nil {
		return berror.Wrapf(err, MemCacheCurdFailed,
			"could not put key-value to memcache, key: %s", key)
	}
	return nil
}

// Delete deletes a value in memcache.
func (doc *DocumentDBCache) Delete(ctx context.Context, key string) error {
	_, err := doc.MongoDBDatabase.Collection(doc.CollectionName).DeleteOne(ctx, bson.M{"key": key})
	if err != nil {
		return berror.Wrapf(err, MemCacheCurdFailed,
			"could not delete key-value from memcache, key: %s", key)
	}
	return nil
}

// Incr increases counter.
func (doc *DocumentDBCache) Incr(ctx context.Context, key string) error {
	return nil
}

// Decr decreases counter.
func (doc *DocumentDBCache) Decr(ctx context.Context, key string) error {
	return nil
}

// IsExist checks if a value exists in memcache.
func (doc *DocumentDBCache) IsExist(ctx context.Context, key string) (bool, error) {
	var item Item
	err := doc.MongoDBDatabase.Collection(doc.CollectionName).FindOne(ctx, bson.M{"key": key}).Decode(&item)
	if err == nil {
		return true, nil
	} else {
		return false, berror.Wrapf(err, MemCacheCurdFailed,
			"could not read data from memcache, please check your key, network and connection. Root cause: %s",
			err.Error())
	}

}

// ClearAll clears all cache in memcache.
func (doc *DocumentDBCache) ClearAll(context.Context) error {
	doc.MongoDBDatabase.Collection(doc.CollectionName).Drop(context.Background())
	return nil
}

// StartAndGC starts the memcache adapter.
// config: must be in the format {"conn":"connection info"}.
// If an error occurs during connecting, an error is returned
func (doc *DocumentDBCache) StartAndGC(config string) error {
	doc.iLog = logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Cache: DocumentDB"}
	var err error

	var cf map[string]string
	if err := json.Unmarshal([]byte(config), &cf); err != nil {
		doc.iLog.Critical(fmt.Sprintf("could not unmarshal this config, it must be valid json stringP: %s with error %s", config, err))
		return berror.Wrapf(err, InvalidMemCacheCfg,
			"could not unmarshal this config, it must be valid json stringP: %s", config)
	}

	if _, ok := cf["conn"]; !ok {
		return berror.Errorf(InvalidMemCacheCfg, `config must contains "conn" field: %s`, config)
	}

	doc.MongoDBClient, err = mongo.NewClient(options.Client().ApplyURI(cf["conn"]))
	if err != nil {
		doc.iLog.Critical(fmt.Sprintf("failed to connect mongodb with error: %s", err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err = doc.MongoDBClient.Connect(ctx)

	if err != nil {
		doc.iLog.Critical(fmt.Sprintf("failed to connect mongodb with error: %s", err))
		return berror.Errorf(InvalidMemCacheCfg, `cannot connec tto mondodbwith configuration: %s`, config)
	}
	if _, ok := cf["db"]; !ok {
		return berror.Errorf(InvalidMemCacheCfg, `config must contains "db" field: %s`, config)
	}
	doc.MongoDBDatabase = doc.MongoDBClient.Database(cf["db"])
	if _, ok := cf["collection"]; !ok {
		return berror.Errorf(InvalidMemCacheCfg, `config must contains "collection" field: %s`, config)
	}

	doc.CollectionName = cf["collection"]
	return nil
}

func init() {
	Register("documentdb", NewDocumentDBCache)
}
