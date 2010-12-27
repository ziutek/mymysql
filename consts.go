package mymy

// Borrowed from GoMySQL

const (
    CLIENT_LONG_PASSWORD = 1 << iota
    CLIENT_FOUND_ROWS
    CLIENT_LONG_FLAG
    CLIENT_CONNECT_WITH_DB
    CLIENT_NO_SCHEMA
    CLIENT_COMPRESS
    CLIENT_ODBC
    CLIENT_LOCAL_FILES
    CLIENT_IGNORE_SPACE
    CLIENT_PROTOCOL_41
    CLIENT_INTERACTIVE
    CLIENT_SSL
    CLIENT_IGNORE_SIGPIPE
    CLIENT_TRANSACTIONS
    CLIENT_RESERVED
    CLIENT_SECURE_CONN
    CLIENT_MULTI_STATEMENTS
    CLIENT_MULTI_RESULTS
)

const (
    COM_QUIT                = 0x01
    COM_INIT_DB             = 0x02
    COM_QUERY               = 0x03
    COM_FIELD_LIST          = 0x04
    COM_CREATE_DB           = 0x05
    COM_DROP_DB             = 0x06
    COM_REFRESH             = 0x07
    COM_SHUTDOWN            = 0x08
    COM_STATISTICS          = 0x09
    COM_PROCESS_INFO        = 0x0a
    COM_CONNECT             = 0x0b
    COM_PROCESS_KILL        = 0x0c
    COM_DEBUG               = 0x0d
    COM_PING                = 0x0e
    COM_TIME                = 0x0f
    COM_DELAYED_INSERT      = 0x10
    COM_CHANGE_USER         = 0x11
    COM_BINLOG_DUMP         = 0x12
    COM_TABLE_DUMP          = 0x13
    COM_CONNECT_OUT         = 0x14
    COM_REGISTER_SLAVE      = 0x15
    COM_STMT_PREPARE        = 0x16
    COM_STMT_EXECUTE        = 0x17
    COM_STMT_SEND_LONG_DATA = 0x18
    COM_STMT_CLOSE          = 0x19
    COM_STMT_RESET          = 0x1a
    COM_SET_OPTION          = 0x1b
    COM_STMT_FETCH          = 0x1c
)

const (
    FIELD_TYPE_DECIMAL     = 0x00
    FIELD_TYPE_TINY        = 0x01
    FIELD_TYPE_SHORT       = 0x02
    FIELD_TYPE_LONG        = 0x03
    FIELD_TYPE_FLOAT       = 0x04
    FIELD_TYPE_DOUBLE      = 0x05
    FIELD_TYPE_NULL        = 0x06
    FIELD_TYPE_TIMESTAMP   = 0x07
    FIELD_TYPE_LONGLONG    = 0x08
    FIELD_TYPE_INT24       = 0x09
    FIELD_TYPE_DATE        = 0x0a
    FIELD_TYPE_TIME        = 0x0b
    FIELD_TYPE_DATETIME    = 0x0c
    FIELD_TYPE_YEAR        = 0x0d
    FIELD_TYPE_NEWDATE     = 0x0e
    FIELD_TYPE_VARCHAR     = 0x0f
    FIELD_TYPE_BIT         = 0x10
    FIELD_TYPE_NEWDECIMAL  = 0xf6
    FIELD_TYPE_ENUM        = 0xf7
    FIELD_TYPE_SET         = 0xf8
    FIELD_TYPE_TINY_BLOB   = 0xf9
    FIELD_TYPE_MEDIUM_BLOB = 0xfa
    FIELD_TYPE_LONG_BLOB   = 0xfb
    FIELD_TYPE_BLOB        = 0xfc
    FIELD_TYPE_VAR_STRING  = 0xfd
    FIELD_TYPE_STRING      = 0xfe
    FIELD_TYPE_GEOMETRY    = 0xff
)

const (
    FLAG_NOT_NULL = 1 << iota
    FLAG_PRI_KEY
    FLAG_UNIQUE_KEY
    FLAG_MULTIPLE_KEY
    FLAG_BLOB
    FLAG_UNSIGNED
    FLAG_ZEROFILL
    FLAG_BINARY
    FLAG_ENUM
    FLAG_AUTO_INCREMENT
    FLAG_TIMESTAMP
    FLAG_SET
)
