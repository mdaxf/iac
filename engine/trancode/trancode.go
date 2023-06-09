package trancode

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"database/sql"

	dbconn "github.com/mdaxf/iac/databases"
	funcgroup "github.com/mdaxf/iac/engine/funcgroup"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
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

	return &TranFlow{
		Tcode:           tcode,
		DBTx:            idbTx,
		Ctx:             ctx,
		CtxCancel:       ctxcancel,
		ilog:            log,
		Externalinputs:  externalinputs,
		externaloutputs: map[string]interface{}{},
		SystemSession:   systemSession,
	}
}

func (t *TranFlow) Execute() (map[string]interface{}, error) {
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
		t.Ctx, t.CtxCancel = context.WithCancel(context.Background())

		defer t.CtxCancel()
	}

	t.ilog.Debug(fmt.Sprintf("Start process transaction code %s's first func group: %s ", t.Tcode.Name, t.Tcode.Firstfuncgroup))
	fgroup, code := t.getFGbyName(t.Tcode.Firstfuncgroup)
	t.ilog.Debug(fmt.Sprintf("start first function group:", logger.ConvertJson(fgroup)))

	for code == 1 {
		fg := funcgroup.NewFGroup(t.DBTx, fgroup, "", systemSession, userSession, externalinputs, externaloutputs, t.Ctx, t.CtxCancel)
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

func (t *TranFlowstr) Execute(tcode string, inputs map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc, dbTx ...*sql.Tx) (map[string]interface{}, error) {
	tc, err := GetTransCode(tcode)

	if err != nil {
		return nil, err
	}
	systemSession := map[string]interface{}{}
	externalinputs := inputs

	idbTx := append(dbTx, nil)[0]

	tf := NewTranFlow(tc, externalinputs, systemSession, ctx, ctxcancel, idbTx)

	return tf.Execute()
}
