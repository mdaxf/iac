package databaseop

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

type DBController struct {
}

type DBData struct {
	TableName string                 `json."tablename"` // table name
	Data      map[string]interface{} `json."data"`
	Operation string                 `json."operation"` // insert, update, delete
	Keys      []string               `json."keys"`      // keys for update and delete
	Where     map[string]interface{} `json."where"`     // where args for update and delete
}

/*
	{
	    "tablename": "EMPLOYEE",
	    "data": {
	        "EMPLOYEE":{
	            "fields":["ID", "Name", "LoginName"],
	            "subtables":{
	                "EMPLOYEE_ROLE":{
	                    "fields":[],
	                    "links":["EMPLOYEE_ROLE.EmployeeID = EMPLOYEE.ID"],
	                    "subtables": {
	                        "ROLE":{
	                            "fields":["ID As RoleID", "ROLE"],
	                            "links": ["ROLE.ID = EMPLOYEE_ROLE.RoleID"]
	                        }
	                    }
	                }
	            }
	        },
	        "RESOURCE_":["facility", "productionline"]
	    },
	    "operation": "detail",
	    "where": {
	        "EMPLOYEE.ResourceID = RESOURCE_.ID":""


	    }
	}
*/
func (db *DBController) GetDataFromTables(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDataFromTable"}
	iLog.Debug(fmt.Sprintf("Get data from table"))

	data, err := db.GetDataFromRequest(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get data from table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get data from table: %s", data.TableName))

	Query, TableNames, err := db.getDataStructForQuery(data.Data)
	Wherestr := ""
	iLog.Debug(fmt.Sprintf("get where condition: %s", data.Where))
	for key, value := range data.Where {
		iLog.Debug(fmt.Sprintf("get where condition: %s %s", key, value))
		if value == "" {
			Wherestr = fmt.Sprintf("%s %s ", Wherestr, key)
		} else {
			Wherestr = fmt.Sprintf("%s %s='%s'", Wherestr, key, value)
		}
	}
	if Wherestr != "" {
		Query = fmt.Sprintf("SELECT %s from %s where %s", Query, TableNames, Wherestr)
	} else {
		Query = fmt.Sprintf("SELECT %s from %s", Query, TableNames)
	}
	iLog.Debug(fmt.Sprintf("Get data from query: %s", Query))

	// get data from database

	result, err := dbconn.NewDBOperation("system", nil, "Execute Query Function").Query_Json(Query)

	if err != nil {
		iLog.Error(fmt.Sprintf("Get data from table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get data from table result: %s", gin.H{"data": result}))
	//jsondata, err := json.Marshal(result)

	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

func (db *DBController) getDataStructForQuery(data map[string]interface{}) (string, string, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDataFromTable"}
	iLog.Debug(fmt.Sprintf("get data struct for query"))
	Query := ""
	TableName := ""
	for k, v := range data {
		tablename := k
		if TableName != "" {
			TableName = fmt.Sprintf(" %s INNER JOIN %s ON 1=1 ", TableName, tablename)
		} else {
			TableName = fmt.Sprintf(" %s ", tablename)
		}

		// get table schema
		var fields []string
		if itemList, ok := v.([]interface{}); ok {
			for _, field := range itemList {
				fields = append(fields, field.(string))
				if Query != "" {
					Query = fmt.Sprintf("%s, %s.%s", Query, tablename, field.(string))
				} else {
					Query = fmt.Sprintf("%s %s.%s", Query, tablename, field.(string))
				}
			}

		} else {
			if item, ok := v.(map[string]interface{}); ok {

				subquery, tablelinks, err := db.getmysqlsubtabls(tablename, item, false)
				if err != nil {
					return "", "", err
				}
				if Query != "" {
					Query = fmt.Sprintf("%s, %s", Query, subquery)
				} else {

					Query = subquery
				}

				TableName = fmt.Sprintf("%s %s", TableName, tablelinks)
			}
		}

	}
	iLog.Debug(fmt.Sprintf("getDataStructForQuery Query: %s, %s", Query, TableName))
	return strings.TrimRight(Query, ","), strings.TrimRight(TableName, ","), nil
}

func (db *DBController) getmysqlsubtabls(tablename string, data map[string]interface{}, markasJson bool) (string, string, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDataFromTable"}
	Links := ""
	Query := " "
	TableLinks := ""
	SubQuery := ""
	SubLinks := ""
	for k, v := range data {
		if k == "fields" {
			if itemList, ok := v.([]interface{}); ok {
				for _, field := range itemList {
					Query = fmt.Sprintf("%s %s.%s,", Query, tablename, field.(string))
				}
			}
		} else if k == "subtables" {
			if item, ok := v.(map[string]interface{}); ok {
				for key, value := range item {
					subquery, subtablelink, err := db.getmysqlsubtabls(key, value.(map[string]interface{}), true)
					if err != nil {
						return "", "", err
					}
					SubQuery = fmt.Sprintf("%s %s,", SubQuery, subquery)
					SubLinks = fmt.Sprintf("%s  %s", SubLinks, subtablelink)
				}
			}
		} else if k == "links" {
			if itemList, ok := v.([]interface{}); ok {
				for _, link := range itemList {
					if Links == "" {
						Links = link.(string)
					} else {
						Links = fmt.Sprintf("%s AND %s", Links, link.(string))
					}
				}
			}
		}
	}

	if SubQuery != "" {
		Query = fmt.Sprintf("%s %s,", Query, SubQuery)
	}

	Query = strings.TrimRight(Query, ",")

	if Links != "" && TableLinks != "" {
		TableLinks = fmt.Sprintf("%s INNER JOIN %s ON %s", TableLinks, tablename, Links)
	} else if Links != "" && TableLinks == "" {
		TableLinks = fmt.Sprintf(" INNER JOIN %s ON %s", tablename, Links)
	}

	if SubLinks != "" {
		TableLinks = fmt.Sprintf("%s  %s", TableLinks, SubLinks)
	}
	/*
		if markasJson {
			if Links != "" {
				Query = fmt.Sprintf("SELECT %s from %s where %s", Query, tablename, Links)
			} else {
				Query = fmt.Sprintf("SELECT %s from %s", Query, tablename)
			}

			if dbconn.DatabaseType == "sqlserver" {
				Query = fmt.Sprintf("%s FOR JSON PATH", Query)
			} else if dbconn.DatabaseType == "mysql" {
				Query = fmt.Sprintf("SELECT json_agg(t)  FROM ( %s ) t ", Query)
			}
			Query = fmt.Sprintf("(%s ) as \"%s\"", Query, tablename)
		}
	*/
	iLog.Debug(fmt.Sprintf("getsubtabls Query: %s", Query))
	return Query, TableLinks, nil

}
func (db *DBController) getsubtabls(tablename string, data map[string]interface{}, markasJson bool) (string, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDataFromTable"}
	/*
		"t1": {
							"fields": ["field1", "field2", "field3"],
							"subtables": {
								"t2":{
									"fields": ["field1", "field2", "field3"]   / fields can be empty for link table
									"links": ["t1.field2 = t2.field1"]
									"subtables": {
										"t3":{
											"fields": ["field1", "field2", "field3"]
											"links": ["t2.field2 = t3.field1"]
										}
								},

							}
						},
	*/

	Links := ""
	Query := " "
	for k, v := range data {
		if k == "fields" {
			if itemList, ok := v.([]interface{}); ok {
				for _, field := range itemList {
					Query = fmt.Sprintf("%s %s.%s,", Query, tablename, field.(string))
				}
			}
		} else if k == "subtables" {
			if item, ok := v.(map[string]interface{}); ok {
				for key, value := range item {
					subquery, err := db.getsubtabls(key, value.(map[string]interface{}), true)
					if err != nil {
						return "", err
					}
					Query = fmt.Sprintf("%s %s,", Query, subquery)

				}
			}
		} else if k == "links" {
			if itemList, ok := v.([]interface{}); ok {
				for _, link := range itemList {
					if Links == "" {
						Links = link.(string)
					} else {
						Links = fmt.Sprintf("%s AND %s", Links, link.(string))
					}
				}
			}
		}
	}

	Query = strings.TrimRight(Query, ",")

	if markasJson {
		if Links != "" {
			Query = fmt.Sprintf("SELECT %s from %s where %s", Query, tablename, Links)
		} else {
			Query = fmt.Sprintf("SELECT %s from %s", Query, tablename)
		}

		if dbconn.DatabaseType == "sqlserver" {
			Query = fmt.Sprintf("%s FOR JSON PATH", Query)
		} else if dbconn.DatabaseType == "mysql" {
			Query = fmt.Sprintf("SELECT json_agg(t)  FROM ( %s ) t ", Query)
		}
		Query = fmt.Sprintf("(%s ) as \"%s\"", Query, tablename)
	}

	iLog.Debug(fmt.Sprintf("getsubtabls Query: %s", Query))
	return Query, nil
}

func (db *DBController) InsertDataToTable(ctx *gin.Context) error {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "InsertDataToTables"}
	iLog.Debug(fmt.Sprintf("Insert data to table"))
	data, err := db.GetDataFromRequest(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Insert data to table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	if data.TableName == "" {
		iLog.Error(fmt.Sprintf("Insert data to table error: %s", "Table name is empty"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Table name is empty"})
		return err
	}

	iLog.Debug(fmt.Sprintf("Insert data to table: %s", data.TableName))

	fields := []string{}
	values := []string{}
	datatype := []int{}
	for key, value := range data.Data {

		iLog.Debug(fmt.Sprintf("Insert data to table: %s %s %s", key, value, reflect.TypeOf(value)))
		if value != nil {
			fields = append(fields, key)

			switch value.(type) {
			case string:
				datatype = append(datatype, 0)
				values = append(values, value.(string))
			case float64:
				datatype = append(datatype, 2)
				v := fmt.Sprintf("%f", value.(float64))
				values = append(values, v)
			case bool:
				datatype = append(datatype, 3)
				v := fmt.Sprintf("%t", value.(bool))
				values = append(values, v)
			case int:
				datatype = append(datatype, 1)
				v := fmt.Sprintf("%d", value.(int))
				values = append(values, v)
			default:
				datatype = append(datatype, 0)
				values = append(values, value.(string))
			}
		}
	}

	if len(fields) == 0 {
		iLog.Error(fmt.Sprintf("Insert data to table error: %s", "No data to insert"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No data to insert"})
		return err
	}

	id, err := dbconn.NewDBOperation("system", nil, "Execute dtable insert").TableInsert(data.TableName, fields, values)

	if err != nil {
		iLog.Error(fmt.Sprintf("Insert data to table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	result := `{"id":` + fmt.Sprintf("%d", id) + `}`
	ctx.JSON(http.StatusOK, gin.H{"data": result})

	return nil
}

func (db *DBController) UpdateDataToTable(ctx *gin.Context) error {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UpdateDataToTables"}
	iLog.Debug(fmt.Sprintf("Update data to table"))

	data, err := db.GetDataFromRequest(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Update data to table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	if data.TableName == "" {
		iLog.Error(fmt.Sprintf("Update data to table error: %s", "Table name is empty"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Table name is empty"})
		return err
	}
	iLog.Debug(fmt.Sprintf("Update data to table: %s", data.TableName))
	fields := []string{}
	values := []string{}
	datatype := []int{}
	for key, value := range data.Data {
		iLog.Debug(fmt.Sprintf("Update data to table: %s %s %s", key, value, reflect.TypeOf(value)))
		if value != nil {
			fields = append(fields, key)

			switch value.(type) {
			case string:
				datatype = append(datatype, 0)
				values = append(values, value.(string))
			case float64:
				datatype = append(datatype, 2)
				v := fmt.Sprintf("%f", value.(float64))
				values = append(values, v)
			case bool:
				datatype = append(datatype, 3)
				v := fmt.Sprintf("%t", value.(bool))
				values = append(values, v)
			case int:
				datatype = append(datatype, 1)
				v := fmt.Sprintf("%d", value.(int))
				values = append(values, v)
			default:
				datatype = append(datatype, 0)
				values = append(values, value.(string))
			}
		}
	}

	if len(fields) == 0 {
		iLog.Error(fmt.Sprintf("Update data to table error: %s", "No data to update"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No data to update"})
		return err
	}

	Wherestr := ""
	iLog.Debug(fmt.Sprintf("get where condition: %s", data.Where))
	for key, value := range data.Where {
		iLog.Debug(fmt.Sprintf("get where condition: %s %s", key, value))
		if value == "" {
			Wherestr = fmt.Sprintf("%s %s ", Wherestr, key)
		} else {
			Wherestr = fmt.Sprintf("%s %s='%s'", Wherestr, key, value)
		}
	}

	if Wherestr == "" {
		iLog.Error(fmt.Sprintf("Update data to table error: %s", "No where condition"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No where condition"})
		return err
	}

	rowcount, err := dbconn.NewDBOperation("system", nil, "Execute dtable update").TableUpdate(data.TableName, fields, values, datatype, Wherestr)

	if err != nil {
		iLog.Error(fmt.Sprintf("Update data to table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	result := `{"rowcount":` + fmt.Sprintf("%d", rowcount) + `}`

	ctx.JSON(http.StatusOK, gin.H{"data": result})
	return nil
}

func (db *DBController) DeleteDataFromTable(ctx *gin.Context) error {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "DeleteDataFromTable"}
	iLog.Debug(fmt.Sprintf("Delete data to table"))

	data, err := db.GetDataFromRequest(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Delete data to table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	if data.TableName == "" {
		iLog.Error(fmt.Sprintf("Delete data to table error: %s", "Table name is empty"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Table name is empty"})
		return err
	}
	iLog.Debug(fmt.Sprintf("Delete data to table: %s", data.TableName))

	Wherestr := ""
	iLog.Debug(fmt.Sprintf("get where condition: %s", data.Where))
	for key, value := range data.Where {
		iLog.Debug(fmt.Sprintf("get where condition: %s %s", key, value))
		if value == "" {
			Wherestr = fmt.Sprintf("%s %s ", Wherestr, key)
		} else {
			Wherestr = fmt.Sprintf("%s %s='%s'", Wherestr, key, value)
		}
	}

	if Wherestr == "" {
		iLog.Error(fmt.Sprintf("Delete data to table error: %s", "No where condition"))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No where condition"})
		return err
	}

	rowcount, err := dbconn.NewDBOperation("system", nil, "Execute dtable delete").TableDelete(data.TableName, Wherestr)

	if err != nil {
		iLog.Error(fmt.Sprintf("Delete data to table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	result := `{"rowcount":` + fmt.Sprintf("%d", rowcount) + `}`
	ctx.JSON(http.StatusOK, gin.H{"data": result})

	return nil

}

func (db *DBController) GetDataFromRequest(ctx *gin.Context) (DBData, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDataFromRequest"}
	iLog.Debug(fmt.Sprintf("GetDataFromRequest"))

	var data DBData
	body, err := ioutil.ReadAll(ctx.Request.Body)
	iLog.Debug(fmt.Sprintf("GetDataFromRequest body: %s", body))
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
		return data, err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest Unmarshal error: %s", err.Error()))
		return data, err
	}
	iLog.Debug(fmt.Sprintf("GetDataFromRequest data: %s", data))
	return data, nil
}
