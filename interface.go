package mysql

type Conn interface {
	Connect() error
	Close() error
	Reconnect() error
	Ping() error
	Use(dbname string) error
	Start(sql string, params ...interface{}) (Result, error)
	Prepare(sql string) (Stmt, error)

	Status() uint16
	SetMaxPktSize(new_size int) int
}

type Stmt interface {
	Run(params ...interface{}) (Result, error)
	Delete() error
	Reset() error
	SendLongData(pnum int, data interface{}, pkt_size int) error

	Map() map[string]int
	FieldCount() int
	ParamCount() int
	WarningCount() int
	Status() uint16
}

type Result interface {
	GetRow() (Row, error)
	MoreResults() bool
	NextResult() (Result, error)

	Fields() []*Field
	Map() map[string]int
	Message() string
	AffectedRows() uint64
	InsertId() uint64
	WarningCount() int
	Status() uint16
}

var New func(proto, laddr, raddr, user, passwd string, db ...string) Conn
