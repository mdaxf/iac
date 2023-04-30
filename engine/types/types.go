package types

type TranCode struct {
	Name           string   "operation name"
	Version        string   "operation version"
	Status         Status   "operation status"
	Inputs         []Input  // trancode inputs
	Outputs        []Output // trancode outputs
	Functiongroups []FuncGroup
	Workflow       map[string]interface{}
	Firstfuncgroup string
}

type FuncGroup struct {
	Name              string     "function group name"
	Functions         []Function "functions"
	Executionsequence string     "functions execution sequence"
	Session           map[string]interface{}
	RouterDef         RouterDef
}

type RouterDef struct {
	Variable         string
	Vartype          string
	Values           []string
	Nextfuncgroups   []string
	Defaultfuncgroup string
}

type Function struct {
	Name     string       "function name"
	Version  string       "function version"
	Status   Status       "function status"
	Functype FunctionType "function type"
	Inputs   []Input      "function inputs"
	Outputs  []Output     "function	outputs"
	Content  string
}

type Input struct {
	Name         string      "input name"
	Source       InputSource "input type"      // 0: constant, 1: function, 2: session
	Datatype     DataType    "input data type" // 0: string, 1: int, 2: float, 3: bool, 4: datetime	5: object (json)
	Inivalue     string      "input initial value constant"
	Defaultvalue string      "input default value"
	Value        string      "input value"
	List         bool        "input is list"
	Aliasname    string      "input session"
}

type Output struct {
	Name         string       "output name"
	Outputdest   []OutputDest "output type"      // 0:none, 1:session, 2:engine
	Datatype     DataType     "output data type" // 0: string, 1: int, 2: float, 3: bool, 4: datetime	5: object (json)
	Inivalue     string       "output initial value constant"
	Defaultvalue string       "output default value"
	Value        string       "input value"
	List         bool         "output is list"
	Aliasname    []string     "input session"
}

type FunctionType int

const (
	InputMap FunctionType = iota
	Csharp
	Javascript
	Query
	StoreProcedure
	SubTranCode
)

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

type OutputDest int

const (
	None OutputDest = iota
	Tosession
	Toexternal
)
