package com

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/mdaxf/signalrsrv/signalr"
	"go.mongodb.org/mongo-driver/mongo"
)

var Instance string
var InstanceType string
var InstanceName string
var SingalRConfig map[string]interface{}
var TransactionTimeout int
var DBTransactionTimeout int

var MongoDBClients []*mongo.Client

var IACMessageBusClient signalr.Client

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
	case string:
		temp, err := strconv.Atoi(value.(string))
		if err != nil {
			return defaultvalue
		}
		return temp
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

func ConverttoString(value interface{}) string {
	return ConverttoStringwithDefault(value, "")
}

func ConverttoStringwithDefault(value interface{}, defaultvalue string) string {
	if value == nil {
		return defaultvalue
	}

	if str, ok := value.(string); ok {
		return str
	} else {
		switch v := value.(type) {
		case string:
			return v
		case int:
			return strconv.Itoa(v)
		case int64:
			return strconv.FormatInt(v, 10)
		case uint:
			return strconv.FormatUint(uint64(v), 10)
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		default:
			return fmt.Sprintf("%v", value)
		}
	}
}

func ConvertstructToMap(input interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	val := reflect.ValueOf(input)
	typ := reflect.TypeOf(input)

	// Check if the input is a struct
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldName := typ.Field(i).Name

			// Add the field to the map
			result[fieldName] = field.Interface()
		}
	}

	return result
}

func ConvertbytesToMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DecodeBase64String(input string) (string, error) {
	decodedData, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(decodedData), nil
}

func Convertbase64ToMap(input string) (map[string]interface{}, error) {
	// Decode Base64 string
	decodedData, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}
	println("data:", string(decodedData))
	return ConvertbytesToMap(decodedData)
	// Unmarshal JSON data
	var resultMap map[string]interface{}
	err = json.Unmarshal(decodedData, &resultMap)
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func ConvertInterfaceToString(input interface{}) (string, error) {
	jsondata, err := json.Marshal(input)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to convert json to map: %v", err))
		return "", err
	}
	return DecodeBase64String(fmt.Sprintf("%s", jsondata))
}

func ConvertMapToString(data map[string]interface{}) (string, error) {
	// Marshal the map to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Convert the byte slice to a string
	jsonString := string(jsonBytes)
	return jsonString, nil
}
