// Package databaseop provides database operation controllers with automatic
// database type support using the factory pattern and dialect system.
//
// Key improvements:
// - Automatic support for multiple database types (MySQL, PostgreSQL, MSSQL, Oracle)
// - Database-specific SQL dialect handling (placeholders, LIMIT/OFFSET, JSON operations)
// - Removes hardcoded database type checks for better maintainability
// - Uses the new DatabaseFactory and Dialect interfaces for database portability
package databaseop

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mdaxf/iac/controllers/common"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/logger"
)

type DBController struct {
}

type TabulatorFilter struct {
	Field string      `json:"field"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"` // can be string, number, bool
}

type TabSorter struct {
	Field string `json:"field"`
	Dir   string `json:"dir"`
}

type DBData struct {
	TableName  string                 `json."tablename"` // table name
	Data       map[string]interface{} `json."data"`
	Operation  string                 `json."operation"` // insert, update, delete
	Keys       []string               `json."keys"`      // keys for update and delete
	Where      map[string]interface{} `json."where"`     // where args for update and delete
	NullValues map[string]interface{} `json."nullvalues"`
	QueryStr   string                 `json."querystr"` // query string for query
	Cursor     *int                   `json:"cursor"`   // last seen id, null for first
	Offset     int                    `json:"offset"`
	Limit      int                    `json:"limit"`
	Sorters    []TabSorter            `json:"sorters"`
	Filters    []TabulatorFilter      `json:"filters"`
}

type QueryInput struct {
	QueryStr string            `json."querystr"` // query string for query
	Offset   int               `json:"offset"`
	Cursor   *int              `json:"cursor"` // last seen id, null for first
	Limit    int               `json:"limit"`
	Sorters  []TabSorter       `json:"sorters"`
	Filters  []TabulatorFilter `json:"filters"`
}

// GetDatabyQuery retrieves data from the database based on the provided query.
// It expects a JSON request body containing the query information.
// The function returns the retrieved data as a JSON response.

func (db *DBController) GetDatabyQuery(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDatabyQuery"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.databaseop.GetDatabyQuery", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("Get data by query error: %s", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()

		_, user, clientid, err := common.GetRequestUser(ctx)
		if err != nil {
			iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		iLog.ClientID = clientid
		iLog.User = user
		iLog.Debug(fmt.Sprintf("Get data by query"))

		var data QueryInput
		body, err := common.GetRequestBody(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	*/
	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var data QueryInput

	iLog.Debug(fmt.Sprintf("GetDatabyQuery from respository with body: %s", body))

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest Unmarshal error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("GetDataFromRequest data: %s", data))

	if err != nil {
		iLog.Error(fmt.Sprintf("Get data by query error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get data by query: %s", data.QueryStr))
	Query := data.QueryStr
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

// the progressive retrieve data for the data list or tabulor list
func (db *DBController) GetDatabyQueryForTabulor(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDatabyQueryForTabulor"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.databaseop.GetDatabyQueryForTabulor", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("Get data by query error: %s", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()

		_, user, clientid, err := common.GetRequestUser(ctx)
		if err != nil {
			iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		iLog.ClientID = clientid
		iLog.User = user
		iLog.Debug(fmt.Sprintf("Get data by query"))

		var data QueryInput
		body, err := common.GetRequestBody(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	*/
	body, clientid, user, err := common.GetRequestBodyandUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var data QueryInput

	iLog.Debug(fmt.Sprintf("GetDatabyQuery from respository with body: %s", body))

	err = json.Unmarshal(body, &data)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest Unmarshal error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("GetDataFromRequest data: %s", data))

	if err != nil {
		iLog.Error(fmt.Sprintf("Get data by query error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get data by query: %s", data.QueryStr))
	Query := data.QueryStr

	// additionaal code for the Tabulor progressive retrieve data

	args := []interface{}{}
	where, args := buildWherePrepared(data.Filters, args) // reuse from prepared version
	sqlStr := Query + where

	// Keyset: assume ordering by id ASC (adapt if other column)
	// If sorters provided and they use id, we can still keyset on id
	sqlStr += " "
	if len(data.Sorters) > 0 {

		for i, sort := range data.Sorters {
			field := sort.Field
			dir := strings.ToUpper(sort.Dir)
			if dir != "ASC" && dir != "DESC" {
				dir = "ASC" // fallback
			}

			// build ORDER BY clause
			if i == 0 {
				sqlStr += fmt.Sprintf(" ORDER BY %s %s", field, dir)
			} else {
				sqlStr += fmt.Sprintf(", %s %s", field, dir)
			}
		}
	}

	if data.Limit <= 0 {
		data.Limit = 1000
	}

	// Get total count for pagination (before adding LIMIT/OFFSET)
	countSQL := "SELECT COUNT(*) as total FROM (" + Query + where + ") as count_query"
	countResult, err := dbconn.NewDBOperation("system", nil, "Execute Count Query").Query_Json(countSQL, args...)

	var totalCount int64 = 0
	if err == nil && len(countResult) > 0 {
		if totalVal, ok := countResult[0]["total"].(float64); ok {
			totalCount = int64(totalVal)
		} else if totalVal, ok := countResult[0]["total"].(int64); ok {
			totalCount = totalVal
		} else if totalVal, ok := countResult[0]["total"].(int); ok {
			totalCount = int64(totalVal)
		}
	}

	// Calculate last_page for Tabulator progressive loading
	var lastPage int64 = 1
	if data.Limit > 0 {
		lastPage = (totalCount + int64(data.Limit) - 1) / int64(data.Limit) // Ceiling division
	}

	// Use dialect-specific LIMIT/OFFSET syntax
	dialect, err := dbconn.GetFactory().GetDialect(dbconn.DBType(dbconn.DatabaseType))
	if err != nil {
		dialect = dbconn.NewMySQLDialect()
	}

	// Add database-specific pagination
	sqlStr += " " + dialect.LimitOffset(data.Limit, data.Offset)
	// Note: Some dialects use placeholders in LimitOffset, others use direct values
	// The current implementation uses direct values in the LimitOffset method

	// get data from database
	result, err := dbconn.NewDBOperation("system", nil, "Execute Query Function").Query_Json(sqlStr, args...)

	if err != nil {
		iLog.Error(fmt.Sprintf("Get data from table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get data from table result: %s", gin.H{"data": result}))

	// Return data with pagination metadata for Tabulator progressive loading
	ctx.JSON(http.StatusOK, gin.H{
		"data":      result,
		"last_page": lastPage,
		"total":     totalCount,
		"page":      (data.Offset / data.Limit) + 1,
		"size":      data.Limit,
	})
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

// GetDataFromTables retrieves data from tables based on the request parameters.
// It first extracts the user and client ID from the request context.
// Then it calls GetDataFromRequest to get the data structure from the request body.
// It constructs a query based on the data structure, user, client ID, and where conditions.
// Finally, it executes the query and returns the result as JSON.

func (db *DBController) GetDataFromTables(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDataFromTable"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.databaseop.GetDataFromTables", elapsed)
	}()

	/*	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Get data from tables error: %s", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()  */
	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	iLog.ClientID = clientid
	iLog.User = user
	iLog.Debug(fmt.Sprintf("Get data from table"))

	data, err := db.GetDataFromRequest(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get data from table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get data from table: %s", data.TableName))

	Query, TableNames, err := db.getDataStructForQuery(data.Data, user, clientid)
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

func (db *DBController) GetDataFromTablesForTabulor(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDataFromTable"}
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.databaseop.GetDataFromTables", elapsed)
	}()

	/*	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Get data from tables error: %s", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()  */
	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	iLog.ClientID = clientid
	iLog.User = user
	iLog.Debug(fmt.Sprintf("Get data from table"))

	data, err := db.GetDataFromRequest(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get data from table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get data from table: %s", data.TableName))

	Query, TableNames, err := db.getDataStructForQuery(data.Data, user, clientid)
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

	// additionaal code for the Tabulor progressive retrieve data

	args := []interface{}{}
	where, args := buildWherePrepared(data.Filters, args) // reuse from prepared version

	sqlStr := Query
	if Wherestr == "" {
		sqlStr += " WHERE 1=1 " + where
	} else {
		sqlStr += where
	}

	// Keyset: assume ordering by id ASC (adapt if other column)
	// If sorters provided and they use id, we can still keyset on id
	sqlStr += " "
	if len(data.Sorters) > 0 {

		for i, sort := range data.Sorters {
			field := sort.Field
			dir := strings.ToUpper(sort.Dir)
			if dir != "ASC" && dir != "DESC" {
				dir = "ASC" // fallback
			}

			// build ORDER BY clause
			if i == 0 {
				sqlStr += fmt.Sprintf(" ORDER BY %s %s", field, dir)
			} else {
				sqlStr += fmt.Sprintf(", %s %s", field, dir)
			}
		}
	}

	if data.Limit <= 0 {
		data.Limit = 1000
	}

	// Use dialect-specific LIMIT/OFFSET syntax
	dialect, err := dbconn.GetFactory().GetDialect(dbconn.DBType(dbconn.DatabaseType))
	if err != nil {
		dialect = dbconn.NewMySQLDialect()
	}

	// Add database-specific pagination
	sqlStr += " " + dialect.LimitOffset(data.Limit, data.Offset)

	result, err := dbconn.NewDBOperation("system", nil, "Execute Query Function").Query_Json(sqlStr, args...)

	if err != nil {
		iLog.Error(fmt.Sprintf("Get data from table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.Debug(fmt.Sprintf("Get data from table result: %s", gin.H{"data": result}))
	//jsondata, err := json.Marshal(result)

	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// getDataStructForQuery retrieves the data structure and table name for a given query.
// It takes a map of data, user string, and client ID string as input parameters.
// It returns the query string, table name string, and an error if any.

func (db *DBController) getDataStructForQuery(data map[string]interface{}, user string, clientid string) (string, string, error) {
	iLog := logger.Log{ModuleName: logger.API, User: user, ClientID: clientid, ControllerName: "GetDataFromTable"}
	/*	startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			iLog.PerformanceWithDuration("controllers.databaseop.getDataStructForQuery", elapsed)
		}()

		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("Get data struct for query error: %s", err))
			}
		}()
	*/
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

				subquery, tablelinks, err := db.getmysqlsubtabls(tablename, item, false, user, clientid)
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

// getmysqlsubtabls is a function that retrieves data from a MySQL table and its subtables.
// It takes the following parameters:
// - tablename: the name of the table to retrieve data from.
// - data: a map containing additional data for the query, such as fields, subtables, and links.
// - markasJson: a boolean indicating whether the result should be marked as JSON.
// - user: the user performing the operation.
// - clientid: the client ID associated with the operation.
// The function returns the query string, table links, and an error (if any).

func (db *DBController) getmysqlsubtabls(tablename string, data map[string]interface{}, markasJson bool, user string, clientid string) (string, string, error) {
	iLog := logger.Log{ModuleName: logger.API, User: user, ClientID: clientid, ControllerName: "GetDataFromTable"}
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
					subquery, subtablelink, err := db.getmysqlsubtabls(key, value.(map[string]interface{}), true, user, clientid)
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
		// Note: JSON aggregation now uses database-specific dialects
		if markasJson {
			if Links != "" {
				Query = fmt.Sprintf("SELECT %s from %s where %s", Query, tablename, Links)
			} else {
				Query = fmt.Sprintf("SELECT %s from %s", Query, tablename)
			}

			// Use dialect-specific JSON aggregation
			dialect, _ := dbconn.GetFactory().GetDialect(dbconn.DBType(dbconn.DatabaseType))
			if dialect != nil && dialect.SupportsJSON() {
				switch dbconn.DBType(dbconn.DatabaseType) {
				case dbconn.DBTypeMSSQL:
					Query = fmt.Sprintf("%s FOR JSON PATH", Query)
				case dbconn.DBTypeMySQL:
					Query = fmt.Sprintf("SELECT JSON_ARRAYAGG(JSON_OBJECT(*)) FROM ( %s ) t", Query)
				case dbconn.DBTypePostgreSQL:
					Query = fmt.Sprintf("SELECT json_agg(t) FROM ( %s ) t", Query)
				case dbconn.DBTypeOracle:
					Query = fmt.Sprintf("SELECT JSON_ARRAYAGG(JSON_OBJECT(*)) FROM ( %s ) t", Query)
				}
			}
			Query = fmt.Sprintf("(%s ) as \"%s\"", Query, tablename)
		}
	*/
	iLog.Debug(fmt.Sprintf("getsubtabls Query: %s", Query))
	return Query, TableLinks, nil

}

// getsubtabls is a recursive function that generates a SQL query for retrieving data from a table and its subtables.
// It takes the table name, data map, markasJson flag, user, and clientid as parameters.
// The function iterates over the data map and constructs the query based on the fields, subtables, and links specified.
// If markasJson is true, the query is formatted to return the result as JSON.
// The function returns the generated SQL query and any error encountered during the process.

func (db *DBController) getsubtabls(tablename string, data map[string]interface{}, markasJson bool, user string, clientid string) (string, error) {
	iLog := logger.Log{ModuleName: logger.API, User: user, ClientID: clientid, ControllerName: "GetDataFromTable"}
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
					subquery, err := db.getsubtabls(key, value.(map[string]interface{}), true, user, clientid)
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

	// Use dialect-specific JSON aggregation for database portability
	if markasJson {
		if Links != "" {
			Query = fmt.Sprintf("SELECT %s from %s where %s", Query, tablename, Links)
		} else {
			Query = fmt.Sprintf("SELECT %s from %s", Query, tablename)
		}

		// Automatically detect database type and use appropriate JSON syntax
		dialect, _ := dbconn.GetFactory().GetDialect(dbconn.DBType(dbconn.DatabaseType))
		if dialect != nil && dialect.SupportsJSON() {
			switch dbconn.DBType(dbconn.DatabaseType) {
			case dbconn.DBTypeMSSQL:
				Query = fmt.Sprintf("%s FOR JSON PATH", Query)
			case dbconn.DBTypeMySQL:
				Query = fmt.Sprintf("SELECT JSON_ARRAYAGG(JSON_OBJECT(*)) FROM ( %s ) t", Query)
			case dbconn.DBTypePostgreSQL:
				Query = fmt.Sprintf("SELECT json_agg(t) FROM ( %s ) t", Query)
			case dbconn.DBTypeOracle:
				Query = fmt.Sprintf("SELECT JSON_ARRAYAGG(JSON_OBJECT(*)) FROM ( %s ) t", Query)
			}
		}
		Query = fmt.Sprintf("(%s ) as \"%s\"", Query, tablename)
	}

	iLog.Debug(fmt.Sprintf("getsubtabls Query: %s", Query))
	return Query, nil
}

// InsertDataToTable inserts data into a table.
// It retrieves the data from the request context and validates it.
// If the table name is empty or there is no data to insert, it returns an error.
// Otherwise, it performs the table insert operation and returns the ID of the inserted data.

func (db *DBController) InsertDataToTable(ctx *gin.Context) error {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "InsertDataToTables"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.databaseop.InsertDataToTable", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("InsertDataToTable error: %s", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	iLog.ClientID = clientid
	iLog.User = user
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

	nullvalues := data.NullValues
	fields := []string{}
	values := []string{}
	datatype := []int{}
	for key, value := range data.Data {

		iLog.Debug(fmt.Sprintf("Insert data to table: %s %s %s", key, value, reflect.TypeOf(value)))
		if value != nil {
			if nullvalues != nil {
				if nullvalue, ok := nullvalues[key]; ok {
					if value == nullvalue {
						continue
					}
				}
			}

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

// UpdateDataToTable updates data in a table based on the request received.
// It retrieves the data from the request, validates it, and performs the update operation.
// If any errors occur during the process, it returns the error and sends an appropriate response to the client.
// The function also logs the performance duration of the operation.

func (db *DBController) UpdateDataToTable(ctx *gin.Context) error {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "UpdateDataToTables"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.databaseop.UpdateDataToTable", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("UpdateDataToTable error: %s", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()  */
	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	iLog.ClientID = clientid
	iLog.User = user
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
	nullvalues := data.NullValues
	fields := []string{}
	values := []interface{}{}
	datatype := []int{}
	for key, value := range data.Data {
		iLog.Debug(fmt.Sprintf("Update data to table: %s %s %s", key, value, reflect.TypeOf(value)))
		if value != nil {
			if nullvalues != nil {
				if nullvalue, ok := nullvalues[key]; ok {
					if value == nullvalue {
						fields = append(fields, key)
						values = append(values, nil)
						datatype = append(datatype, -1) // -1 = null type
						continue
					}
				}
			}
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
		} else {
			fields = append(fields, key)
			values = append(values, nil)
			datatype = append(datatype, -1) // -1 = null type
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

	rowcount, err := dbconn.NewDBOperation("system", nil, "Execute dtable update").TableUpdate_v2(data.TableName, fields, values, datatype, Wherestr)

	if err != nil {
		iLog.Error(fmt.Sprintf("Update data to table error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	result := `{"rowcount":` + fmt.Sprintf("%d", rowcount) + `}`

	ctx.JSON(http.StatusOK, gin.H{"data": result})
	return nil
}

// DeleteDataFromTable deletes data from a table based on the provided conditions.
// It takes a gin.Context as input and returns an error if any.
// The function retrieves the user and client ID from the request context and logs the operation.
// It then gets the data from the request and checks if the table name is empty.
// If the table name is empty, it returns an error.
// Otherwise, it constructs the WHERE condition based on the provided data and deletes the matching rows from the table.
// The function returns the number of rows deleted as a JSON response.

func (db *DBController) DeleteDataFromTable(ctx *gin.Context) error {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "DeleteDataFromTable"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.databaseop.DeleteDataFromTable", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("DeleteDataFromTable error: %s", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()  */
	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	iLog.ClientID = clientid
	iLog.User = user
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

// GetDataFromRequest retrieves data from the request body and returns it as a DBData struct.
// It also logs the performance duration of the function.
// If there is an error during the process, it logs the error and returns an empty DBData struct.

func (db *DBController) GetDataFromRequest(ctx *gin.Context) (DBData, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "GetDataFromRequest"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.databaseop.GetDataFromRequest", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("GetDataFromRequest error: %s", err.Error()))
		return DBData{}, err
	}

	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug(fmt.Sprintf("GetDataFromRequest"))

	var data DBData
	body, err := common.GetRequestBody(ctx)
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

// buildWherePrepared builds WHERE clause with database-specific placeholders
// Uses the new database model to automatically support different database types
func buildWherePrepared(filters []TabulatorFilter, args []interface{}) (string, []interface{}) {
	// Get dialect from the global factory for placeholder generation
	dialect, err := dbconn.GetFactory().GetDialect(dbconn.DBType(dbconn.DatabaseType))
	if err != nil {
		// Fallback to MySQL-style placeholders if dialect not found
		dialect = dbconn.NewMySQLDialect()
	}

	clauses := []string{}
	paramIndex := len(args) + 1 // Start from the next parameter index

	for _, f := range filters {
		if f.Field == "" {
			continue
		}

		placeholder := dialect.Placeholder(paramIndex)
		quotedField := dialect.QuoteIdentifier(f.Field)

		switch strings.ToLower(f.Type) {
		case "like":
			clauses = append(clauses, fmt.Sprintf("%s LIKE %s", quotedField, placeholder))
			args = append(args, "%"+ConvertInterfaceToString(f.Value)+"%")
			paramIndex++
		case "starts":
			clauses = append(clauses, fmt.Sprintf("%s LIKE %s", quotedField, placeholder))
			args = append(args, ConvertInterfaceToString(f.Value)+"%")
			paramIndex++
		case "ends":
			clauses = append(clauses, fmt.Sprintf("%s LIKE %s", quotedField, placeholder))
			args = append(args, "%"+ConvertInterfaceToString(f.Value))
			paramIndex++
		default:
			clauses = append(clauses, fmt.Sprintf("%s = %s", quotedField, placeholder))
			args = append(args, f.Value)
			paramIndex++
		}
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " AND " + strings.Join(clauses, " AND "), args
}

func ConvertInterfaceToString(value interface{}) string {
	var valStr string

	switch v := value.(type) {
	case string:
		valStr = v
	case int:
		valStr = fmt.Sprintf("%d", v)
	case float64:
		valStr = fmt.Sprintf("%f", v)
	case bool:
		valStr = fmt.Sprintf("%t", v)
	default:
		valStr = fmt.Sprint(v) // fallback for other types
	}

	return valStr
}

func escape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
