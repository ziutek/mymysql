package godrv

import (
	"errors"
	"strings"
	"exp/sql"
	"exp/sql/driver"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
)

type conn struct {
	my mysql.Conn
}

func (c conn) Close() error {
	return c.my.Close()
}

func (c conn) Prepare(query string) (driver.Stmt, error) {
	st, err := c.my.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt{st}, nil
}

type stmt struct {
	my mysql.Stmt
}

func (s stmt) Close() error {
	return s.my.Delete()
}

func (s stmt) Close() error {
	return s.my.Delete()
}


type drv struct {
	// Defaults
	proto, laddr, raddr, user, passwd, db string
}

// Establish a connection using URI of following syntax:
//   DBNAME/USER/PASSWD
//   unix:SOCKPATH*DBNAME/USER/PASSWD
//   tcp:ADDR*DBNAME/USER/PASSWD
func (d *drv) Open(uri string) (driver.Conn, error) {
	var db string
	pd := strings.SplitN(uri, "*", 2)
	if len(pd) == 2 {
		// Parse protocol part of URI
		p := strings.SplitN(pd[0], ":", 2)
		if len(p) != 2 {
			return nil, errors.New("Wrong protocol part of URI")
		}
		d.proto = p[0]
		d.raddr = p[1]
		// Remove protocol part
		pd = pd[1:]
	}
	// Parse database part of URI
	dup := strings.SplitN(pd[0], "/", 3)
	if len(dup) != 3 {
		return nil, errors.New("Wrong database part of URI")
	}
	d.db = dup[0]
	d.user = dup[1]
	d.passwd = dup[2]

	// Establish the connection
	c := conn{mysql.New(d.proto, d.laddr, d.raddr, d.user, d.passwd, d.db)}
	if err := c.my.Connect(); err != nil {
		return nil, err
	}
	return c, nil
}


func init() {
	sql.Register("mymysql", &drv{proto: "tcp", raddr: "127.0.0.1:3306"})
}
