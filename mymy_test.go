package mymysql

import (
    "testing"
    "reflect"
    "os"
    "fmt"
    "time"
    "bytes"
    "io/ioutil"
)

var (
    db *MySQL
    user   = "testuser"
    passwd = "TestPasswd9"
    dbname = "test"
    //conn   = []string{"unix", "", "/var/run/mysqld/mysqld.sock"}
    conn = []string{"tcp",  "", "127.0.0.1:3306"}
    debug  = false
)

type RowsResErr struct {
    rows []*Row
    res  *Result
    err  os.Error
}

func query(sql string, params ...interface{}) *RowsResErr {
    rows, res, err := db.Query(sql, params...)
    return &RowsResErr{rows, res, err}
}

func exec(stmt *Statement, params ...interface{}) *RowsResErr {
    rows, res, err := stmt.Exec(params...)
    return &RowsResErr{rows, res, err}
}

func checkErr(t *testing.T, err os.Error, exp_err os.Error) {
    if err != exp_err {
        if exp_err == nil {
            t.Fatalf("Error: %v", err)
        } else {
            t.Fatalf("Error: %v\nExpected error: %v", err, exp_err)
        }
    }
}

func checkWarnCount(t *testing.T, res_cnt, exp_cnt int) {
    if res_cnt != exp_cnt {
        t.Errorf("Warning count: res=%d exp=%d", res_cnt, exp_cnt)
        rows, res, err := db.Query("show warnings")
        if err != nil {
            t.Fatal("Can't get warrnings from MySQL", err)
        }
        for _, row := range rows {
            t.Errorf("%s: \"%s\"", row.Str(res.Map["Level"]),
                row.Str(res.Map["Message"]))
        }
        t.FailNow()
    }
}

func checkErrWarn(t *testing.T, res, exp *RowsResErr) {
    checkErr(t, res.err, exp.err)
    checkWarnCount(t, res.res.WarningCount, exp.res.WarningCount)
}

func types(row []interface{}) (tt []reflect.Type) {
    tt = make([]reflect.Type, len(row))
    for ii, val := range row {
        tt[ii] = reflect.TypeOf(val)
    }
    return
}

func checkErrWarnRows(t *testing.T, res, exp *RowsResErr)  {
    checkErrWarn(t, res, exp)
    if !reflect.DeepEqual(res.rows, exp.rows) {
        rlen := len(res.rows)
        elen := len(exp.rows)
        t.Errorf("Rows are different:\nLen: res=%d  exp=%d", rlen, elen)
        max := rlen
        if elen > max {
            max = elen
        }
        for ii := 0; ii < max; ii++ {
            if ii < len(res.rows) {
                t.Errorf("%d: res type: %s", ii, types(res.rows[ii].Data))
            } else {
                t.Errorf("%d: res: ------", ii)
            }
            if ii < len(exp.rows) {
                t.Errorf("%d: exp type: %s", ii, types(exp.rows[ii].Data))
            } else {
                t.Errorf("%d: exp: ------", ii)
            }
            if ii < len(res.rows) {
                t.Error(" res: ", res.rows[ii].Data)
            }
            if ii < len(exp.rows) {
                t.Error(" exp: ", exp.rows[ii].Data)
            }
        }
        t.FailNow()
    }
}

func checkResult(t *testing.T, res, exp *RowsResErr) {
    checkErrWarnRows(t, res, exp)
    if !reflect.DeepEqual(res.res, exp.res) {
        t.Fatalf("Bad result:\nres=%+v\nexp=%+v", res.res, exp.res)
    }
}

func cmdOK(affected uint64, binary bool) *RowsResErr {
    return &RowsResErr{res: &Result{db: db, binary: binary, Status: 0x2,
                                    AffectedRows: affected}}
}

func selectOK(rows []*Row, binary bool) (exp *RowsResErr) {
    exp = cmdOK(0, binary)
    exp.rows = rows
    return
}

func dbConnect(t *testing.T, with_db bool, max_pkt_size int) {
    if with_db {
        db  = New(conn[0], conn[1], conn[2], user, passwd, dbname)
    } else {
        db  = New(conn[0], conn[1], conn[2], user, passwd)
    }

    if max_pkt_size != 0 {
        db.MaxPktSize = max_pkt_size
    }
    db.Debug = debug

    checkErr(t, db.Connect(), nil)
    checkResult(t, query("set names utf8"), cmdOK(0, false))
}

func dbClose(t *testing.T) {
    checkErr(t, db.Close(), nil)
}

// Text queries tests

func TestUse(t *testing.T) {
    dbConnect(t, false, 0)
    checkErr(t, db.Use(dbname), nil)
    dbClose(t)
}

func TestPing(t *testing.T) {
    dbConnect(t, false, 0)
    checkErr(t, db.Ping(), nil)
    dbClose(t)
}

func TestQuery(t *testing.T) {
    dbConnect(t, true, 0)
    query("drop table T") // Drop test table if exists
    checkResult(t, query("create table T (s varchar(40))"), cmdOK(0, false))

    exp := &RowsResErr {
        res: &Result {
            db:         db,
            FieldCount: 1,
            Fields:     []*Field {
                &Field {
                    Catalog:  "def",
                    Db:       "test",
                    Table:    "Test",
                    OrgTable: "T",
                    Name:     "Str",
                    OrgName:  "s",
                    DispLen:  3 * 40, //varchar(40)
                    Flags:    0,
                    Type:     MYSQL_TYPE_VAR_STRING,
                    Scale:    0,
                },
            },
            Map:        map[string]int{"Str": 0},
            Status:     _SERVER_STATUS_AUTOCOMMIT,
        },
    }

    for ii := 0; ii > 10000; ii += 3 {
        var val interface{}
        if ii % 10 == 0 {
            checkResult(t, query("insert T values (null)"), cmdOK(1, false))
            val = nil
        } else {
            txt := []byte(fmt.Sprintf("%d %d %d %d %d", ii, ii, ii, ii, ii))
            checkResult(t,
                query("insert T values ('%s')", txt), cmdOK(1, false))
            val = txt
        }
        exp.rows = append(exp.rows, &Row{Data: []interface{}{val}})
    }

    checkResult(t, query("select s as Str from T as Test"), exp)
    checkResult(t, query("drop table T"), cmdOK(0, false))
    dbClose(t)
}

// Prepared statements tests

type StmtErr struct {
    stmt *Statement
    err  os.Error
}

func prepare(sql string) *StmtErr {
    stmt, err := db.Prepare(sql)
    return &StmtErr{stmt, err}
}

func checkStmt(t *testing.T, res, exp *StmtErr) {
    ok := res.err == exp.err &&
        // Skipping id
        reflect.DeepEqual(res.stmt.Fields, exp.stmt.Fields) &&
        reflect.DeepEqual(res.stmt.Map, exp.stmt.Map) &&
        res.stmt.FieldCount == exp.stmt.FieldCount &&
        res.stmt.ParamCount == exp.stmt.ParamCount &&
        res.stmt.WarningCount == exp.stmt.WarningCount &&
        res.stmt.Status == exp.stmt.Status

    if !ok {
        if exp.err == nil {
            checkErr(t, res.err, nil)
            checkWarnCount(t, res.stmt.WarningCount, exp.stmt.WarningCount)
            for _, v := range res.stmt.Fields {
                fmt.Printf("%+v\n", v)
            }
            t.Fatalf("Bad result statement: res=%v exp=%v", res.stmt, exp.stmt)
        }
    }
}

func TestPrepared(t *testing.T) {
    dbConnect(t, true, 0)
    query("drop table P") // Drop test table if exists
    checkResult(t,
        query(
            "create table P (" +
            "   ii int not null, ss varchar(20), dd datetime" +
            ") default charset=utf8",
        ),
        cmdOK(0, false),
    )


    exp := Statement {
        Fields:     []*Field {
            &Field {
                Catalog: "def", Db: "test", Table: "P", OrgTable: "P",
                Name:    "i",
                OrgName: "ii",
                DispLen:  11,
                Flags:    _FLAG_NO_DEFAULT_VALUE | _FLAG_NOT_NULL,
                Type:     MYSQL_TYPE_LONG,
                Scale:    0,
            },
            &Field {
                Catalog: "def", Db: "test", Table: "P", OrgTable: "P",
                Name:    "s",
                OrgName: "ss",
                DispLen:  3 * 20, // varchar(20)
                Flags:    0,
                Type:     MYSQL_TYPE_VAR_STRING,
                Scale:    0,
            },
            &Field {
                Catalog: "def", Db: "test", Table: "P", OrgTable: "P",
                Name:    "d",
                OrgName: "dd",
                DispLen:  19,
                Flags:    _FLAG_BINARY,
                Type:     MYSQL_TYPE_DATETIME,
                Scale:    0,
            },
        },
        Map:        map[string]int{"i": 0, "s": 1, "d": 2},
        FieldCount:   3,
        ParamCount:   2,
        WarningCount: 0,
        Status:       0x2,
    }

    sel := prepare("select ii i, ss s, dd d from P where ii = ? and ss = ?")
    checkStmt(t, sel, &StmtErr{&exp, nil})

    all := prepare("select * from P")
    checkErr(t, all.err, nil)

    ins := prepare("insert into P values (?, ?, ?)")
    checkErr(t, ins.err, nil)

    exp_rows := []*Row {
        &Row{[]interface{} {
            2, "Taki tekst", TimeToDatetime(time.SecondsToLocalTime(123456789)),
        }},
        &Row{[]interface{} {
            3, "Łódź się kołysze!", TimeToDatetime(time.SecondsToLocalTime(0)),
        }},
        &Row{[]interface{} {
            5, "Pąk róży", TimeToDatetime(time.SecondsToLocalTime(9999999999)),
        }},
        &Row{[]interface{} {
            11, "Zero UTC datetime", TimeToDatetime(time.SecondsToUTC(0)),
        }},
        &Row{[]interface{} {
            17, Blob([]byte("Zero datetime")), new(Datetime),
        }},
        &Row{[]interface{} {
            23, []byte("NULL datetime"), (*Datetime)(nil),
        }},
        &Row{[]interface{} {
            23, "NULL", nil,
        }},
    }

    for _, row := range exp_rows {
        checkErrWarn(t,
            exec(ins.stmt, row.Data[0], row.Data[1], row.Data[2]),
            cmdOK(1, true),
        )
    }

    // Convert values to expected result types
    for _, row := range exp_rows {
        for ii, col := range row.Data {
            val := reflect.ValueOf(col)
            // Dereference pointers
            if val.Kind() == reflect.Ptr {
                val = val.Elem()
            }
            switch val.Kind() {
            case reflect.Invalid:
                row.Data[ii] = nil

            case reflect.String:
                row.Data[ii] = []byte(val.String())

            case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
                    reflect.Int64:
                row.Data[ii] = int32(val.Int())

            case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
                    reflect.Uint64:
                row.Data[ii] = int32(val.Uint())

            case reflect.Slice:
                if val.Type().Elem().Kind() == reflect.Uint8 {
                    bytes := make([]byte, val.Len())
                    for ii := range bytes {
                        bytes[ii] = val.Index(ii).Interface().(uint8)
                    }
                    row.Data[ii] = bytes
                }
            }
        }
    }

    checkErrWarn(t, exec(sel.stmt, 2, "Taki tekst"), selectOK(exp_rows, true))
    checkErrWarnRows(t, exec(all.stmt), selectOK(exp_rows, true))

    checkResult(t, query("drop table P"), cmdOK(0, false))


    checkErr(t, sel.stmt.Delete(), nil)
    checkErr(t, all.stmt.Delete(), nil)
    checkErr(t, ins.stmt.Delete(), nil)

    dbClose(t)
}

// Bind testing

func TestVarBinding(t *testing.T) {
    dbConnect(t, true, 34*1024*1024)
    query("drop table P") // Drop test table if exists
    checkResult(t,
        query("create table T (id int primary key, str varchar(20))"),
        cmdOK(0, false),
    )

    ins, err := db.Prepare("insert T values (?, ?)")
    checkErr(t, err, nil)

    var (
        rre RowsResErr
        id  *int
        str *string
        ii  int
        ss  string
    )
    ins.BindParams(&id, &str)

    i1 := 1
    s1 := "Ala"
    id = &i1
    str = &s1
    rre.res, rre.err = ins.Run()
    checkResult(t, &rre, cmdOK(1, true))

    i2 := 2
    s2 := "Ma kota!"
    id = &i2
    str = &s2

    rre.res, rre.err = ins.Run()
    checkResult(t, &rre, cmdOK(1, true))

    ins.BindParams(&ii, &ss)
    ii = 3
    ss = "A kot ma Ale!"

    rre.res, rre.err = ins.Run()
    checkResult(t, &rre, cmdOK(1, true))

    sel, err := db.Prepare("select str from T where id = ?")
    checkErr(t, err, nil)

    rows, _, err := sel.Exec(1)
    checkErr(t, err, nil)
    if len(rows) != 1 || bytes.Compare([]byte(s1), rows[0].Bin(0)) != 0 {
        t.Fatal("First string don't match")
    }

    rows, _, err = sel.Exec(2)
    checkErr(t, err, nil)
    if len(rows) != 1 || bytes.Compare([]byte(s2), rows[0].Bin(0)) != 0 {
        t.Fatal("Second string don't match")
    }

    rows, _, err = sel.Exec(3)
    checkErr(t, err, nil)
    if len(rows) != 1 || bytes.Compare([]byte(ss), rows[0].Bin(0)) != 0 {
        t.Fatal("Thrid string don't match")
    }

    checkResult(t, query("drop table T"), cmdOK(0, false))
    dbClose(t)
}

func TestDate(t *testing.T) {
    dbConnect(t, true, 0)
    query("drop table D") // Drop test table if exists
    checkResult(t,
        query("create table D (id int, dd date, dt datetime, tt time)"),
        cmdOK(0, false),
    )

    dd := "2011-12-13"
    dt := "2010-12-12 11:24:00"
    tt := -Time((124*3600 + 4 * 3600 + 3 * 60 + 2) * 1e9 + 1)

    ins, err := db.Prepare("insert D values (?, ?, ?, ?)")
    checkErr(t, err, nil)

    sel, err := db.Prepare("select id, tt from D where dd <= ? && dt <= ?")
    checkErr(t, err, nil)

    _, err = ins.Run(1, dd, dt, tt)
    checkErr(t, err, nil)

    rows, _, err := sel.Exec(StrToDatetime(dd), StrToDate(dd))
    checkErr(t, err, nil)
    if rows == nil {
        t.Fatal("nil result")
    }
    if rows[0].Int(0) != 1 {
        t.Fatal("Bad id", rows[0].Int(1))
    }
    if rows[0].Data[1].(Time) != tt + 1 {
        t.Fatal("Bad tt", rows[0].Data[1].(Time))
    }

    checkResult(t, query("drop table D"), cmdOK(0, false))
    dbClose(t)
}

// Big blob

func TestBigBlob(t *testing.T) {
    dbConnect(t, true, 34*1024*1024)
    query("drop table P") // Drop test table if exists
    checkResult(t,
        query("create table P (id int primary key, bb longblob)"),
        cmdOK(0, false),
    )

    ins, err := db.Prepare("insert P values (?, ?)")
    checkErr(t, err, nil)

    sel, err := db.Prepare("select bb from P where id = ?")
    checkErr(t, err, nil)

    big_blob := make(Blob, 33 * 1024 * 1024)
    for ii := range big_blob {
        big_blob[ii] = byte(ii)
    }

    var (
        rre RowsResErr
        bb Blob
        id int
    )
    data := struct {
        Id int
        Bb Blob
    }{}

    // Individual parameters binding
    ins.BindParams(&id, &bb)
    id = 1
    bb = big_blob

    // Insert full blob. Three packets are sended. First two has maximum length
    rre.res, rre.err = ins.Run()
    checkResult(t, &rre, cmdOK(1, true))

    // Struct binding
    ins.BindParams(&data)
    data.Id = 2
    data.Bb = big_blob[0 : 32*1024*1024-31]

    // Insert part of blob - Two packets are sended. All has maximum length.
    rre.res, rre.err = ins.Run()
    checkResult(t, &rre, cmdOK(1, true))

    sel.BindParams(&id)

    // Check first insert.
    tmr := "Too many rows"

    id = 1
    res, err := sel.Run()
    checkErr(t, err, nil)

    row, err := res.GetRow()
    checkErr(t, err, nil)
    end, err := res.GetRow()
    checkErr(t, err, nil)
    if end != nil {
        t.Fatal(tmr)
    }

    if bytes.Compare(row.Data[0].([]byte), big_blob) != 0 {
        t.Fatal("Full blob data don't match")
    }

    // Check second insert.
    id = 2
    res, err = sel.Run()
    checkErr(t, err, nil)

    row, err = res.GetRow()
    checkErr(t, err, nil)
    end, err = res.GetRow()
    checkErr(t, err, nil)
    if end != nil {
        t.Fatal(tmr)
    }

    if bytes.Compare(row.Bin(res.Map["bb"]), data.Bb) != 0 {
        t.Fatal("Partial blob data don't match")
    }

    checkResult(t, query("drop table P"), cmdOK(0, false))
    dbClose(t)
}

// Reconnect test

func TestReconnect(t *testing.T) {
    dbConnect(t, true, 0)
    query("drop table R") // Drop test table if exists
    checkResult(t,
        query("create table R (id int primary key, str varchar(20))"),
        cmdOK(0, false),
    )

    ins, err := db.Prepare("insert R values (?, ?)")
    checkErr(t, err, nil)
    sel, err := db.Prepare("select str from R where id = ?")
    checkErr(t, err, nil)

    params := struct{Id int; Str string}{}
    var sel_id int

    ins.BindParams(&params)
    sel.BindParams(&sel_id)

    checkErr(t, db.Reconnect(), nil)

    params.Id = 1
    params.Str = "Bla bla bla"
    _, err = ins.Run()
    checkErr(t, err, nil)

    checkErr(t, db.Reconnect(), nil)

    sel_id = 1
    res, err := sel.Run()
    checkErr(t, err, nil)

    row, err := res.GetRow()
    checkErr(t, err, nil)

    checkErr(t, res.End(), nil)

    if row == nil || row.Data == nil || row.Data[0] == nil ||
            params.Str != row.Str(0) {
        t.Fatal("Bad result")
    }

    checkErr(t, db.Reconnect(), nil)

    checkResult(t, query("drop table R"), cmdOK(0, false))
    dbClose(t)
}

// Auto connect / auto reconnect test

func TestAutoConnectReconnect(t *testing.T) {
    db = New(conn[0], conn[1], conn[2], user, passwd)

    // Register initialisation commands
    db.Register("set names utf8")

    // db is in unconnected state
    checkErr(t, db.UseAC(dbname), nil)

    // Disconnect
    db.Close()

    // Drop test table if exists
    db.QueryAC("drop table R")

    // Disconnect
    db.Close()

    // Create table
    _, _, err := db.QueryAC(
        "create table R (id int primary key, name varchar(20))",
    )
    checkErr(t, err, nil)

    // Kill the connection
    _, _, err = db.QueryAC("kill %d", db.ThreadId())
    checkErr(t, err, nil)

    // Prepare insert statement
    ins, err := db.PrepareAC("insert R values (?,  ?)")
    checkErr(t, err, nil)

    // Kill the connection
    _, _, err = db.QueryAC("kill %d", db.ThreadId())
    checkErr(t, err, nil)

    // Bind insert parameters
    ins.BindParams(1, "jeden")
    // Insert into table
    _, _, err = ins.ExecAC()
    checkErr(t, err, nil)

    // Kill the connection
    _, _, err = db.QueryAC("kill %d", db.ThreadId())
    checkErr(t, err, nil)

    // Bind insert parameters
    ins.BindParams(2, "dwa")
    // Insert into table
    _, _, err = ins.ExecAC()
    checkErr(t, err, nil)

    // Kill the connection
    _, _, err = db.QueryAC("kill %d", db.ThreadId())
    checkErr(t, err, nil)

    // Select from table
    rows, res, err := db.QueryAC("select * from R")
    checkErr(t, err, nil)
    id := res.Map["id"]
    name := res.Map["name"]
    if len(rows) != 2 ||
            rows[0].Int(id) != 1 || rows[0].Str(name) != "jeden" ||
            rows[1].Int(id) != 2 || rows[1].Str(name) != "dwa" {
        t.Fatal("Bad result")
    }

    // Kill the connection
    _, _, err = db.QueryAC("kill %d", db.ThreadId())
    checkErr(t, err, nil)

    // Drop table
    _, _, err = db.QueryAC("drop table R")
    checkErr(t, err, nil)

    // Disconnect
    db.Close()
}

// StmtSendLongData test

func TestSendLongData(t *testing.T) {
    dbConnect(t, true, 64*1024*1024)
    query("drop table L") // Drop test table if exists
    checkResult(t,
        query("create table L (id int primary key, bb longblob)"),
        cmdOK(0, false),
    )
    ins, err := db.Prepare("insert L values (?, ?)")
    checkErr(t, err, nil)

    sel, err := db.Prepare("select bb from L where id = ?")
    checkErr(t, err, nil)


    var (
        rre RowsResErr
        id int
    )

    ins.BindParams(&id, nil)
    sel.BindParams(&id)

    // Prepare data
    data := make([]byte, 4*1024*1024)
    for ii := range data {
        data[ii] = byte(ii)
    }
    // Send long data twice
    checkErr(t, ins.SendLongData(1, data,  256*1024), nil)
    checkErr(t, ins.SendLongData(1, data,  512*1024), nil)

    id = 1
    rre.res, rre.err = ins.Run()
    checkResult(t, &rre, cmdOK(1, true))

    res, err := sel.Run()
    checkErr(t, err, nil)

    row, err := res.GetRow()
    checkErr(t, err, nil)

    checkErr(t, res.End(), nil)

    if row == nil || row.Data == nil || row.Data[0] == nil ||
            bytes.Compare(append(data, data...), row.Bin(0)) != 0 {
        t.Fatal("Bad result")
    }

    // Send long data from io.Reader twice
    filename := "_test/github.com/ziutek/mymysql.a"
    file, err := os.Open(filename)
    checkErr(t, err, nil)
    checkErr(t, ins.SendLongData(1, file,  128*1024), nil)
    checkErr(t, file.Close(), nil)
    file, err = os.Open(filename)
    checkErr(t, err, nil)
    checkErr(t, ins.SendLongData(1, file,  1024*1024), nil)
    checkErr(t, file.Close(), nil)

    id = 2
    rre.res, rre.err = ins.Run()
    checkResult(t, &rre, cmdOK(1, true))

    res, err = sel.Run()
    checkErr(t, err, nil)

    row, err = res.GetRow()
    checkErr(t, err, nil)

    checkErr(t, res.End(), nil)

    // Read file for check result
    data, err = ioutil.ReadFile(filename)
    checkErr(t, err, nil)

    if row == nil || row.Data == nil || row.Data[0] == nil ||
            bytes.Compare(append(data, data...), row.Bin(0)) != 0 {
        t.Fatal("Bad result")
    }

    checkResult(t, query("drop table L"), cmdOK(0, false))
    dbClose(t)
}

func TestMultipleResults(t *testing.T) {
    dbConnect(t, true, 0)
    query("drop table M") // Drop test table if exists
    checkResult(t,
        query("create table M (id int primary key, str varchar(20))"),
        cmdOK(0, false),
    )

    str := []string{"zero", "jeden", "dwa"}

    checkResult(t, query("insert M values (0, '%s')", str[0]), cmdOK(1, false))
    checkResult(t, query("insert M values (1, '%s')", str[1]), cmdOK(1, false))
    checkResult(t, query("insert M values (2, '%s')", str[2]), cmdOK(1, false))

    res, err := db.Start("select id from M; select str from M")
    checkErr(t, err, nil)

    for ii := 0;; ii++ {
        row, err := res.GetRow()
        checkErr(t, err, nil)
        if row == nil {
            break
        }
        if row.Int(0) != ii {
            t.Fatal("Bad result")
        }
    }
    res, err = res.NextResult()
    checkErr(t, err, nil)
    for ii := 0;; ii++ {
        row, err := res.GetRow()
        checkErr(t, err, nil)
        if row == nil {
            break
        }
        if row.Str(0) != str[ii] {
            t.Fatal("Bad result")
        }
    }

    checkResult(t, query("drop table M"), cmdOK(0, false))
    dbClose(t)
}

// Benchamrks

func check(err os.Error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func BenchmarkInsertSelect(b *testing.B) {
    b.StopTimer()

    db := New(conn[0], conn[1], conn[2], user, passwd, dbname)
    check(db.Connect())

    db.Start("drop table B") // Drop test table if exists

    _, err := db.Start("create table B (s varchar(40), i int)")
    check(err)

    for ii := 0; ii < 10000; ii++ {
        _, err := db.Start("insert B values ('%d-%d-%d', %d)", ii, ii, ii, ii)
        check(err)
    }

    b.StartTimer()

    for ii := 0; ii < b.N; ii++ {
        res, err := db.Start("select * from B")
        check(err)
        for {
            row, err := res.GetRow()
            check(err)
            if row == nil {
                break
            }
        }
    }

    b.StopTimer()

    _, err = db.Start("drop table B")
    check(err)
    check(db.Close())
}

func BenchmarkPreparedInsertSelect(b *testing.B) {
    b.StopTimer()

    db := New(conn[0], conn[1], conn[2], user, passwd, dbname)
    check(db.Connect())

    db.Start("drop table B") // Drop test table if exists

    _, err := db.Start("create table B (s varchar(40), i int)")
    check(err)

    ins, err := db.Prepare("insert B values (?, ?)")
    check(err)

    sel, err := db.Prepare("select * from B")
    check(err)

    for ii := 0; ii < 10000; ii++ {
        _, err := ins.Run(fmt.Sprintf("%d-%d-%d", ii, ii, ii), ii)
        check(err)
    }

    b.StartTimer()

    for ii := 0; ii < b.N; ii++ {
        res, err := sel.Run()
        check(err)
        for {
            row, err := res.GetRow()
            check(err)
            if row == nil {
                break
            }
        }
    }

    b.StopTimer()

    _, err = db.Start("drop table B")
    check(err)
    check(db.Close())
}
