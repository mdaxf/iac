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
var TransactionTimeout int
var DBTransactionTimeout int

func ConverttoInt(value interface{}) int {
	if value == nil {
		return 0
	}
	switch value.(type) {
	case int:
		return value.(int)
	case float64:
		return int(value.(float64))
	default:
		return 0
	}
}

func ConverttoIntwithDefault(value interface{}, defaultvalue int) int {
	if value == nil {
		return defaultvalue
	}
	switch value.(type) {
	case int:
		return value.(int)
	case float64:
		return int(value.(float64))
	default:
		return defaultvalue
	}
}
