package main

import (
    "os"
    "fmt"
    "mymy"
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

func checkResult(rows []*mymy.TextRow, res *mymy.Result, err os.Error) (
        []*mymy.TextRow, *mymy.Result) {
    checkError(err)
    return rows, res
}

func checkResNM(rows []*mymy.TextRow, res *mymy.Result, err os.Error) (
        []*mymy.TextRow, *mymy.Result) {
    if err != nil {
        if e, ok := err.(*mymy.Error); ok {
            // Error from MySQL server
            fmt.Println(e)
            return nil, nil
        }
    } else {
        printOK()
        return rows, res
    }
    // Other error
    return checkResult(rows, res, err)
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
    checkResNM(db.Query("drop table A"))

    fmt.Print("Create A table... ")
    checkResult(db.Query("create table A (txt varchar(40), number int)"))
    printOK()

    fmt.Print("Insert into A... ")
    for ii := 0; ii < 10; ii++ {
        if ii % 5 == 0 {
            checkResult(db.Query("insert A values (null, null)"))
        } else {
            checkResult(db.Queryf(
                "insert A values ('%d * 10 = %d', %d)", ii, ii*10, ii*100,
            ))
        }
    }
    printOK()

    fmt.Println("Select from A... ")
    rows, res := checkResult(db.Queryf("select * from A"))
    txt    := res.Map["name"]
    number := res.Map["number"]
    for _, row := range rows {
        fmt.Printf(
            "txt: {%s} '%s'  number: {%s} %d\n",
            row.Data[txt], row.Str(txt), row.Data[number], row.Int(number),
        )
    }

    fmt.Print("Remove A... ")
    //checkResult(db.Query("drop table A"))
    printOK()

    fmt.Print("Close connection... ")
    checkError(db.Close())
    printOK()
}
