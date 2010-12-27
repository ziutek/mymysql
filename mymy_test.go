package mymy

import (
    "testing"
    "reflect"
    "os"
    "fmt"
)

var (
    db *MySQL
    user   = "testuser"
    passwd = "TestPasswd9"
    dbname = "test"
    conn   = []string{"unix", "", "/var/run/mysqld/mysqld.sock"}
    //conn = []string{"tcp",  "", "127.0.0.1:3306"}
    debug  = false
)

// Utils

type RowsResErr struct {
    rows []*TextRow
    res  *Result
    err  os.Error
}

func query(sql string) *RowsResErr {
    rows, res, err := db.Query(sql)
    return &RowsResErr{rows, res, err}
}

func queryf(format string, a ...interface{}) *RowsResErr {
    rows, res, err := db.Queryf(format, a...)
    return &RowsResErr{rows, res, err}
}

func checkRes(t *testing.T, res, exp *RowsResErr) {
    /*if !reflect.DeepEqual(res.rows, exp.rows) {
        fmt.Println("r:", len(res.rows)m len(exp.rows))
        for ii := range res.rows {
            fmt.Println("v:", len(res.rows[ii].Data), len(exp.rows[ii].Data))
        }
    }*/
    if !reflect.DeepEqual(res, exp) {
        t.Fatalf(
            "Bad result:\nres=%v %v %v\nexp=%v %v %v",
            res.rows, *res.res, res.err, exp.rows, *exp.res, exp.err,
        )
    }
}

func checkErr(t *testing.T, err os.Error) {
    if err != nil {
        t.Fatal("Error:", err)
    }
}

func queryOK(affected uint64) *RowsResErr {
    return &RowsResErr {
        res: &Result{db: db, Status: 0x2, AffectedRows: affected},
    }
}

func dbConnect(t *testing.T, with_db bool) {
    if with_db {
        db  = New(conn[0], conn[1], conn[2], user, passwd, dbname)
    } else {
        db  = New(conn[0], conn[1], conn[2], user, passwd)
    }

    db.Debug = debug

    checkErr(t, db.Connect())
    checkRes(t, query("set names utf8"), queryOK(0))
}

func dbClose(t *testing.T) {
    checkErr(t, db.Close())
}


// Tests

func TestUse(t *testing.T) {
    dbConnect(t, false)
    checkErr(t, db.Use(dbname))
    dbClose(t)
}

func TestPing(t *testing.T) {
    dbConnect(t, false)
    checkErr(t, db.Ping())

func TestQuery(t *testing.T) {
    dbConnect(t, true)
    query("drop table T") // Drop test table if exists
    checkRes(t, query("create table T (s varchar(40))"), queryOK(0))
    exp := &RowsResErr {
        res: &Result {
            db:         db,
            Status:     0x22,
            FieldCount: 1,
            Fields:     []*Field {
                &Field {
                    Catalog:  "def",
                    Db:       "test",
                    Table:    "T",
                    OrgTable: "T",
                    Name:     "s",
                    OrgName:  "s",
                    DispLen:  3 * 40, //varchar(40)
                    Flags:    0,
                    Type:     FIELD_TYPE_VAR_STRING,
                    Scale:    0,
                },
            },
            Map:        map[string]int{"s": 0},
        },
    }
    for ii := 0; ii < 100; ii++ {
        var val Nbin
        if ii % 10 == 0 {
                checkRes(t, query("insert T values (null)"), queryOK(1))
                val = nil
        } else {
                txt := []byte(fmt.Sprintf("%d-%d-%d", ii, ii, ii))
                checkRes(t, queryf("insert T values ('%s')", txt), queryOK(1))
                val = &txt
        }
        exp.rows = append(exp.rows, &TextRow{Data: []Nbin{val}})
    }
    checkRes(t, query("select s from T"), exp)
    checkRes(t, query("drop table T"), queryOK(0))
    dbClose(t)
}

func BenchmarkSelect(b *testing.B) {
    b.StopTimer()
    // Szykujemy dane
    // Startujemy test
    b.StartTimer()
    // Kod testu
}
