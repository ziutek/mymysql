This package contains MySQL client API written entirely in Go. It was created
due to lack of properly working MySQL API package, ready for my production
application (December 2010).

The code of this package is carefuly written and has internal error handling
using *panic()* exceptions, thus the probability of Go bugs or an unhandled
internal errors should be very small. Unfortunately I'm not a MySQL expert, 
so bugs in the protocol handling are possible.

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

In *GODOC.html* or *GODOC.txt* you can find the full documentation of this package in godoc format.

## Example 1

    import "mymy"

    db := mymy.New("tcp", "", "127.0.0.1:3306", user, pass, dbname)
    db.Debug = true

    err := db.Connect()
    if err != nil {
        panic(err)
    }

    rows, res, err := db.Query("select * from X")
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
        row, err := res.GetTextRow()
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
                os.Stdout.Write(col)
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
parameters and use *Execute* method to avoid these unnecessary rebinds. The
simplest way to bind parameters is:

    stmt.BindParams(data.id, data.tax)

but you can't use it in our example, becouse parameters binded this wah can't
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
        _, err = stmt.Execute()
        if err != nil {
            panic(err)
        }
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

If one thread is calling *Query* method, other threads will be blocked if they
call *Query*, *Start*, *Execute* or other method which send data to the server,
until *Query* return in first thread.

If one thread is calling *Start* or *Execute* method, other threads will be
blocked if they call *Query*, *Start*, *Execute* or other method which send
data to the server,  until all rows will be readed from a connection in first
thread.

## TODO

1. Complete GODOC documentation
2. stmt.SendLongData
3. stmt.BindResult
4. Multiple results
5. io.Reader as bind paremeter, io.Writer as bind result variable
`import "mymysql"`

#### Package files

[addons.go][1] [binding.go][2] [codecs.go][3] [command.go][4] [common.go][5]
[consts.go][6] [errors.go][7] [mysql.go][8] [packet.go][9] [prepared.go][10]
[result.go][11] [unsafe.go][12] [utils.go][13]

## Constants

MySQL protocol types.

mymysql uses only some of them for send data to the MySQL server. Used MySQL
types are marked with a comment contains mymysql type that uses it.


    const (

        MYSQL_TYPE_DECIMAL     = 0x00

        MYSQL_TYPE_TINY        = 0x01 // int8, uint8

        MYSQL_TYPE_SHORT       = 0x02 // int16, uint16

        MYSQL_TYPE_LONG        = 0x03 // int32, uint32

        MYSQL_TYPE_FLOAT       = 0x04 // float32

        MYSQL_TYPE_DOUBLE      = 0x05 // float64

        MYSQL_TYPE_NULL        = 0x06 // nil

        MYSQL_TYPE_TIMESTAMP   = 0x07 // *Timestamp

        MYSQL_TYPE_LONGLONG    = 0x08 // int64, uint64

        MYSQL_TYPE_INT24       = 0x09

        MYSQL_TYPE_DATE        = 0x0a

        MYSQL_TYPE_TIME        = 0x0b

        MYSQL_TYPE_DATETIME    = 0x0c // *Datetime

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

        MYSQL_TYPE_BLOB        = 0xfc // Blob

        MYSQL_TYPE_VAR_STRING  = 0xfd // []byte

        MYSQL_TYPE_STRING      = 0xfe // string

        MYSQL_TYPE_GEOMETRY    = 0xff


        MYSQL_UNSIGNED_MASK = uint16(1 << 15)

    )

Mapping of MySQL types to (prefered) protocol types. Use it if you create your
own Raw value.

Comments contains corresponding types used by mymysql. string type may be
replaced by []byte type and vice versa. []byte type is native for sending on a
network, so any string is converted to it before sending. Than for better
preformance use []byte.


    const (

        // Client send and receive, mymysql representation for send / receive

        TINYINT   = MYSQL_TYPE_TINY      // int8 / int8

        SMALLINT  = MYSQL_TYPE_SHORT     // int16 / int16

        INT       = MYSQL_TYPE_LONG      // int32 / int32

        BIGINT    = MYSQL_TYPE_LONGLONG  // int64 / int64

        FLOAT     = MYSQL_TYPE_FLOAT     // float32 / float32

        DOUBLE    = MYSQL_TYPE_DOUBLE    // float64 / float32

        TIME      = MYSQL_TYPE_TIME      // *Datetime / *Datetime

        DATE      = MYSQL_TYPE_DATE      // *Datetime / *Datetime

        DATETIME  = MYSQL_TYPE_DATETIME  // *Datetime / *Datetime

        TIMESTAMP = MYSQL_TYPE_TIMESTAMP // *Timestamp / *Datetime

        CHAR      = MYSQL_TYPE_STRING    // string / []byte

        BLOB      = MYSQL_TYPE_BLOB      // Blob / []byte

        NULL      = MYSQL_TYPE_NULL      // nil



        // Client send only, mymysql representation for send

        OUT_TEXT      = MYSQL_TYPE_STRING // string

        OUT_VARCHAR   = MYSQL_TYPE_STRING // string

        OUT_BINARY    = MYSQL_TYPE_BLOB   // Blob

        OUT_VARBINARY = MYSQL_TYPE_BLOB   // Blob



        // Client receive only, mymysql representation for receive

        IN_MEDIUMINT  = MYSQL_TYPE_INT24       // int32

        IN_YEAR       = MYSQL_TYPE_SHORT       // int16

        IN_BINARY     = MYSQL_TYPE_STRING      // []byte

        IN_VARCHAR    = MYSQL_TYPE_VAR_STRING  // []byte

        IN_VARBINARY  = MYSQL_TYPE_VAR_STRING  // []byte

        IN_TINYBLOB   = MYSQL_TYPE_TINY_BLOB   // []byte

        IN_TINYTEXT   = MYSQL_TYPE_TINY_BLOB   // []byte

        IN_TEXT       = MYSQL_TYPE_BLOB        // []byte

        IN_MEDIUMBLOB = MYSQL_TYPE_MEDIUM_BLOB // []byte

        IN_MEDIUMTEXT = MYSQL_TYPE_MEDIUM_BLOB // []byte

        IN_LONGBLOB   = MYSQL_TYPE_LONG_BLOB   // []byte

        IN_LONGTEXT   = MYSQL_TYPE_LONG_BLOB   // []byte



        // MySQL 5.x specific

        IN_DECIMAL = MYSQL_TYPE_NEWDECIMAL // TODO

        IN_BIT     = MYSQL_TYPE_BIT        // []byte

    )

## Variables


    var (

        WR_BUF_ERROR          = os.NewError("write buffer/packet too short")

        SEQ_ERROR             = os.NewError("packet sequence error")

        PKT_ERROR             = os.NewError("malformed packet")

        PKT_LONG_ERROR        = os.NewError("packet too long")

        UNEXP_NULL_LCS_ERROR  = os.NewError("unexpected null LCS")

        UNEXP_NULL_LCB_ERROR  = os.NewError("unexpected null LCB")

        UNEXP_NULL_DATE_ERROR = os.NewError("unexpected null datetime")

        UNK_RESULT_PKT_ERROR  = os.NewError("unexpected or unknown result
packet")

        NOT_CONN_ERROR        = os.NewError("not connected")

        ALREDY_CONN_ERROR     = os.NewError("not connected")

        BAD_RESULT_ERROR      = os.NewError("unexpected result")

        UNREADED_ROWS_ERROR   = os.NewError("there are unreaded rows")

        BIND_COUNT_ERROR      = os.NewError("wrong number of values for bind")

        BIND_UNK_TYPE         = os.NewError("unknown bind value type")

        RESULT_COUNT_ERROR    = os.NewError("wrong number of result columns")

        BAD_COMMAND_ERROR     = os.NewError("comand isn't text SQL nor
*Statement")

        WRONG_DATE_LEN_ERROR  = os.NewError("wrong datetime/timestamp length")

        UNK_MYSQL_TYPE_ERROR  = os.NewError("unknown MySQL type")

    )

## func [DecodeU16][14]

`func DecodeU16(buf []byte) uint16`

## func [DecodeU24][15]

`func DecodeU24(buf []byte) uint32`

## func [DecodeU32][16]

`func DecodeU32(buf []byte) uint32`

## func [DecodeU64][17]

`func DecodeU64(buf []byte) (rv uint64)`

## func [EncodeDatetime][18]

`func EncodeDatetime(dt *Datetime) *[]byte`

## func [EncodeU16][19]

`func EncodeU16(val uint16) *[]byte`

## func [EncodeU24][20]

`func EncodeU24(val uint32) *[]byte`

## func [EncodeU32][21]

`func EncodeU32(val uint32) *[]byte`

## func [EncodeU64][22]

`func EncodeU64(val uint64) *[]byte`

## func [IsDatetimeZero][23]

`func IsDatetimeZero(dt *Datetime) bool`

## func [NbinToNstr][24]

`func NbinToNstr(nbin *[]byte) *string`

## func [NstrToNbin][25]

`func NstrToNbin(nstr *string) *[]byte`

## type [Blob][26]


    type Blob []byte

## type [Datetime][27]


    type Datetime struct {

        Year                             int16

        Month, Day, Hour, Minute, Second uint8

        Nanosec                          uint32

    }

### func [TimeToDatetime][28]

`func TimeToDatetime(tt *time.Time) *Datetime`

### func (*Datetime) [String][29]

`func (dt *Datetime) String() string`

## type [Error][30]


    type Error struct {

        // contains unexported fields

    }

### func (Error) [String][31]

`func (err Error) String() string`

## type [Field][32]


    type Field struct {

        Catalog  string

        Db       string

        Table    string

        OrgTable string

        Name     string

        OrgName  string

        DispLen  uint32

        //  Charset  uint16

        Flags uint16

        Type  byte

        Scale byte

    }

## type [MySQL][33]

MySQL connection handler


    type MySQL struct {


        // Maximum packet size that client can accept from server.

        // Default 16*1024*1024-1. You may change it before connect.

        MaxPktSize int


        // Debug logging. You may change it at any time.

        Debug bool

        // contains unexported fields

    }

### func [New][34]

`func New(proto, laddr, raddr, user, passwd string, db ...string) (my *MySQL)`

Create new MySQL handler. The first three arguments are passed to net.Bind for
create connection. user and passwd are for authentication. Optional db is
database name (you may not specifi it and use Use() method later).

### func (*MySQL) [Close][35]

`func (my *MySQL) Close() (err os.Error)`

Close connection to the server

### func (*MySQL) [Connect][36]

`func (my *MySQL) Connect() (err os.Error)`

Establishes a connection with MySQL server version 4.1 or later.

### func (*MySQL) [Ping][37]

`func (my *MySQL) Ping() (err os.Error)`

Send PING packet to server.

### func (*MySQL) [Prepare][38]

`func (my *MySQL) Prepare(sql string) (stmt *Statement, err os.Error)`

Prepare server side statement. Return statement handler.

### func (*MySQL) [Query][39]

`func (my *MySQL) Query(command interface{}, params ...interface{}) (rows
[]*Row, res *Result, err os.Error)`

This call Start and next call GetTextRow once or more times. It read all rows
from connection and returns they as a slice.

### func (*MySQL) [Start][40]

`func (my *MySQL) Start(command interface{}, params ...interface{}) (res
*Result, err os.Error)`

Start new query session.

command can be SQL query (string) or a prepared statement (*Statement).

If the command is a string and you specify the parameters, the SQL string will
be a result of fmt.Sprintf(command, params...).

If the command is a prepared statement, params will be binded to this
statement before execution.

You must get all result rows (if they exists) before next query.

### func (*MySQL) [Use][41]

`func (my *MySQL) Use(dbname string) (err os.Error)`

Change database

## type [Raw][42]


    type Raw struct {

        // contains unexported fields

    }

## type [Result][43]


    type Result struct {

        FieldCount int

        Fields     []*Field       // Fields table

        Map        map[string]int // Maps field name to column number


        Message      []byte

        AffectedRows uint64

        InsertId     uint64

        WarningCount int

        Status       uint16

        // contains unexported fields

    }

### func (*Result) [End][44]

`func (res *Result) End() (err os.Error)`

Read all unreaded rows and discard them. All rows must be read before next
query or other command.

### func (*Result) [GetRow][45]

`func (res *Result) GetRow() (row *Row, err os.Error)`

Get data row from a server. This method reads one row of result directly from
network connection (without rows buffering on client side).

## type [Row][46]

Result row. Data field is a slice that contains values for any column of
received row.

If row is a result of ordinary text query, an element of Data field can be
[]byte slice, contained result text or nil if NULL is returned.

If it is result of prepared statement execution, an element of Data field can
be: intXX, uintXX, floatXX, []byte, *Datetime or nil.


    type Row struct {

        Data []interface{}

    }

### func (*Row) [Bin][47]

`func (tr *Row) Bin(nn int) (bin []byte)`

Get the nn-th value and return it as []byte ([]byte{} if NULL)

### func (*Row) [Int][48]

`func (tr *Row) Int(nn int) (val int)`

Get the nn-th value and return it as int. Return 0 if value is NULL or
conversion is impossible.

### func (*Row) [IntErr][49]

`func (tr *Row) IntErr(nn int) (val int, err os.Error)`

Get the nn-th value and return it as int (0 if NULL). Return error if
conversion is impossible.

### func (*Row) [MustInt][50]

`func (tr *Row) MustInt(nn int) (val int)`

Get the nn-th value and return it as int (0 if NULL). Panic if conversion is
impossible.

### func (*Row) [MustUint][51]

`func (tr *Row) MustUint(nn int) (val uint)`

Get the nn-th value and return it as uint (0 if NULL). Panic if conversion is
impossible.

### func (*Row) [Str][52]

`func (tr *Row) Str(nn int) (str string)`

Get the nn-th value and return it as string ("" if NULL)

### func (*Row) [Uint][53]

`func (tr *Row) Uint(nn int) (val uint)`

Get the nn-th value and return it as uint. Return 0 if value is NULL or
conversion is impossible.

### func (*Row) [UintErr][54]

`func (tr *Row) UintErr(nn int) (val uint, err os.Error)`

Get the nn-th value and return it as uint (0 if NULL). Return error if
conversion is impossible.

## type [ServerInfo][55]


    type ServerInfo struct {

        // contains unexported fields

    }

## type [Statement][56]


    type Statement struct {

        Fields []*Field

        Map    map[string]int // Maps field name to column number


        FieldCount   int

        ParamCount   int

        WarningCount int

        Status       uint16

        // contains unexported fields

    }

### func (*Statement) [BindParams][57]

`func (stmt *Statement) BindParams(params ...interface{})`

Bind input data for the parameter markers in the SQL statement that was passed
to Prepare.

params may be a parameter list (slice), a struct or a pointer to the struct. A
struct field can by value or pointer to value. A parameter (slice element) can
be value, pointer to value or pointer to pointer of value. Values may be of
the folowind types: intXX, uintXX, floatXX, []byte, Blob, string, Datetime,
Timestamp, Raw.

### func (*Statement) [Delete][58]

`func (stmt *Statement) Delete() (err os.Error)`

Destroy statement on server side. Client side handler is invalid after this
command.

### func (*Statement) [Execute][59]

`func (stmt *Statement) Execute() (res *Result, err os.Error)`

### func (*Statement) [Reset][60]

`func (stmt *Statement) Reset() (err os.Error)`

Resets a prepared statement on server: data sent to the server, unbuffered
result sets and current errors.

## type [Timestamp][61]


    type Timestamp Datetime

### func [TimeToTimestamp][62]

`func TimeToTimestamp(tt *time.Time) *Timestamp`

### func (*Timestamp) [String][63]

`func (ts *Timestamp) String() string`

## type [Value][64]


    type Value struct {

        // contains unexported fields

    }

### func (*Value) [Len][65]

`func (val *Value) Len() int`

## Subdirectories

Name


Synopsis

[..][66]

[.git][67]

[examples][68]

[godoc][69]

   [1]: /mymysql/addons.go

   [2]: /mymysql/binding.go

   [3]: /mymysql/codecs.go

   [4]: /mymysql/command.go

   [5]: /mymysql/common.go

   [6]: /mymysql/consts.go

   [7]: /mymysql/errors.go

   [8]: /mymysql/mysql.go

   [9]: /mymysql/packet.go

   [10]: /mymysql/prepared.go

   [11]: /mymysql/result.go

   [12]: /mymysql/unsafe.go

   [13]: /mymysql/utils.go

   [14]: /mymysql/codecs.go#L9

   [15]: /mymysql/codecs.go#L18

   [16]: /mymysql/codecs.go#L27

   [17]: /mymysql/codecs.go#L37

   [18]: /mymysql/codecs.go#L328

   [19]: /mymysql/codecs.go#L49

   [20]: /mymysql/codecs.go#L56

   [21]: /mymysql/codecs.go#L63

   [22]: /mymysql/codecs.go#L70

   [23]: /mymysql/common.go#L80

   [24]: /mymysql/addons.go#L3

   [25]: /mymysql/addons.go#L11

   [26]: /mymysql/binding.go#L46

   [27]: /mymysql/binding.go#L16

   [28]: /mymysql/common.go#L85

   [29]: /mymysql/binding.go#L21

   [30]: /mymysql/errors.go#L29

   [31]: /mymysql/errors.go#L34

   [32]: /mymysql/result.go#L12

   [33]: /mymysql/mysql.go#L22

   [34]: /mymysql/mysql.go#L53

   [35]: /mymysql/mysql.go#L97

   [36]: /mymysql/mysql.go#L72

   [37]: /mymysql/mysql.go#L243

   [38]: /mymysql/mysql.go#L263

   [39]: /mymysql/mysql.go#L223

   [40]: /mymysql/mysql.go#L149

   [41]: /mymysql/mysql.go#L117

   [42]: /mymysql/binding.go#L48

   [43]: /mymysql/result.go#L26

   [44]: /mymysql/mysql.go#L214

   [45]: /mymysql/mysql.go#L187

   [46]: /mymysql/result.go#L49

   [47]: /mymysql/result.go#L54

   [48]: /mymysql/result.go#L135

   [49]: /mymysql/result.go#L87

   [50]: /mymysql/result.go#L125

   [51]: /mymysql/result.go#L168

   [52]: /mymysql/result.go#L71

   [53]: /mymysql/result.go#L178

   [54]: /mymysql/result.go#L142

   [55]: /mymysql/mysql.go#L12

   [56]: /mymysql/prepared.go#L7

   [57]: /mymysql/mysql.go#L301

   [58]: /mymysql/mysql.go#L360

   [59]: /mymysql/mysql.go#L339

   [60]: /mymysql/mysql.go#L380

   [61]: /mymysql/binding.go#L41

   [62]: /mymysql/common.go#L96

   [63]: /mymysql/binding.go#L42

   [64]: /mymysql/binding.go#L8

   [65]: /mymysql/unsafe.go#L8

   [66]: ..

   [67]: .git

   [68]: examples

   [69]: godoc

