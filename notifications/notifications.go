package notifications

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/documents"
)

type Notification struct {
}

func (n *Notification) GetNotificationsbyUser(user string) (interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.GetNotificationsbyUser", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
		}
	}()

	var filter bson.M
	filter = bson.M{
		"$or": []bson.M{
			{"receipts." + user: bson.M{"$in": []int{1, 2, 3}}},
			{
				"$and": []bson.M{
					{"receipts." + user: bson.M{"$exists": false}},
					{"receipts.all": bson.M{"$exists": true}},
				},
			},
		},
		"status": bson.M{"$in": []int{1, 2}},
	}

	collectionitems, err := documents.DocDBCon.QueryCollection("Notifications", filter, nil)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to retrieve the list from notification: %v", err))
		return nil, err
	}
	iLog.Debug(fmt.Sprintf("Get notification list from respository with data: %s", logger.ConvertJson(collectionitems)))

	return collectionitems, nil
}

func (n *Notification) SaveNotification(ndata interface{}, user string) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "Notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.GetNotificationsbyUser", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
		}
	}()

	_, err := documents.DocDBCon.InsertCollection("Notifications", ndata)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to save notification: %v", err))
		return err
	}
	return nil
}

func (n *Notification) UpdateNotification(ndata interface{}, user string, comments string, status int) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "Notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.UpdateNotification", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
		}
	}()

	newdata := ndata.(bson.M)
	newdata["system.updatedby"] = user
	newdata["system.updatedon"] = time.Now()

	if newdata["sender"] != user && newdata["receipts"].(bson.M)["all"] == 1 && (newdata["receipts."+user] == nil || newdata["receipts."+user] == 1) {
		newdata["receipts."+user] = 2
	}
	userhisitem := make(map[string]interface{})
	userhisitem["status"] = newdata["receipts."+user]
	userhisitem["updatedby"] = user
	userhisitem["updatedon"] = time.Now()
	userhisitem["comments"] = comments
	userhistory := newdata["histories"]
	if userhistory == nil {
		userhistory = make([]map[string]interface{}, 1)
		userhistory.([]map[string]interface{})[0] = userhisitem
		newdata["histories"] = userhistory
	} else {
		userhistory = append(userhistory.([]map[string]interface{}), userhisitem)
		newdata["histories"] = userhistory
	}
	if status != 0 {
		newdata["status"] = status
	}
	var filter bson.M
	filter = bson.M{"_id": newdata["_id"]}

	delete(newdata, "_id")
	err := documents.DocDBCon.UpdateCollection("Notifications", filter, nil, newdata)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to update notification: %v", err))
		return err
	}
	return nil
}

func (n *Notification) CreateNewNotification(ndata interface{}, user string) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.CreateNewNotification", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
		}
	}()

	ndata.(bson.M)["system.createdby"] = user
	ndata.(bson.M)["system.createdon"] = time.Now()
	ndata.(bson.M)["system.updatedby"] = user
	ndata.(bson.M)["system.updatedon"] = time.Now()
	ndata.(bson.M)["status"] = 1
	ndata.(bson.M)["sender"] = user
	if ndata.(bson.M)["receipts"] == nil {
		ndata.(bson.M)["receipts"] = bson.M{"all": 1}
	}
	history := make([]map[string]interface{}, 1)
	history[0] = make(map[string]interface{})
	history[0]["status"] = 1
	history[0]["updatedby"] = user
	history[0]["updatedon"] = time.Now()
	history[0]["comments"] = "New Notification"
	ndata.(bson.M)["histories"] = history
	_, err := documents.DocDBCon.InsertCollection("Notifications", ndata)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to save notification: %v", err))
		return err
	}
	return nil
}
