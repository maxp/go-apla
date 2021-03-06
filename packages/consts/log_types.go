package consts

type LogEventType int

const (
	NetworkError             = "Network"
	JSONMarshallError        = "JSONMarshall"
	JSONUnmarshallError      = "JSONUnmarshall"
	CommandExecutionError    = "CommandExecution"
	ConvertionError          = "Convertion"
	TypeError                = "Type"
	ProtocolError            = "Protocol"
	MarshallingError         = "Marshall"
	UnmarshallingError       = "Unmarshall"
	ParseError               = "Parse"
	IOError                  = "IO"
	CryptoError              = "Crypto"
	ContractError            = "Contract"
	DBError                  = "DB"
	PanicRecoveredError      = "Panic"
	ConnectionError          = "Connection"
	ConfigError              = "Config"
	VMError                  = "VM"
	JustWaiting              = "JustWaiting"
	BlockError               = "Block"
	ParserError              = "Parser"
	ContextError             = "Context"
	SessionError             = "Session"
	RouteError               = "Route"
	NotFound                 = "NotFound"
	Found                    = "Found"
	EmptyObject              = "EmptyObject"
	InvalidObject            = "InvalidObject"
	DuplicateObject          = "DuplicateObject"
	UnknownObject            = "UnknownObject"
	ParameterExceeded        = "ParameterExceeded"
	DivisionByZero           = "DivisionByZero"
	EvalError                = "Eval"
	JWTError                 = "JWT"
	AccessDenied             = "AccessDenied"
	SizeDoesNotMatch         = "SizeDoesNotMatch"
	NoIndex                  = "NoIndex"
	NoFunds                  = "NoFunds"
	BlockIsFirst             = "BlockIsFirst"
	IncorrectCallingContract = "IncorrectCallingContract"
	WritingFile = "WritingFile"
)
