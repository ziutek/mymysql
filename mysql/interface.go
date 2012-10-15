// MySQL Client API written entirely in Go without any external dependences.
package mysql

type RegFunc func(my Conn)

type ConnCommon interface {
	Start(sql string, params ...interface{}) (Result, error)
	Prepare(sql string) (Stmt, error)

	Ping() error
	ThreadId() uint32
	EscapeString(txt string) string

	Query(sql string, params ...interface{}) ([]Row, Result, error)
	QueryFirst(sql string, params ...interface{}) (Row, Result, error)
	QueryLast(sql string, params ...interface{}) (Row, Result, error)
}

type Conn interface {
	ConnCommon

	Clone() Conn
	Connect() error
	Close() error
	IsConnected() bool
	Reconnect() error
	Use(dbname string) error
	Register(sql string)
        RegisterFunc(f RegFunc)
	SetMaxPktSize(new_size int) int

	Begin() (Transaction, error)
}

type Transaction interface {
	ConnCommon

	Commit() error
	Rollback() error
	Do(st Stmt) Stmt
	IsValid() bool
}

type Stmt interface {
	Bind(params ...interface{})
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
	ExecFirst(params ...interface{}) (Row, Result, error)
	ExecLast(params ...interface{}) (Row, Result, error)
}

type Result interface {
	StatusOnly() bool
	ScanRow(Row) error
	GetRow() (Row, error)

	MoreResults() bool
	NextResult() (Result, error)

	Fields() []*Field
	Map(string) int
	Message() string
	AffectedRows() uint64
	InsertId() uint64
	WarnCount() int

	MakeRow() Row
	GetRows() ([]Row, error)
	End() error
	GetFirstRow() (Row, error)
	GetLastRow() (Row, error)
}

var New func(proto, laddr, raddr, user, passwd string, db ...string) Conn
