package mymy

import (
    "os"
    "net"
    "bufio"
    "sync"
    "fmt"
    "reflect"
)

type ServerInfo struct {
    prot_ver byte
    serv_ver string
    thr_id   uint32
    scramble []byte
    caps     uint16
    lang     byte
}

// MySQL connection handler
type MySQL struct {
    proto string // Network protocol
    laddr string // Local address
    raddr string // Remote (server) address

    user   string // MySQL username
    passwd string // MySQL password
    dbname string // Database name

    conn net.Conn      // MySQL connection
    rd   *bufio.Reader
    wr   *bufio.Writer

    info ServerInfo // MySQL server information
    seq  byte       // MySQL sequence number

    mutex           *sync.Mutex // For concurency
    unreaded_rows   bool
    reconnect_count int         // Reconnect counter

    // Maximum packet size that client can accept from server.
    // Default 16*1024*1024-1. You may change it before connect.
    MaxPktSize int

    // Debug logging. You may change it at any time.
    Debug bool
}

// Create new MySQL handler. The first three arguments are passed to net.Bind
// for create connection. user and passwd are for authentication. Optional db
// is database name (you may not specifi it and use Use() method later).
func New(proto, laddr, raddr, user, passwd string, db ...string) (my *MySQL) {
    my = &MySQL{
        proto:      proto,
        laddr:      laddr,
        raddr:      raddr,
        user:       user,
        passwd:     passwd,
        mutex:      new(sync.Mutex),
        MaxPktSize: 16*1024*1024-1,
    }
    if len(db) == 1 {
        my.dbname = db[0]
    } else if len(db) > 1 {
        panic("mymy.New: too many arguments")
    }
    return
}

// Establishes a connection with MySQL server version 4.1 or later.
func (my *MySQL) Connect() (err os.Error) {
    if my.conn != nil {
        return ALREDY_CONN_ERROR
    }
    defer my.unlock()
    defer catchOsError(&err)
    my.lock()

    // Make connection
    my.conn, err = net.Dial(my.proto, my.laddr, my.raddr)
    if err != nil {
        return
    }
    my.rd = bufio.NewReader(my.conn)
    my.wr = bufio.NewWriter(my.conn)

    // Initialisation
    my.init()
    my.auth()
    my.getResult(nil)

    return
}

// Close connection to the server
func (my *MySQL) Close() (err os.Error) {
    if my.conn == nil {
        return NOT_CONN_ERROR
    }
    if my.unreaded_rows {
        return UNREADED_ROWS_ERROR
    }
    defer my.unlock()
    defer catchOsError(&err)
    my.lock()

    // Close the connection
    my.sendCmd(_COM_QUIT)
    err = my.conn.Close()
    my.conn = nil // Mark that we disconnect

    return
}

// Change database
func (my *MySQL) Use(dbname string) (err os.Error) {
    if my.conn == nil {
        return NOT_CONN_ERROR
    }
    if my.unreaded_rows {
        return UNREADED_ROWS_ERROR
    }
    defer my.unlock()
    defer catchOsError(&err)
    my.lock()

    // Send command
    my.sendCmd(_COM_INIT_DB, dbname)
    // Get server response
    my.getResult(nil)
    // Save new database name
    my.dbname = dbname

    return
}

// Start new query session.
//
// command can be SQL query (string) or a prepared statement (*Statement).
//
// If the command is a string and you specify the parameters, the SQL string
// will be a result of fmt.Sprintf(command, params...).
//
// If the command is a prepared statement, params will be binded to this
// statement before execution.
//
// You must get all result rows (if they exists) before next query.
func (my *MySQL) Start(command interface{}, params ...interface{}) (
        res *Result, err os.Error) {

    // Check type of command
    switch cmd := command.(type) {
    case *Statement:
        // Prepared statement
        cmd.BindParams(params...)
        return cmd.Execute()

    case string:
        // Text SQL
        if my.conn == nil {
            return nil, NOT_CONN_ERROR
        }
        if my.unreaded_rows {
            return nil, UNREADED_ROWS_ERROR
        }
        defer my.unlockIfError(&err)
        defer catchOsError(&err)
        my.lock()

        // Send query
        if len(params) == 0 {
            my.sendCmd(_COM_QUERY, cmd)
        } else {
            my.sendCmd(_COM_QUERY, fmt.Sprintf(cmd, params...))
        }

        // Get command response
        res = my.getResponse()
        return
    }
    return nil, BAD_COMMAND_ERROR
}

// Get data row from a server. This method reads one row of result directly
// from network connection (without rows buffering on client side).
func (res *Result) GetRow() (row *Row, err os.Error) {
    if res.FieldCount == 0 {
        // There is no fields in result (OK result)
        return
    }
    defer res.db.unlockIfError(&err)
    defer catchOsError(&err)

    switch result := res.db.getResult(res).(type) {
    case *Row:
        // Row of data
        row = result

    case *Result:
        // EOF result
        res.db.unreaded_rows = false
        res.db.unlock()

    default:
        err = BAD_RESULT_ERROR
    }
    return
    // TODO: Check (res.Status & SERVER_MORE_RESULTS_EXISTS) for more results
}

// Read all unreaded rows and discard them. All rows must be read before next
// query or other command.
func (res *Result) End() (err os.Error) {
    for err == nil && res.db.unreaded_rows {
        _, err = res.GetRow()
    }
    return
}

// This call Start and next call GetTextRow once or more times. It read
// all rows from connection and returns they as a slice.
func (my *MySQL) Query(command interface{}, params ...interface{}) (
        rows []*Row, res *Result, err os.Error) {

    res, err = my.Start(command, params...)
    if err != nil {
        return
    }
    // Read rows
    var row *Row
    for {
        row, err = res.GetRow()
        if err != nil || row == nil {
            break
        }
        rows = append(rows, row)
    }
    return
}

// Send PING packet to server.
func (my *MySQL) Ping() (err os.Error) {
    if my.conn == nil {
        return NOT_CONN_ERROR
    }
    if my.unreaded_rows {
        return UNREADED_ROWS_ERROR
    }
    defer my.unlock()
    defer catchOsError(&err)
    my.lock()

    // Send command
    my.sendCmd(_COM_PING)
    // Get server response
    my.getResult(nil)

    return
}

// Prepare server side statement. Return statement handler.
func (my *MySQL) Prepare(sql string) (stmt *Statement, err os.Error) {
    if my.conn == nil {
        return nil, NOT_CONN_ERROR
    }
    if my.unreaded_rows {
        return nil, UNREADED_ROWS_ERROR
    }
    defer my.unlock()
    defer catchOsError(&err)
    my.lock()

    // Send command
    my.sendCmd(_COM_STMT_PREPARE, sql)
    // Get server response
    stmt, ok := my.getPrepareResult(nil).(*Statement)
    if !ok {
        return nil, BAD_RESULT_ERROR
    }
    if len(stmt.params) > 0 {
        // Get param fields
        my.getPrepareResult(stmt)
    }
    if len(stmt.Fields) > 0 {
        // Get column fields
        my.getPrepareResult(stmt)
    }
    stmt.db = my
    return
}

// Bind input data for the parameter markers in the SQL statement that was
// passed to Prepare.
// 
// params may be a parameter list (slice), a struct or a pointer to the struct.
// A struct field can by value or pointer to value. A parameter (slice element)
// can be value, pointer to value or pointer to pointer to value.
// Values may be of the folowind types: intXX, uintXX, floatXX, []byte, Blob,
// string, Datetime, Timestamp, Raw.
func (stmt *Statement) BindParams(params ...interface{}) {
    stmt.rebind = true

    // Check for struct
    if len(params) == 1 {
        pval := reflect.NewValue(params[0])
        // Dereference pointer
        if vv, ok := pval.(*reflect.PtrValue); ok {
            pval = vv.Elem()
        }
        val, ok := pval.(*reflect.StructValue)
        if ok && val.Type() != reflectDatetimeType &&
                val.Type() != reflectTimestampType {
            // We have struct to bind
            if val.NumField() != stmt.ParamCount {
                panic(BIND_COUNT_ERROR)
            }
            for ii := 0; ii < stmt.ParamCount; ii ++ {
                stmt.params[ii] = bindValue(val.Field(ii))
            }
            return
        }

    }

    if len(params) != stmt.ParamCount {
        panic(BIND_COUNT_ERROR)
    }
    for ii, par := range params {
        pval := reflect.NewValue(par)
        // Dereference pointer
        if vv, ok := pval.(*reflect.PtrValue); ok {
            pval = vv.Elem()
        }
        stmt.params[ii] = bindValue(pval)
    }
}

func (stmt *Statement) Execute() (res *Result, err os.Error) {
    if stmt.db.conn == nil {
        return nil, NOT_CONN_ERROR
    }
    if stmt.db.unreaded_rows {
        return nil, UNREADED_ROWS_ERROR
    }
    defer stmt.db.unlockIfError(&err)
    defer catchOsError(&err)
    stmt.db.lock()

    // Send EXEC command with binded parameters
    stmt.sendCmdExec()
    // Get response
    res = stmt.db.getResponse()
    res.binary = true
    return
}

// Destroy statement on server side. Client side handler is invalid after this
// command.
func (stmt *Statement) Delete() (err os.Error) {
    if stmt.db.conn == nil {
        return NOT_CONN_ERROR
    }
    if stmt.db.unreaded_rows {
        return UNREADED_ROWS_ERROR
    }
    defer stmt.db.unlock()
    defer catchOsError(&err)
    stmt.db.lock()

    // Send command
    stmt.db.sendCmd(_COM_STMT_CLOSE, stmt.id)
    // Invalidate handler
    *stmt = Statement{}
    return
}

// Resets a prepared statement on server: data sent to the server, unbuffered
// result sets and current errors.
func (stmt *Statement) Reset() (err os.Error) {
    if stmt.db.conn == nil {
        return NOT_CONN_ERROR
    }
    if stmt.db.unreaded_rows {
        return UNREADED_ROWS_ERROR
    }
    defer stmt.db.unlock()
    defer catchOsError(&err)
    stmt.db.lock()

    // Send command
    stmt.db.sendCmd(_COM_STMT_RESET, stmt.id)
    // Get result
    stmt.db.getResult(nil)
    // Next exec must send type information.
    stmt.rebind = true
    return
}
