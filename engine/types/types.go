package types

type TranCode struct {
	ID             string                 "json:'_id'"
	UUID           string                 "json:'uuid'"
	Name           string                 "json:'trancodename'"
	Version        string                 "json:'version'"
	IsDefault      bool                   "json:'isdefault'"
	Status         Status                 "json:'status'"
	Inputs         []Input                "json:'inputs'"
	Outputs        []Output               "json:'outputs'"
	Functiongroups []FuncGroup            "json:'functiongroups'"
	Workflow       map[string]interface{} "json:'workflow'"
	Firstfuncgroup string                 "json:'firstfuncgroup'"
	SystemData     SystemData             "json:'system'"
	Description    string                 "json:'description'"
	TestDatas      []TestData             "json:'testdatas'"
}

type TestData struct {
	Name      string   "json:'name'"
	Inputs    []Input  "json:'inputs'"
	Outputs   []Output "json:'outputs'"
	WantErr   bool     "json:'wanterr'"
	WantedErr string   "json:'wantederr'"
}

type SystemData struct {
	CreatedBy string "json:'createdby'"
	CreatedOn string "json:'createdon'"
	UpdatedBy string "json:'updatedby'"
	UpdatedOn string "json:'updatedon'"
}

type FuncGroup struct {
	ID                string                 "json:'id'"
	Name              string                 "json:'name'"
	Functions         []Function             "json:'functions'"
	Executionsequence string                 "json:'sequence'"
	Session           map[string]interface{} "json:'session'"
	RouterDef         RouterDef              "json:'routerdef'"
	functiongroupname string                 "json:'functiongroupname'"
	Description       string                 "json:'description'"
	routing           bool                   "json:'routing'"
	Type              string                 "json:'type'"
	x                 int                    "json:'x'"
	y                 int                    "json:'y'"
	width             int                    "json:'width'"
	height            int                    "json:'height'"
}

type RouterDef struct {
	Variable         string   "json:'variable'"
	Vartype          string   "json:'vartype'"
	Values           []string "json:'values'"
	Nextfuncgroups   []string "json:'nextfuncgroups'"
	Defaultfuncgroup string   "json:'defaultfuncgroup'"
}

type Function struct {
	ID           string                 "json:'id'"
	Name         string                 "json:'name'"
	Version      string                 "json:'version'"
	Status       Status                 "json:'status'"
	Functype     FunctionType           "json:'functype'"
	Inputs       []Input                "json:'inputs'"
	Outputs      []Output               "json:'outputs'"
	Content      string                 "json:'content'"
	Script       string                 "json:'script'"
	Mapdata      map[string]interface{} "json:'mapdata'"
	FunctionName string                 "json:'functionname'"
	Description  string                 "json:'description'"
	Type         string                 "json:'type'"
	x            int                    "json:'x'"
	y            int                    "json:'y'"
	width        int                    "json:'width'"
	height       int                    "json:'height'"
}

type Input struct {
	ID           string      "json:'id'"
	Name         string      "json:'name'"
	Source       InputSource "json:'source'"   // 0: constant, 1: function, 2: session
	Datatype     DataType    "json:'datatype'" // 0: string, 1: int, 2: float, 3: bool, 4: datetime	5: object (json)
	Inivalue     string      "json:'initialvalue'"
	Defaultvalue string      "json:'defaultvalue'"
	Value        string      "json:'value'"
	List         bool        "json:'list'"
	Repeat       bool        "json:'repeat'"
	Aliasname    string      "json:'aliasname'"
	Description  string      "json:'description'"
}

type Output struct {
	ID           string       "json:'id'"
	Name         string       "json:'name'"
	Outputdest   []OutputDest "json:'outputdest'" // 0:none, 1:session, 2:engine
	Datatype     DataType     "json:'datatype'"   // 0: string, 1: int, 2: float, 3: bool, 4: datetime	5: object (json)
	Inivalue     string       "json:'initialvalue'"
	Defaultvalue string       "json:'defaultvalue'"
	Value        string       "json:'value'"
	List         bool         "json:'list'"
	Aliasname    []string     "json:'aliasname'"
	Description  string       "json:'description'"
}

type FunctionType int

const (
	InputMap FunctionType = iota
	GoExpr
	Javascript
	Query
	StoreProcedure
	SubTranCode
	TableInsert
	TableUpdate
	TableDelete
	CollectionInsert
	CollectionUpdate
	CollectionDelete
	ThrowError
	SendMessage
	SendEmail
	ExplodeWorkFlow
	StartWorkFlowTask
	CompleteWorkFlowTask
	SendMessagebyKafka
	SendMessagebyMQTT
	SendMessagebyAQMP
	WebServiceCall
	PythonExpr
	PythonScript
)

// String returns the string representation of FunctionType
func (ft FunctionType) String() string {
	switch ft {
	case InputMap:
		return "InputMap"
	case GoExpr:
		return "GoExpr"
	case Javascript:
		return "Javascript"
	case Query:
		return "Query"
	case StoreProcedure:
		return "StoreProcedure"
	case SubTranCode:
		return "SubTranCode"
	case TableInsert:
		return "TableInsert"
	case TableUpdate:
		return "TableUpdate"
	case TableDelete:
		return "TableDelete"
	case CollectionInsert:
		return "CollectionInsert"
	case CollectionUpdate:
		return "CollectionUpdate"
	case CollectionDelete:
		return "CollectionDelete"
	case ThrowError:
		return "ThrowError"
	case SendMessage:
		return "SendMessage"
	case SendEmail:
		return "SendEmail"
	case ExplodeWorkFlow:
		return "ExplodeWorkFlow"
	case StartWorkFlowTask:
		return "StartWorkFlowTask"
	case CompleteWorkFlowTask:
		return "CompleteWorkFlowTask"
	case SendMessagebyKafka:
		return "SendMessagebyKafka"
	case SendMessagebyMQTT:
		return "SendMessagebyMQTT"
	case SendMessagebyAQMP:
		return "SendMessagebyAQMP"
	case WebServiceCall:
		return "WebServiceCall"
	case PythonExpr:
		return "PythonExpr"
	case PythonScript:
		return "PythonScript"
	default:
		return "Unknown"
	}
}

type Status int

const (
	Design Status = iota
	Test
	Prototype
	Production
)

type DataType int

const (
	String DataType = iota
	Integer
	Float
	Bool
	DateTime
	Object
)

type InputSource int

const (
	Constant InputSource = iota
	Prefunction
	Fromsyssession
	Fromusersession
	Fromexternal
)

// String returns the string representation of InputSource
func (is InputSource) String() string {
	switch is {
	case Constant:
		return "Constant"
	case Prefunction:
		return "Prefunction"
	case Fromsyssession:
		return "Fromsyssession"
	case Fromusersession:
		return "Fromusersession"
	case Fromexternal:
		return "Fromexternal"
	default:
		return "Unknown"
	}
}

type OutputDest int

const (
	None OutputDest = iota
	Tosession
	Toexternal
)

var DateTimeFormat string = "2006-01-02 15:04:05"

const TranCodeTestProcessMessageBus string = "IAC_TRANCODE_TEST_PROCESS"
const TranCodeTestResultMessageBus string = "IAC_TRANCODE_TEST_RESULT"
