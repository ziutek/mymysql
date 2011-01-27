package main

import (
    "fmt"
    "os"
    "time"
    "io"
    "net"
    "log"
    mymy "github.com/ziutek/mymysql"
)

type ReconDB struct {
    // MySQL handler
    my *mymy.MySQL

    // Maximum reconnect retries
    MaxRetries int
}

func NewRDB(proto, laddr, raddr, user, passwd string, db ...string) *ReconDB {
    return &ReconDB {
        my: mymy.New(proto, laddr, raddr, user, passwd, db...),
        MaxRetries: 6,
    }
}

func isNetErr(err os.Error) bool {
    if err == io.ErrUnexpectedEOF {
        // Probably network error
        return true
    } else if _, ok := err.(*net.OpError); ok {
        // Network error
        return true
    }
    return false
}

func (rdb *ReconDB) reconnect(err *os.Error) {
    for nn := 0; *err != nil && isNetErr(*err); nn++ {
        log.Println("Reconnecting...")
        time.Sleep(int64(1e9) * int64(nn))
        *err = rdb.my.Reconnect()
        if nn > rdb.MaxRetries {
            return
        }
        if *err != nil {
            log.Println("Can't reconnect:", *err)
        }
    }
}

func (rdb *ReconDB) Connect() (err os.Error) {
    err = rdb.my.Connect()
    if err != nil {
        log.Println("Can't connect:", err)
        rdb.reconnect(&err)
    }
    return
}

func (rdb *ReconDB) Close() (err os.Error) {
    return rdb.my.Close()
}

func (rdb *ReconDB) Query(sql string , params ...interface{}) (
        rows []*mymy.Row, res *mymy.Result, err os.Error) {
    for nn := 0; nn < rdb.MaxRetries; nn++ {
        rows, res, err = rdb.my.Query(sql, params...)
        if err == nil {
            break
        }
        log.Println("Query error:", err)
        rdb.reconnect(&err)
        if err != nil {
            // Can't reconnect
            break
        }
    }
    return
}

func check(err os.Error) {
    if err != nil {
        fmt.Println("", err)
        os.Exit(1)
    } else {
        fmt.Println(" OK")
    }
}

func main() {
    user   := "testuser"
    passwd := "TestPasswd9"
    dbname := "test"
    //conn := []string{"unix", "", "/var/run/mysqld/mysqld.sock"}
    conn   := []string{"tcp",  "", "127.0.0.1:3306"}

    db := NewRDB(conn[0], conn[1], conn[2], user, passwd, dbname)

    fmt.Print("Connect to MySQL...")
    check(db.Connect())

    sec := 9
    fmt.Println(
        "You may temporarily stop MySQL daemon or make network failure.",
    )
    fmt.Printf("Waiting %ds...", sec)
    for sec--; sec >= 0; sec-- {
        time.Sleep(1e9)
        fmt.Printf("\b\b\b\b\b%ds...", sec)
    }
    fmt.Println("\b\b\b.  ")

    data := "qwertyuiopasdfghjklzxcvbnm1234567890"

    fmt.Printf("Select (len=%d)...", len(data) + 9)

    rows, _, err := db.Query("select '%s'", data)
    check(err)

    fmt.Println("Result:", rows[0].Str(0))

    fmt.Print("Disconnect...")
    check(db.Close())
}
