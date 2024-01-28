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
	tcom "github.com/mdaxf/iac/engine/com"
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
	TestwithSc      bool
	TestResults     map[string]interface{}
}

// ExecuteUnitTest executes a unit test for a given trancode with the provided systemsessions.
// It returns a map[string]interface{} containing the result of the unit test and an error, if any.
// The result contains the following fields:
// - Name: the name of the unit test
// - Inputs: the inputs of the unit test
// - ExpectedOutputs: the expected outputs of the unit test
// - ExpectError: a boolean value indicating whether the unit test is expected to return an error
// - ExpectedError: the expected error message
// - ActualOutputs: the actual outputs of the unit test
// - ActualError: the actual error message
// - Result: the result of the unit test, either "Pass" or "Fail"

func ExecuteUnitTest(trancode string, systemsessions map[string]interface{}) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.TranCode, User: "System", ControllerName: "TransCode"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.TranCode.ExecuteUnitTest", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to engine.TranCode.ExecuteUnitTest with error: %s", err))
			//	f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err)

		}
	}()

	tranobj, err := getTranCodeData(trancode, documents.DocDBCon)
	if err != nil {
		return nil, err
	}
	tf := NewTranFlow(tranobj, map[string]interface{}{}, systemsessions, nil, nil)
	tf.TestwithSc = true

	result, err := tf.UnitTest()

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ExecuteUnitTestWithTestData executes a unit test with the given test data for a specific trancode.
// It takes the trancode string, testcase map, and systemsessions map as input parameters.
// It returns a map[string]interface{} containing the test result and an error if any.
// The test result contains the following fields:
// - Name: the name of the unit test
// - Inputs: the inputs of the unit test
// - ExpectedOutputs: the expected outputs of the unit test
// - ExpectError: a boolean value indicating whether the unit test is expected to return an error
// - ExpectedError: the expected error message
// - ActualOutputs: the actual outputs of the unit test
// - ActualError: the actual error message
// - Result: the result of the unit test, either "Pass" or "Fail"

func ExecuteUnitTestWithTestData(trancode string, testcase map[string]interface{}, systemsessions map[string]interface{}) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.TranCode, User: "System", ControllerName: "TransCode"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.TranCode.ExecuteUnitTestWithTestData", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to engine.TranCode.ExecuteUnitTestWithTestData with error: %s", err))
			//	f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err)

		}
	}()
	tranobj, err := getTranCodeData(trancode, documents.DocDBCon)
	if err != nil {
		return nil, err
	}
	tf := NewTranFlow(tranobj, map[string]interface{}{}, systemsessions, nil, nil)
	tf.TestwithSc = true

	var testdata types.TestData

	testdata.Inputs = testcase["inputs"].([]types.Input)
	testdata.Outputs = testcase["outputs"].([]types.Output)
	testdata.Name = testcase["name"].(string)
	testdata.WantErr = testcase["wanterr"].(bool)
	testdata.WantedErr = testcase["wantederr"].(string)

	result, err := tf.UnitTestbyTestData(testdata)

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ExecutebyExternal executes a transaction code by calling an external service.
// It takes a trancode string, a data map, a DBTx transaction, a DBCon database connection,
// and a sc signalr client as input parameters.
// It returns a map of outputs and an error.
// The outputs map contains the outputs of the transaction code.
// The error contains the error message if any.
// The function also logs the performance of the transaction code execution.

func ExecutebyExternal(trancode string, data map[string]interface{}, DBTx *sql.Tx, DBCon *documents.DocDB, sc signalr.Client) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.TranCode, User: "System", ControllerName: "TransCode"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.TranCode.ExecutebyExternal", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to engine.TranCode.ExecutebyExternal with error: %s", err))
			//	f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err)

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

// NewTranFlow creates a new instance of TranFlow.
// It takes the following parameters:
// - tcode: the transaction code (types.TranCode)
// - externalinputs: a map of external inputs (map[string]interface{})
// - systemSession: a map of system session data (map[string]interface{})
// - ctx: the context (context.Context)
// - ctxcancel: the cancel function for the context (context.CancelFunc)
// - dbTx: optional parameter for the database transaction (*sql.Tx)
// It returns a pointer to TranFlow.

func NewTranFlow(tcode types.TranCode, externalinputs, systemSession map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc, dbTx ...*sql.Tx) *TranFlow {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Trancode"
	if systemSession["UserNo"] != nil {
		log.User = systemSession["UserNo"].(string)
	} else {
		log.User = "System"
	}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("engine.TranCode.ExecutebyExternal", elapsed)
	}()

	/*	defer func() {
			if err := recover(); err != nil {
				log.Error(fmt.Sprintf("There is error to engine.TranCode.ExecutebyExternal with error: %s", err))
				//	f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err)

			}
		}()
	*/
	idbTx := append(dbTx, nil)[0]

	/*
		tfr := TranFlowstr{}
		callback.RegisterCallBack("TranFlowstr_Execute", tfr.Execute)
	*/
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
		TestwithSc:      false,
		TestResults:     map[string]interface{}{},
	}
}

// Execute executes the transaction flow.
// It starts the timer to measure the execution time and logs the performance duration.
// It recovers from any panics and logs the error message.
// It retrieves the system session, external inputs, and external outputs from the transaction flow.
// It starts a new database transaction if one doesn't exist.
// It starts a new context with a timeout if one doesn't exist.
// It executes the first function group of the transaction code and iterates through subsequent function groups until the code is no longer 1.
// It commits the database transaction if it was started in this function.
// It returns the external outputs and nil error if successful.
func (t *TranFlow) Execute() (map[string]interface{}, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		t.ilog.PerformanceWithDuration("engine.TranCode.Execute", elapsed)
	}()
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

	if t.TestwithSc {
		t.TestResults = map[string]interface{}{}
		t.TestResults["Name"] = t.Tcode.Name
		t.TestResults["Version"] = t.Tcode.Version
		t.TestResults["Inputs"] = t.Externalinputs
		t.TestResults["SystemSession"] = t.SystemSession
		t.TestResults["UserSession"] = userSession
		t.TestResults["Outputs"] = t.externaloutputs
		t.TestResults["FunctionGroups"] = []map[string]interface{}{}

		tcom.SendTestResultMessageBus(t.Tcode.Name, "", "", "UnitTest", "Start",
			t.Externalinputs, t.externaloutputs, t.SystemSession, map[string]interface{}{}, nil, t.SystemSession["ClientID"].(string), t.SystemSession["UserNo"].(string))
	}

	t.ilog.Debug(fmt.Sprintf("Start process transaction code %s's first func group: %s ", t.Tcode.Name, t.Tcode.Firstfuncgroup))
	fgroup, code := t.getFGbyName(t.Tcode.Firstfuncgroup)
	t.ilog.Debug(fmt.Sprintf("start first function group:", logger.ConvertJson(fgroup)))

	for code == 1 {
		fg := funcgroup.NewFGroup(t.DocDBCon, t.SignalRClient, t.DBTx, fgroup, "", systemSession, userSession, externalinputs, externaloutputs, t.Ctx, t.CtxCancel)

		fg.TestwithSc = t.TestwithSc

		fg.Execute()

		if t.TestwithSc {
			t.TestResults["FunctionGroups"] = append(t.TestResults["FunctionGroups"].([]map[string]interface{}), fg.TestResults)
		}

		externalinputs = fg.Externalinputs
		externaloutputs = fg.Externaloutputs
		userSession = fg.UserSession

		if fg.Nextfuncgroup == "" {
			code = 0
			break
		} else {
			fgroup, code = t.getFGbyName(fg.Nextfuncgroup)
			t.ilog.Debug(fmt.Sprintf("function group:%s, Code:%d", logger.ConvertJson(fgroup), code))
		}
	}

	if newTransaction {
		err := t.DBTx.Commit()
		if err != nil {
			t.ilog.Error(fmt.Sprintf("Error in Trancode.Execute during DB transaction commit: %s", err.Error()))
			t.CtxCancel()
			return map[string]interface{}{}, err
		}
	}

	if t.TestwithSc {
		t.TestResults["Outputs"] = externaloutputs
	}
	return externaloutputs, nil

}

// getFGbyName retrieves the FuncGroup by its name from the TranFlow.
// It returns the found FuncGroup and a flag indicating whether the FuncGroup was found or not.

func (t *TranFlow) getFGbyName(name string) (types.FuncGroup, int) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		t.ilog.PerformanceWithDuration("engine.TranCode.getFGbyName", elapsed)
	}()
	/*	defer func() {
			if r := recover(); r != nil {
				t.ilog.Error(fmt.Sprintf("Error in Trancode.getFGbyName: %s", r))
				t.ErrorMessage = fmt.Sprintf("Error in Trancode.getFGbyName: %s", r)
				t.DBTx.Rollback()
				t.CtxCancel()
				return
			}
		}()
	*/
	t.ilog.Debug(fmt.Sprintf("Get the Func group by name: %s", name))
	for _, fgroup := range t.Tcode.Functiongroups {
		if fgroup.Name == name {

			return fgroup, 1
		}
	}
	t.ilog.Debug(fmt.Sprintf("Can't find the Func group by name: %s", name))
	return types.FuncGroup{}, 0
}

// GetTransCode retrieves the transaction code for the given name.
// It reads the transaction code configuration file and returns the corresponding TranCode object.
// If an error occurs during the process, it returns an empty TranCode object and an error.
func GetTransCode(name string) (types.TranCode, error) {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Trancode"
	log.User = "System"

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("engine.TranCode.GetTranCode", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("Error in Trancode.GetTranCode: %s", r))
			return
		}
	}()

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

// Bytetoobj converts a byte slice to a TranCode object.
// It parses the transaction code configuration from the provided byte slice and returns a TranCode object.
// If there is an error during parsing, it returns an empty TranCode object and an error.

func Bytetoobj(config []byte) (types.TranCode, error) {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Trancode"
	log.User = "System"
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("engine.TranCode.Bytetoobj", elapsed)
	}()
	/*	defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("Error in Trancode.Byetoobj: %s", r))
				return
			}
		}()
	*/
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
	TestwithSc bool `json:"TestwithSc,omitempty"`
}

func (t *TranFlowstr) Execute(tcode string, inputs map[string]interface{}, sc signalr.Client, docdbconn *documents.DocDB, ctx context.Context, ctxcancel context.CancelFunc, dbTx ...*sql.Tx) (map[string]interface{}, error) {
	log := logger.Log{ModuleName: logger.TranCode, User: "System", ControllerName: "TransCode.TranFlow"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("engine.TranCode.Tranflow.Execute", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("Error in Trancode.TranFLow.Execute: %s", r))
			return
		}
	}()

	tc, err := GetTranCodeDatabyCode(tcode)

	if err != nil {
		return nil, err
	}
	systemSession := map[string]interface{}{}
	externalinputs := inputs

	idbTx := append(dbTx, nil)[0]

	tf := NewTranFlow(tc, externalinputs, systemSession, ctx, ctxcancel, idbTx)
	tf.SignalRClient = sc
	tf.TestwithSc = t.TestwithSc

	tf.DocDBCon = docdbconn

	return tf.Execute()
}

func (t *TranFlow) UnitTestbyTestData(testdata types.TestData) (map[string]interface{}, error) {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		t.ilog.PerformanceWithDuration("engine.TranCode.Tranflow.UnitTestbyTestData", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			t.ilog.Error(fmt.Sprintf("Error in Trancode.UnitTestbyTestData: %s", r))
			t.ErrorMessage = fmt.Sprintf("Error in Trancode.UnitTestbyTestData: %s", r)
			t.DBTx.Rollback()
			t.CtxCancel()
			return
		}
	}()

	t.ilog.Debug(fmt.Sprintf("Start process transaction code %s's with test data: %s ", t.Tcode.Name, testdata))
	t.ilog.Debug(fmt.Sprintf("externalinputs: %s", logger.ConvertJson(testdata.Inputs)))
	t.ilog.Debug(fmt.Sprintf("expected externaloutputs: %s", logger.ConvertJson(testdata.Outputs)))

	t.Externalinputs = convertInputsToMap(testdata.Inputs)
	t.externaloutputs = map[string]interface{}{}
	testresult := map[string]interface{}{}
	testresult["Name"] = testdata.Name
	testresult["Inputs"] = testdata.Inputs
	testresult["ExpectedOutputs"] = testdata.Outputs
	testresult["ExpectError"] = testdata.WantErr
	testresult["ExpectedError"] = testdata.WantedErr

	tcom.SendTestResultMessageBus(t.Tcode.Name, "", "", "UnitTest", "Start",
		t.Externalinputs, t.externaloutputs, t.SystemSession, map[string]interface{}{}, nil, t.SystemSession["ClientID"].(string), t.SystemSession["UserNo"].(string))

	outputs, err := t.Execute()

	tcom.SendTestResultMessageBus(t.Tcode.Name, "", "", "UnitTest", "Complete",
		t.Externalinputs, outputs, t.SystemSession, map[string]interface{}{}, err, t.SystemSession["ClientID"].(string), t.SystemSession["UserNo"].(string))

	t.ilog.Debug(fmt.Sprintf("actual externaloutputs: %v, expected outputs: %v", outputs, testdata.Outputs))
	if err != nil {
		t.ilog.Error(fmt.Sprintf("Error in Trancode.Execute: %s", err.Error()))

		if testdata.WantErr {
			if testdata.WantedErr == err.Error() {
				testresult["ActualError"] = err.Error()
				testresult["Result"] = "Pass"

			} else {
				testresult["ActualError"] = err.Error()
				testresult["Result"] = "Pass"

			}

		} else {
			testresult["ActualError"] = err.Error()
			testresult["Result"] = "Fail"

		}
	}

	if !testdata.WantErr {
		if !compareMap(outputs, convertOutputsToMap(testdata.Outputs)) {
			testresult["ActualOutputs"] = outputs
			testresult["Result"] = "Fail"

		} else {
			testresult["ActualOutputs"] = outputs
			testresult["Result"] = "Pass"

		}
	} else {
		testresult["Result"] = "Fail"
		testresult["ActualOutputs"] = outputs
		testresult["ActualError"] = ""

	}

	return testresult, nil
}

func (t *TranFlow) UnitTest() (map[string]interface{}, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		t.ilog.PerformanceWithDuration("engine.TranCode.Tranflow.UnitTest", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			t.ilog.Error(fmt.Sprintf("Error in Trancode.UnitTest: %s", r))
			t.ErrorMessage = fmt.Sprintf("Error in Trancode.UnitTest: %s", r)
			t.DBTx.Rollback()
			t.CtxCancel()
			return
		}
	}()

	result := make(map[string]interface{})

	t.ilog.Info(fmt.Sprintf("Start Process for transaction code %s's %s ", t.Tcode.Name, "Unit Test"))
	t.ilog.Debug(fmt.Sprintf("systemSession: %s", logger.ConvertJson(t.SystemSession)))
	testdatalist := t.Tcode.TestDatas

	for _, testdata := range testdatalist {

		testresult, _ := t.UnitTestbyTestData(testdata)
		result[testdata.Name] = testresult

	}

	return result, nil
}

func GetTranCodeDatabyCode(Code string) (types.TranCode, error) {
	log := logger.Log{ModuleName: logger.TranCode, User: "System", ControllerName: "TransCode"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("engine.TranCode.GetTranCodeDatabyCode", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("Error in Trancode.GetTranCodeDatabyCode: %s", r))
			return
		}
	}()

	trancodeobj, err := getTranCodeData(Code, documents.DocDBCon)
	if err != nil {
		return types.TranCode{}, err
	}
	return trancodeobj, nil
}

func getTranCodeData(Code string, DBConn *documents.DocDB) (types.TranCode, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.TranCode.getTranCodeData", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Error in Trancode.getTranCodeData: %s", r))
			return
		}
	}()

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

func convertInputsToMap(inputs []types.Input) map[string]interface{} {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.TranCode.convertInputsToMap", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Error in Trancode.convertInputsToMap: %s", r))
			return
		}
	}()

	result := map[string]interface{}{}

	for _, input := range inputs {
		result[input.Name] = input.Value
	}

	return result
}

func convertOutputsToMap(outputs []types.Output) map[string]interface{} {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "TranCode"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.TranCode.convertOutputsToMap", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Error in Trancode.convertOutputsToMap: %s", r))
			return
		}
	}()

	result := map[string]interface{}{}

	for _, output := range outputs {
		result[output.Name] = output.Value
	}

	return result
}

func compareMap(map1, map2 map[string]interface{}) bool {
	iLog := logger.Log{ModuleName: logger.TranCode, User: "System", ControllerName: "TranCode"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.TranCode.compareMap", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Error in Trancode.compareMap: %s", r))
			return
		}
	}()

	if len(map1) != len(map2) {
		return false
	}

	for key1, value1 := range map1 {
		value2, ok := map2[key1]
		if !ok {
			return false
		}

		if value1 != value2 {
			return false
		}
	}

	return true
}
