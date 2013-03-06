package godrv

import (
	"database/sql"
	"testing"
	"time"
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

func TestAll(t *testing.T) {
	data := []string{"jeden", "dwa"}

	db, err := sql.Open("mymysql", "test/testuser/TestPasswd9")
	checkErr(t, err)
	defer db.Close()
	defer db.Exec("DROP TABLE go")

	db.Exec("DROP TABLE go")

	_, err = db.Exec(
		`CREATE TABLE go (
			id  INT PRIMARY KEY AUTO_INCREMENT,
			txt TEXT
		) ENGINE=InnoDB`)
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

	checkErr(t, tx.Commit())

	tx, err = db.Begin()
	checkErr(t, err)

	res, err = tx.Exec("INSERT go SET txt=?", "trzy")
	checkErr(t, err)
	id, err = res.LastInsertId()
	checkErrId(t, err, id, 3)

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
		checkErr(t, err)
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

func TestTypes(t *testing.T) {
	db, err := sql.Open("mymysql", "test/testuser/TestPasswd9")
	checkErr(t, err)
	defer db.Close()
	defer db.Exec("DROP TABLE t")

	db.Exec("DROP TABLE t")

	_, err = db.Exec(
		`CREATE TABLE t (
			i INT NOT NULL,
			f DOUBLE NOT NULL, 
			b BOOL NOT NULL,
			s VARCHAR(8) NOT NULL,
			d DATETIME NOT NULL,
			n INT
		) ENGINE=InnoDB`)
	checkErr(t, err)

	_, err = db.Exec(
		"INSERT t VALUES (23, 0.25, true, 'test', '2013-03-06 21:07', NULL)",
	)
	checkErr(t, err)
	l, err := time.LoadLocation("Local")
	td := time.Date(2013, 3, 6, 21, 7, 0, 0, l)
	checkErr(t, err)
	_, err = db.Exec(
		"INSERT t VALUES (?, ?, ?, ?, ?, ?)",
		23, 0.25, true, "test", td, nil,
	)

	rows, err := db.Query("SELECT * FROM t")
	checkErr(t, err)
	var (
		i int64
		f float64
		b bool
		s string
		d time.Time
		n sql.NullInt64
	)

	for rows.Next() {
		checkErr(t, rows.Scan(&i, &f, &b, &s, &d, &n))
		if i != 23 {
			t.Fatal("int64", i)
		}
		if f != 0.25 {
			t.Fatal("float64", f)
		}
		if b != true {
			t.Fatal("bool", b)
		}
		if s != "test" {
			t.Fatal("string", s)
		}
		if d != td {
			t.Fatal("time.Time", d)

		}
		if n.Valid {
			t.Fatal("mysql.NullInt64", n)
		}
	}
}
