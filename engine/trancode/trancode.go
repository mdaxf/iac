package trancode

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"database/sql"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	funcgroup "github.com/mdaxf/iac/engine/funcgroup"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/engine/callback"
	"github.com/mdaxf/signalrsrv/signalr"
)

type TranFlow struct {
	Tcode           types.TranCode
	DBTx            *sql.Tx
	Ctx             context.Context
	CtxCancel       context.CancelFunc
	Externalinputs  map[string]interface{} // {sessionanme: value}
	externaloutputs map[string]interface{} // {sessionanme: value}
	SystemSession   map[string]interface{}
	ilog            logger.Log
	DocDBCon        *documents.DocDB
	SignalRClient   signalr.Client
	ErrorMessage    string
}

func ExecutebyExternal(trancode string, data map[string]interface{}, DBTx *sql.Tx, DBCon *documents.DocDB, sc signalr.Client) (map[string]interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			//outputs := make(map[string]interface{})
			return
		}
	}()

	tranobj, err := getTranCodeData(trancode, DBCon)
	if err != nil {
		return nil, err
	}
	tf := NewTranFlow(tranobj, data, map[string]interface{}{}, nil, nil, DBTx)
	tf.DocDBCon = DBCon
	tf.SignalRClient = sc

	outputs, err := tf.Execute()

	if err != nil {
		return nil, err
	}
	return outputs, nil
}

func NewTranFlow(tcode types.TranCode, externalinputs, systemSession map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc, dbTx ...*sql.Tx) *TranFlow {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Trancode"
	if systemSession["User"] != nil {
		log.User = systemSession["User"].(string)
	} else {
		log.User = "System"
	}

	idbTx := append(dbTx, nil)[0]

	tfr := TranFlowstr{}
	callback.RegisterCallBack("TranFlowstr_Execute", tfr.Execute)

	return &TranFlow{
		Tcode:           tcode,
		DBTx:            idbTx,
		Ctx:             ctx,
		CtxCancel:       ctxcancel,
		ilog:            log,
		Externalinputs:  externalinputs,
		externaloutputs: map[string]interface{}{},
		SystemSession:   systemSession,
		DocDBCon:        documents.DocDBCon,
		SignalRClient:   com.IACMessageBusClient,
		ErrorMessage:    "",
	}
}

func (t *TranFlow) Execute() (map[string]interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			t.ilog.Error(fmt.Sprintf("Error in Trancode.Execute: %s", r))
			t.ErrorMessage = fmt.Sprintf("Error in Trancode.Execute: %s", r)
			t.DBTx.Rollback()
			t.CtxCancel()
			return
		}
	}()

	t.ilog.Info(fmt.Sprintf("Start process transaction code %s's %s ", t.Tcode.Name, "Execute"))
	t.ilog.Debug(fmt.Sprintf("systemSession: %s", logger.ConvertJson(t.SystemSession)))
	t.ilog.Debug(fmt.Sprintf("externalinputs: %s", logger.ConvertJson(t.Externalinputs)))
	t.ilog.Debug(fmt.Sprintf("externaloutputs: %s", logger.ConvertJson(t.externaloutputs)))
	systemSession := t.SystemSession
	externalinputs := t.Externalinputs
	externaloutputs := t.externaloutputs
	userSession := map[string]interface{}{}
	var err error
	newTransaction := false

	if t.DBTx == nil {

		t.DBTx, err = dbconn.DB.Begin()
		newTransaction = true
		if err != nil {
			t.ilog.Error(fmt.Sprintf("Error in Trancode.Execute during DB transaction beginning: %s", err.Error()))
			return map[string]interface{}{}, err
		}

		defer t.DBTx.Rollback()
	}

	if t.Ctx == nil {
		t.Ctx, t.CtxCancel = context.WithTimeout(context.Background(), time.Second*time.Duration(com.TransactionTimeout))

		defer t.CtxCancel()
	}

	t.ilog.Debug(fmt.Sprintf("Start process transaction code %s's first func group: %s ", t.Tcode.Name, t.Tcode.Firstfuncgroup))
	fgroup, code := t.getFGbyName(t.Tcode.Firstfuncgroup)
	t.ilog.Debug(fmt.Sprintf("start first function group:", logger.ConvertJson(fgroup)))

	for code == 1 {
		fg := funcgroup.NewFGroup(t.DocDBCon, t.SignalRClient, t.DBTx, fgroup, "", systemSession, userSession, externalinputs, externaloutputs, t.Ctx, t.CtxCancel)
		fg.Execute()
		externalinputs = fg.Externalinputs
		externaloutputs = fg.Externaloutputs
		userSession = fg.UserSession

		fgroup, code = t.getFGbyName(fg.Nextfuncgroup)
		t.ilog.Debug(fmt.Sprintf("function group:%s, Code:%d", logger.ConvertJson(fgroup), code))
	}

	if newTransaction {
		err := t.DBTx.Commit()
		if err != nil {
			t.ilog.Error(fmt.Sprintf("Error in Trancode.Execute during DB transaction commit: %s", err.Error()))
			t.CtxCancel()
			return map[string]interface{}{}, err
		}
	}

	return externaloutputs, nil

}

func (t *TranFlow) getFGbyName(name string) (types.FuncGroup, int) {
	t.ilog.Debug(fmt.Sprintf("Get the Func group by name: %s", name))
	for _, fgroup := range t.Tcode.Functiongroups {
		if fgroup.Name == name {

			return fgroup, 1
		}
	}
	t.ilog.Debug(fmt.Sprintf("Can't find the Func group by name: %s", name))
	return types.FuncGroup{}, 0
}

func GetTransCode(name string) (types.TranCode, error) {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Trancode"
	log.User = "System"
	log.Info(fmt.Sprintf("Start get transaction code %s", name))

	log.Info(fmt.Sprintf("./%s/%s%s", "trancodes", name, ".json"))
	data, err := ioutil.ReadFile(fmt.Sprintf("./%s/%s%s", "trancodes", name, ".json"))
	if err != nil {
		log.Error(fmt.Sprintf("failed to read configuration file: %v", err))
		return types.TranCode{}, fmt.Errorf("failed to read configuration file: %v", err)
	}
	log.Debug(fmt.Sprintf("Read the tran code configuration:%s", string(data)))
	//	fmt.Println(string(data))
	return Bytetoobj(data)
}

func Bytetoobj(config []byte) (types.TranCode, error) {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Trancode"
	log.User = "System"
	log.Info(fmt.Sprintf("Start parse transaction code configuration"))

	var tranCode types.TranCode
	if err := json.Unmarshal(config, &tranCode); err != nil {
		return types.TranCode{}, fmt.Errorf("failed to parse configuration file: %v", err)
	}
	log.Debug(fmt.Sprintf("Parse the tran code configuration:%s", logger.ConvertJson(tranCode)))
	return tranCode, nil
}

func Configtoobj(config string) (types.TranCode, error) {

	return Bytetoobj([]byte(config))
}

type TranCodeData struct {
	TranCode string                 `json:"code"`
	inputs   map[string]interface{} `json:"Inputs"`
}

type TranFlowstr struct {
}

func (t *TranFlowstr) Execute(tcode string, inputs map[string]interface{}, sc signalr.Client, docdbconn *documents.DocDB, ctx context.Context, ctxcancel context.CancelFunc, dbTx ...*sql.Tx) (map[string]interface{}, error) {
	tc, err := GetTranCodeData(tcode)

	if err != nil {
		return nil, err
	}
	systemSession := map[string]interface{}{}
	externalinputs := inputs

	idbTx := append(dbTx, nil)[0]

	tf := NewTranFlow(tc, externalinputs, systemSession, ctx, ctxcancel, idbTx)
	tf.SignalRClient = sc
	tf.DocDBCon = docdbconn

	return tf.Execute()
}

func GetTranCodeData(Code string) (types.TranCode, error) {
	/*	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}
		iLog.Info(fmt.Sprintf("Get the trancode code for %s ", Code))

		iLog.Info(fmt.Sprintf("Start process transaction code %s's %s ", Code, "Execute"))

		filter := bson.M{"trancodename": Code, "isdefault": true}

		tcode, err := documents.DocDBCon.QueryCollection("Transaction_Code", filter, nil)

		if err != nil {
			iLog.Error(fmt.Sprintf("Get transaction code %s's error", Code))

			return types.TranCode{}, err
		}
		iLog.Debug(fmt.Sprintf("transaction code %s's data: %s", Code, tcode))
		jsonString, err := json.Marshal(tcode[0])
		if err != nil {

			iLog.Error(fmt.Sprintf("Error marshaling json:", err.Error()))
			return types.TranCode{}, err
		}

		trancodeobj, err := Configtoobj(string(jsonString))
		if err != nil {
			iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
			return types.TranCode{}, err
		}

		iLog.Debug(fmt.Sprintf("transaction code %s's json: %s", trancodeobj, string(jsonString)))

		if err != nil {
			iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
			return types.TranCode{}, err
		}
	*/
	trancodeobj, err := getTranCodeData(Code, documents.DocDBCon)
	if err != nil {
		return types.TranCode{}, err
	}
	return trancodeobj, nil
}

func getTranCodeData(Code string, DBConn *documents.DocDB) (types.TranCode, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}
	iLog.Info(fmt.Sprintf("Get the trancode code for %s ", Code))

	iLog.Info(fmt.Sprintf("Start process transaction code %s's %s ", Code, "Execute"))

	filter := bson.M{"trancodename": Code, "isdefault": true}

	tcode, err := DBConn.QueryCollection("Transaction_Code", filter, nil)

	if err != nil {
		iLog.Error(fmt.Sprintf("Get transaction code %s's error", Code))

		return types.TranCode{}, err
	}
	iLog.Debug(fmt.Sprintf("transaction code %s's data: %s", Code, tcode))
	jsonString, err := json.Marshal(tcode[0])
	if err != nil {

		iLog.Error(fmt.Sprintf("Error marshaling json:", err.Error()))
		return types.TranCode{}, err
	}

	trancodeobj, err := Configtoobj(string(jsonString))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
		return types.TranCode{}, err
	}

	iLog.Debug(fmt.Sprintf("transaction code %s's json: %s", trancodeobj, string(jsonString)))

	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling json:", err.Error()))
		return types.TranCode{}, err
	}

	return trancodeobj, nil
}
