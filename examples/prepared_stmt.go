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

    fmt.Print("Prepare insert statement... ")
    ins, err := db.Prepare("insert A values (?, ?)")
    checkError(err)
    printOK()

    fmt.Print("Prepare select statement... ")
    sel, err := db.Prepare("select * from A where number > ?")
    checkError(err)
    printOK()

    params := struct {txt *string; number *int}{}

    fmt.Print("Bind insert parameters... ")
    ins.BindParams(&params)
    printOK()

    fmt.Print("Insert into A... ")
    for ii := 0; ii < 1000; ii += 100 {
        if ii % 500 == 0 {
            // Assign NULL values to the parameters
            params.txt    = nil
            params.number = nil
        } else {
            // Modify parameters
            str := fmt.Sprintf("%d*10= %d", ii / 100, ii / 10)
            params.txt = &str
            params.number = &ii
        }
        // Execute statement with modified data
        _, err = ins.Run()
        checkError(err)
    }
    printOK()

    fmt.Println("Select from A... ")
    rows, res := checkedResult(sel.Exec(0))
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
