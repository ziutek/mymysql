package mymysql

import (
    "os"
    "io"
    "net"
    "log"
    "time"
)

// Return true if error is network error or UnexpectedEOF.
func IsNetErr(err os.Error) bool {
    if err == io.ErrUnexpectedEOF {
        return true
    } else if _, ok := err.(*net.OpError); ok {
        return true
    }
    return false
}

func (my *MySQL) reconnectIfNetErr(nn *int, err *os.Error) {
    for *err != nil && IsNetErr(*err) && *nn <= my.MaxRetries {
        if my.Debug {
            log.Printf("Error: '%s' - reconnecting...", *err)
        }
        time.Sleep(int64(1e9) * int64(*nn))
        *err = my.Reconnect()
        if my.Debug && *err != nil {
            log.Println("Can't reconnect:", *err)
        }
        *nn++
    }
}


func (my *MySQL) connectIfNotConnected() (err os.Error) {
    if my.conn != nil {
        return
    }
    err = my.Connect()
    nn := 0
    my.reconnectIfNetErr(&nn, &err)
    return
}

// Automatic connect/reconnect/repeat version of Use
func (my *MySQL) UseAC(dbname string) (err os.Error) {
    if err = my.connectIfNotConnected(); err != nil {
        return
    }
    nn := 0
    for {
        if err = my.Use(dbname); err == nil {
            return
        }
        if my.reconnectIfNetErr(&nn, &err); err != nil {
            return
        }
    }
    return
}

// Automatic connect/reconnect/repeat version of Query
func (my *MySQL) QueryAC(sql string, params ...interface{}) (
        rows []*Row, res *Result, err os.Error) {

    if err = my.connectIfNotConnected(); err != nil {
        return
    }
    nn := 0
    for {
        if rows, res, err = my.Query(sql, params...); err == nil {
            return
        }
        if my.reconnectIfNetErr(&nn, &err); err != nil {
            return
        }
    }
    return
}

// Automatic connect/reconnect/repeat version of Prepare
func (my *MySQL) PrepareAC(sql string) (stmt *Statement, err os.Error) {
    if err = my.connectIfNotConnected(); err != nil {
        return
    }
    nn := 0
    for {
        if stmt, err = my.Prepare(sql); err == nil {
            return
        }
        if my.reconnectIfNetErr(&nn, &err); err != nil {
            return
        }
    }
    return
}

// Automatic connect/reconnect/repeat version of Exec
func (stmt *Statement) ExecAC(params ...interface{}) (
        rows []*Row, res *Result, err os.Error) {

    if err = stmt.db.connectIfNotConnected(); err != nil {
        return
    }
    nn := 0
    for {
        if rows, res, err = stmt.Exec(params...); err == nil {
            return
        }
        if stmt.db.reconnectIfNetErr(&nn, &err); err != nil {
            return
        }
    }
    return
}
