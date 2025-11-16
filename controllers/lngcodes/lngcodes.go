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
	IDs         []int    `json:"ids"`
	Lngcodes    []string `json:"lngcodes"`
	Texts       []string `json:"texts"`
	Shorts      []string `json:"shorts"`
	Languages   []string `json:"languages"`
	Language    string   `json:"language"`
	Languageid  int64    `json:"languageid"`
	Lngcodeids  []int64  `json:"lngcodeids"`
	Languageids []int64  `json:"languageids"`
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

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var lcdata LCData
	err = json.Unmarshal(body, &lcdata)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get LngCodes get the message body error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Get LngCodes lcdata: %s", lcdata))
	/*
		language := "en"

		if lcdata.Language == "" {
			language = "en"
		} else {
			language = lcdata.Language
		} */
	languageid := int64(1)
	if lcdata.Languageid > 0 {
		languageid = lcdata.Languageid
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
			f.populatelngcodes(lcdata.Lngcodes, lcdata.Texts, languageid, user)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Your second task code goes here
		querytemp := `SELECT lnc.id as id, lnc.name as lngcode, COALESCE(lc.mediumtext_,lc.shorttext,lnc.name) as text FROM lngcodes lnc 
		INNER JOIN lngcode_contents lc ON lc.lngcodeid = lnc.id
		WHERE lnc.name IN ('%s') AND lc.languageid = '%i'`

		query := fmt.Sprintf(querytemp, strings.Join(lcdata.Lngcodes, "','"), languageid)
		//query := fmt.Sprintf("SELECT lngcode, text FROM language_codes Where language = '%s'", language)

		iLog.Debug(fmt.Sprintf("Get LngCodes query: %s", query))
		idbtx, err := dbconn.DB.Begin()
		if err != nil {
			iLog.Error(fmt.Sprintf("Get LngCodes error: %s", err.Error()))
			return
		}
		defer idbtx.Rollback()
		db := dbconn.NewDBOperation(user, idbtx, logger.Framework)

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

func (f *LCController) InsertLngCode(db *dbconn.DBOperation, lngcode string, User string) (int64, error) {

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
	Columns := make([]string, 3)
	Values := make([]string, 3)

	n := 0
	Columns[n] = "name"
	Values[n] = lngcode

	n += 1
	Columns[n] = "createdon"
	Values[n] = formattedTime

	n += 1
	Columns[n] = "createdby"
	Values[n] = User

	iLog.Debug(fmt.Sprintf("insertlngcode: %s , %s", Columns, Values))
	id, err := db.TableInsert("lngcodes", Columns, Values)
	if err != nil {
		iLog.Error(fmt.Sprintf("inert a new lngcode record error: %s", err.Error()))
		return 0, err
	}
	return id, nil

}

func (f *LCController) UpdateLngCode(c *gin.Context) {

}

func (f *LCController) UpdateLngContent(c *gin.Context) {
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
			lngcodeid := lcdata.Lngcodeids[index]
			if lngcodeid == 0 {
				lngcodeid, err = f.InsertLngCode(db, lcdata.Lngcodes[index], LoginName)

				if err != nil {
					iLog.Error(fmt.Sprintf("Create Lng Code error: %s", err.Error()))
				}
			}

			err = f.insertlngcontent(db, lngcodeid, lcdata.Texts[index], lcdata.Languageids[index], LoginName)
			if err != nil {
				iLog.Error(fmt.Sprintf("Insert LngCode error: %s", err.Error()))
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else {
			err = f.updatelngcontent(db, id, lcdata.Texts[index], lcdata.Languageids[index], LoginName)
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
func (f *LCController) insertlngcontent(db *dbconn.DBOperation, lngcodeid int64, text string, languageid int64, User string) error {
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
	Columns := make([]string, 8)
	Values := make([]string, 8)

	n := 0
	Columns[n] = "lngcodeid"
	Values[n] = fmt.Sprintf("%d", lngcodeid)

	n += 1
	Columns[n] = "languageid"
	Values[n] = fmt.Sprintf("%d", languageid)

	n += 1
	Columns[n] = "shorttext"
	Values[n] = text

	n += 1
	Columns[n] = "mediumtext_"
	Values[n] = text

	n += 1
	Columns[n] = "modifiedon"
	Values[n] = formattedTime

	n += 1
	Columns[n] = "modifiedby"
	Values[n] = User

	n += 1
	Columns[n] = "createdon"
	Values[n] = formattedTime

	n += 1
	Columns[n] = "createdby"
	Values[n] = User

	iLog.Debug(fmt.Sprintf("insertlngcode: %s , %s", Columns, Values))
	_, err := db.TableInsert("lngcode_contents", Columns, Values)
	if err != nil {
		iLog.Error(fmt.Sprintf("inert a new lngcode record error: %s", err.Error()))
		return err
	}
	return nil
}

func (f *LCController) updatelngcontent(db *dbconn.DBOperation, id int, text string, languageid int64, User string) error {
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

	currentTimeUTC := time.Now().UTC()

	// Format the time as a string in the MySQL date and time format
	formattedTime := currentTimeUTC.Format("2006-01-02 15:04:05")
	Columns := make([]string, 5)
	Values := make([]string, 5)
	datatypes := make([]int, 5)
	Where := ""
	n := 0

	Columns[n] = "shorttext"
	Values[n] = text
	datatypes[n] = 1

	n += 1
	Columns[n] = "mediumtext_"
	Values[n] = text
	datatypes[n] = 1

	n += 1
	Columns[n] = "languageid"
	Values[n] = fmt.Sprintf("%d", languageid)
	datatypes[n] = 1

	n += 1
	Columns[n] = "modifiedon"
	Values[n] = formattedTime
	datatypes[n] = 2

	n += 1
	Columns[n] = "modifiedby"
	Values[n] = User
	datatypes[n] = 1

	Where = fmt.Sprintf("id = '%d' ", id)

	_, err := db.TableUpdate("language_codes", Columns, Values, datatypes, Where)
	if err != nil {
		iLog.Error(fmt.Sprintf("update the lngcode error: %s", err.Error()))
		return err
	}
	return nil
}

func (f *LCController) populatesinglelngcodes(db *dbconn.DBOperation, lngcode string, text string, languageid int64, User string) {
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

	if languageid == 0 {
		iLog.Error(fmt.Sprintf("populatelngcodes error: %s", "language is empty"))
		return
	}

	query := fmt.Sprintf("SELECT lngcodes.name as lngcode, lngcodes.id as lngcodeid, lngcode_contents.id as lngcodecontentid, lngcode_contents.mediumtext_ as text FROM lngcodes LEFT JOIN lngcode_contents ON lngcode_contents.lngcodeid = lngcodes.id AND lngcode_contents.languageid ='%d' WHERE lngcodes.name = '%s' ", languageid, lngcode)
	iLog.Debug(fmt.Sprintf("populatesinglelngcodes query: %s", query))

	result, err := db.Query_Json(query)
	if err != nil {
		iLog.Error(fmt.Sprintf("populatesinglelngcodes error: %s", err.Error()))
		return
	}
	iLog.Debug(fmt.Sprintf("get lng code and conents rows: %s", result))

	var existlngcode string
	var existtext string
	lngcodeid := int64(0)
	lngcodecontentid := int64(0)
	update := false
	insertcontent := true
	inertlngcode := true
	updatecontent := false

	if result != nil {

		if len(result) > 0 {
			inertlngcode = false
			insertcontent = false
			existlngcode = result[0]["lngcode"].(string)
			existtext = result[0]["text"].(string)
			lngcodeid = result[0]["lngcodeid"].(int64)
			lngcodecontentid = result[0]["lngcodecontentid"].(int64)

			if existlngcode == lngcode && existtext == text {
				update = false
				updatecontent = false
			} else {
				update = false
				if lngcodecontentid == 0 {
					insertcontent = true
				} else {
					updatecontent = true
					insertcontent = false
				}
			}
		}
	}

	iLog.Debug(fmt.Sprintf("populatesinglelngcodes update: %t insert: %t, %t, %t, %d", update, inertlngcode, updatecontent, insertcontent, lngcodeid))
	//currentTimeUTC := time.Now().UTC()

	// Format the time as a string in the MySQL date and time format
	//formattedTime := currentTimeUTC.Format("2006-01-02 15:04:05")
	if inertlngcode {
		lngcodeid, err = f.InsertLngCode(db, lngcode, User) // db.TableInsert("lngcodes", Columns, Values)
		if err != nil {
			iLog.Error(fmt.Sprintf("insert lng code error: %s", err.Error()))
			return
		}
	}
	iLog.Debug(fmt.Sprintf("data for the lngcode content:%d,%d", lngcodeid, languageid))
	if insertcontent && lngcodeid > 0 {

		err := f.insertlngcontent(db, lngcodeid, text, languageid, User)
		if err != nil {
			iLog.Error(fmt.Sprintf("insert lng content error: %s", err.Error()))
			return
		}
	}

}

func (f *LCController) populatelngcodes(lngcodes []string, text []string, languageid int64, User string) {
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

	if languageid == 0 {
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
		iLog.Debug(fmt.Sprintf("populatelngcodes lngcode: %s %s %d %s", lngcodes[i], text[i], languageid, User))
		f.populatesinglelngcodes(db, lngcodes[i], text[i], languageid, User)

	}
	idbtx.Commit()
}
