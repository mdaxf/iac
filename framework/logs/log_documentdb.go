package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mdaxf/iac/com"
)

// fileLogWriter implements LoggerInterface.
// Writes messages by lines limit, file size limit, or time frequency.
type docdbLogWriter struct {
	sync.RWMutex // write log order by order and  atomic incr maxLinesCurLines and maxSizeCurSize

	MongoDBClient        *mongo.Client
	MongoDBDatabase      *mongo.Database
	MongoDBCollection_TC *mongo.Collection
	CollectionName       string

	Level int `json:"level"`

	logFormatter LogFormatter
	Formatter    string `json:"formatter"`

	DatabaseConnection string
	DatabaseName       string
	monitoring         bool
}

// newDocDBLogger creates a documentLogWriter returning as LoggerInterface.
func newDocDBLogger() Logger {
	w := &docdbLogWriter{
		Level: LevelTrace,
	}
	w.logFormatter = w
	return w
}

func (*docdbLogWriter) Format(lm *LogMsg) string {
	msg := lm.OldStyleFormat()
	hd, _, _ := formatTimeHeader(lm.When)
	msg = fmt.Sprintf("%s %s\n", string(hd), msg)
	return msg
}

func (w *docdbLogWriter) SetFormatter(f LogFormatter) {
	w.logFormatter = f
}

func (doc *docdbLogWriter) Init(config string) error {

	var err error
	//	fmt.Println("init docdblogwriter:", config)
	var cf map[string]string
	if err := json.Unmarshal([]byte(config), &cf); err != nil {
		fmt.Sprintln("could not unmarshal this config, it must be valid json stringP: %s with error %s", config, err)
		return err
	}

	if _, ok := cf["conn"]; !ok {
		return fmt.Errorf(`config must contains "conn" field: %s`, config)
	}

	if _, ok := cf["db"]; !ok {
		fmt.Errorf(`config must contains "db" field: %s`, config)
		return err
	}

	if _, ok := cf["collection"]; !ok {
		fmt.Errorf(`config must contains "collection" field: %s`, config)
		return fmt.Errorf(`config must contains "collection" field: %s`, config)
	}

	doc.DatabaseConnection = cf["conn"]
	doc.DatabaseName = cf["db"]
	doc.CollectionName = cf["collection"]
	doc.monitoring = false

	err = doc.ConnectMongoDB()

	if err != nil {
		fmt.Errorf("There is error to connect to Mongodb for logger")
		return err
	}

	if doc.monitoring == false {
		go func() {
			doc.MonitorAndReconnect()
		}()

	}
	return nil
}

func (doc *docdbLogWriter) ConnectMongoDB() error {

	var err error

	// Create direct MongoDB connection
	// Note: Factory registration is handled separately to avoid import cycles
	doc.MongoDBClient, err = mongo.NewClient(options.Client().ApplyURI(doc.DatabaseConnection))
	if err != nil {
		fmt.Errorf(fmt.Sprintf("failed to connect mongodb with error: %s", err))
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = doc.MongoDBClient.Connect(ctx)
	if err != nil {
		fmt.Errorf("failed to connect mongodb with error: %s", err)
		return err
	}

	// Register with global array for backward compatibility (deprecated)
	// Factory registration will be handled in initialization code
	if com.MongoDBClients == nil {
		com.MongoDBClients = make([]*mongo.Client, 0)
	}
	com.MongoDBClients = append(com.MongoDBClients, doc.MongoDBClient)

	doc.MongoDBDatabase = doc.MongoDBClient.Database(doc.DatabaseName)
	doc.MongoDBCollection_TC = doc.MongoDBDatabase.Collection(doc.CollectionName)

	err = doc.MongoDBClient.Ping(context.Background(), nil)
	if err != nil {
		fmt.Errorf("failed to connect mongodb with error: %s", err)
		return err
	}

	return nil
}

func (doc *docdbLogWriter) MonitorAndReconnect() {
	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			fmt.Errorf("monitorAndReconnect defer error: %s", err)
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	doc.monitoring = true
	for {
		err := doc.MongoDBClient.Ping(context.Background(), nil)
		if err != nil {
			fmt.Errorf("MongoDB connection lost, reconnecting...")

			err := doc.ConnectMongoDB()

			if err != nil {
				fmt.Errorf("Failed to reconnect to MongoDB %s with err :%v", doc.DatabaseConnection, err)
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			} else {
				time.Sleep(1 * time.Second)
				fmt.Errorf("MongoDB reconnected successfully")
			}
		} else {
			time.Sleep(1 * time.Second) // Check connection every 60 seconds
		}
	}

}

// WriteMsg writes logger message into file.
func (doc *docdbLogWriter) WriteMsg(lm *LogMsg) error {
	//	fmt.Println("fileLogWriter.WriteMsg, %s", lm)

	if lm.Level > doc.Level && lm.Level != LeverPerformance {
		return nil
	}

	y, mo, d := lm.When.Date()
	h, mi, s := lm.When.Clock()

	msg := doc.logFormatter.Format(lm)

	docmsg := make(map[string]interface{})
	docmsg["instance"] = com.Instance
	docmsg["level"] = lm.Level
	docmsg["year"] = y
	docmsg["month"] = mo
	docmsg["date"] = d
	docmsg["hour"] = h
	docmsg["minute"] = mi
	docmsg["second"] = s
	docmsg["when"] = lm.When

	docmsg["message"] = msg

	_, err := doc.MongoDBCollection_TC.InsertOne(context.Background(), docmsg)

	return err
}

func (doc *docdbLogWriter) deleteOldLog() {
	doc.MongoDBCollection_TC.DeleteMany(context.Background(), bson.M{"when": bson.M{"$lt": time.Now().AddDate(0, 0, -7)}})
	return
}

// Destroy all collection.
func (doc *docdbLogWriter) Destroy() {
	doc.MongoDBCollection_TC.Drop(context.Background())
	return
}

func (w *docdbLogWriter) Flush() {
	return
}

func init() {
	Register(AdapterDocumentDB, newDocDBLogger)
}
