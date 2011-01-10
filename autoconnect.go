package mymy

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

func (my *MySQL) reconnectIfNetErr(err *os.Error) {
    for nn := 0; *err != nil && IsNetErr(*err) && nn <= my.MaxRetries; nn++ {
        if my.Debug {
            log.Println("Reconnecting...")
        }
        time.Sleep(int64(1e9) * int64(nn))
        *err = my.Reconnect()
        if my.Debug && *err != nil {
            log.Println("Can't reconnect:", *err)
        }
    }
}


func (my *MySQL) connectIfNotConnected() (err os.Error) {
    if my.conn != nil {
        return
    }
    err = my.Connect()
    my.reconnectIfNetErr(&err)
    return
}

// Autoconnect/reconnect version of USE
func (my *MySQL) UseAC(dbname string) (err os.Error) {
    if err = my.connectIfNotConnected(); err != nil {
        return
    }
    err = my.Use(dbname)
    if err == nil {
        return
    }
    if my.reconnectIfNetErr(&err); err != nil {
        return
    }
    return my.Use(dbname)
}

// Autoconnect/reconnect version of Query
func (my *MySQL) QueryAC(command interface{}, params ...interface{}) (
        rows []*Row, res *Result, err os.Error) {

    if err = my.connectIfNotConnected(); err != nil {
        return
    }
    rows, res, err = my.Query(command, params...)
    if err == nil {
        return
    }
    if my.reconnectIfNetErr(&err); err != nil {
        return
    }
    return my.Query(command, params...)
}


