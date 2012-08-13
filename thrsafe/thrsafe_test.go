/*
Copyright (c) 2012, Guihong Bai
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:
1. Redistributions of source code must retain the above copyright
   notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright
   notice, this list of conditions and the following disclaimer in the
   documentation and/or other materials provided with the distribution.
3. The name of the author may not be used to endorse or promote products
   derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT,
INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package thrsafe

//UnitTest For thrsafe api 
// TestCase Naming
// S : stand for one Status only row
// D : stand for one data row


import (
	"testing"
)

var (
	conn   = []string{"tcp", "", "127.0.0.1:3306"}
	user   = "testuser"
	passwd = "TestPasswd9"
	dbname = "test"
)

func checkErr(t *testing.T, err error, exp_err error) {
	if err != exp_err {
		if exp_err == nil {
			t.Fatalf("Error: %v", err)
		} else {
			t.Fatalf("Error: %v\nExpected error: %v", err, exp_err)
		}
	}
}


func TestNormalData(t *testing.T) {
	c := New(conn[0], conn[1], conn[2], user, passwd)

	var err error
	err = c.Connect()
	checkErr(t, err, nil)

	// Register initialisation commands
	c.Register("set names utf8")

	// my is in unconnected state
	checkErr(t, c.Use(dbname), nil)

	// Drop test table if exists
	c.Query("drop table R")

	// Create table
	_, _, err = c.Query(
		"create table R (id int primary key, name varchar(20))",
	)
	checkErr(t, err, nil)

	// Prepare insert statement
	ins, err := c.Prepare("insert R values (?,  ?)")
	checkErr(t, err, nil)

	// Bind insert parameters
	ins.Bind(1, "jeden")
	// Insert into table
	_, _, err = ins.Exec()
	checkErr(t, err, nil)

	// Bind insert parameters
	ins.Bind(2, "dwa")
	// Insert into table
	_, _, err = ins.Exec()
	checkErr(t, err, nil)

	// Select from table
	rows, res, err := c.Query("select * from R")
	checkErr(t, err, nil)
	id := res.Map("id")
	name := res.Map("name")
	if len(rows) != 2 ||
		rows[0].Int(id) != 1 || rows[0].Str(name) != "jeden" ||
		rows[1].Int(id) != 2 || rows[1].Str(name) != "dwa" {
		t.Fatal("Bad result")
	}
	// Drop table
	_, _, err = c.Query("drop table R")
	checkErr(t, err, nil)

	// Disconnect
	c.Close()

}

func TestNormal_S(t *testing.T) {
	c := New(conn[0], conn[1], conn[2], user, passwd)

	var err error
	err = c.Connect()
	checkErr(t, err, nil)

	// Register initialisation commands
	c.Register("set names utf8")

	// my is in unconnected state
	checkErr(t, c.Use(dbname), nil)

	res, err := c.Start("set @rtn= 0;")
	checkErr(t, err, nil)

	if !res.StatusOnly() {
		t.Fatal("No Stauts Only")
	}

	if res.MoreResults() {
		t.Fatal("More Result")
	}

	// Disconnect
	c.Close()
}
func TestMutli_SS(t *testing.T) {
	c := New(conn[0], conn[1], conn[2], user, passwd)
	var err error
	err = c.Connect()
	checkErr(t, err, nil)

	// my is in unconnected state
	checkErr(t, c.Use(dbname), nil)

	res, err := c.Start("set @rtn= 0; set @test=1;")
	checkErr(t, err, nil)

	for {
		//use GetRows to handle lock issue
		_, err := res.GetRows()
		checkErr(t, err, nil)

		if !res.StatusOnly() {
			t.Fatal("Bad result")
		}
		if res.MoreResults() {
			res, err = res.NextResult()
			checkErr(t, err, nil)

		} else {

			break
		}
	}
	// Disconnect
	c.Close()
}

func TestMutli_SD(t *testing.T) {

	c := New(conn[0], conn[1], conn[2], user, passwd)
	var err error
	err = c.Connect()
	checkErr(t, err, nil)

	// my is in unconnected state
	checkErr(t, c.Use(dbname), nil)

	res, err := c.Start("set @rtn = 0; select @rtn as result;")
	checkErr(t, err, nil)

	for {
		rows, err := res.GetRows()
		checkErr(t, err, nil)
		if rows != nil {

			if len(rows) != 1 ||
				rows[0].Int(res.Map("result")) != 0 {
				t.Fatal("Bad result")
			}

		}
		if res.MoreResults() {
			res, err = res.NextResult()
			checkErr(t, err, nil)

		} else {
			break
		}
	}
	// Disconnect
	c.Close()
}

func TestMutli_DD(t *testing.T) {
	c := New(conn[0], conn[1], conn[2], user, passwd)
	var err error
	err = c.Connect()
	checkErr(t, err, nil)

	// my is in unconnected state
	checkErr(t, c.Use(dbname), nil)

	res, err := c.Start("select 1 as rtn;select 2 as test;")
	checkErr(t, err, nil)

	i := 0
	for {
		i++
		rows, err := res.GetRows()
		checkErr(t, err, nil)
		if rows != nil {

			if len(rows) != 1 {

				if (i == 1 && rows[0].Int(res.Map("result")) != 1) || (i == 2 && rows[0].Int(res.Map("test")) != 2) {
					t.Fatal("Bad result")
				}
			}

		}

		if res.MoreResults() {
			res, err = res.NextResult()
			checkErr(t, err, nil)

		} else {

			break
		}
	}
	// Disconnect
	c.Close()
}

func TestMutli_SDS(t *testing.T) {
	c := New(conn[0], conn[1], conn[2], user, passwd)

	var err error
	err = c.Connect()
	checkErr(t, err, nil)

	// Register initialisation commands
	c.Register("set names utf8")

	// my is in unconnected state
	checkErr(t, c.Use(dbname), nil)

	res, err := c.Start("set @rtn= 0; select @rtn as result;set @aa=1;")
	checkErr(t, err, nil)
	for {

		rows, err := res.GetRows()
		checkErr(t, err, nil)
		if rows != nil {

			if len(rows) != 1 ||
				rows[0].Int(res.Map("result")) != 0 {
				t.Fatal("Bad result")
			}

		}
		if res.MoreResults() {
			res, err = res.NextResult()
			checkErr(t, err, nil)
		} else {
			break
		}

	}
	// Disconnect
	c.Close()
}
