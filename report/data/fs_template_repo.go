package data

import (
	"fmt"
	//"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type filesystemTemplateRepo struct {
	templatesFolder string
}

func NewFilesystemTemplateRepo(templatesFolder string) *filesystemTemplateRepo {
	return &filesystemTemplateRepo{templatesFolder: templatesFolder}
}

func (ftr *filesystemTemplateRepo) ListAll() ([]map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.data.LoadTemplate"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("documents.UpdateCollection", elapsed)
	}()

	collectionName := "UI_Reports"

	projection := bson.M{
		"_id":         1,
		"name":        1,
		"description": 1,
		"category":    1,
		"version":     1,
		"status":      1,
		"isdefault":   1,
		"system":      1,
	}
	items, err := documents.DocDBCon.QueryCollection(collectionName, nil, projection)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the transaction code list: %v", err))
		return nil, err
	}
	//var result []map[string]interface{}
	result := make([]map[string]interface{}, 0)
	for _, item := range items {
		result = append(result, map[string]interface{}(item))
	}

	return result, nil
}

func (ftr *filesystemTemplateRepo) LoadTemplate(reportName string) ([]byte, error) {
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.data.LoadTemplate"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("documents.UpdateCollection", elapsed)
	}()

	collectionName := "UI_Reports"

	iLog.Debug(fmt.Sprintf("Collection Name: %s, Name: %s ", collectionName, reportName))

	collectionitems, err := documents.DocDBCon.GetDefaultItembyName(collectionName, reportName)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the detail data from collection: %v", err))
		return nil, err
	}

	byteData, err := bson.Marshal(collectionitems)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error marshalling BSON:", err))
		return nil, err
	}

	return byteData, nil
	// return ioutil.ReadFile(filepath.Join(ftr.templatesFolder, templateId+".html"))
}

func (ftr *filesystemTemplateRepo) listFiles(root string) ([]string, error) {
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.data.listFiles"}

	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			key := strings.ReplaceAll(strings.TrimPrefix(strings.TrimPrefix(path, ftr.templatesFolder), "/"), ".html", "")
			files = append(files, key)
		}
		return nil
	})
	if err != nil {
		iLog.ErrorLog(err)
		return nil, err
	}

	return files, err
}
