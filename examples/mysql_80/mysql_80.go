package main

import (
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
)

func main() {

	user := "testuser"
	pass := "TestPasswd9"
	dbname := "test"
	//proto  := "unix"
	//addr   := "/var/run/mysqld/mysqld.sock"
	proto := "tcp"
	addr := "127.0.0.1:3306"

	db := mysql.New(proto, "", addr, user, pass, dbname, "caching_sha2_password")

	err := db.Connect()
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
}
