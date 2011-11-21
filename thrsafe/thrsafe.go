// MySQL Client API written entirely in Go without any external dependences.
// This is thread safe wrapper over native engine.
// See documentation of mymysql/native for details/
package thrsafe

import (
	"sync"
	//"log"
	"github.com/ziutek/mymysql/mysql"
	"github.com/ziutek/mymysql/native"
)

type Conn struct {
	*native.Conn
	mutex *sync.Mutex
}

func (c *Conn) lock() {
	//log.Println("lock")
	c.mutex.Lock()
}

func (c *Conn) unlock() {
	c.mutex.Unlock()
	//log.Println("unlock")
}

type Result struct {
	*native.Result
	conn *Conn
}

type Stmt struct {
	*native.Stmt
	conn *Conn
}

func New(proto, laddr, raddr, user, passwd string, db ...string) mysql.Conn {
	return &Conn{
		native.New(proto, laddr, raddr, user, passwd, db...).(*native.Conn),
		new(sync.Mutex),
	}
}

func (c *Conn) Connect() error {
	c.lock()
	defer c.unlock()
	return c.Conn.Connect()
}

func (c *Conn) Close() error {
	c.lock()
	defer c.unlock()
	return c.Conn.Close()
}

func (c *Conn) Reconnect() error {
	c.lock()
	defer c.unlock()
	return c.Conn.Reconnect()
}

func (c *Conn) Use(dbname string) error {
	c.lock()
	defer c.unlock()
	return c.Conn.Use(dbname)
}

func (c *Conn) Start(sql string, params ...interface{}) (mysql.Result, error) {
	c.lock()
	res, err := c.Conn.Start(sql, params...)
	if err != nil || len(res.Fields()) == 0 {
		// Unlock if error or OK result (which doesn't provide any fields)
		c.unlock()
	}
	return &Result{res.(*native.Result), c}, err
}

func (res *Result) GetRow() (mysql.Row, error) {
	row, err := res.Result.GetRow()
	if err != nil || row == nil && !res.MoreResults() {
		res.conn.unlock()
	}
	return row, err
}

func (res *Result) NextResult() (mysql.Result, error) {
	next, err := res.Result.NextResult()
	if err != nil {
		return nil, err
	}
	return &Result{next.(*native.Result), res.conn}, nil
}

func (c *Conn) Ping() error {
	c.lock()
	defer c.unlock()
	return c.Conn.Ping()
}

func (c *Conn) Prepare(sql string) (mysql.Stmt, error) {
	c.lock()
	defer c.unlock()
	stmt, err := c.Conn.Prepare(sql)
	if err != nil {
		return nil, err
	}
	return &Stmt{stmt.(*native.Stmt), c}, nil
}

func (stmt *Stmt) Run(params ...interface{}) (mysql.Result, error) {
	stmt.conn.lock()
	res, err := stmt.Stmt.Run()
	if err != nil {
		stmt.conn.unlock()
		return nil, err
	}
	return &Result{res.(*native.Result), stmt.conn}, nil
}

func (stmt *Stmt) Delete() error {
	stmt.conn.lock()
	defer stmt.conn.unlock()
	return stmt.Stmt.Delete()
}

func (stmt *Stmt) Reset() error {
	stmt.conn.lock()
	defer stmt.conn.unlock()

	return stmt.Stmt.Delete()
}

func (stmt *Stmt) SendLongData(pnum int, data interface{}, pkt_size int) error {
	stmt.conn.lock()
	defer stmt.conn.unlock()
	return stmt.Stmt.SendLongData(pnum, data, pkt_size)
}

func init() {
	mysql.New = New
}
