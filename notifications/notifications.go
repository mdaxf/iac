package notifications

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/documents"
)

// GetNotificationsbyUser retrieves notifications for a specific user.
// It takes a user string as input and returns a list of notifications and an error.
// The function first logs the start time and defer logs the performance duration.
// It also recovers from any panics and logs the error if any.
// The function constructs a filter using the user parameter and other conditions.
// It queries the "Notifications" collection using the constructed filter.
// If there is an error in querying the collection, it logs the error and returns nil and the error.
// Otherwise, it logs the retrieved notification list and returns it along with nil error.

func GetNotificationsbyUser(user string) (interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.GetNotificationsbyUser", elapsed)
	}()

	/*	defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error: %v", err))
			}
		}()
	*/
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
			{"sender": user},
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

// SaveNotification saves a notification for a specific user.
// It takes ndata, which is the data of the notification, and user, which is the username of the user.
// It returns an error if there was a problem saving the notification.

func SaveNotification(ndata interface{}, user string) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "Notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.GetNotificationsbyUser", elapsed)
	}()

	/*	defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error: %v", err))
			}
		}()
	*/
	_, err := documents.DocDBCon.InsertCollection("Notifications", ndata)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to save notification: %v", err))
		return err
	}
	return nil
}

func UpdateNotification(ndata interface{}, user string, comments string, status int) error {
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

	newdata := ndata.(map[string]interface{})
	newdata["system.updatedby"] = user
	newdata["system.updatedon"] = time.Now()

	if newdata["sender"] != user && newdata["receipts.all"] == 1 && (newdata["receipts."+user] == nil || newdata["receipts."+user] == 1) {
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
		userhistory = append(userhistory.([]interface{}), userhisitem)
		newdata["histories"] = userhistory
	}
	if status != 0 {
		newdata["status"] = status
	}
	var filter bson.M
	filter = bson.M{"uuid": newdata["uuid"]}
	/*	objectid, err := primitive.ObjectIDFromHex(newdata["_id"])
		if err != nil {
			doc.iLog.Error(fmt.Sprintf("failed to convert id to objectid with error: %s", err))
			return err
		}

		filter := bson.M{"_id": objectid}  */

	delete(newdata, "_id")
	err := documents.DocDBCon.UpdateCollection("Notifications", filter, nil, newdata)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to update notification: %v", err))
		return err
	}

	go com.IACMessageBusClient.Invoke("send", "IAC_SERVER_NOTIICATION_UPDATE", newdata, "")

	return nil
}

func UpdateNotificationbyUUID(uuid string, user string, comments string) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "Notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.UpdateNotificationbyUUID", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
		}
	}()

	var filter bson.M
	filter = bson.M{"uuid": uuid}

	update := bson.M{
		"$set": bson.M{
			"system.updatedby": user,
			"system.updatedon": time.Now(),
			"receipts." + user: 2,
			"histories": bson.M{
				"$push": bson.M{
					"status":    1,
					"updatedby": user,
					"updatedon": time.Now(),
					"comments":  comments,
				},
			},
		},
	}

	err := documents.DocDBCon.UpdateCollection("Notifications", filter, nil, update)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to update notification: %v", err))
		return err
	}

	data := make(map[string]interface{})
	data["uuid"] = uuid
	data["comments"] = comments
	data["user"] = user

	go com.IACMessageBusClient.Invoke("send", "IAC_SERVER_NOTIICATION_REPLY", data, "")

	return nil
}

// CreateNewNotification creates a new notification with the given notification data and user.
// It sets various system fields such as created by, created on, updated by, updated on, and status.
// If receipts are not provided, it sets the receipts to "all".
// It also adds a history entry with status, updated by, updated on, and comments.
// The notification data is then inserted into the "Notifications" collection in the database.
// Returns an error if there was a failure in saving the notification.

func CreateNewNotification(notificationdata interface{}, user string) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Notifications"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("notification.CreateNewNotification", elapsed)
	}()

	/*	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
		}
	}()  */
	ndata := notificationdata.(map[string]interface{})
	ndata["system.createdby"] = user
	ndata["system.createdon"] = time.Now()
	ndata["system.updatedby"] = user
	ndata["system.updatedon"] = time.Now()
	ndata["status"] = 1
	ndata["sender"] = user
	if ndata["receipts"] == nil {
		ndata["receipts"] = map[string]interface{}{"all": 1}
	}
	history := make([]map[string]interface{}, 1)
	history[0] = make(map[string]interface{})
	history[0]["status"] = 1
	history[0]["updatedby"] = user
	history[0]["updatedon"] = time.Now()
	history[0]["comments"] = "New Notification"
	ndata["histories"] = history
	_, err := documents.DocDBCon.InsertCollection("Notifications", ndata)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to save notification: %v", err))
		return err
	}

	go com.IACMessageBusClient.Invoke("send", "IAC_SERVER_NOTIICATION", ndata, "")

	return nil
}
