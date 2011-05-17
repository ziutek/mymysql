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


<dl>
<dt><a href='#DecodeU16'>func DecodeU16</a></dt>
<dt><a href='#DecodeU24'>func DecodeU24</a></dt>
<dt><a href='#DecodeU32'>func DecodeU32</a></dt>
<dt><a href='#DecodeU64'>func DecodeU64</a></dt>
<dt><a href='#EncodeDate'>func EncodeDate</a></dt>
<dt><a href='#EncodeDatetime'>func EncodeDatetime</a></dt>
<dt><a href='#EncodeTime'>func EncodeTime</a></dt>
<dt><a href='#EncodeU16'>func EncodeU16</a></dt>
<dt><a href='#EncodeU24'>func EncodeU24</a></dt>
<dt><a href='#EncodeU32'>func EncodeU32</a></dt>
<dt><a href='#EncodeU64'>func EncodeU64</a></dt>
<dt><a href='#IsDateZero'>func IsDateZero</a></dt>
<dt><a href='#IsDatetimeZero'>func IsDatetimeZero</a></dt>
<dt><a href='#IsNetErr'>func IsNetErr</a></dt>
<dt><a href='#NbinToNstr'>func NbinToNstr</a></dt>
<dt><a href='#NstrToNbin'>func NstrToNbin</a></dt>
<dt><a href='#Blob'>type Blob</a></dt>
<dt><a href='#Date'>type Date</a></dt>
<dd><a href='#Date.StrToDate'>func StrToDate</a></dd>
<dd><a href='#Date.String'>func (*Date) String</a></dd>
<dt><a href='#Datetime'>type Datetime</a></dt>
<dd><a href='#Datetime.DateToDatetime'>func DateToDatetime</a></dd>
<dd><a href='#Datetime.StrToDatetime'>func StrToDatetime</a></dd>
<dd><a href='#Datetime.TimeToDatetime'>func TimeToDatetime</a></dd>
<dd><a href='#Datetime.String'>func (*Datetime) String</a></dd>
<dt><a href='#Error'>type Error</a></dt>
<dd><a href='#Error.String'>func (Error) String</a></dd>
<dt><a href='#Field'>type Field</a></dt>
<dt><a href='#MySQL'>type MySQL</a></dt>
<dd><a href='#MySQL.New'>func New</a></dd>
<dd><a href='#MySQL.Close'>func (*MySQL) Close</a></dd>
<dd><a href='#MySQL.Connect'>func (*MySQL) Connect</a></dd>
<dd><a href='#MySQL.EscapeString'>func (*MySQL) EscapeString</a></dd>
<dd><a href='#MySQL.IsConnected'>func (*MySQL) IsConnected</a></dd>
<dd><a href='#MySQL.Ping'>func (*MySQL) Ping</a></dd>
<dd><a href='#MySQL.Prepare'>func (*MySQL) Prepare</a></dd>
<dd><a href='#MySQL.PrepareAC'>func (*MySQL) PrepareAC</a></dd>
<dd><a href='#MySQL.Query'>func (*MySQL) Query</a></dd>
<dd><a href='#MySQL.QueryAC'>func (*MySQL) QueryAC</a></dd>
<dd><a href='#MySQL.Reconnect'>func (*MySQL) Reconnect</a></dd>
<dd><a href='#MySQL.Register'>func (*MySQL) Register</a></dd>
<dd><a href='#MySQL.Start'>func (*MySQL) Start</a></dd>
<dd><a href='#MySQL.ThreadId'>func (*MySQL) ThreadId</a></dd>
<dd><a href='#MySQL.Use'>func (*MySQL) Use</a></dd>
<dd><a href='#MySQL.UseAC'>func (*MySQL) UseAC</a></dd>
<dt><a href='#Raw'>type Raw</a></dt>
<dt><a href='#Result'>type Result</a></dt>
<dd><a href='#Result.End'>func (*Result) End</a></dd>
<dd><a href='#Result.GetRow'>func (*Result) GetRow</a></dd>
<dd><a href='#Result.NextResult'>func (*Result) NextResult</a></dd>
<dt><a href='#Row'>type Row</a></dt>
<dd><a href='#Row.Bin'>func (*Row) Bin</a></dd>
<dd><a href='#Row.Date'>func (*Row) Date</a></dd>
<dd><a href='#Row.DateErr'>func (*Row) DateErr</a></dd>
<dd><a href='#Row.Datetime'>func (*Row) Datetime</a></dd>
<dd><a href='#Row.DatetimeErr'>func (*Row) DatetimeErr</a></dd>
<dd><a href='#Row.Int'>func (*Row) Int</a></dd>
<dd><a href='#Row.IntErr'>func (*Row) IntErr</a></dd>
<dd><a href='#Row.MustDate'>func (*Row) MustDate</a></dd>
<dd><a href='#Row.MustDatetime'>func (*Row) MustDatetime</a></dd>
<dd><a href='#Row.MustInt'>func (*Row) MustInt</a></dd>
<dd><a href='#Row.MustTime'>func (*Row) MustTime</a></dd>
<dd><a href='#Row.MustUint'>func (*Row) MustUint</a></dd>
<dd><a href='#Row.Str'>func (*Row) Str</a></dd>
<dd><a href='#Row.Time'>func (*Row) Time</a></dd>
<dd><a href='#Row.TimeErr'>func (*Row) TimeErr</a></dd>
<dd><a href='#Row.Uint'>func (*Row) Uint</a></dd>
<dd><a href='#Row.UintErr'>func (*Row) UintErr</a></dd>
<dt><a href='#Statement'>type Statement</a></dt>
<dd><a href='#Statement.BindParams'>func (*Statement) BindParams</a></dd>
<dd><a href='#Statement.Delete'>func (*Statement) Delete</a></dd>
<dd><a href='#Statement.Exec'>func (*Statement) Exec</a></dd>
<dd><a href='#Statement.ExecAC'>func (*Statement) ExecAC</a></dd>
<dd><a href='#Statement.Reset'>func (*Statement) Reset</a></dd>
<dd><a href='#Statement.ResetParams'>func (*Statement) ResetParams</a></dd>
<dd><a href='#Statement.Run'>func (*Statement) Run</a></dd>
<dd><a href='#Statement.SendLongData'>func (*Statement) SendLongData</a></dd>
<dt><a href='#Time'>type Time</a></dt>
<dd><a href='#Time.StrToTime'>func StrToTime</a></dd>
<dd><a href='#Time.String'>func (*Time) String</a></dd>
<dt><a href='#Timestamp'>type Timestamp</a></dt>
<dd><a href='#Timestamp.String'>func (*Timestamp) String</a></dd>
</dl>
<!--
Copyright 2009 The Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
-->

<!-- PackageName is printed as title by the top-level template -->
<p><code>import "mymysql"</code></p>

<p>
<h4>Package files</h4>
<span style="font-size:90%">
<a href="/mymysql/addons.go">addons.go</a>
<a href="/mymysql/autoconnect.go">autoconnect.go</a>
<a href="/mymysql/binding.go">binding.go</a>
<a href="/mymysql/codecs.go">codecs.go</a>
<a href="/mymysql/command.go">command.go</a>
<a href="/mymysql/common.go">common.go</a>
<a href="/mymysql/consts.go">consts.go</a>
<a href="/mymysql/errors.go">errors.go</a>
<a href="/mymysql/init.go">init.go</a>
<a href="/mymysql/mysql.go">mysql.go</a>
<a href="/mymysql/packet.go">packet.go</a>
<a href="/mymysql/prepared.go">prepared.go</a>
<a href="/mymysql/result.go">result.go</a>
<a href="/mymysql/unsafe.go">unsafe.go</a>
</span>
</p>
<h2 id="Constants">Constants</h2>
<p>
MySQL error codes
</p>

<pre>const (
ER_HASHCHK                                 = 1000
ER_NISAMCHK                                = 1001
ER_NO                                      = 1002
ER_YES                                     = 1003
ER_CANT_CREATE_FILE                        = 1004
ER_CANT_CREATE_TABLE                       = 1005
ER_CANT_CREATE_DB                          = 1006
ER_DB_CREATE_EXISTS                        = 1007
ER_DB_DROP_EXISTS                          = 1008
ER_DB_DROP_DELETE                          = 1009
ER_DB_DROP_RMDIR                           = 1010
ER_CANT_DELETE_FILE                        = 1011
ER_CANT_FIND_SYSTEM_REC                    = 1012
ER_CANT_GET_STAT                           = 1013
ER_CANT_GET_WD                             = 1014
ER_CANT_LOCK                               = 1015
ER_CANT_OPEN_FILE                          = 1016
ER_FILE_NOT_FOUND                          = 1017
ER_CANT_READ_DIR                           = 1018
ER_CANT_SET_WD                             = 1019
ER_CHECKREAD                               = 1020
ER_DISK_FULL                               = 1021
ER_DUP_KEY                                 = 1022
ER_ERROR_ON_CLOSE                          = 1023
ER_ERROR_ON_READ                           = 1024
ER_ERROR_ON_RENAME                         = 1025
ER_ERROR_ON_WRITE                          = 1026
ER_FILE_USED                               = 1027
ER_FILSORT_ABORT                           = 1028
ER_FORM_NOT_FOUND                          = 1029
ER_GET_ERRNO                               = 1030
ER_ILLEGAL_HA                              = 1031
ER_KEY_NOT_FOUND                           = 1032
ER_NOT_FORM_FILE                           = 1033
ER_NOT_KEYFILE                             = 1034
ER_OLD_KEYFILE                             = 1035
ER_OPEN_AS_READONLY                        = 1036
ER_OUTOFMEMORY                             = 1037
ER_OUT_OF_SORTMEMORY                       = 1038
ER_UNEXPECTED_EOF                          = 1039
ER_CON_COUNT_ERROR                         = 1040
ER_OUT_OF_RESOURCES                        = 1041
ER_BAD_HOST_ERROR                          = 1042
ER_HANDSHAKE_ERROR                         = 1043
ER_DBACCESS_DENIED_ERROR                   = 1044
ER_ACCESS_DENIED_ERROR                     = 1045
ER_NO_DB_ERROR                             = 1046
ER_UNKNOWN_COM_ERROR                       = 1047
ER_BAD_NULL_ERROR                          = 1048
ER_BAD_DB_ERROR                            = 1049
ER_TABLE_EXISTS_ERROR                      = 1050
ER_BAD_TABLE_ERROR                         = 1051
ER_NON_UNIQ_ERROR                          = 1052
ER_SERVER_SHUTDOWN                         = 1053
ER_BAD_FIELD_ERROR                         = 1054
ER_WRONG_FIELD_WITH_GROUP                  = 1055
ER_WRONG_GROUP_FIELD                       = 1056
ER_WRONG_SUM_SELECT                        = 1057
ER_WRONG_VALUE_COUNT                       = 1058
ER_TOO_LONG_IDENT                          = 1059
ER_DUP_FIELDNAME                           = 1060
ER_DUP_KEYNAME                             = 1061
ER_DUP_ENTRY                               = 1062
ER_WRONG_FIELD_SPEC                        = 1063
ER_PARSE_ERROR                             = 1064
ER_EMPTY_QUERY                             = 1065
ER_NONUNIQ_TABLE                           = 1066
ER_INVALID_DEFAULT                         = 1067
ER_MULTIPLE_PRI_KEY                        = 1068
ER_TOO_MANY_KEYS                           = 1069
ER_TOO_MANY_KEY_PARTS                      = 1070
ER_TOO_LONG_KEY                            = 1071
ER_KEY_COLUMN_DOES_NOT_EXITS               = 1072
ER_BLOB_USED_AS_KEY                        = 1073
ER_TOO_BIG_FIELDLENGTH                     = 1074
ER_WRONG_AUTO_KEY                          = 1075
ER_READY                                   = 1076
ER_NORMAL_SHUTDOWN                         = 1077
ER_GOT_SIGNAL                              = 1078
ER_SHUTDOWN_COMPLETE                       = 1079
ER_FORCING_CLOSE                           = 1080
ER_IPSOCK_ERROR                            = 1081
ER_NO_SUCH_INDEX                           = 1082
ER_WRONG_FIELD_TERMINATORS                 = 1083
ER_BLOBS_AND_NO_TERMINATED                 = 1084
ER_TEXTFILE_NOT_READABLE                   = 1085
ER_FILE_EXISTS_ERROR                       = 1086
ER_LOAD_INFO                               = 1087
ER_ALTER_INFO                              = 1088
ER_WRONG_SUB_KEY                           = 1089
ER_CANT_REMOVE_ALL_FIELDS                  = 1090
ER_CANT_DROP_FIELD_OR_KEY                  = 1091
ER_INSERT_INFO                             = 1092
ER_UPDATE_TABLE_USED                       = 1093
ER_NO_SUCH_THREAD                          = 1094
ER_KILL_DENIED_ERROR                       = 1095
ER_NO_TABLES_USED                          = 1096
ER_TOO_BIG_SET                             = 1097
ER_NO_UNIQUE_LOGFILE                       = 1098
ER_TABLE_NOT_LOCKED_FOR_WRITE              = 1099
ER_TABLE_NOT_LOCKED                        = 1100
ER_BLOB_CANT_HAVE_DEFAULT                  = 1101
ER_WRONG_DB_NAME                           = 1102
ER_WRONG_TABLE_NAME                        = 1103
ER_TOO_BIG_SELECT                          = 1104
ER_UNKNOWN_ERROR                           = 1105
ER_UNKNOWN_PROCEDURE                       = 1106
ER_WRONG_PARAMCOUNT_TO_PROCEDURE           = 1107
ER_WRONG_PARAMETERS_TO_PROCEDURE           = 1108
ER_UNKNOWN_TABLE                           = 1109
ER_FIELD_SPECIFIED_TWICE                   = 1110
ER_INVALID_GROUP_FUNC_USE                  = 1111
ER_UNSUPPORTED_EXTENSION                   = 1112
ER_TABLE_MUST_HAVE_COLUMNS                 = 1113
ER_RECORD_FILE_FULL                        = 1114
ER_UNKNOWN_CHARACTER_SET                   = 1115
ER_TOO_MANY_TABLES                         = 1116
ER_TOO_MANY_FIELDS                         = 1117
ER_TOO_BIG_ROWSIZE                         = 1118
ER_STACK_OVERRUN                           = 1119
ER_WRONG_OUTER_JOIN                        = 1120
ER_NULL_COLUMN_IN_INDEX                    = 1121
ER_CANT_FIND_UDF                           = 1122
ER_CANT_INITIALIZE_UDF                     = 1123
ER_UDF_NO_PATHS                            = 1124
ER_UDF_EXISTS                              = 1125
ER_CANT_OPEN_LIBRARY                       = 1126
ER_CANT_FIND_DL_ENTRY                      = 1127
ER_FUNCTION_NOT_DEFINED                    = 1128
ER_HOST_IS_BLOCKED                         = 1129
ER_HOST_NOT_PRIVILEGED                     = 1130
ER_PASSWORD_ANONYMOUS_USER                 = 1131
ER_PASSWORD_NOT_ALLOWED                    = 1132
ER_PASSWORD_NO_MATCH                       = 1133
ER_UPDATE_INFO                             = 1134
ER_CANT_CREATE_THREAD                      = 1135
ER_WRONG_VALUE_COUNT_ON_ROW                = 1136
ER_CANT_REOPEN_TABLE                       = 1137
ER_INVALID_USE_OF_NULL                     = 1138
ER_REGEXP_ERROR                            = 1139
ER_MIX_OF_GROUP_FUNC_AND_FIELDS            = 1140
ER_NONEXISTING_GRANT                       = 1141
ER_TABLEACCESS_DENIED_ERROR                = 1142
ER_COLUMNACCESS_DENIED_ERROR               = 1143
ER_ILLEGAL_GRANT_FOR_TABLE                 = 1144
ER_GRANT_WRONG_HOST_OR_USER                = 1145
ER_NO_SUCH_TABLE                           = 1146
ER_NONEXISTING_TABLE_GRANT                 = 1147
ER_NOT_ALLOWED_COMMAND                     = 1148
ER_SYNTAX_ERROR                            = 1149
ER_DELAYED_CANT_CHANGE_LOCK                = 1150
ER_TOO_MANY_DELAYED_THREADS                = 1151
ER_ABORTING_CONNECTION                     = 1152
ER_NET_PACKET_TOO_LARGE                    = 1153
ER_NET_READ_ERROR_FROM_PIPE                = 1154
ER_NET_FCNTL_ERROR                         = 1155
ER_NET_PACKETS_OUT_OF_ORDER                = 1156
ER_NET_UNCOMPRESS_ERROR                    = 1157
ER_NET_READ_ERROR                          = 1158
ER_NET_READ_INTERRUPTED                    = 1159
ER_NET_ERROR_ON_WRITE                      = 1160
ER_NET_WRITE_INTERRUPTED                   = 1161
ER_TOO_LONG_STRING                         = 1162
ER_TABLE_CANT_HANDLE_BLOB                  = 1163
ER_TABLE_CANT_HANDLE_AUTO_INCREMENT        = 1164
ER_DELAYED_INSERT_TABLE_LOCKED             = 1165
ER_WRONG_COLUMN_NAME                       = 1166
ER_WRONG_KEY_COLUMN                        = 1167
ER_WRONG_MRG_TABLE                         = 1168
ER_DUP_UNIQUE                              = 1169
ER_BLOB_KEY_WITHOUT_LENGTH                 = 1170
ER_PRIMARY_CANT_HAVE_NULL                  = 1171
ER_TOO_MANY_ROWS                           = 1172
ER_REQUIRES_PRIMARY_KEY                    = 1173
ER_NO_RAID_COMPILED                        = 1174
ER_UPDATE_WITHOUT_KEY_IN_SAFE_MODE         = 1175
ER_KEY_DOES_NOT_EXITS                      = 1176
ER_CHECK_NO_SUCH_TABLE                     = 1177
ER_CHECK_NOT_IMPLEMENTED                   = 1178
ER_CANT_DO_THIS_DURING_AN_TRANSACTION      = 1179
ER_ERROR_DURING_COMMIT                     = 1180
ER_ERROR_DURING_ROLLBACK                   = 1181
ER_ERROR_DURING_FLUSH_LOGS                 = 1182
ER_ERROR_DURING_CHECKPOINT                 = 1183
ER_NEW_ABORTING_CONNECTION                 = 1184
ER_DUMP_NOT_IMPLEMENTED                    = 1185
ER_FLUSH_MASTER_BINLOG_CLOSED              = 1186
ER_INDEX_REBUILD                           = 1187
ER_MASTER                                  = 1188
ER_MASTER_NET_READ                         = 1189
ER_MASTER_NET_WRITE                        = 1190
ER_FT_MATCHING_KEY_NOT_FOUND               = 1191
ER_LOCK_OR_ACTIVE_TRANSACTION              = 1192
ER_UNKNOWN_SYSTEM_VARIABLE                 = 1193
ER_CRASHED_ON_USAGE                        = 1194
ER_CRASHED_ON_REPAIR                       = 1195
ER_WARNING_NOT_COMPLETE_ROLLBACK           = 1196
ER_TRANS_CACHE_FULL                        = 1197
ER_SLAVE_MUST_STOP                         = 1198
ER_SLAVE_NOT_RUNNING                       = 1199
ER_BAD_SLAVE                               = 1200
ER_MASTER_INFO                             = 1201
ER_SLAVE_THREAD                            = 1202
ER_TOO_MANY_USER_CONNECTIONS               = 1203
ER_SET_CONSTANTS_ONLY                      = 1204
ER_LOCK_WAIT_TIMEOUT                       = 1205
ER_LOCK_TABLE_FULL                         = 1206
ER_READ_ONLY_TRANSACTION                   = 1207
ER_DROP_DB_WITH_READ_LOCK                  = 1208
ER_CREATE_DB_WITH_READ_LOCK                = 1209
ER_WRONG_ARGUMENTS                         = 1210
ER_NO_PERMISSION_TO_CREATE_USER            = 1211
ER_UNION_TABLES_IN_DIFFERENT_DIR           = 1212
ER_LOCK_DEADLOCK                           = 1213
ER_TABLE_CANT_HANDLE_FT                    = 1214
ER_CANNOT_ADD_FOREIGN                      = 1215
ER_NO_REFERENCED_ROW                       = 1216
ER_ROW_IS_REFERENCED                       = 1217
ER_CONNECT_TO_MASTER                       = 1218
ER_QUERY_ON_MASTER                         = 1219
ER_ERROR_WHEN_EXECUTING_COMMAND            = 1220
ER_WRONG_USAGE                             = 1221
ER_WRONG_NUMBER_OF_COLUMNS_IN_SELECT       = 1222
ER_CANT_UPDATE_WITH_READLOCK               = 1223
ER_MIXING_NOT_ALLOWED                      = 1224
ER_DUP_ARGUMENT                            = 1225
ER_USER_LIMIT_REACHED                      = 1226
ER_SPECIFIC_ACCESS_DENIED_ERROR            = 1227
ER_LOCAL_VARIABLE                          = 1228
ER_GLOBAL_VARIABLE                         = 1229
ER_NO_DEFAULT                              = 1230
ER_WRONG_VALUE_FOR_VAR                     = 1231
ER_WRONG_TYPE_FOR_VAR                      = 1232
ER_VAR_CANT_BE_READ                        = 1233
ER_CANT_USE_OPTION_HERE                    = 1234
ER_NOT_SUPPORTED_YET                       = 1235
ER_MASTER_FATAL_ERROR_READING_BINLOG       = 1236
ER_SLAVE_IGNORED_TABLE                     = 1237
ER_INCORRECT_GLOBAL_LOCAL_VAR              = 1238
ER_WRONG_FK_DEF                            = 1239
ER_KEY_REF_DO_NOT_MATCH_TABLE_REF          = 1240
ER_OPERAND_COLUMNS                         = 1241
ER_SUBQUERY_NO_1_ROW                       = 1242
ER_UNKNOWN_STMT_HANDLER                    = 1243
ER_CORRUPT_HELP_DB                         = 1244
ER_CYCLIC_REFERENCE                        = 1245
ER_AUTO_CONVERT                            = 1246
ER_ILLEGAL_REFERENCE                       = 1247
ER_DERIVED_MUST_HAVE_ALIAS                 = 1248
ER_SELECT_REDUCED                          = 1249
ER_TABLENAME_NOT_ALLOWED_HERE              = 1250
ER_NOT_SUPPORTED_AUTH_MODE                 = 1251
ER_SPATIAL_CANT_HAVE_NULL                  = 1252
ER_COLLATION_CHARSET_MISMATCH              = 1253
ER_SLAVE_WAS_RUNNING                       = 1254
ER_SLAVE_WAS_NOT_RUNNING                   = 1255
ER_TOO_BIG_FOR_UNCOMPRESS                  = 1256
ER_ZLIB_Z_MEM_ERROR                        = 1257
ER_ZLIB_Z_BUF_ERROR                        = 1258
ER_ZLIB_Z_DATA_ERROR                       = 1259
ER_CUT_VALUE_GROUP_CONCAT                  = 1260
ER_WARN_TOO_FEW_RECORDS                    = 1261
ER_WARN_TOO_MANY_RECORDS                   = 1262
ER_WARN_NULL_TO_NOTNULL                    = 1263
ER_WARN_DATA_OUT_OF_RANGE                  = 1264
WARN_DATA_TRUNCATED                        = 1265
ER_WARN_USING_OTHER_HANDLER                = 1266
ER_CANT_AGGREGATE_2COLLATIONS              = 1267
ER_DROP_USER                               = 1268
ER_REVOKE_GRANTS                           = 1269
ER_CANT_AGGREGATE_3COLLATIONS              = 1270
ER_CANT_AGGREGATE_NCOLLATIONS              = 1271
ER_VARIABLE_IS_NOT_STRUCT                  = 1272
ER_UNKNOWN_COLLATION                       = 1273
ER_SLAVE_IGNORED_SSL_PARAMS                = 1274
ER_SERVER_IS_IN_SECURE_AUTH_MODE           = 1275
ER_WARN_FIELD_RESOLVED                     = 1276
ER_BAD_SLAVE_UNTIL_COND                    = 1277
ER_MISSING_SKIP_SLAVE                      = 1278
ER_UNTIL_COND_IGNORED                      = 1279
ER_WRONG_NAME_FOR_INDEX                    = 1280
ER_WRONG_NAME_FOR_CATALOG                  = 1281
ER_WARN_QC_RESIZE                          = 1282
ER_BAD_FT_COLUMN                           = 1283
ER_UNKNOWN_KEY_CACHE                       = 1284
ER_WARN_HOSTNAME_WONT_WORK                 = 1285
ER_UNKNOWN_STORAGE_ENGINE                  = 1286
ER_WARN_DEPRECATED_SYNTAX                  = 1287
ER_NON_UPDATABLE_TABLE                     = 1288
ER_FEATURE_DISABLED                        = 1289
ER_OPTION_PREVENTS_STATEMENT               = 1290
ER_DUPLICATED_VALUE_IN_TYPE                = 1291
ER_TRUNCATED_WRONG_VALUE                   = 1292
ER_TOO_MUCH_AUTO_TIMESTAMP_COLS            = 1293
ER_INVALID_ON_UPDATE                       = 1294
ER_UNSUPPORTED_PS                          = 1295
ER_GET_ERRMSG                              = 1296
ER_GET_TEMPORARY_ERRMSG                    = 1297
ER_UNKNOWN_TIME_ZONE                       = 1298
ER_WARN_INVALID_TIMESTAMP                  = 1299
ER_INVALID_CHARACTER_STRING                = 1300
ER_WARN_ALLOWED_PACKET_OVERFLOWED          = 1301
ER_CONFLICTING_DECLARATIONS                = 1302
ER_SP_NO_RECURSIVE_CREATE                  = 1303
ER_SP_ALREADY_EXISTS                       = 1304
ER_SP_DOES_NOT_EXIST                       = 1305
ER_SP_DROP_FAILED                          = 1306
ER_SP_STORE_FAILED                         = 1307
ER_SP_LILABEL_MISMATCH                     = 1308
ER_SP_LABEL_REDEFINE                       = 1309
ER_SP_LABEL_MISMATCH                       = 1310
ER_SP_UNINIT_VAR                           = 1311
ER_SP_BADSELECT                            = 1312
ER_SP_BADRETURN                            = 1313
ER_SP_BADSTATEMENT                         = 1314
ER_UPDATE_LOG_DEPRECATED_IGNORED           = 1315
ER_UPDATE_LOG_DEPRECATED_TRANSLATED        = 1316
ER_QUERY_INTERRUPTED                       = 1317
ER_SP_WRONG_NO_OF_ARGS                     = 1318
ER_SP_COND_MISMATCH                        = 1319
ER_SP_NORETURN                             = 1320
ER_SP_NORETURNEND                          = 1321
ER_SP_BAD_CURSOR_QUERY                     = 1322
ER_SP_BAD_CURSOR_SELECT                    = 1323
ER_SP_CURSOR_MISMATCH                      = 1324
ER_SP_CURSOR_ALREADY_OPEN                  = 1325
ER_SP_CURSOR_NOT_OPEN                      = 1326
ER_SP_UNDECLARED_VAR                       = 1327
ER_SP_WRONG_NO_OF_FETCH_ARGS               = 1328
ER_SP_FETCH_NO_DATA                        = 1329
ER_SP_DUP_PARAM                            = 1330
ER_SP_DUP_VAR                              = 1331
ER_SP_DUP_COND                             = 1332
ER_SP_DUP_CURS                             = 1333
ER_SP_CANT_ALTER                           = 1334
ER_SP_SUBSELECT_NYI                        = 1335
ER_STMT_NOT_ALLOWED_IN_SF_OR_TRG           = 1336
ER_SP_VARCOND_AFTER_CURSHNDLR              = 1337
ER_SP_CURSOR_AFTER_HANDLER                 = 1338
ER_SP_CASE_NOT_FOUND                       = 1339
ER_FPARSER_TOO_BIG_FILE                    = 1340
ER_FPARSER_BAD_HEADER                      = 1341
ER_FPARSER_EOF_IN_COMMENT                  = 1342
ER_FPARSER_ERROR_IN_PARAMETER              = 1343
ER_FPARSER_EOF_IN_UNKNOWN_PARAMETER        = 1344
ER_VIEW_NO_EXPLAIN                         = 1345
ER_FRM_UNKNOWN_TYPE                        = 1346
ER_WRONG_OBJECT                            = 1347
ER_NONUPDATEABLE_COLUMN                    = 1348
ER_VIEW_SELECT_DERIVED                     = 1349
ER_VIEW_SELECT_CLAUSE                      = 1350
ER_VIEW_SELECT_VARIABLE                    = 1351
ER_VIEW_SELECT_TMPTABLE                    = 1352
ER_VIEW_WRONG_LIST                         = 1353
ER_WARN_VIEW_MERGE                         = 1354
ER_WARN_VIEW_WITHOUT_KEY                   = 1355
ER_VIEW_INVALID                            = 1356
ER_SP_NO_DROP_SP                           = 1357
ER_SP_GOTO_IN_HNDLR                        = 1358
ER_TRG_ALREADY_EXISTS                      = 1359
ER_TRG_DOES_NOT_EXIST                      = 1360
ER_TRG_ON_VIEW_OR_TEMP_TABLE               = 1361
ER_TRG_CANT_CHANGE_ROW                     = 1362
ER_TRG_NO_SUCH_ROW_IN_TRG                  = 1363
ER_NO_DEFAULT_FOR_FIELD                    = 1364
ER_DIVISION_BY_ZERO                        = 1365
ER_TRUNCATED_WRONG_VALUE_FOR_FIELD         = 1366
ER_ILLEGAL_VALUE_FOR_TYPE                  = 1367
ER_VIEW_NONUPD_CHECK                       = 1368
ER_VIEW_CHECK_FAILED                       = 1369
ER_PROCACCESS_DENIED_ERROR                 = 1370
ER_RELAY_LOG_FAIL                          = 1371
ER_PASSWD_LENGTH                           = 1372
ER_UNKNOWN_TARGET_BINLOG                   = 1373
ER_IO_ERR_LOG_INDEX_READ                   = 1374
ER_BINLOG_PURGE_PROHIBITED                 = 1375
ER_FSEEK_FAIL                              = 1376
ER_BINLOG_PURGE_FATAL_ERR                  = 1377
ER_LOG_IN_USE                              = 1378
ER_LOG_PURGE_UNKNOWN_ERR                   = 1379
ER_RELAY_LOG_INIT                          = 1380
ER_NO_BINARY_LOGGING                       = 1381
ER_RESERVED_SYNTAX                         = 1382
ER_WSAS_FAILED                             = 1383
ER_DIFF_GROUPS_PROC                        = 1384
ER_NO_GROUP_FOR_PROC                       = 1385
ER_ORDER_WITH_PROC                         = 1386
ER_LOGGING_PROHIBIT_CHANGING_OF            = 1387
ER_NO_FILE_MAPPING                         = 1388
ER_WRONG_MAGIC                             = 1389
ER_PS_MANY_PARAM                           = 1390
ER_KEY_PART_0                              = 1391
ER_VIEW_CHECKSUM                           = 1392
ER_VIEW_MULTIUPDATE                        = 1393
ER_VIEW_NO_INSERT_FIELD_LIST               = 1394
ER_VIEW_DELETE_MERGE_VIEW                  = 1395
ER_CANNOT_USER                             = 1396
ER_XAER_NOTA                               = 1397
ER_XAER_INVAL                              = 1398
ER_XAER_RMFAIL                             = 1399
ER_XAER_OUTSIDE                            = 1400
ER_XAER_RMERR                              = 1401
ER_XA_RBROLLBACK                           = 1402
ER_NONEXISTING_PROC_GRANT                  = 1403
ER_PROC_AUTO_GRANT_FAIL                    = 1404
ER_PROC_AUTO_REVOKE_FAIL                   = 1405
ER_DATA_TOO_LONG                           = 1406
ER_SP_BAD_SQLSTATE                         = 1407
ER_STARTUP                                 = 1408
ER_LOAD_FROM_FIXED_SIZE_ROWS_TO_VAR        = 1409
ER_CANT_CREATE_USER_WITH_GRANT             = 1410
ER_WRONG_VALUE_FOR_TYPE                    = 1411
ER_TABLE_DEF_CHANGED                       = 1412
ER_SP_DUP_HANDLER                          = 1413
ER_SP_NOT_VAR_ARG                          = 1414
ER_SP_NO_RETSET                            = 1415
ER_CANT_CREATE_GEOMETRY_OBJECT             = 1416
ER_FAILED_ROUTINE_BREAK_BINLOG             = 1417
ER_BINLOG_UNSAFE_ROUTINE                   = 1418
ER_BINLOG_CREATE_ROUTINE_NEED_SUPER        = 1419
ER_EXEC_STMT_WITH_OPEN_CURSOR              = 1420
ER_STMT_HAS_NO_OPEN_CURSOR                 = 1421
ER_COMMIT_NOT_ALLOWED_IN_SF_OR_TRG         = 1422
ER_NO_DEFAULT_FOR_VIEW_FIELD               = 1423
ER_SP_NO_RECURSION                         = 1424
ER_TOO_BIG_SCALE                           = 1425
ER_TOO_BIG_PRECISION                       = 1426
ER_M_BIGGER_THAN_D                         = 1427
ER_WRONG_LOCK_OF_SYSTEM_TABLE              = 1428
ER_CONNECT_TO_FOREIGN_DATA_SOURCE          = 1429
ER_QUERY_ON_FOREIGN_DATA_SOURCE            = 1430
ER_FOREIGN_DATA_SOURCE_DOESNT_EXIST        = 1431
ER_FOREIGN_DATA_STRING_INVALID_CANT_CREATE = 1432
ER_FOREIGN_DATA_STRING_INVALID             = 1433
ER_CANT_CREATE_FEDERATED_TABLE             = 1434
ER_TRG_IN_WRONG_SCHEMA                     = 1435
ER_STACK_OVERRUN_NEED_MORE                 = 1436
ER_TOO_LONG_BODY                           = 1437
ER_WARN_CANT_DROP_DEFAULT_KEYCACHE         = 1438
ER_TOO_BIG_DISPLAYWIDTH                    = 1439
ER_XAER_DUPID                              = 1440
ER_DATETIME_FUNCTION_OVERFLOW              = 1441
ER_CANT_UPDATE_USED_TABLE_IN_SF_OR_TRG     = 1442
ER_VIEW_PREVENT_UPDATE                     = 1443
ER_PS_NO_RECURSION                         = 1444
ER_SP_CANT_SET_AUTOCOMMIT                  = 1445
ER_MALFORMED_DEFINER                       = 1446
ER_VIEW_FRM_NO_USER                        = 1447
ER_VIEW_OTHER_USER                         = 1448
ER_NO_SUCH_USER                            = 1449
ER_FORBID_SCHEMA_CHANGE                    = 1450
ER_ROW_IS_REFERENCED_2                     = 1451
ER_NO_REFERENCED_ROW_2                     = 1452
ER_SP_BAD_VAR_SHADOW                       = 1453
ER_TRG_NO_DEFINER                          = 1454
ER_OLD_FILE_FORMAT                         = 1455
ER_SP_RECURSION_LIMIT                      = 1456
ER_SP_PROC_TABLE_CORRUPT                   = 1457
ER_SP_WRONG_NAME                           = 1458
ER_TABLE_NEEDS_UPGRADE                     = 1459
ER_SP_NO_AGGREGATE                         = 1460
ER_MAX_PREPARED_STMT_COUNT_REACHED         = 1461
ER_VIEW_RECURSIVE                          = 1462
ER_NON_GROUPING_FIELD_USED                 = 1463
ER_TABLE_CANT_HANDLE_SPKEYS                = 1464
ER_NO_TRIGGERS_ON_SYSTEM_SCHEMA            = 1465
ER_REMOVED_SPACES                          = 1466
ER_AUTOINC_READ_FAILED                     = 1467
ER_USERNAME                                = 1468
ER_HOSTNAME                                = 1469
ER_WRONG_STRING_LENGTH                     = 1470
ER_NON_INSERTABLE_TABLE                    = 1471
)</pre>
<p>
MySQL protocol types.
</p>
<p>
mymysql uses only some of them for send data to the MySQL server. Used
MySQL types are marked with a comment contains mymysql type that uses it.
</p>

<pre>const (
MYSQL_TYPE_DECIMAL     = 0x00
MYSQL_TYPE_TINY        = 0x01 <span class="comment">// int8, uint8</span>
MYSQL_TYPE_SHORT       = 0x02 <span class="comment">// int16, uint16</span>
MYSQL_TYPE_LONG        = 0x03 <span class="comment">// int32, uint32</span>
MYSQL_TYPE_FLOAT       = 0x04 <span class="comment">// float32</span>
MYSQL_TYPE_DOUBLE      = 0x05 <span class="comment">// float64</span>
MYSQL_TYPE_NULL        = 0x06 <span class="comment">// nil</span>
MYSQL_TYPE_TIMESTAMP   = 0x07 <span class="comment">// *Timestamp</span>
MYSQL_TYPE_LONGLONG    = 0x08 <span class="comment">// int64, uint64</span>
MYSQL_TYPE_INT24       = 0x09
MYSQL_TYPE_DATE        = 0x0a <span class="comment">// *Date</span>
MYSQL_TYPE_TIME        = 0x0b <span class="comment">// Time</span>
MYSQL_TYPE_DATETIME    = 0x0c <span class="comment">// *Datetime</span>
MYSQL_TYPE_YEAR        = 0x0d
MYSQL_TYPE_NEWDATE     = 0x0e
MYSQL_TYPE_VARCHAR     = 0x0f
MYSQL_TYPE_BIT         = 0x10
MYSQL_TYPE_NEWDECIMAL  = 0xf6
MYSQL_TYPE_ENUM        = 0xf7
MYSQL_TYPE_SET         = 0xf8
MYSQL_TYPE_TINY_BLOB   = 0xf9
MYSQL_TYPE_MEDIUM_BLOB = 0xfa
MYSQL_TYPE_LONG_BLOB   = 0xfb
MYSQL_TYPE_BLOB        = 0xfc <span class="comment">// Blob</span>
MYSQL_TYPE_VAR_STRING  = 0xfd <span class="comment">// []byte</span>
MYSQL_TYPE_STRING      = 0xfe <span class="comment">// string</span>
MYSQL_TYPE_GEOMETRY    = 0xff

MYSQL_UNSIGNED_MASK = uint16(1 &lt;&lt; 15)
)</pre>
<p>
Mapping of MySQL types to (prefered) protocol types. Use it if you create
your own Raw value.
</p>
<p>
Comments contains corresponding types used by mymysql. string type may be
replaced by []byte type and vice versa. []byte type is native for sending
on a network, so any string is converted to it before sending. Than for
better preformance use []byte.
</p>

<pre>const (
<span class="comment">// Client send and receive, mymysql representation for send / receive</span>
TINYINT   = MYSQL_TYPE_TINY      <span class="comment">// int8 / int8</span>
SMALLINT  = MYSQL_TYPE_SHORT     <span class="comment">// int16 / int16</span>
INT       = MYSQL_TYPE_LONG      <span class="comment">// int32 / int32</span>
BIGINT    = MYSQL_TYPE_LONGLONG  <span class="comment">// int64 / int64</span>
FLOAT     = MYSQL_TYPE_FLOAT     <span class="comment">// float32 / float32</span>
DOUBLE    = MYSQL_TYPE_DOUBLE    <span class="comment">// float64 / float32</span>
TIME      = MYSQL_TYPE_TIME      <span class="comment">// Time / Time</span>
DATE      = MYSQL_TYPE_DATE      <span class="comment">// *Date / *Date</span>
DATETIME  = MYSQL_TYPE_DATETIME  <span class="comment">// *Datetime / *Datetime</span>
TIMESTAMP = MYSQL_TYPE_TIMESTAMP <span class="comment">// *Timestamp / *Datetime</span>
CHAR      = MYSQL_TYPE_STRING    <span class="comment">// string / []byte</span>
BLOB      = MYSQL_TYPE_BLOB      <span class="comment">// Blob / []byte</span>
NULL      = MYSQL_TYPE_NULL      <span class="comment">// nil</span>


<span class="comment">// Client send only, mymysql representation for send</span>
OUT_TEXT      = MYSQL_TYPE_STRING <span class="comment">// string</span>
OUT_VARCHAR   = MYSQL_TYPE_STRING <span class="comment">// string</span>
OUT_BINARY    = MYSQL_TYPE_BLOB   <span class="comment">// Blob</span>
OUT_VARBINARY = MYSQL_TYPE_BLOB   <span class="comment">// Blob</span>


<span class="comment">// Client receive only, mymysql representation for receive</span>
IN_MEDIUMINT  = MYSQL_TYPE_INT24       <span class="comment">// int32</span>
IN_YEAR       = MYSQL_TYPE_SHORT       <span class="comment">// int16</span>
IN_BINARY     = MYSQL_TYPE_STRING      <span class="comment">// []byte</span>
IN_VARCHAR    = MYSQL_TYPE_VAR_STRING  <span class="comment">// []byte</span>
IN_VARBINARY  = MYSQL_TYPE_VAR_STRING  <span class="comment">// []byte</span>
IN_TINYBLOB   = MYSQL_TYPE_TINY_BLOB   <span class="comment">// []byte</span>
IN_TINYTEXT   = MYSQL_TYPE_TINY_BLOB   <span class="comment">// []byte</span>
IN_TEXT       = MYSQL_TYPE_BLOB        <span class="comment">// []byte</span>
IN_MEDIUMBLOB = MYSQL_TYPE_MEDIUM_BLOB <span class="comment">// []byte</span>
IN_MEDIUMTEXT = MYSQL_TYPE_MEDIUM_BLOB <span class="comment">// []byte</span>
IN_LONGBLOB   = MYSQL_TYPE_LONG_BLOB   <span class="comment">// []byte</span>
IN_LONGTEXT   = MYSQL_TYPE_LONG_BLOB   <span class="comment">// []byte</span>


<span class="comment">// MySQL 5.x specific</span>
IN_DECIMAL = MYSQL_TYPE_NEWDECIMAL <span class="comment">// TODO</span>
IN_BIT     = MYSQL_TYPE_BIT        <span class="comment">// []byte</span>
)</pre>
<h2 id="Variables">Variables</h2>

<pre>var (
SEQ_ERROR             = os.NewError(&#34;packet sequence error&#34;)
PKT_ERROR             = os.NewError(&#34;malformed packet&#34;)
PKT_LONG_ERROR        = os.NewError(&#34;packet too long&#34;)
UNEXP_NULL_LCS_ERROR  = os.NewError(&#34;unexpected null LCS&#34;)
UNEXP_NULL_LCB_ERROR  = os.NewError(&#34;unexpected null LCB&#34;)
UNEXP_NULL_DATE_ERROR = os.NewError(&#34;unexpected null datetime&#34;)
UNEXP_NULL_TIME_ERROR = os.NewError(&#34;unexpected null time&#34;)
UNK_RESULT_PKT_ERROR  = os.NewError(&#34;unexpected or unknown result packet&#34;)
NOT_CONN_ERROR        = os.NewError(&#34;not connected&#34;)
ALREDY_CONN_ERROR     = os.NewError(&#34;not connected&#34;)
BAD_RESULT_ERROR      = os.NewError(&#34;unexpected result&#34;)
UNREADED_ROWS_ERROR   = os.NewError(&#34;there are unreaded rows&#34;)
BIND_COUNT_ERROR      = os.NewError(&#34;wrong number of values for bind&#34;)
BIND_UNK_TYPE         = os.NewError(&#34;unknown value type for bind&#34;)
RESULT_COUNT_ERROR    = os.NewError(&#34;wrong number of result columns&#34;)
BAD_COMMAND_ERROR     = os.NewError(&#34;comand isn&#39;t text SQL nor *Statement&#34;)
WRONG_DATE_LEN_ERROR  = os.NewError(&#34;wrong datetime/timestamp length&#34;)
WRONG_TIME_LEN_ERROR  = os.NewError(&#34;wrong time length&#34;)
UNK_MYSQL_TYPE_ERROR  = os.NewError(&#34;unknown MySQL type&#34;)
WRONG_PARAM_NUM_ERROR = os.NewError(&#34;wrong parameter number&#34;)
UNK_DATA_TYPE_ERROR   = os.NewError(&#34;unknown data source type&#34;)
SMALL_PKT_SIZE_ERROR  = os.NewError(&#34;specified packet size is to small&#34;)
)</pre>
<h2 id="DecodeU16">func <a href="/mymysql/codecs.go?s=68:101#L1">DecodeU16</a></h2>
<p><code>func DecodeU16(buf []byte) uint16</code></p>

<h2 id="DecodeU24">func <a href="/mymysql/codecs.go?s=268:301#L8">DecodeU24</a></h2>
<p><code>func DecodeU24(buf []byte) uint32</code></p>

<h2 id="DecodeU32">func <a href="/mymysql/codecs.go?s=492:525#L17">DecodeU32</a></h2>
<p><code>func DecodeU32(buf []byte) uint32</code></p>

<h2 id="DecodeU64">func <a href="/mymysql/codecs.go?s=748:786#L27">DecodeU64</a></h2>
<p><code>func DecodeU64(buf []byte) (rv uint64)</code></p>

<h2 id="EncodeDate">func <a href="/mymysql/codecs.go?s=9907:9939#L471">EncodeDate</a></h2>
<p><code>func EncodeDate(dd *Date) []byte</code></p>

<h2 id="EncodeDatetime">func <a href="/mymysql/codecs.go?s=8541:8581#L409">EncodeDatetime</a></h2>
<p><code>func EncodeDatetime(dt *Datetime) []byte</code></p>

<h2 id="EncodeTime">func <a href="/mymysql/codecs.go?s=6484:6516#L313">EncodeTime</a></h2>
<p><code>func EncodeTime(tt *Time) []byte</code></p>

<h2 id="EncodeU16">func <a href="/mymysql/codecs.go?s=998:1031#L39">EncodeU16</a></h2>
<p><code>func EncodeU16(val uint16) []byte</code></p>

<h2 id="EncodeU24">func <a href="/mymysql/codecs.go?s=1156:1189#L46">EncodeU24</a></h2>
<p><code>func EncodeU24(val uint32) []byte</code></p>

<h2 id="EncodeU32">func <a href="/mymysql/codecs.go?s=1331:1364#L53">EncodeU32</a></h2>
<p><code>func EncodeU32(val uint32) []byte</code></p>

<h2 id="EncodeU64">func <a href="/mymysql/codecs.go?s=1523:1556#L60">EncodeU64</a></h2>
<p><code>func EncodeU64(val uint64) []byte</code></p>

<h2 id="IsDateZero">func <a href="/mymysql/common.go?s=2512:2542#L112">IsDateZero</a></h2>
<p><code>func IsDateZero(dd *Date) bool</code></p>
<p>
True if date is 0000-00-00
</p>

<h2 id="IsDatetimeZero">func <a href="/mymysql/common.go?s=1687:1725#L79">IsDatetimeZero</a></h2>
<p><code>func IsDatetimeZero(dt *Datetime) bool</code></p>
<p>
True if datetime is 0000-00-00 00:00:00
</p>

<h2 id="IsNetErr">func <a href="/mymysql/autoconnect.go?s=137:169#L2">IsNetErr</a></h2>
<p><code>func IsNetErr(err os.Error) bool</code></p>
<p>
Return true if error is network error or UnexpectedEOF.
</p>

<h2 id="NbinToNstr">func <a href="/mymysql/addons.go?s=17:54#L1">NbinToNstr</a></h2>
<p><code>func NbinToNstr(nbin *[]byte) *string</code></p>

<h2 id="NstrToNbin">func <a href="/mymysql/addons.go?s=147:184#L1">NstrToNbin</a></h2>
<p><code>func NstrToNbin(nstr *string) *[]byte</code></p>

<h2 id="Blob">type <a href="/mymysql/binding.go?s=1505:1521#L61">Blob</a></h2>

<p><pre>type Blob []byte</pre></p>
<h2 id="Date">type <a href="/mymysql/binding.go?s=594:651#L20">Date</a></h2>

<p><pre>type Date struct {
Year       int16
Month, Day uint8
}</pre></p>
<h3 id="Date.StrToDate">func <a href="/mymysql/common.go?s=2729:2766#L118">StrToDate</a></h3>
<p><code>func StrToDate(str string) (dd *Date)</code></p>
<p>
Convert string date in format YYYY-MM-DD to Date.
Leading and trailing spaces are ignored. If format is invalid returns nil.
</p>

<h3 id="Date.String">func (*Date) <a href="/mymysql/binding.go?s=652:683#L24">String</a></h3>
<p><code>func (dd *Date) String() string</code></p>

<h2 id="Datetime">type <a href="/mymysql/binding.go?s=53:155#L1">Datetime</a></h2>

<p><pre>type Datetime struct {
Year                             int16
Month, Day, Hour, Minute, Second uint8
Nanosec                          uint32
}</pre></p>
<h3 id="Datetime.DateToDatetime">func <a href="/mymysql/common.go?s=2292:2331#L100">DateToDatetime</a></h3>
<p><code>func DateToDatetime(dd *Date) *Datetime</code></p>
<p>
Convert *Date to *Datetime. Return nil if dd is nil
</p>

<h3 id="Datetime.StrToDatetime">func <a href="/mymysql/common.go?s=3387:3432#L139">StrToDatetime</a></h3>
<p><code>func StrToDatetime(str string) (dt *Datetime)</code></p>
<p>
Convert string datetime in format YYYY-MM-DD[ HH:MM:SS] to Datetime.
Leading and trailing spaces are ignored. If format is invalid returns nil.
</p>

<h3 id="Datetime.TimeToDatetime">func <a href="/mymysql/common.go?s=1918:1962#L85">TimeToDatetime</a></h3>
<p><code>func TimeToDatetime(tt *time.Time) *Datetime</code></p>
<p>
Convert *time.Time to *Datetime. Return nil if tt is nil
</p>

<h3 id="Datetime.String">func (*Datetime) <a href="/mymysql/binding.go?s=156:191#L3">String</a></h3>
<p><code>func (dt *Datetime) String() string</code></p>

<h2 id="Error">type <a href="/mymysql/errors.go?s=1713:1768#L26">Error</a></h2>
<p>
If function/method returns error you can check returned error type, and if
it is *mymy.Error it is error received from MySQL server. Next you can check
Code for error number.
</p>

<p><pre>type Error struct {
Code uint16
Msg  []byte
}</pre></p>
<h3 id="Error.String">func (Error) <a href="/mymysql/errors.go?s=1770:1802#L31">String</a></h3>
<p><code>func (err Error) String() string</code></p>

<h2 id="Field">type <a href="/mymysql/result.go?s=95:332#L2">Field</a></h2>

<p><pre>type Field struct {
Catalog  string
Db       string
Table    string
OrgTable string
Name     string
OrgName  string
DispLen  uint32
<span class="comment">//  Charset  uint16</span>
Flags uint16
Type  byte
Scale byte
}</pre></p>
<h2 id="MySQL">type <a href="/mymysql/mysql.go?s=276:1359#L13">MySQL</a></h2>
<p>
MySQL connection handler
</p>

<p><pre>type MySQL struct {

<span class="comment">// Current status of MySQL server connection</span>
Status uint16

<span class="comment">// Maximum packet size that client can accept from server.</span>
<span class="comment">// Default 16*1024*1024-1. You may change it before connect.</span>
MaxPktSize int

<span class="comment">// Debug logging. You may change it at any time.</span>
Debug bool

<span class="comment">// Maximum reconnect retries - for XxxAC methods. Default is 5 which</span>
<span class="comment">// means 1+2+3+4+5 = 15 seconds before return an error.</span>
MaxRetries int
<span class="comment">// contains unexported fields</span>
}</pre></p>
<h3 id="MySQL.New">func <a href="/mymysql/mysql.go?s=1590:1666#L53">New</a></h3>
<p><code>func New(proto, laddr, raddr, user, passwd string, db ...string) (my *MySQL)</code></p>
<p>
Create new MySQL handler. The first three arguments are passed to net.Bind
for create connection. user and passwd are for authentication. Optional db
is database name (you may not specifi it and use Use() method later).
</p>

<h3 id="MySQL.Close">func (*MySQL) <a href="/mymysql/mysql.go?s=4965:5004#L191">Close</a></h3>
<p><code>func (my *MySQL) Close() (err os.Error)</code></p>
<p>
Close connection to the server
</p>

<h3 id="MySQL.Connect">func (*MySQL) <a href="/mymysql/mysql.go?s=4290:4331#L158">Connect</a></h3>
<p><code>func (my *MySQL) Connect() (err os.Error)</code></p>
<p>
Establishes a connection with MySQL server version 4.1 or later.
</p>

<h3 id="MySQL.EscapeString">func (*MySQL) <a href="/mymysql/mysql.go?s=20010:20058#L744">EscapeString</a></h3>
<p><code>func (my *MySQL) EscapeString(txt string) string</code></p>
<p>
Escapes special characters in the txt, so it is safe to place returned string
to Query or Start method.
</p>

<h3 id="MySQL.IsConnected">func (*MySQL) <a href="/mymysql/mysql.go?s=4500:4535#L170">IsConnected</a></h3>
<p><code>func (my *MySQL) IsConnected() bool</code></p>
<p>
Check if connection is established
</p>

<h3 id="MySQL.Ping">func (*MySQL) <a href="/mymysql/mysql.go?s=10075:10113#L399">Ping</a></h3>
<p><code>func (my *MySQL) Ping() (err os.Error)</code></p>
<p>
Send MySQL PING to the server.
</p>

<h3 id="MySQL.Prepare">func (*MySQL) <a href="/mymysql/mysql.go?s=10988:11056#L441">Prepare</a></h3>
<p><code>func (my *MySQL) Prepare(sql string) (stmt *Statement, err os.Error)</code></p>
<p>
Prepare server side statement. Return statement handler.
</p>

<h3 id="MySQL.PrepareAC">func (*MySQL) <a href="/mymysql/autoconnect.go?s=1834:1904#L73">PrepareAC</a></h3>
<p><code>func (my *MySQL) PrepareAC(sql string) (stmt *Statement, err os.Error)</code></p>
<p>
Automatic connect/reconnect/repeat version of Prepare
</p>

<h3 id="MySQL.Query">func (*MySQL) <a href="/mymysql/mysql.go?s=9654:9761#L379">Query</a></h3>
<p><code>func (my *MySQL) Query(sql string, params ...interface{}) (rows []*Row, res *Result, err os.Error)</code></p>
<p>
This call Start and next call GetRow as long as it reads all rows from the
result. Next it returns all readed rows as the slice of rows.
</p>

<h3 id="MySQL.QueryAC">func (*MySQL) <a href="/mymysql/autoconnect.go?s=1365:1474#L54">QueryAC</a></h3>
<p><code>func (my *MySQL) QueryAC(sql string, params ...interface{}) (rows []*Row, res *Result, err os.Error)</code></p>
<p>
Automatic connect/reconnect/repeat version of Query
</p>

<h3 id="MySQL.Reconnect">func (*MySQL) <a href="/mymysql/mysql.go?s=5324:5367#L207">Reconnect</a></h3>
<p><code>func (my *MySQL) Reconnect() (err os.Error)</code></p>
<p>
Close and reopen connection in one, thread-safe operation.
Ignore unreaded rows, reprepare all prepared statements.
</p>

<h3 id="MySQL.Register">func (*MySQL) <a href="/mymysql/mysql.go?s=19812:19849#L738">Register</a></h3>
<p><code>func (my *MySQL) Register(sql string)</code></p>
<p>
Register MySQL command/query to be executed immediately after connecting to
the server. You may register multiple commands. They will be executed in
the order of registration. Yhis method is mainly useful for reconnect.
</p>

<h3 id="MySQL.Start">func (*MySQL) <a href="/mymysql/mysql.go?s=7314:7408#L293">Start</a></h3>
<p><code>func (my *MySQL) Start(sql string, params ...interface{}) (res *Result, err os.Error)</code></p>
<p>
Start new query.
</p>
<p>
If you specify the parameters, the SQL string will be a result of
fmt.Sprintf(sql, params...).
You must get all result rows (if they exists) before next query.
</p>

<h3 id="MySQL.ThreadId">func (*MySQL) <a href="/mymysql/mysql.go?s=19517:19551#L731">ThreadId</a></h3>
<p><code>func (my *MySQL) ThreadId() uint32</code></p>
<p>
Returns the thread ID of the current connection.
</p>

<h3 id="MySQL.Use">func (*MySQL) <a href="/mymysql/mysql.go?s=6196:6246#L243">Use</a></h3>
<p><code>func (my *MySQL) Use(dbname string) (err os.Error)</code></p>
<p>
Change database
</p>

<h3 id="MySQL.UseAC">func (*MySQL) <a href="/mymysql/autoconnect.go?s=977:1029#L37">UseAC</a></h3>
<p><code>func (my *MySQL) UseAC(dbname string) (err os.Error)</code></p>
<p>
Automatic connect/reconnect/repeat version of Use
</p>

<h2 id="Raw">type <a href="/mymysql/binding.go?s=1523:1573#L63">Raw</a></h2>

<p><pre>type Raw struct {
Typ uint16
Val *[]byte
}</pre></p>
<h2 id="Result">type <a href="/mymysql/result.go?s=334:912#L16">Result</a></h2>

<p><pre>type Result struct {
FieldCount int
Fields     []*Field       <span class="comment">// Fields table</span>
Map        map[string]int <span class="comment">// Maps field name to column number</span>

Message      []byte
AffectedRows uint64

<span class="comment">// Primary key value (useful for AUTO_INCREMENT primary keys)</span>
InsertId uint64

<span class="comment">// Number of warinigs during command execution</span>
<span class="comment">// You can use the SHOW WARNINGS query for details.</span>
WarningCount int

<span class="comment">// MySQL server status immediately after the query execution</span>
Status uint16
<span class="comment">// contains unexported fields</span>
}</pre></p>
<h3 id="Result.End">func (*Result) <a href="/mymysql/mysql.go?s=9374:9413#L370">End</a></h3>
<p><code>func (res *Result) End() (err os.Error)</code></p>
<p>
Read all unreaded rows and discard them. This function is useful if you
don&#39;t want to use the remaining rows. It has an impact only on current
result. If there is multi result query, you must use NextResult method and
read/discard all rows in this result, before use other method that sends
data to the server.
</p>

<h3 id="Result.GetRow">func (*Result) <a href="/mymysql/mysql.go?s=8276:8328#L338">GetRow</a></h3>
<p><code>func (res *Result) GetRow() (row *Row, err os.Error)</code></p>
<p>
Get the data row from a server. This method reads one row of result directly
from network connection (without rows buffering on client side).
</p>

<h3 id="Result.NextResult">func (*Result) <a href="/mymysql/mysql.go?s=8859:8919#L357">NextResult</a></h3>
<p><code>func (res *Result) NextResult() (next *Result, err os.Error)</code></p>
<p>
This function is used when last query was the multi result query.
Return the next result or nil if no more resuts exists.
</p>

<h2 id="Row">type <a href="/mymysql/result.go?s=1308:1350#L46">Row</a></h2>
<p>
Result row. Data field is a slice that contains values for any column of
received row.
</p>
<p>
If row is a result of ordinary text query, an element of Data field can be
[]byte slice, contained result text or nil if NULL is returned.
</p>
<p>
If it is result of prepared statement execution, an element of Data field can
be: intXX, uintXX, floatXX, []byte, *Date, *Datetime, Time or nil
</p>

<p><pre>type Row struct {
Data []interface{}
}</pre></p>
<h3 id="Row.Bin">func (*Row) <a href="/mymysql/result.go?s=1418:1457#L51">Bin</a></h3>
<p><code>func (tr *Row) Bin(nn int) (bin []byte)</code></p>
<p>
Get the nn-th value and return it as []byte ([]byte{} if NULL)
</p>

<h3 id="Row.Date">func (*Row) <a href="/mymysql/result.go?s=5316:5355#L205">Date</a></h3>
<p><code>func (tr *Row) Date(nn int) (val *Date)</code></p>
<p>
It is like DateErr but return 0000-00-00 if conversion is impossible.
</p>

<h3 id="Row.DateErr">func (*Row) <a href="/mymysql/result.go?s=4667:4723#L178">DateErr</a></h3>
<p><code>func (tr *Row) DateErr(nn int) (val *Date, err os.Error)</code></p>
<p>
Get the nn-th value and return it as Date (0000-00-00 if NULL). Return error
if conversion is impossible.
</p>

<h3 id="Row.Datetime">func (*Row) <a href="/mymysql/result.go?s=6356:6403#L245">Datetime</a></h3>
<p><code>func (tr *Row) Datetime(nn int) (val *Datetime)</code></p>
<p>
It is like DatetimeErr but return 0000-00-00 00:00:00 if conversion is
impossible.
</p>

<h3 id="Row.DatetimeErr">func (*Row) <a href="/mymysql/result.go?s=5608:5672#L215">DatetimeErr</a></h3>
<p><code>func (tr *Row) DatetimeErr(nn int) (val *Datetime, err os.Error)</code></p>
<p>
Get the nn-th value and return it as Datetime (0000-00-00 00:00:00 if NULL).
Return error if conversion is impossible. It can convert Date to Datetime.
</p>

<h3 id="Row.Int">func (*Row) <a href="/mymysql/result.go?s=3308:3344#L128">Int</a></h3>
<p><code>func (tr *Row) Int(nn int) (val int)</code></p>
<p>
Get the nn-th value and return it as int. Return 0 if value is NULL or
conversion is impossible.
</p>

<h3 id="Row.IntErr">func (*Row) <a href="/mymysql/result.go?s=2084:2137#L80">IntErr</a></h3>
<p><code>func (tr *Row) IntErr(nn int) (val int, err os.Error)</code></p>
<p>
Get the nn-th value and return it as int (0 if NULL). Return error if
conversion is impossible.
</p>

<h3 id="Row.MustDate">func (*Row) <a href="/mymysql/result.go?s=5107:5150#L196">MustDate</a></h3>
<p><code>func (tr *Row) MustDate(nn int) (val *Date)</code></p>
<p>
It is like DateErr but panics if conversion is impossible.
</p>

<h3 id="Row.MustDatetime">func (*Row) <a href="/mymysql/result.go?s=6119:6170#L235">MustDatetime</a></h3>
<p><code>func (tr *Row) MustDatetime(nn int) (val *Datetime)</code></p>
<p>
As DatetimeErr but panics if conversion is impossible.
</p>

<h3 id="Row.MustInt">func (*Row) <a href="/mymysql/result.go?s=3073:3113#L118">MustInt</a></h3>
<p><code>func (tr *Row) MustInt(nn int) (val int)</code></p>
<p>
Get the nn-th value and return it as int (0 if NULL). Panic if conversion is
impossible.
</p>

<h3 id="Row.MustTime">func (*Row) <a href="/mymysql/result.go?s=7103:7145#L277">MustTime</a></h3>
<p><code>func (tr *Row) MustTime(nn int) (val Time)</code></p>
<p>
It is like TimeErr but panics if conversion is impossible.
</p>

<h3 id="Row.MustUint">func (*Row) <a href="/mymysql/result.go?s=4233:4275#L161">MustUint</a></h3>
<p><code>func (tr *Row) MustUint(nn int) (val uint)</code></p>
<p>
Get the nn-th value and return it as uint (0 if NULL). Panic if conversion is
impossible.
</p>

<h3 id="Row.Str">func (*Row) <a href="/mymysql/result.go?s=1758:1797#L66">Str</a></h3>
<p><code>func (tr *Row) Str(nn int) (str string)</code></p>
<p>
Get the nn-th value and return it as string (&#34;&#34; if NULL)
</p>

<h3 id="Row.Time">func (*Row) <a href="/mymysql/result.go?s=7308:7346#L286">Time</a></h3>
<p><code>func (tr *Row) Time(nn int) (val Time)</code></p>
<p>
It is like TimeErr but return 0:00:00 if conversion is impossible.
</p>

<h3 id="Row.TimeErr">func (*Row) <a href="/mymysql/result.go?s=6615:6670#L255">TimeErr</a></h3>
<p><code>func (tr *Row) TimeErr(nn int) (val Time, err os.Error)</code></p>
<p>
Get the nn-th value and return it as Time (0:00:00 if NULL). Return error
if conversion is impossible.
</p>

<h3 id="Row.Uint">func (*Row) <a href="/mymysql/result.go?s=4472:4510#L171">Uint</a></h3>
<p><code>func (tr *Row) Uint(nn int) (val uint)</code></p>
<p>
Get the nn-th value and return it as uint. Return 0 if value is NULL or
conversion is impossible.
</p>

<h3 id="Row.UintErr">func (*Row) <a href="/mymysql/result.go?s=3491:3546#L135">UintErr</a></h3>
<p><code>func (tr *Row) UintErr(nn int) (val uint, err os.Error)</code></p>
<p>
Get the nn-th value and return it as uint (0 if NULL). Return error if
conversion is impossible.
</p>

<h2 id="Statement">type <a href="/mymysql/prepared.go?s=39:379#L1">Statement</a></h2>

<p><pre>type Statement struct {
Fields []*Field
Map    map[string]int <span class="comment">// Maps field name to column number</span>

FieldCount   int
ParamCount   int
WarningCount int
Status       uint16
<span class="comment">// contains unexported fields</span>
}</pre></p>
<h3 id="Statement.BindParams">func (*Statement) <a href="/mymysql/mysql.go?s=12240:12296#L478">BindParams</a></h3>
<p><code>func (stmt *Statement) BindParams(params ...interface{})</code></p>
<p>
Bind input data for the parameter markers in the SQL statement that was
passed to Prepare.
</p>
<p>
params may be a parameter list (slice), a struct or a pointer to the struct.
A struct field can by value or pointer to value. A parameter (slice element)
can be value, pointer to value or pointer to pointer to value.
Values may be of the folowind types: intXX, uintXX, floatXX, []byte, Blob,
string, Datetime, Date, Time, Timestamp, Raw.
</p>
<p>
Warning! This method isn&#39;t thread safe. If you use the same prepared
statement in multiple threads, you should not use this method unless you know
exactly what you are doing. For each thread you may prepare its own statement
or use Run, Exec or ExecAC method with parameters (but they rebind parameters
on each call).
</p>

<h3 id="Statement.Delete">func (*Statement) <a href="/mymysql/mysql.go?s=15471:15517#L595">Delete</a></h3>
<p><code>func (stmt *Statement) Delete() (err os.Error)</code></p>
<p>
Destroy statement on server side. Client side handler is invalid after this
command.
</p>

<h3 id="Statement.Exec">func (*Statement) <a href="/mymysql/mysql.go?s=15005:15105#L574">Exec</a></h3>
<p><code>func (stmt *Statement) Exec(params ...interface{}) (rows []*Row, res *Result, err os.Error)</code></p>
<p>
This call Run and next call GetRow once or more times. It read all rows
from connection and returns they as a slice.
</p>

<h3 id="Statement.ExecAC">func (*Statement) <a href="/mymysql/autoconnect.go?s=2246:2348#L90">ExecAC</a></h3>
<p><code>func (stmt *Statement) ExecAC(params ...interface{}) (rows []*Row, res *Result, err os.Error)</code></p>
<p>
Automatic connect/reconnect/repeat version of Exec
</p>

<h3 id="Statement.Reset">func (*Statement) <a href="/mymysql/mysql.go?s=16195:16240#L623">Reset</a></h3>
<p><code>func (stmt *Statement) Reset() (err os.Error)</code></p>
<p>
Resets a prepared statement on server: data sent to the server, unbuffered
result sets and current errors.
</p>

<h3 id="Statement.ResetParams">func (*Statement) <a href="/mymysql/mysql.go?s=13987:14023#L537">ResetParams</a></h3>
<p><code>func (stmt *Statement) ResetParams()</code></p>
<p>
Resets the previous parameter binding
</p>

<h3 id="Statement.Run">func (*Statement) <a href="/mymysql/mysql.go?s=14306:14383#L547">Run</a></h3>
<p><code>func (stmt *Statement) Run(params ...interface{}) (res *Result, err os.Error)</code></p>
<p>
Execute prepared statement. If statement requires parameters you may bind
them first or specify directly. After this command you may use GetRow to
retrieve data.
</p>

<h3 id="Statement.SendLongData">func (*Statement) <a href="/mymysql/mysql.go?s=17738:17839#L664">SendLongData</a></h3>
<p><code>func (stmt *Statement) SendLongData(pnum int, data interface{}, pkt_size int) (err os.Error)</code></p>
<p>
Send long data to MySQL server in chunks.
You can call this method after Bind and before Run/Execute. It can be called
multiple times for one parameter to send TEXT or BLOB data in chunks.
</p>
<p>
pnum     - Parameter number to associate the data with.
</p>
<p>
data     - Data source string, []byte or io.Reader.
</p>
<p>
pkt_size - It must be must be greater than 6 and less or equal to MySQL
max_allowed_packet variable. You can obtain value of this variable
using such query: SHOW variables WHERE Variable_name = &#39;max_allowed_packet&#39;
If data source is io.Reader then (pkt_size - 6) is size of a buffer that
will be allocated for reading.
</p>
<p>
If you have data source of type string or []byte in one piece you may
properly set pkt_size and call this method once. If you have data in
multiple pieces you can call this method multiple times. If data source is
io.Reader you should properly set pkt_size. Data will be readed from
io.Reader and send in pieces to the server until EOF.
</p>

<h2 id="Time">type <a href="/mymysql/binding.go?s=1036:1051#L38">Time</a></h2>
<p>
MySQL TIME in nanoseconds. Note that MySQL doesn&#39;t store fractional part
of second but it is permitted for temporal values.
</p>

<p><pre>type Time int64</pre></p>
<h3 id="Time.StrToTime">func <a href="/mymysql/common.go?s=4066:4098#L167">StrToTime</a></h3>
<p><code>func StrToTime(str string) *Time</code></p>
<p>
Convert string time in format [+-]H+:MM:SS[.UUUUUUUUU] to Time.
Leading and trailing spaces are ignored. If format is invalid returns nil.
</p>

<h3 id="Time.String">func (*Time) <a href="/mymysql/binding.go?s=1052:1083#L39">String</a></h3>
<p><code>func (tt *Time) String() string</code></p>

<h2 id="Timestamp">type <a href="/mymysql/binding.go?s=804:827#L31">Timestamp</a></h2>

<p><pre>type Timestamp Datetime</pre></p>
<h3 id="Timestamp.String">func (*Timestamp) <a href="/mymysql/binding.go?s=828:864#L32">String</a></h3>
<p><code>func (ts *Timestamp) String() string</code></p>

<h2 id="Subdirectories">Subdirectories</h2>
<p>
<table class="layout">
<tr>
<th align="left" colspan="1">Name</th>
<td width="25">&nbsp;</td>
<th align="left">Synopsis</th>
</tr>
<tr>
<th align="left"><a href="..">..</a></th>
</tr>
<tr>

<td align="left" colspan="1"><a href=".git">.git</a></td>
<td></td>
<td align="left"></td>
</tr>
<tr>

<td align="left" colspan="1"><a href="doc">doc</a></td>
<td></td>
<td align="left"></td>
</tr>
<tr>

<td align="left" colspan="1"><a href="examples">examples</a></td>
<td></td>
<td align="left"></td>
</tr>
</table>
</p>
