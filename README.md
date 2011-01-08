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

## Example 3 - using SendLongData in conjunction with http.Get

    _, err = db.Start("CREATE TABLE web (url VARCHAR(80), content LONGBLOB)")
    checkError(err)

    ins, err := db.Prepare("INSERT INTO web VALUES (?, ?)")
    checkError(err)

    var url string

    ins.BindParams(&url, nil)

    for  {
        // Get URL from stdin
        url = ""
        fmt.Scanln(&url)
        if len(url) == 0 {
            break
        }

        resp, _, err := http.Get(url)
        checkError(err)

        // Retrieve response directly into database (resp.Body is io.Reader)
        err = ins.SendLongData(1, resp.Body, 4092)
        checkError(err)

        _, err = ins.Execute()
        checkError(err)
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
2. Multiple results
3. stmt.BindResult
4. io.Writer as bind result variable

# Package documentation generated by godoc

It is converted to Markdown from GODOC.html. Unfortunately, after conversion
links doesn't works. 

# Package mymy

[func DecodeU16][1]

[func DecodeU24][2]

[func DecodeU32][3]

[func DecodeU64][4]

[func EncodeDatetime][5]

[func EncodeU16][6]

[func EncodeU24][7]

[func EncodeU32][8]

[func EncodeU64][9]

[func IsDatetimeZero][10]

[func NbinToNstr][11]

[func NstrToNbin][12]

[type Blob][13]

[type Datetime][14]

> [func TimeToDatetime][15]

> [func (*Datetime) String][16]

[type Error][17]

> [func (Error) String][18]

[type Field][19]

[type MySQL][20]

> [func New][21]

> [func (*MySQL) Close][22]

> [func (*MySQL) Connect][23]

> [func (*MySQL) Ping][24]

> [func (*MySQL) Prepare][25]

> [func (*MySQL) Query][26]

> [func (*MySQL) Reconnect][27]

> [func (*MySQL) Start][28]

> [func (*MySQL) Use][29]

[type Raw][30]

[type Result][31]

> [func (*Result) End][32]

> [func (*Result) GetRow][33]

[type Row][34]

> [func (*Row) Bin][35]

> [func (*Row) Int][36]

> [func (*Row) IntErr][37]

> [func (*Row) MustInt][38]

> [func (*Row) MustUint][39]

> [func (*Row) Str][40]

> [func (*Row) Uint][41]

> [func (*Row) UintErr][42]

[type ServerInfo][43]

[type Statement][44]

> [func (*Statement) BindParams][45]

> [func (*Statement) Delete][46]

> [func (*Statement) Execute][47]

> [func (*Statement) Reset][48]

> [func (*Statement) SendLongData][49]

[type Timestamp][50]

> [func TimeToTimestamp][51]

> [func (*Timestamp) String][52]

`import "mymysql"`

#### Package files

[addons.go][53] [binding.go][54] [codecs.go][55] [command.go][56]
[common.go][57] [consts.go][58] [errors.go][59] [mysql.go][60] [packet.go][61]
[prepared.go][62] [result.go][63] [unsafe.go][64] [utils.go][65]

## Constants

MySQL error codes


    const (

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

    )

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

        BIND_UNK_TYPE         = os.NewError("unknown value type for bind")

        RESULT_COUNT_ERROR    = os.NewError("wrong number of result columns")

        BAD_COMMAND_ERROR     = os.NewError("comand isn't text SQL nor
*Statement")

        WRONG_DATE_LEN_ERROR  = os.NewError("wrong datetime/timestamp length")

        UNK_MYSQL_TYPE_ERROR  = os.NewError("unknown MySQL type")

        WRONG_PARAM_NUM_ERROR = os.NewError("wrong parameter number")

        UNK_DATA_TYPE_ERROR   = os.NewError("unknown data source type")

        SMALL_PKT_SIZE_ERROR  = os.NewError("specified packet size is to
small")

    )

## func [DecodeU16][66]

`func DecodeU16(buf []byte) uint16`

## func [DecodeU24][67]

`func DecodeU24(buf []byte) uint32`

## func [DecodeU32][68]

`func DecodeU32(buf []byte) uint32`

## func [DecodeU64][69]

`func DecodeU64(buf []byte) (rv uint64)`

## func [EncodeDatetime][70]

`func EncodeDatetime(dt *Datetime) *[]byte`

## func [EncodeU16][71]

`func EncodeU16(val uint16) *[]byte`

## func [EncodeU24][72]

`func EncodeU24(val uint32) *[]byte`

## func [EncodeU32][73]

`func EncodeU32(val uint32) *[]byte`

## func [EncodeU64][74]

`func EncodeU64(val uint64) *[]byte`

## func [IsDatetimeZero][75]

`func IsDatetimeZero(dt *Datetime) bool`

## func [NbinToNstr][76]

`func NbinToNstr(nbin *[]byte) *string`

## func [NstrToNbin][77]

`func NstrToNbin(nstr *string) *[]byte`

## type [Blob][78]


    type Blob []byte

## type [Datetime][79]


    type Datetime struct {

        Year                             int16

        Month, Day, Hour, Minute, Second uint8

        Nanosec                          uint32

    }

### func [TimeToDatetime][80]

`func TimeToDatetime(tt *time.Time) *Datetime`

### func (*Datetime) [String][81]

`func (dt *Datetime) String() string`

## type [Error][82]

If function/method returns error you can check returned error type, and if it
is *mymy.Error it is error received from MySQL server. Next you can check Code
for error number.


    type Error struct {

        Code uint16

        Msg  []byte

    }

### func (Error) [String][83]

`func (err Error) String() string`

## type [Field][84]


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

## type [MySQL][85]

MySQL connection handler


    type MySQL struct {


        // Maximum packet size that client can accept from server.

        // Default 16*1024*1024-1. You may change it before connect.

        MaxPktSize int


        // Debug logging. You may change it at any time.

        Debug bool

        // contains unexported fields

    }

### func [New][86]

`func New(proto, laddr, raddr, user, passwd string, db ...string) (my *MySQL)`

Create new MySQL handler. The first three arguments are passed to net.Bind for
create connection. user and passwd are for authentication. Optional db is
database name (you may not specifi it and use Use() method later).

### func (*MySQL) [Close][87]

`func (my *MySQL) Close() (err os.Error)`

Close connection to the server

### func (*MySQL) [Connect][88]

`func (my *MySQL) Connect() (err os.Error)`

Establishes a connection with MySQL server version 4.1 or later.

### func (*MySQL) [Ping][89]

`func (my *MySQL) Ping() (err os.Error)`

Send MySQL PING to the server.

### func (*MySQL) [Prepare][90]

`func (my *MySQL) Prepare(sql string) (stmt *Statement, err os.Error)`

Prepare server side statement. Return statement handler.

### func (*MySQL) [Query][91]

`func (my *MySQL) Query(command interface{}, params ...interface{}) (rows
[]*Row, res *Result, err os.Error)`

This call Start and next call GetRow once or more times. It read all rows from
connection and returns they as a slice.

### func (*MySQL) [Reconnect][92]

`func (my *MySQL) Reconnect() (err os.Error)`

Close and reopen connection in one thread safe operation. Ignore unreaded
rows, reprepare all prepared statements.

### func (*MySQL) [Start][93]

`func (my *MySQL) Start(command interface{}, params ...interface{}) (res
*Result, err os.Error)`

Start new query session.

command can be SQL query (string) or a prepared statement (*Statement).

If the command is a string and you specify the parameters, the SQL string will
be a result of fmt.Sprintf(command, params...).

If the command is a prepared statement, params will be binded to this
statement before execution.

You must get all result rows (if they exists) before next query.

### func (*MySQL) [Use][94]

`func (my *MySQL) Use(dbname string) (err os.Error)`

Change database

## type [Raw][95]


    type Raw struct {

        Typ uint16

        Val *[]byte

    }

## type [Result][96]


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

### func (*Result) [End][97]

`func (res *Result) End() (err os.Error)`

Read all unreaded rows and discard them. All rows must be read before next
query or other command.

### func (*Result) [GetRow][98]

`func (res *Result) GetRow() (row *Row, err os.Error)`

Get data row from a server. This method reads one row of result directly from
network connection (without rows buffering on client side).

## type [Row][99]

Result row. Data field is a slice that contains values for any column of
received row.

If row is a result of ordinary text query, an element of Data field can be
[]byte slice, contained result text or nil if NULL is returned.

If it is result of prepared statement execution, an element of Data field can
be: intXX, uintXX, floatXX, []byte, *Datetime or nil.


    type Row struct {

        Data []interface{}

    }

### func (*Row) [Bin][100]

`func (tr *Row) Bin(nn int) (bin []byte)`

Get the nn-th value and return it as []byte ([]byte{} if NULL)

### func (*Row) [Int][101]

`func (tr *Row) Int(nn int) (val int)`

Get the nn-th value and return it as int. Return 0 if value is NULL or
conversion is impossible.

### func (*Row) [IntErr][102]

`func (tr *Row) IntErr(nn int) (val int, err os.Error)`

Get the nn-th value and return it as int (0 if NULL). Return error if
conversion is impossible.

### func (*Row) [MustInt][103]

`func (tr *Row) MustInt(nn int) (val int)`

Get the nn-th value and return it as int (0 if NULL). Panic if conversion is
impossible.

### func (*Row) [MustUint][104]

`func (tr *Row) MustUint(nn int) (val uint)`

Get the nn-th value and return it as uint (0 if NULL). Panic if conversion is
impossible.

### func (*Row) [Str][105]

`func (tr *Row) Str(nn int) (str string)`

Get the nn-th value and return it as string ("" if NULL)

### func (*Row) [Uint][106]

`func (tr *Row) Uint(nn int) (val uint)`

Get the nn-th value and return it as uint. Return 0 if value is NULL or
conversion is impossible.

### func (*Row) [UintErr][107]

`func (tr *Row) UintErr(nn int) (val uint, err os.Error)`

Get the nn-th value and return it as uint (0 if NULL). Return error if
conversion is impossible.

## type [ServerInfo][108]


    type ServerInfo struct {

        // contains unexported fields

    }

## type [Statement][109]


    type Statement struct {

        Fields []*Field

        Map    map[string]int // Maps field name to column number


        FieldCount   int

        ParamCount   int

        WarningCount int

        Status       uint16

        // contains unexported fields

    }

### func (*Statement) [BindParams][110]

`func (stmt *Statement) BindParams(params ...interface{})`

Bind input data for the parameter markers in the SQL statement that was passed
to Prepare.

params may be a parameter list (slice), a struct or a pointer to the struct. A
struct field can by value or pointer to value. A parameter (slice element) can
be value, pointer to value or pointer to pointer to value. Values may be of
the folowind types: intXX, uintXX, floatXX, []byte, Blob, string, Datetime,
Timestamp, Raw.

### func (*Statement) [Delete][111]

`func (stmt *Statement) Delete() (err os.Error)`

Destroy statement on server side. Client side handler is invalid after this
command.

### func (*Statement) [Execute][112]

`func (stmt *Statement) Execute() (res *Result, err os.Error)`

Execute prepared statement. If statement requires parameters you must bind
them first.

### func (*Statement) [Reset][113]

`func (stmt *Statement) Reset() (err os.Error)`

Resets a prepared statement on server: data sent to the server, unbuffered
result sets and current errors.

### func (*Statement) [SendLongData][114]

`func (stmt *Statement) SendLongData(pnum int, data interface{}, pkt_size int)
(err os.Error)`

Send long data to MySQL server in chunks. You can call this method after Bind
and before Execute. It can be called multiple times for one parameter to send
TEXT or BLOB data in chunks.

pnum - Parameter number to associate the data with.

data - Data source string, []byte or io.Reader.

pkt_size - It must be must be greater than 6 and less or equal to MySQL
max_allowed_packet variable. You can obtain value of this variable using such
query: SHOW variables WHERE Variable_name = 'max_allowed_packet' If data
source is io.Reader then (pkt_size - 6) is size of a buffer that will be
allocated for reading.

If you have data source of type string or []byte in one piece you may properly
set pkt_size and call this method once. If you have data in multiple pieces
you can call this method multiple times. If data source is io.Reader you
should properly set pkt_size. Data will be readed from io.Reader and send in
pieces to the server until EOF.

## type [Timestamp][115]


    type Timestamp Datetime

### func [TimeToTimestamp][116]

`func TimeToTimestamp(tt *time.Time) *Timestamp`

### func (*Timestamp) [String][117]

`func (ts *Timestamp) String() string`

## Subdirectories

Name


Synopsis

[..][118]

[.git][119]

[doc][120]

[examples][121]

   [1]: #DecodeU16

   [2]: #DecodeU24

   [3]: #DecodeU32

   [4]: #DecodeU64

   [5]: #EncodeDatetime

   [6]: #EncodeU16

   [7]: #EncodeU24

   [8]: #EncodeU32

   [9]: #EncodeU64

   [10]: #IsDatetimeZero

   [11]: #NbinToNstr

   [12]: #NstrToNbin

   [13]: #Blob

   [14]: #Datetime

   [15]: #Datetime.TimeToDatetime

   [16]: #Datetime.String

   [17]: #Error

   [18]: #Error.String

   [19]: #Field

   [20]: #MySQL

   [21]: #MySQL.New

   [22]: #MySQL.Close

   [23]: #MySQL.Connect

   [24]: #MySQL.Ping

   [25]: #MySQL.Prepare

   [26]: #MySQL.Query

   [27]: #MySQL.Reconnect

   [28]: #MySQL.Start

   [29]: #MySQL.Use

   [30]: #Raw

   [31]: #Result

   [32]: #Result.End

   [33]: #Result.GetRow

   [34]: #Row

   [35]: #Row.Bin

   [36]: #Row.Int

   [37]: #Row.IntErr

   [38]: #Row.MustInt

   [39]: #Row.MustUint

   [40]: #Row.Str

   [41]: #Row.Uint

   [42]: #Row.UintErr

   [43]: #ServerInfo

   [44]: #Statement

   [45]: #Statement.BindParams

   [46]: #Statement.Delete

   [47]: #Statement.Execute

   [48]: #Statement.Reset

   [49]: #Statement.SendLongData

   [50]: #Timestamp

   [51]: #Timestamp.TimeToTimestamp

   [52]: #Timestamp.String

   [53]: /mymysql/addons.go

   [54]: /mymysql/binding.go

   [55]: /mymysql/codecs.go

   [56]: /mymysql/command.go

   [57]: /mymysql/common.go

   [58]: /mymysql/consts.go

   [59]: /mymysql/errors.go

   [60]: /mymysql/mysql.go

   [61]: /mymysql/packet.go

   [62]: /mymysql/prepared.go

   [63]: /mymysql/result.go

   [64]: /mymysql/unsafe.go

   [65]: /mymysql/utils.go

   [66]: /mymysql/codecs.go#L9

   [67]: /mymysql/codecs.go#L18

   [68]: /mymysql/codecs.go#L27

   [69]: /mymysql/codecs.go#L37

   [70]: /mymysql/codecs.go#L324

   [71]: /mymysql/codecs.go#L49

   [72]: /mymysql/codecs.go#L56

   [73]: /mymysql/codecs.go#L63

   [74]: /mymysql/codecs.go#L70

   [75]: /mymysql/common.go#L86

   [76]: /mymysql/addons.go#L3

   [77]: /mymysql/addons.go#L11

   [78]: /mymysql/binding.go#L46

   [79]: /mymysql/binding.go#L16

   [80]: /mymysql/common.go#L91

   [81]: /mymysql/binding.go#L21

   [82]: /mymysql/errors.go#L34

   [83]: /mymysql/errors.go#L39

   [84]: /mymysql/result.go#L12

   [85]: /mymysql/mysql.go#L23

   [86]: /mymysql/mysql.go#L54

   [87]: /mymysql/mysql.go#L117

   [88]: /mymysql/mysql.go#L94

   [89]: /mymysql/mysql.go#L293

   [90]: /mymysql/mysql.go#L334

   [91]: /mymysql/mysql.go#L273

   [92]: /mymysql/mysql.go#L132

   [93]: /mymysql/mysql.go#L199

   [94]: /mymysql/mysql.go#L167

   [95]: /mymysql/binding.go#L48

   [96]: /mymysql/result.go#L26

   [97]: /mymysql/mysql.go#L264

   [98]: /mymysql/mysql.go#L237

   [99]: /mymysql/result.go#L49

   [100]: /mymysql/result.go#L54

   [101]: /mymysql/result.go#L135

   [102]: /mymysql/result.go#L87

   [103]: /mymysql/result.go#L125

   [104]: /mymysql/result.go#L168

   [105]: /mymysql/result.go#L71

   [106]: /mymysql/result.go#L178

   [107]: /mymysql/result.go#L142

   [108]: /mymysql/mysql.go#L13

   [109]: /mymysql/prepared.go#L7

   [110]: /mymysql/mysql.go#L365

   [111]: /mymysql/mysql.go#L426

   [112]: /mymysql/mysql.go#L405

   [113]: /mymysql/mysql.go#L448

   [114]: /mymysql/mysql.go#L487

   [115]: /mymysql/binding.go#L41

   [116]: /mymysql/common.go#L102

   [117]: /mymysql/binding.go#L42

   [118]: ..

   [119]: .git

   [120]: doc

   [121]: examples

