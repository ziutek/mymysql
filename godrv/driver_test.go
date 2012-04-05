package godrv

import (
    "database/sql"
    "testing"
    "strings"
)

func init() {
    Register("set names utf8")
}

func checkErr(t *testing.T, err error) {
    if err != nil {
        t.Fatalf("Error: %v", err)
    }
}
func checkErrId(t *testing.T, err error, rid, eid int64) {
    checkErr(t, err)
    if rid != eid {
        t.Fatal("res.LastInsertId() ==", rid, "but should be", eid)
    }
}

var (
    dsn []string = []string{
        "db/user/passwd",
        "tcp://127.0.0.1/db/user/passwd",
        "tcp://127.0.0.1/db/user/passwd?charset=utf8",
        "tcp://127.0.0.1:3307/db/user/passwd?charset=utf8&keepalive=3600",
        "db/user",
        "tcp://127.0.0.1/db/user?charset=utf8",
        "db/user/pass*wd",
        "db/user/pass**wd",
    }

    dsnr []string = []string{
        "tcp://127.0.0.1:3306/db/user/passwd",
        "tcp://127.0.0.1:3306/db/user/passwd",
        "tcp://127.0.0.1:3306/db/user/passwd",
        "tcp://127.0.0.1:3307/db/user/passwd",
        "tcp://127.0.0.1:3306/db/user/",
        "tcp://127.0.0.1:3306/db/user/",
        "tcp://127.0.0.1:3306/db/user/pass/wd",
        "tcp://127.0.0.1:3306/db/user/pass*wd",
    }

    p []string = []string{
        "",
        "",
        "charset",
        "charset keepalive",
        "",
        "charset",
        "",
        "",
        "",
    }

    e []error = []error{
        nil,
        nil,
        nil,
        nil,
        nil,
        nil,
        nil,
        nil,
    }
)

func checkDSN(t *testing.T, i int, proto, addr, db, user, passwd string, params map[string]string, err error) {
    if dsnr[i] != "" {
        ndsn := proto + "://" + addr + "/" + db + "/" + user + "/" + passwd
        if ndsn != dsnr[i] {
            t.Fatal("dsn ==", ndsn, "but should be(", i, ")" , dsnr[i])
        }
    }
    if p[i] != "" {
        tp := strings.Split(p[i], " ")
        for _, v := range tp {
            if _, ok := params[v]; !ok {
                t.Fatal("param", v, "not found")
            }
        }
    }
    if e[i] != err {
        t.Fatal(err)
    }
}

func TestDSN(t *testing.T) {
    for i, d := range dsn {
        proto, addr, db, user, passwd, params, err := parseDSN(d)
        checkDSN(t, i, proto, addr, db, user, passwd, params, err)
    }
}

func TestAll(t *testing.T) {
    data := []string{"jeden", "dwa", "中文"}

    db, err := sql.Open("mymysql", "test/testuser/TestPasswd9?charset=utf8")
    checkErr(t, err)
    defer db.Close()
	defer db.Exec("DROP TABLE go")

    db.Exec("DROP TABLE go")

    _, err = db.Exec(
        `CREATE TABLE go (
            id  INT PRIMARY KEY AUTO_INCREMENT,
            txt TEXT
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8`)
        checkErr(t, err)

        ins, err := db.Prepare("INSERT go SET txt=?")
        checkErr(t, err)

        tx, err := db.Begin()
        checkErr(t, err)

        res, err := ins.Exec(data[0])
        checkErr(t, err)
        id, err := res.LastInsertId()
        checkErrId(t, err, id, 1)

        res, err = ins.Exec(data[1])
        checkErr(t, err)
        id, err = res.LastInsertId()
        checkErrId(t, err, id, 2)

        res, err = ins.Exec(data[2])
        checkErr(t, err)
        id, err = res.LastInsertId()
        checkErrId(t, err, id, 3)

        checkErr(t, tx.Commit())

        tx, err = db.Begin()
        checkErr(t, err)

        res, err = tx.Exec("INSERT go SET txt=?", "trzy")
        checkErr(t, err)
        id, err = res.LastInsertId()
        checkErrId(t, err, id, 4)

        checkErr(t, tx.Rollback())

        rows, err := db.Query("SELECT * FROM go")
        checkErr(t, err)
        for rows.Next() {
            var id int
            var txt string
            checkErr(t, rows.Scan(&id, &txt))
            if id > len(data) {
                t.Fatal("To many rows in table")
            }
            if data[id-1] != txt {
                t.Fatalf("txt[%d] == '%s' != '%s'", id, txt, data[id-1])
            }
        }
	    checkErr(t, rows.Err())

        sql := "select sum(41) as test"
        row := db.QueryRow(sql)
        var vi int64
        checkErr(t, row.Scan(&vi))
        if vi != 41 {
            t.Fatal(sql)
        }
        sql = "select sum(4123232323232) as test"
        row = db.QueryRow(sql)
        var vf float64
        checkErr(t, row.Scan(&vf))
        if vf != 4123232323232 {
            t.Fatal(sql)
        }
    }

func TestMediumInt(t *testing.T) {
	db, err := sql.Open("mymysql", "test/testuser/TestPasswd9")
	checkErr(t, err)
	defer db.Exec("DROP TABLE mi")
	defer db.Close()

	db.Exec("DROP TABLE mi")

	_, err = db.Exec(
		`CREATE TABLE mi (
			id INT PRIMARY KEY AUTO_INCREMENT,
			m MEDIUMINT
		)`)
	checkErr(t, err)

	const n = 9

	for i := 0; i < n; i++ {
		_, err = db.Exec("INSERT mi VALUES (0, ?)", i)
	}

	rows, err := db.Query("SELECT * FROM mi")
	checkErr(t, err)

	var i int
	for i = 0; rows.Next(); i++ {
		var id, m int
		checkErr(t, rows.Scan(&id, &m))
		if id != i+1 || m != i {
			t.Fatalf("i=%d id=%d m=%d", i, id, m)
		}
	}
	checkErr(t, rows.Err())
	if i != n {
		t.Fatalf("%d rows read, %d expected", i, n)
	}
}

