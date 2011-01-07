package main

import (
    "mymy"
    "fmt"
    "os"
    "time"
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

    fmt.Print("Select...")
    buf := [1500]byte{}
    _, _, err := db.Query("select '%s'", buf[:])
    check(err)
    //fmt.Println("Result:", rows[0].Str(0))

    fmt.Print("Disconnect...")
    check(db.Close())
}
