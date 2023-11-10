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
	return ConverttoIntwithDefault(value, 0)
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

func ConverttoFloat64(value interface{}) float64 {
	return ConverttoFloat64withDefault(value, 0)
}

func ConverttoFloat64withDefault(value interface{}, defaultvalue float64) float64 {
	if value == nil {
		return defaultvalue
	}
	switch value.(type) {
	case int:
		return float64(value.(int))
	case float64:
		return value.(float64)
	default:
		return defaultvalue
	}
}

func ConverttoBoolean(value interface{}) bool {
	return ConverttoBooleanwithDefault(value, false)
}

func ConverttoBooleanwithDefault(value interface{}, defaultvalue bool) bool {
	if value == nil {
		return defaultvalue
	}
	switch value.(type) {
	case bool:
		return value.(bool)
	case int:
		return value.(int) != 0
	case float64:
		return value.(float64) != 0
	default:
		return defaultvalue
	}
}
