package lngcodes

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	config "github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/controllers/common"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/framework/auth"
	"github.com/mdaxf/iac/logger"
)

type LCController struct {
}

type LCData struct {
	IDs       []int    `json:"ids"`
	Lngcodes  []string `json:"lngcodes"`
	Texts     []string `json:"texts"`
	Shorts    []string `json:"shorts"`
	Languages []string `json:"languages"`
	Language  string   `json:"language"`
}

func (f *LCController) GetLngCodes(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "LngCodes"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.lngcodes.GetLngCodes", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("GetLngCodes error: %s", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	_, LoginName, _, err := auth.GetUserInformation(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get LngCode error: %s", err.Error()))
	}
	if LoginName == "" {
		LoginName = "System"
	}

	iLog.User = LoginName

	iLog.Debug(fmt.Sprintf("Get LngCodes"))

	body, err := common.GetRequestBody(c)

	if err != nil {
		iLog.Error(fmt.Sprintf("Get LngCodes error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var lcdata LCData
	err = json.Unmarshal(body, &lcdata)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get LngCodes get the message body error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Get LngCodes lcdata: %s", lcdata))

	language := "en"
	if lcdata.Language == "" {
		language = "en"
	} else {
		language = lcdata.Language
	}

	if len(lcdata.Lngcodes) == 0 {
		iLog.Error(fmt.Sprintf("Get LngCodes error: %s", "lngcodes is empty"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "lngcodes is empty"})
		return
	}

	var wg sync.WaitGroup
	iLog.Debug(fmt.Sprintf("Get LngCodes autopopulate: %s, %d, %d", config.GlobalConfiguration.TranslationConfig["autopopulate"].(bool), len(lcdata.Texts), len(lcdata.Lngcodes)))
	if config.GlobalConfiguration.TranslationConfig["autopopulate"].(bool) && len(lcdata.Texts) > 0 && len(lcdata.Lngcodes) > 0 {
		wg.Add(1) // Increment the wait group counter
		go func() {
			defer wg.Done() // Decrement the wait group counter when the goroutine exits
			// Your task code goes here
			f.populatelngcodes(lcdata.Lngcodes, lcdata.Texts, language, LoginName)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Your second task code goes here
		querytemp := `SELECT lnc.id as id, lnc.name as lngcode, COALESCE(lc.mediumtext_,lc.shorttext,lnc.name) as text FROM lngcodes lnc 
		INNER JOIN lngcode_contents lc ON lc.lngcodeid = lnc.id
		WHERE lnc.name IN ('%s') AND lc.languageid = (SELECT id FROM languages WHERE name = '%s' LIMIT 1)`

		query := fmt.Sprintf(querytemp, strings.Join(lcdata.Lngcodes, "','"), language)
		//query := fmt.Sprintf("SELECT lngcode, text FROM language_codes Where language = '%s'", language)

		iLog.Debug(fmt.Sprintf("Get LngCodes query: %s", query))
		idbtx, err := dbconn.DB.Begin()
		if err != nil {
			iLog.Error(fmt.Sprintf("Get LngCodes error: %s", err.Error()))
			return
		}
		defer idbtx.Rollback()
		db := dbconn.NewDBOperation(LoginName, idbtx, logger.Framework)

		result, err := db.Query_Json(query)

		if err != nil {
			iLog.Error(fmt.Sprintf("Get LngCodes error: %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		iLog.Debug(fmt.Sprintf("Get LngCodes rows: %s", result))
		idbtx.Commit()
		c.JSON(http.StatusOK, gin.H{"data": result})
	}()

	wg.Wait()
}

func (f *LCController) UpdateLngCode(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "LngCodes"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.lngcodes.UpdateLngCode", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("UpdateLngCode error: %s", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	_, LoginName, _, err := auth.GetUserInformation(c)

	if err != nil {
		iLog.Error(fmt.Sprintf("Update LngCode error: %s", err.Error()))
	}
	if LoginName == "" {
		LoginName = "sys"
	}
	iLog.User = LoginName

	iLog.Debug(fmt.Sprintf("Update LngCode"))
	body, err := common.GetRequestBody(c)

	if err != nil {
		iLog.Error(fmt.Sprintf("Insert LngCode error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var lcdata LCData
	err = json.Unmarshal(body, &lcdata)
	if err != nil {
		iLog.Error(fmt.Sprintf("Insert LngCode get the message body error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("update LngCode lcdata: %s, %s", body, lcdata))
	idbtx, err := dbconn.DB.Begin()
	if err != nil {
		iLog.Error(fmt.Sprintf("populatelngcodes error: %s", err.Error()))
		return
	}
	defer idbtx.Rollback()
	db := dbconn.NewDBOperation(LoginName, idbtx, logger.Framework)
	for index, id := range lcdata.IDs {
		if id == 0 {
			err = f.insertlngcode(db, lcdata.Lngcodes[index], lcdata.Texts[index], lcdata.Languages[index], LoginName)
			if err != nil {
				iLog.Error(fmt.Sprintf("Insert LngCode error: %s", err.Error()))
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else {
			err = f.updatelngcodbyid(db, id, lcdata.Lngcodes[index], lcdata.Texts[index], lcdata.Languages[index], LoginName)
			if err != nil {
				iLog.Error(fmt.Sprintf("Insert LngCode error: %s", err.Error()))
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
	}
	idbtx.Commit()
	c.JSON(http.StatusOK, gin.H{"data": "success"})

}
func (f *LCController) insertlngcode(db *dbconn.DBOperation, lngcode string, text string, language string, User string) error {
	iLog := logger.Log{ModuleName: logger.API, User: User, ControllerName: "LngCodes"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.lngcodes.insertlngcode", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("insertlngcode error: %s", err))
		}
	}()
	currentTimeUTC := time.Now().UTC()

	// Format the time as a string in the MySQL date and time format
	formattedTime := currentTimeUTC.Format("2006-01-02 15:04:05")
	Columns := make([]string, 5)
	Values := make([]string, 5)

	n := 0
	Columns[n] = "name"
	Values[n] = lngcode

	n += 1
	Columns[n] = "UpdatedOn"
	Values[n] = formattedTime

	n += 1
	Columns[n] = "UpdatedBy"
	Values[n] = User

	n += 1
	Columns[n] = "CreatedOn"
	Values[n] = formattedTime

	n += 1
	Columns[n] = "CreatedBy"
	Values[n] = User

	iLog.Debug(fmt.Sprintf("insertlngcode: %s , %s", Columns, Values))
	lngcodeid, err := db.TableInsert("lngcodes", Columns, Values)
	if err != nil {
		iLog.Error(fmt.Sprintf("inert a new lngcode record error: %s", err.Error()))
		return err
	}

	if lngcodeid == 0 {
		iLog.Error(fmt.Sprintf("inert a new lngcode record error: %s", "lngcodeid is 0"))
		return fmt.Errorf("lngcodeid is 0")
	}

	querytext := fmt.Sprintf("Select id from languages where language='%s' limit 1", language)
	result, err := db.Query_Json(querytext)
	if err != nil {
		iLog.Error(fmt.Sprintf("get the languageid error: %s", err.Error()))
		return err
	}
	if result == nil || len(result) == 0 {
		iLog.Error(fmt.Sprintf("get the languageid error: %s", "no languageid found"))
		return fmt.Errorf("no languageid found")
	}
	languageid := int(result[0]["id"].(float64))

	if result[0]["id"] == nil {
		iLog.Error(fmt.Sprintf("get the languageid error: %s", "languageid is nil"))
		return fmt.Errorf("languageid is nil")
	}

	Columns = make([]string, 8)
	Values = make([]string, 8)

	n = 0
	Columns[n] = "shorttext"
	Values[n] = text

	n += 1
	Columns[n] = "mediumtext_"
	Values[n] = text

	n += 1
	Columns[n] = "lngcodeid"
	Values[n] = string(lngcodeid)

	n += 1
	Columns[n] = "languageid"
	Values[n] = string(languageid)

	n += 1
	Columns[n] = "UpdatedOn"
	Values[n] = formattedTime

	n += 1
	Columns[n] = "UpdatedBy"
	Values[n] = User

	n += 1
	Columns[n] = "CreatedOn"
	Values[n] = formattedTime

	n += 1
	Columns[n] = "CreatedBy"
	Values[n] = User

	iLog.Debug(fmt.Sprintf("insertlngcode: %s , %s", Columns, Values))
	_, err = db.TableInsert("lngcode_contents", Columns, Values)
	if err != nil {
		iLog.Error(fmt.Sprintf("inert a new lngcode record error: %s", err.Error()))
		return err
	}
	return nil
}

func (f *LCController) updatelngcodbyid(db *dbconn.DBOperation, id int, lngcode string, text string, language string, User string) error {
	iLog := logger.Log{ModuleName: logger.API, User: User, ControllerName: "LngCodes"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.lngcodes.updatelngcodbyid", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("insertlngcode error: %s", err))
		}
	}()

	// get the languageid
	querytext := fmt.Sprintf("Select id from languages where language='%s' limit 1", language)
	result, err := db.Query_Json(querytext)
	if err != nil {
		iLog.Error(fmt.Sprintf("get the languageid error: %s", err.Error()))
		return err
	}
	if result == nil || len(result) == 0 {
		iLog.Error(fmt.Sprintf("get the languageid error: %s", "no languageid found"))
		return fmt.Errorf("no languageid found")
	}

	if result[0]["id"] == nil {
		iLog.Error(fmt.Sprintf("get the languageid error: %s", "languageid is nil"))
		return fmt.Errorf("languageid is nil")
	}

	languageid := int(result[0]["id"].(float64))

	if languageid == 0 {
		iLog.Error(fmt.Sprintf("get the languageid error: %s", "languageid is 0"))
		return fmt.Errorf("languageid is 0")
	}

	currentTimeUTC := time.Now().UTC()

	// Format the time as a string in the MySQL date and time format
	formattedTime := currentTimeUTC.Format("2006-01-02 15:04:05")
	Columns := make([]string, 5)
	Values := make([]string, 5)
	datatypes := make([]int, 5)
	Where := ""
	n := 0
	Columns[n] = "text"
	Values[n] = text
	datatypes[n] = 1
	n += 1
	Columns[n] = "lngcode"
	Values[n] = lngcode
	datatypes[n] = 1
	n += 1
	Columns[n] = "languageid"
	Values[n] = string(languageid)
	datatypes[n] = 1
	n += 1
	Columns[n] = "UpdatedOn"
	Values[n] = formattedTime
	datatypes[n] = 2
	n += 1
	Columns[n] = "UpdatedBy"
	Values[n] = User
	datatypes[n] = 1

	Where = fmt.Sprintf("lngcodeid = '%d' AND langaugeid = %d ", id, languageid)

	_, err = db.TableUpdate("language_codes", Columns, Values, datatypes, Where)
	if err != nil {
		iLog.Error(fmt.Sprintf("update the lngcode error: %s", err.Error()))
		return err
	}
	return nil
}

func (f *LCController) populatesinglelngcodes(db *dbconn.DBOperation, lngcode string, text string, language string, User string) {
	iLog := logger.Log{ModuleName: logger.API, User: User, ControllerName: "LngCodes"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.lngcodes.populatesinglelngcodes", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("populatesinglelngcodes error: %s", err))
		}
	}()
	iLog.Debug(fmt.Sprintf("populatesinglelngcodes"))

	if lngcode == "" {
		iLog.Error(fmt.Sprintf("populatelngcodes error: %s", "lngcodes is empty"))
		return
	}

	if language == "" {
		iLog.Error(fmt.Sprintf("populatelngcodes error: %s", "language is empty"))
		return
	}

	query := fmt.Sprintf("SELECT lngcode, text FROM language_codes WHERE lngcode = '%s' AND language = '%s'", lngcode, language)
	iLog.Debug(fmt.Sprintf("populatesinglelngcodes query: %s", query))

	result, err := db.Query_Json(query)
	if err != nil {
		iLog.Error(fmt.Sprintf("populatesinglelngcodes error: %s", err.Error()))
		return
	}
	iLog.Debug(fmt.Sprintf("populatesinglelngcodes rows: %s", result))

	var existlngcode string
	var existtext string

	update := false
	insert := true
	if result != nil {

		if len(result) > 0 {
			insert = false
			existlngcode = result[0]["lngcode"].(string)
			existtext = result[0]["text"].(string)
			if existlngcode == lngcode && existtext == text {
				update = false
			} else {
				update = true
			}
		}
	} else {
		insert = true
	}

	iLog.Debug(fmt.Sprintf("populatesinglelngcodes update: %s insert: %s", update, insert))
	currentTimeUTC := time.Now().UTC()

	// Format the time as a string in the MySQL date and time format
	formattedTime := currentTimeUTC.Format("2006-01-02 15:04:05")
	if update {
		return
		/*	Columns := make([]string, 3)
			Values := make([]string, 3)
			datatypes := make([]int, 3)
			Where := ""
			n := 0
			Columns[n] = "text"
			Values[n] = text
			datatypes[n] = 1

			n += 1
			Columns[n] = "UpdatedOn"
			Values[n] = formattedTime
			datatypes[n] = 2
			n += 1
			Columns[n] = "UpdatedBy"
			Values[n] = User
			datatypes[n] = 1

			Where = fmt.Sprintf("lngcode = '%s' ANG language ='%s'", lngcode, language)

			_, err := db.TableUpdate("language_codes", Columns, Values, datatypes, Where)
			if err != nil {
				iLog.Error(fmt.Sprintf("populatelngcodes error: %s", err.Error()))
				return
			} */
	} else if insert {

		Columns := make([]string, 5)
		Values := make([]string, 5)
		datatypes := make([]int, 5)
		n := 0
		Columns[n] = "lngcode"
		Values[n] = lngcode
		datatypes[n] = 1
		n += 1
		Columns[n] = "text"
		Values[n] = text
		datatypes[n] = 1
		/*	n += 1
			Columns[n] = "short"
			Values[n] = short
			datatypes[n] = 1  */
		n += 1
		Columns[n] = "language"
		Values[n] = language
		datatypes[n] = 1
		n += 1
		Columns[n] = "CreatedOn"
		Values[n] = formattedTime
		datatypes[n] = 2
		n += 1
		Columns[n] = "CreatedBy"
		Values[n] = User
		datatypes[n] = 1
		n += 1

		_, err := db.TableInsert("language_codes", Columns, Values)
		if err != nil {
			iLog.Error(fmt.Sprintf("populatelngcodes error: %s", err.Error()))
			return
		}

	}

}

func (f *LCController) populatelngcodes(lngcodes []string, text []string, language string, User string) {
	iLog := logger.Log{ModuleName: logger.API, User: User, ControllerName: "LngCodes"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.lngcodes.populatelngcodes", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("populatelngcodes error: %s", err))
		}
	}()
	iLog.Debug(fmt.Sprintf("populatelngcodes"))

	if len(lngcodes) == 0 {
		iLog.Error(fmt.Sprintf("populatelngcodes error: %s", "lngcodes is empty"))
		return
	}

	if len(text) == 0 {
		iLog.Error(fmt.Sprintf("populatelngcodes error: %s", "text is empty"))
		return
	}
	/*
		if len(short) == 0 {
			iLog.Error(fmt.Sprintf("populatelngcodes error: %s", "short is empty"))
			return
		} */

	if language == "" {
		iLog.Error(fmt.Sprintf("populatelngcodes error: %s", "language is empty"))
		return
	}

	idbtx, err := dbconn.DB.Begin()
	if err != nil {
		iLog.Error(fmt.Sprintf("populatelngcodes error: %s", err.Error()))
		return
	}
	defer idbtx.Rollback()
	db := dbconn.NewDBOperation(User, idbtx, logger.Framework)
	for i := 0; i < len(lngcodes); i++ {
		iLog.Debug(fmt.Sprintf("populatelngcodes lngcode: %s %s %s %s", lngcodes[i], text[i], language, User))
		f.populatesinglelngcodes(db, lngcodes[i], text[i], language, User)

	}
	idbtx.Commit()
}
