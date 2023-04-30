package trancode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"database/sql"

	dbconn "github.com/mdaxf/iac/databases"
	funcgroup "github.com/mdaxf/iac/engine/funcgroup"
	"github.com/mdaxf/iac/engine/types"
)

type TranFlow struct {
	Tcode           types.TranCode
	DBTx            *sql.Tx
	Externalinputs  map[string]interface{} // {sessionanme: value}
	externaloutputs map[string]interface{} // {sessionanme: value}
	SystemSession   map[string]interface{}
}

func NewTranFlow(tcode types.TranCode, externalinputs, systemSession map[string]interface{}, dbTx ...*sql.Tx) *TranFlow {
	idbTx := append(dbTx, nil)[0]
	return &TranFlow{
		Tcode:           tcode,
		DBTx:            idbTx,
		Externalinputs:  externalinputs,
		externaloutputs: map[string]interface{}{},
		SystemSession:   systemSession,
	}
}

func (t *TranFlow) Execute() (map[string]interface{}, error) {
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
			return map[string]interface{}{}, err
		}

		defer t.DBTx.Rollback()
	}

	fgroup, code := t.getFGbyName(t.Tcode.Firstfuncgroup)
	for code == 1 {
		fg := funcgroup.NewFGroup(t.DBTx, fgroup, "", systemSession, userSession, externalinputs, externaloutputs)
		/*	funcgroup.FGroup{
			fgobj: fgroup,
			nextfuncgroup: "",
			systemSession: systemSession,
			userSession: userSession,
			externalinputs: externalinputs,
			externaloutputs: externaloutputs,
		} */
		fg.Execute()
		externalinputs = fg.Externalinputs
		externaloutputs = fg.Externaloutputs
		userSession = fg.UserSession

		fgroup, code = t.getFGbyName(fg.Nextfuncgroup)

	}

	if newTransaction {
		err := t.DBTx.Commit()
		if err != nil {
			return map[string]interface{}{}, err
		}
	}

	return externaloutputs, nil

}

func (t *TranFlow) getFGbyName(name string) (types.FuncGroup, int) {
	for _, fgroup := range t.Tcode.Functiongroups {
		if fgroup.Name == name {
			return fgroup, 1
		}
	}
	return types.FuncGroup{}, 0
}

func GetTransCode(name string) (types.TranCode, error) {
	log.Println(fmt.Sprintf("./%s/%s%s", "trancodes", name, ".json"))
	data, err := ioutil.ReadFile(fmt.Sprintf("./%s/%s%s", "trancodes", name, ".json"))
	if err != nil {
		log.Println(fmt.Errorf("failed to read configuration file: %v", err))
		return types.TranCode{}, fmt.Errorf("failed to read configuration file: %v", err)
	}
	log.Println(string(data))
	fmt.Println(string(data))
	return Bytetoobj(data)
}

func Bytetoobj(config []byte) (types.TranCode, error) {
	var tranCode types.TranCode
	if err := json.Unmarshal(config, &tranCode); err != nil {
		return types.TranCode{}, fmt.Errorf("failed to parse configuration file: %v", err)
	}
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

func (t *TranFlowstr) Execute(tcode string, inputs map[string]interface{}, dbTx ...*sql.Tx) (map[string]interface{}, error) {
	tc, err := GetTransCode(tcode)

	if err != nil {
		return nil, err
	}
	systemSession := map[string]interface{}{}
	externalinputs := inputs

	idbTx := append(dbTx, nil)[0]

	tf := NewTranFlow(tc, externalinputs, systemSession, idbTx)

	return tf.Execute()
}
