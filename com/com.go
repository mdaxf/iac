package com

import (
	"go.mongodb.org/mongo-driver/mongo"
)

var Instance string
var InstanceType string
var InstanceName string
var MongoDBClients []*mongo.Client
