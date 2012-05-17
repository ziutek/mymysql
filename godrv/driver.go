//MySQL driver for Go sql package
package godrv

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/mikespook/mymysql/mysql"
	"github.com/mikespook/mymysql/native"
	"io"
	"math"
	"net"
	"reflect"
	"strings"
	"time"
	"unsafe"
        "strconv"
)

var (
    ErrDSN = errors.New("Wrong URI")
    ErrMaxIdle = errors.New("Wrong max idle value")
)

type conn struct {
	my mysql.Conn
}

func errFilter(err error) error {
	if err == io.ErrUnexpectedEOF {
		return driver.ErrBadConn
	} else if e, ok := err.(net.Error); ok && e.Temporary() {
		return driver.ErrBadConn
	}
	return err
}

func (c conn) Prepare(query string) (driver.Stmt, error) {
	st, err := c.my.Prepare(query)
	if err != nil {
		return nil, errFilter(err)
	}
	return stmt{st}, nil
}

func (c conn) Close() error {
	err := c.my.Close()
	c.my = nil
	return errFilter(err)
}

func (c conn) Begin() (driver.Tx, error) {
	t, err := c.my.Begin()
	if err != nil {
		return tx{nil}, errFilter(err)
	}
	return tx{t}, nil
}

type tx struct {
	my mysql.Transaction
}

func (t tx) Commit() error {
	return t.my.Commit()
}

func (t tx) Rollback() error {
	return t.my.Rollback()
}

type stmt struct {
	my mysql.Stmt
}

func (s stmt) Close() error {
	err := s.my.Delete()
	s.my = nil
	return errFilter(err)
}

func (s stmt) NumInput() int {
	return s.my.NumParam()
}

func (s stmt) run(args []driver.Value) (*rowsRes, error) {
	a := (*[]interface{})(unsafe.Pointer(&args))
	res, err := s.my.Run(*a...)
	if err != nil {
		return nil, errFilter(err)
	}
	return &rowsRes{res, res.MakeRow()}, nil
}

func (s stmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.run(args)
}

func (s stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.run(args)
}

type rowsRes struct {
	my  mysql.Result
	row mysql.Row
}

func (r rowsRes) LastInsertId() (int64, error) {
	return int64(r.my.InsertId()), nil
}

func (r rowsRes) RowsAffected() (int64, error) {
	return int64(r.my.AffectedRows()), nil
}

func (r rowsRes) Columns() []string {
	flds := r.my.Fields()
	cls := make([]string, len(flds))
	for i, f := range flds {
		cls[i] = f.Name
	}
	return cls
}

func (r rowsRes) Close() error {
	err := r.my.End()
	r.my = nil
	r.row = nil
	if err != native.READ_AFTER_EOR_ERROR {
		return errFilter(err)
	}
	return nil
}

// DATE, DATETIME, TIMESTAMP are treated as they are in Local time zone
func (r rowsRes) Next(dest []driver.Value) error {
	err := r.my.ScanRow(r.row)
	if err != nil {
		return errFilter(err)
	}
	for i, col := range r.row {
		if col == nil {
			dest[i] = nil
			continue
		}
		switch c := col.(type) {
		case time.Time:
			dest[i] = c
			continue
		case mysql.Timestamp:
			dest[i] = c.Time
			continue
		case mysql.Date:
			dest[i] = c.Localtime()
			continue
		}
		v := reflect.ValueOf(col)
		switch v.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// this contains time.Duration to
			dest[i] = v.Int()
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u := v.Uint()
			if u > math.MaxInt64 {
				panic("Value to large for int64 type")
			}
			dest[i] = int64(u)
		case reflect.Float32, reflect.Float64:
			dest[i] = v.Float()
		case reflect.Slice:
			if v.Type().Elem().Kind() == reflect.Uint8 {
				dest[i] = v.Interface().([]byte)
				break
			}
			fallthrough
		default:
			panic(fmt.Sprint("Unknown type of column: ", v.Type()))
		}
	}
	return nil
}

type Driver struct {
	// Defaults
	proto, laddr, raddr, user, passwd, db string

	initCmds []string
        initFuncs []mysql.RegFunc
}

// Open new connection. The uri need to have the following syntax:
//   [tcp://addr/]dbname/user/password[?params]
//   [unix://sockpath/]dbname/user/password[?params]
// 
// Params need to have the following syntax:
//   key1=val1&key2=val2
//
// Key need to have the following value:
//   charset - used by 'set names'
//   keepalive - send a PING to mysql server after every keepalive seconds.
//
// where protocol spercific part may be empty (this means connection to
// local server using default protocol). Currently possible forms:
//   DBNAME/USER/PASSWD?charset=utf8
//   unix://SOCKPATH/DBNAME/USER/PASSWD
//   tcp://ADDR/DBNAME/USER/PASSWD?maxidle=3600
//
// If a password contains the slashes (/), use a star (*) to repleace it.
// If a password contains the star (*), use double stars (**) to repleace it.
//   pass/wd => pass*wd
//   pass*wd => pass**wd
func (d *Driver) Open(uri string) (driver.Conn, error) {
        proto, addr, dbname, user, passwd, params, err := parseDSN(uri)
	if err != nil {
	    return nil, err
        }
	d.proto = proto
        d.raddr = addr
        d.user = user
	d.passwd = passwd
        d.db = dbname

	// Establish the connection
	c := conn{mysql.New(d.proto, d.laddr, d.raddr, d.user, d.passwd, d.db)}

        if v, ok := params["charset"]; ok {
            Register("SET NAMES " + v)
        }
        if v, ok := params["keepalive"]; ok {
            t, err := strconv.Atoi(v)
            if err != nil {
                return nil, ErrMaxIdle
            }
            RegisterFunc(func(my mysql.Conn){
                go func() {
                    for my.IsConnected() {
                        time.Sleep(time.Duration(t) * time.Second)
                        if err := my.Ping(); err != nil {
                            break
                        }
                    }
                }()
            })
        }
	for _, q := range d.initCmds {
		c.my.Register(q) // Register initialisation commands
	}
        for _, f := range d.initFuncs {
                c.my.RegisterFunc(f)
        }
	if err := c.my.Connect(); err != nil {
		return nil, errFilter(err)
	}
	return &c, nil
}

func parseDSN(uri string) (proto, addr, dbname, user, passwd string, params map[string]string, err error) {
    proto = "tcp"; addr = "127.0.0.1:3306"
    // [tcp:addr/]dbname/user/password[?params]
    s := strings.SplitN(uri, "?", 2)
    // dsn and params
    if len(s) == 2 {
        uri = s[0]
        params = parseParams(s[1])
    }
    s = strings.SplitN(uri, "://", 2)
    hasProto := (len(s) == 2)
    if hasProto {
        proto = s[0]
        uri = s[1]
    }
    s = strings.SplitN(uri, "/", 4)
    switch(len(s)) {
    case 1:
        dbname = s[0]
    case 2:
        dbname = s[0]
        user = s[1]
    case 3:
        if hasProto {
            if strings.Contains(s[0], ":") {
                addr = s[0]
            } else {
                addr = s[0] + ":3306"
            }
            dbname = s[1]
            user = s[2]
        } else {
            dbname = s[0]
            user = s[1]
            passwd = strings.Replace(s[2], "*", "/", -1)
            passwd = strings.Replace(passwd, "//", "*", -1)
        }
    // protocol has been specifieded.
    case 4 :
        if strings.Contains(s[0], ":") {
            addr = s[0]
        } else {
            addr = s[0] + ":3306"
        }
        dbname = s[1]
        user = s[2]
        passwd = strings.Replace(s[3], "*", "/", -1)
        passwd = strings.Replace(passwd, "//", "*", -1)
    //
    default:
        err = ErrDSN
        return
   }
   return
}

func parseParams(str string) (params map[string]string) {
    params = make(map[string]string, 2)
    s := strings.Split(str, "&")
    for _, v := range s {
        p := strings.SplitN(v, "=", 2)
        if len(p) != 2 {
            continue
        }
        params[p[0]] = p[1]
    }
    return params
}

// Driver automatically registered in database/sql
var d = Driver{proto: "tcp", raddr: "127.0.0.1:3306"}

// Registers initialisation commands.
// This is workaround, see http://codereview.appspot.com/5706047
func Register(query string) {
	d.initCmds = append(d.initCmds, query)
}

// Registers initialisation functions.
func RegisterFunc(f mysql.RegFunc) {
	d.initFuncs = append(d.initFuncs, f)
}


func init() {
	sql.Register("mymysql", &d)
}
