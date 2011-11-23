// MySQL Client API written entirely in Go without any external dependences.
package mysql

type conn interface {
	Connect() error
	IsConnected() bool
	Close() error
	Reconnect() error
	Ping() error
	Use(dbname string) error
	Start(sql string, params ...interface{}) (Result, error)
	Prepare(sql string) (Stmt, error)

	ThreadId() uint32
	Register(sql string)
	EscapeString(txt string) string
	SetMaxPktSize(new_size int) int

	Query(sql string, params ...interface{}) ([]Row, Result, error)
}

type Conn interface {
	conn
	Begin() (Transaction, error)
}

type Transaction interface {
	conn
	Commit() error
	Rollback() error
	Do(st Stmt) Stmt
}

type Stmt interface {
	BindParams(params ...interface{})
	ResetParams()
	Run(params ...interface{}) (Result, error)
	Delete() error
	Reset() error
	SendLongData(pnum int, data interface{}, pkt_size int) error

	Map(string) int
	NumField() int
	NumParam() int
	WarnCount() int

	Exec(params ...interface{}) ([]Row, Result, error)
}

type Result interface {
	GetRow() (Row, error)
	MoreResults() bool
	NextResult() (Result, error)

	Fields() []*Field
	Map(string) int
	Message() string
	AffectedRows() uint64
	InsertId() uint64
	WarnCount() int

	GetRows() ([]Row, error)
	End() error
}

var New func(proto, laddr, raddr, user, passwd string, db ...string) Conn
