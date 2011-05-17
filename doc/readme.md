Sorry for my poor English. If you can help in improving English in this
documentation, please contact me.

## MyMySQL v0.3.7 (2011-05-17)

This package contains MySQL client API written entirely in Go. It was created
due to lack of properly working MySQL client API package, ready for my
production application (December 2010).

The code of this package is carefuly written and has internal error handling
using *panic()* exceptions, thus the probability of bugs in Go code or an
unhandled internal errors should be very small.

This package works with the MySQL protocol version 4.1 or greater. It definitely
works well with MySQL 5.0 and 5.1 (I use these versions of MySQL for my
production application).

The package includes an extensive set of automated tests that ensure that any
code changes during development will not break the package itself.

## Differences betwen version 0.2 and 0.3.7

1. There is one change in v0.3, which doesn't preserve backwards compatibility
with v0.2: the name of *Execute* method was changed to *Run*. A new *Exec*
method was added. It is similar in result to *Query* method.
2. *Reconnect* method was added. After reconnect it re-prepare all prepared
statements, related to database handler that was reconnected.
3. Autoreconn interface was added. It allows not worry about making the
connection, and not wory about re-establish connection after network error or
MySQL server restart. It is certainly safe to use it with *select* queries and
to prepare statements. You must be careful with *insert* queries. I'm not sure
whether the server performs an insert: immediately after receive query or after successfull sending OK packet. Even if it is the second option, server may not
immediately notice the network failure, becouse of network buffers in kernel.
Therefore query repetitions may cause additional unnecessary inserts into
database. This interface does not appear to be useful with local transactions.
4. *Register* method was added in v0.3.2. It allows to register commands which
will be executed immediately after connect. It is mainly useful with
*Reconnect* method and autoreconn interface.
5. Multi statements / multi results were added.
6. Types *ENUM* and *SET* were added for prepared statements results.
7. *Time* and *Date* types added in v0.3.3.
8. Since v0.3.3 *Run*, *Exec* and *ExecAC* accept parameters, *Start*, *Query*,
*QueryAC* no longer accept prepared statement as first argument.
9. In v0.3.4 float type disappeared because Go release.2011-01-20. If you use
older Go release use mymysql v0.3.3 
10. *IsConnected()* method was added in v0.3.5.
11. In v0.3.5 package name was changed from *mymy* to *mymysql*. Now the
package name corresponds to the name of Github repository.
12. The *EscapeString* method was added in v0.3.6.
13. v0.3.7 works with Go release.r57.1

## Installing

### Using *goinstall* - preferred way:

     $ goinstall github.com/ziutek/mymysql

After this command *mymysql* is ready to use. You may find source in

    $GOROOT/src/pkg/github.com/ziutek/mymysql

directory.

You can use `goinstall -u -a` for update all installed packages.

### Using *git clone* command:

    $ git clone git://github.com/ziutek/mymysql
    $ cd mymysql
    $ make install

### Version for Go weekly releases

If master branch can't be compiled with Go weekly release, try clone MyMySQL weekly branch:

    $ git clone -b weekly git://github.com/ziutek/mymysql
    $ cd mymysql
    $ make install

## Testing

For testing you need test database and test user:

    mysql> create database test;
    mysql> grant all privileges on test.* to testuser@localhost;
    mysql> set password for testuser@localhost = password("TestPasswd9")

Make sure that MySQL *max_allowed_packet* variable in *my.cnf is greater than
33M (needed to test long packets) and logging is disabled. If logging is enabled
test may fail with this message:

	--- FAIL: mymy.TestSendLongData
	Error: Received #1210 error from MySQL server: "Incorrect arguments to mysqld_stmt_execute"

The default MySQL test server address is *127.0.0.1:3306*. You may change it in
*mymy_test.go* file.

Next run tests:

    $ cd $GOROOT/src/pkg/github.com/ziutek/mymysql
    $ gotest -v

## Interface

In *GODOC.html* or *GODOC.txt* you can find the full documentation of this package in godoc format.

## Example 1

    import (
        mymy "github.com/ziutek/mymysql"
    )

    db := mymy.New("tcp", "", "127.0.0.1:3306", user, pass, dbname)
    db.Debug = true

    err := db.Connect()
    if err != nil {
        panic(err)
    }

    rows, res, err := db.Query("select * from X where id > %d", 20)
    if err != nil {
        panic(err)
    }

    for _, row := range rows {
        for _, col := range row.Data {
            if col == nil {
                // col has NULL value
            } else {
                // Do something with text in col (type []byte)
            }
        }
        // You can get specific value from a row
        val1 := row.Data[1].([]byte)

        // You can use it directly if conversion isn't needed
        os.Stdout.Write(val1)

        // You can get converted value
        number := row.Int(0)      // Zero value
        str    := row.Str(1)      // First value
        bignum := row.MustUint(2) // Second value

        // You may get value by column name
        val2 := row.Data[res.Map["FirstColumn"]].([]byte)
    }

If you do not want to load the entire result into memory you may use
*Start* and *GetRow* methods:

    res, err := db.Start("select * from X")
    checkError(err)

    // Print fields names
    for _, field := range res.Fields {
        fmt.Print(field.Name, " ")
    }
    fmt.Println()

    // Print all rows
    for {
        row, err := res.GetRow()
        checkError(err)

        if row == nil {
            // No more rows
            break
        }

        // Print all cols
        for _, col := range row.Data {
            if col == nil {
                fmt.Print("<NULL>")
            } else {
                os.Stdout.Write(col.([]byte))
            }
            fmt.Print(" ")
        }
        fmt.Println()
    }

## Example 2 - prepared statements

You can use *Run* or *Exec* method for prepared statements:

    stmt, err := db.Prepare("insert into X values (?, ?)")
    checkError(err)

    type Data struct {
        Id  int
        Tax *float32 // nil means NULL
    }

    data = new(Data)

    for {
        err := getData(data)
        if err == endOfData {
            break       
        }
        checkError(err)

        _, err = stmt.Run(data.Id, data.Tax)
        checkError(err)
    }

*getData* is your function which retrieves data from somewhere and set *Id* and
*Tax* fields of the Data struct. In the case of *Tax* field *getData* may
assign pointer to retieved variable or nil if NULL should be stored in
database.

If you pass parameters to *Run* or *Exec* method, data are rebinded on every
method call. It isn't efficient if statement is executed more than once. You
can bind parameters and use *Run* or *Exec* method without parameters, to avoid
these unnecessary rebinds. Warning! If you use *Bind* in multithreaded
application, you should be sure that no other thread will use *Bind*, until you
no longer need binded parameters.

The simplest way to bind parameters is:

    stmt.BindParams(data.Id, data.Tax)

but you can't use it in our example, becouse parameters binded this way can't
be changed by *getData* function. You may modify bind like this:

    stmt.BindParams(&data.Id, &data.Tax)

and now it should work properly. But in our example there is better solution:

    stmt.BindParams(data)

If *BindParams* method has one parameter, and this parameter is a struct or
a pointer to the struct, it treats all fields of this struct as parameters and
bind them,

This is improved part of previous example:

    data = new(Data)
    stmt.BindParams(data)

    for {
        err := getData(data)
        if err == endOfData {
            break       
        }
        checkError(err)

        _, err = stmt.Run()
        checkError(err)
    }

## Example 3 - using SendLongData in conjunction with http.Get

    _, err = db.Start("CREATE TABLE web (url VARCHAR(80), content LONGBLOB)")
    checkError(err)

    ins, err := db.Prepare("INSERT INTO web VALUES (?, ?)")
    checkError(err)

    var url string

    ins.BindParams(&url, nil)

    for  {
        // Read URL from stdin
        url = ""
        fmt.Scanln(&url)
        if len(url) == 0 {
            // Stop reading if URL is blank line
            break
        }

        // Make connection
        resp, _, err := http.Get(url)
        checkError(err)

        // We can retrieve response directly into database because 
        // the resp.Body implements io.Reader. Use 8 kB buffer.
        err = ins.SendLongData(1, resp.Body, 8192)
        checkError(err)

        // Execute insert statement
        _, err = ins.Run()
        checkError(err)
    }

## Example 4 - multi statement / multi result

    res, err := db.Start("select id from M; select name from M")
    checkError(err)

    // Get result from first select
    for {
        row, err := res.GetRow()
        checkError(err)
        if row == nil {
            // End of first result
            break
        }

        // Do something with with the data
        functionThatUseId(row.Int(0))
    }

    // Get result from second select
    res, err = res.NextResult()
    checkError(err)
    if res == nil {
        panic("Hmm, there is no result. Why?!")
    }
    for {
        row, err := res.GetRow()
        checkError(err)
        if row == nil {
            // End of second result
            break
        }

        // Do something with with the data
        functionThatUseName(row.Str(0))
    }

## Example 5 - autoreconn interface

    db := mymy.New("tcp", "", "127.0.0.1:3306", user, pass, dbname)

    // Register initilisation command. It will be executed after each connect.
    db.Register("set names utf8")

    // There is no need to explicity connect to the MySQL server
    rows, res, err := db.QueryAC("SELECT * FROM R")
    checkError(err)

    // Now we are connected.

    // It does not matter if connection will be interrupted during sleep, eg
    // due to server reboot or network down.
    time.Sleep(9e9)

    // If we can reconnect in no more than db.MaxRetries attempts this
    // statement will be prepared.
    sel, err := db.PrepareAC("SELECT name FROM R where id > ?")
    checkError(err)

    // We can destroy our connection on server side
    _, _, err = db.QueryAC("kill %d", db.ThreadId())
    checkError(err)

    // But it doesn't matter
    sel.BindParams(2)
    rows, res, err = sel.ExecAC()
    checkError(err)

More examples are in *examples* directory.

## Type mapping

In the case of classic text queries, all variables that are sent to the MySQL
server are embded in text query. Thus you allways convert them to a string and
send embded in SQL query:

    rows, res, err := db.Query("select * from X where id > %d", id)

After text query you always receive a text result. Mysql text result
corresponds to *[]byte* type in mymysql. It isn't *string* type due to
avoidance of unnecessary type conversions. You can allways convert *[]byte* to
*string* yourself:

    fmt.Print(string(rows[0].Data[1].([]byte)))

or usnig *Str* helper method:

    fmt.Print(rows[0].Str(1))

There are other helper methods, for data conversion like *Int* or *Uint*:

    fmt.Print(rows[0].Int(1))

All three above examples return value received in row 0 column 1. If you prefer
to use the column names, you can use *res.Map* which maps result field names to
corresponding indexes:

    name := res.Map["name"]
    fmt.Print(rows[0].Str(name))

In case of prepared statements, the type mapping is slightly more complicated.
For parameters sended from the client to the server, Go/mymysql types are
mapped for MySQL protocol types as below:

             string  -->  MYSQL_TYPE_STRING
             []byte  -->  MYSQL_TYPE_VAR_STRING
        int8, uint8  -->  MYSQL_TYPE_TINY
      int16, uint16  -->  MYSQL_TYPE_SHORT
      int32, uint32  -->  MYSQL_TYPE_LONG
      int64, uint64  -->  MYSQL_TYPE_LONGLONG
            float32  -->  MYSQL_TYPE_FLOAT
            float64  -->  MYSQL_TYPE_DOUBLE
    *mymy.Timestamp  -->  MYSQL_TYPE_TIMESTAMP
     *mymy.Datetime  -->  MYSQL_TYPE_DATETIME
         *mymy.Date  -->  MYSQL_TYPE_DATE
         *mymy.Time  -->  MYSQL_TYPE_TIME
          mymy.Blob  -->  MYSQL_TYPE_BLOB
                nil  -->  MYSQL_TYPE_NULL

The MySQL server maps/converts them to a particular MySQL storage type.

For received results MySQL storage types are mapped to Go/mymysql types as
below:

                                 TINYINT  -->  int8
                        UNSIGNED TINYINT  -->  uint8
                                SMALLINT  -->  int16
                       UNSIGNED SMALLINT  -->  uint16
                          MEDIUMINT, INT  -->  int32
        UNSIGNED MEDIUMINT, UNSIGNED INT  -->  uint32
                                  BIGINT  -->  int64
                         UNSIGNED BIGINT  -->  uint64
                                   FLOAT  -->  float32
                                  DOUBLE  -->  float64
                     TIMESTAMP, DATETIME  -->  *mymy.Datetime
                                    DATE  -->  *mymy.Date
                                    TIME  -->  *mymy.Time
                                    YEAR  -->  int16
        CHAR, VARCHAR, BINARY, VARBINARY  -->  []byte
     TEXT, TINYTEXT, MEDIUMTEXT, LONGTEX  -->  []byte
    BLOB, TINYBLOB, MEDIUMBLOB, LONGBLOB  -->  []byte
                            DECIMAL, BIT  -->  []byte
                               SET, ENUM  -->  []byte
                                    NULL  -->  nil

## Big packets

This package can send and receive MySQL data packets that are biger than 16 MB.
This means that you can receive response rows biger than 16 MB and can execute
prepared statements with parameter data biger than 16 MB without using
SendLongData method. If you want to use this feature you must set *MaxPktSize*
field in database handler to appropriate value before connect, and change
*max_allowed_packet* value in MySQL server configuration.

## Thread safety

You can use this package in multithreading enviroment. All methods are thread
safe, unless the description of the method says something else.

If one thread is calling *Query* or *Exec* method, other threads will be
blocked if they call *Query*, *Start*, *Exec*, *Run* or other method which send
data to the server, until *Query*/*Exec* return in first thread.

If one thread is calling *Start* or *Run* method, other threads will be
blocked if they call *Query*, *Start*, *Exec*, *Run* or other method which send
data to the server,  until all results and all rows  will be readed from
the connection in first thread.

Multithreading was tested on my production web application. It uses *http*
package to serve dynamic web pages. *http* package creates one gorutine for any
HTTP connection. Any GET request during connection causes 4-8 select queries to
MySQL database (some of them are prepared statements). Database contains ca.
30 tables (three largest have 82k, 73k and 3k rows). There is one persistant
connection to MySQL server which is shared by all gorutines. Application is
running on dual-core machine with GOMAXPROCS=2. It was tested using *siege*:

    # siege my.httpserver.pl -c25 -d0 -t 30s
    ** SIEGE 2.69
    ** Preparing 25 concurrent users for battle.
    The server is now under siege...
    Lifting the server siege..      done.
    Transactions:                   3212 hits
    Availability:                 100.00 %
    Elapsed time:                  29.83 secs
    Data transferred:               3.88 MB
    Response time:                  0.22 secs
    Transaction rate:             107.68 trans/sec
    Throughput:	                    0.13 MB/sec
    Concurrency:                   23.43
    Successful transactions:        3218
    Failed transactions:               0
    Longest transaction:            9.28
    Shortest transaction:           0.01

Thanks to *siege* stress tests I fixed some multi-threading bugs in v0.3.2.

## TODO

1. Complete GODOC documentation
2. stmt.BindResult
3. io.Writer as bind result variable

# Package documentation generated by godoc

It is *GODOC.html* file embded in *README.md*.  Unfortunately, after embding links no longer work.


