package main

import (
    "mymy"
    "fmt"
    "os"
    "time"
    "io"
)

var (
    user   = "testuser"
    passwd = "TestPasswd9"
    dbname = "test"
    //conn   = []string{"unix", "", "/var/run/mysqld/mysqld.sock"}
    conn = []string{"tcp",  "", "127.0.0.1:3306"}
)

func check(err os.Error) {
    if err != nil {
        fmt.Println("", err)
        os.Exit(1)
    } else {
        fmt.Println(" OK")
    }
}

func main() {
    db := mymy.New(conn[0], conn[1], conn[2], user,  passwd, dbname)

    fmt.Print("Connect to MySQL...")
    check(db.Connect())

    sec := 9
    fmt.Printf("You may restart MySQL server. Waiting %ds...", sec)
    for sec--; sec >= 0; sec-- {
        time.Sleep(1e9)
        fmt.Printf("\b\b\b\b\b%ds...", sec)
    }
    fmt.Println("\b\b\b.  ")


loop:
    fmt.Print("Select...")
    rows, _, err := db.Query("select 'qwertyuiopasdfghjklzxcvbnm1234567890'")
    if ie, ok := err.(*io.Error); ok && ie == io.ErrUnexpectedEOF {
        fmt.Println(" Error:", ie)
        fmt.Print("Reconnecting...")
        check(db.Reconnect())
        goto loop
    }
    check(err)

    fmt.Println("Result:", rows[0].Str(0))

    fmt.Print("Disconnect...")
    check(db.Close())
}
