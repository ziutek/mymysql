## MyMySQL v0.3

This package contains MySQL client API written entirely in Go. It was created
due to lack of properly working MySQL API package, ready for my production
application (December 2010).

The code of this package is carefuly written and has internal error handling
using *panic()* exceptions, thus the probability of Go bugs or an unhandled
internal errors should be very small. Unfortunately I'm not a MySQL protocol
expert, so bugs in the protocol handling are possible.

## Instaling

    $ git clone git://github.com/ziutek/mymysql
    $ cd mymysql
    $ make install

## Testing

For testing you need test database and test user:

    mysql> create database test;
    mysql> grant all privileges on test.* to testuser@localhost;
    mysql> set password for testuser@localhost = password("TestPasswd9")

Make sure that MySQL variable *max_allowed_packet* is greater than 33M (needed
to test long packets). If not, change it in *my.cnf* file and restart MySQL
daemon. The default MySQL server addres is *127.0.0.1:3306*. You can change it
by edit *mymy_test.go* file.

Next run tests:

    $ cd mymysql
    $ gotest -v

## Interface

There is one change in v0.3, which doesn't preserve backwards compatibility
with v0.2: the name of *Execute* method was changed to *Run*. A new *Exec*
method for Statement struct was added. It is similar in result to *Query*
method.

In *GODOC.html* or *GODOC.txt* you can find the full documentation of this package in godoc format.

## Example 1

    import "mymy"

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
    if err != nil {
        panic(err)
    }

    // Print fields names
    for _, field := range res.Fields {
        fmt.Print(field.Name, " ")
    }
    fmt.Println()

    // Print all rows
    for {
        row, err := res.GetRow()
        if err != nil {
            panic(err)
        }
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

You can use *Start* or *Query* method for prepared statements:

    stmt, err := db.Prepare("insert into X values (?, ?)")
    if err != nil {
        panic(err)
    }

    type Data struct {
        id  int
        tax *float // nil means NULL
    }

    data = new(Data)

    for {
        err := getData(data)
        if err == endOfData {
            break       
        } else if err != nil {
            panic(err)
        }
        _, err = db.Start(stmt, data.id, data.tax)
        if err != nil {
            panic(err)
        }
    }

*getData* is your function which retrieves data from somewhere and set *id* and
*tax* fields of the Data struct. In the case of *tax* field *getData* may
assign pointer to retieved variable or nil if NULL should be stored in
database.

With *Start* and *Query* methods data are rebinded on every method call. It
isn't efficient if statement is executer more than once. You can bind
parameters and use *Run* method to avoid these unnecessary rebinds. The
simplest way to bind parameters is:

    stmt.BindParams(data.id, data.tax)

but you can't use it in our example, becouse parameters binded this way can't
be changed by *getData* function. You may modify bind like this:

    stmt.BindParams(&data.id, &data.tax)

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
        } else if err != nil {
            panic(err)
        }
        _, err = stmt.Run()
        if err != nil {
            panic(err)
        }
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
        // the resp.Body implements io.Reader
        err = ins.SendLongData(1, resp.Body, 4092)
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
         TIME, DATE, DATETIME, TIMESTAMP  -->  *mymy.Datetime
                                    YEAR  -->  int16
        CHAR, VARCHAR, BINARY, VARBINARY  -->  []byte
     TEXT, TINYTEXT, MEDIUMTEXT, LONGTEX  -->  []byte
    BLOB, TINYBLOB, MEDIUMBLOB, LONGBLOB  -->  []byte
                            DECIMAL, BIT  -->  []byte
                                    NULL  -->  nil

## Big packets

This package can send and receive MySQL data packets that are biger than 16 MB.
This means that you can receive response rows biger than 16 MB and can execute
prepared statements with parameter data biger than 16 MB without using
SEND_LONG_DATA command. If you want to use this feature you must set
*MySQL.MaxPktSize* to appropriate value before connect and change
*max_allowed_packet* value in MySQL server configuration.

## Thread safety

You can use this package in multithreading enviroment. All functions are thread
safe.

If one thread is calling *Query* or *Exec* method, other threads will be
blocked if they call *Query*, *Start*, *Exec*, *Run* or other method which send
data to the server, until *Query*/*Exec* return in first thread.

If one thread is calling *Start* or *Run* method, other threads will be
blocked if they call *Query*, *Start*, *Exec*, *Run* or other method which send
data to the server,  until all results and all rows  will be readed from
a connection in first thread.

## TODO

1. Complete GODOC documentation
2. stmt.BindResult
3. io.Writer as bind result variable

# Package documentation generated by godoc

It is converted to Markdown from GODOC.html. Unfortunately, after conversion
links doesn't works. 

