package com

import (
	"github.com/mdaxf/signalrsrv/signalr"
	"go.mongodb.org/mongo-driver/mongo"
)

var Instance string
var InstanceType string
var InstanceName string
var MongoDBClients []*mongo.Client
var SingalRConfig map[string]interface{}

var IACMessageBusClient signalr.Client
