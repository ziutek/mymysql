package main

import (
    "os"
    "fmt"
    mymy "github.com/ziutek/mymysql"
)

func printOK() {
    fmt.Println("OK")
}

func checkError(err os.Error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func checkedResult(rows []*mymy.Row, res *mymy.Result, err os.Error) (
        []*mymy.Row, *mymy.Result) {
    checkError(err)
    return rows, res
}

func main() {
    user   := "testuser"
    pass   := "TestPasswd9"
    dbname := "test"
    //proto  := "unix"
    //addr   := "/var/run/mysqld/mysqld.sock"
    proto := "tcp"
    addr  := "127.0.0.1:3306"

    db := mymy.New(proto, "", addr, user, pass, dbname)
    //db.Debug = true

    fmt.Printf("Connect to %s:%s... ", proto, addr)
    checkError(db.Connect())
    printOK()

    fmt.Print("Drop A table if exists... ")
    _, err := db.Start("drop table A")
    if err == nil {
        printOK()
    } else if e, ok := err.(*mymy.Error); ok {
        // Error from MySQL server
        fmt.Println(e)
    } else {
        checkError(err)
    }

    fmt.Print("Create A table... ")
    checkedResult(db.Query("create table A (txt varchar(40), number int)"))
    printOK()

    fmt.Print("Insert into A... ")
    for ii := 0; ii < 10; ii++ {
        if ii % 5 == 0 {
            checkedResult(db.Query("insert A values (null, null)"))
        } else {
            checkedResult(db.Query(
                "insert A values ('%d*10= %d', %d)", ii, ii*10, ii*100,
            ))
        }
    }
    printOK()

    fmt.Println("Select from A... ")
    rows, res := checkedResult(db.Query("select * from A"))
    name   := res.Map["name"]
    number := res.Map["number"]
    for ii, row := range rows {
        fmt.Printf(
            "Row: %d\n name:  %-10s {%#v}\n number: %-8d  {%#v}\n", ii,
            "'" + row.Str(name) + "'", row.Data[name],
            row.Int(number), row.Data[number],
        )
    }

    fmt.Print("Remove A... ")
    checkedResult(db.Query("drop table A"))
    printOK()

    fmt.Print("Close connection... ")
    checkError(db.Close())
    printOK()
}
