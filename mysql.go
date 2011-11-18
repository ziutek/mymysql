package mysql

import (
	"io"
	"net"
	"bufio"
	"fmt"
	"reflect"
)

// Read all unreaded rows and discard them. This function is useful if you
// don't want to use the remaining rows. It has an impact only on current
// result. If there is multi result query, you must use NextResult method and
// read/discard all rows in this result, before use other method that sends
// data to the server.
func (res *Result) End() (err error) {
	var row native.Row
	for {
		row, err = res.GetRow();
		if err != nil || row == nil {
			break
		}
	}
	return
}


// This call Start and next call GetRow as long as it reads all rows from the
// result. Next it returns all readed rows as the slice of rows.
func (my *Conn) Query(sql string, params ...interface{}) (rows []Row, res *Result, err error) {

	res, err = my.Start(sql, params...)
	if err != nil {
		return
	}
	// Read rows
	var row Row
	for {
		row, err = res.GetRow()
		if err != nil || row == nil {
			break
		}
		rows = append(rows, row)
	}
	return
}

// This call Run and next call GetRow once or more times. It read all rows
// from connection and returns they as a slice.
func (stmt *Stmt) Exec(params ...interface{}) (rows []Row, res *Result, err error) {

	res, err = stmt.Run(params...)
	if err != nil {
		return
	}
	// Read rows
	var row Row
	for {
		row, err = res.GetRow()
		if err != nil || row == nil {
			break
		}
		rows = append(rows, row)
	}
	return
}
