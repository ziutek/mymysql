package mymy

import (
    "os"
    "net"
    "bufio"
    "sync"
)

type ServerInfo struct {
    prot_ver byte
    serv_ver string
    thr_id   uint32
    scramble []byte
    caps     uint16
    lang     byte
}

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

    mutex         *sync.Mutex // For concurency
    unreaded_rows bool

    Debug bool // Logging
}

// New MySQL object
func New(proto, laddr, raddr, user, passwd string, db ...string) (my *MySQL) {
    my = &MySQL{
        proto:  proto,
        laddr:  laddr,
        raddr:  raddr,
        user:   user,
        passwd: passwd,
        mutex:  new(sync.Mutex),
    }
    if len(db) == 1 {
        my.dbname = db[0]
    } else if len(db) > 1 {
        panic("mymy.New: too many arguments")
    }
    return
}

// Connect to the server
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
    my.sendCmd(COM_QUIT)
    my.conn.Close()
    my.conn = nil // Mark that we disconnect

    return
}

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
    my.sendCmd(COM_INIT_DB, dbname)
    // Get server response
    my.getResult(nil)
    // Save new database name
    my.dbname = dbname

    return
}

func (my *MySQL) Start(sql string) (res *Result, err os.Error) {
    if my.conn == nil {
        return nil, NOT_CONN_ERROR
    }
    if my.unreaded_rows {
        return nil, UNREADED_ROWS_ERROR
    }
    defer my.unlockIfError(&err)
    defer catchOsError(&err)
    my.lock()

    // Send command
    my.sendCmd(COM_QUERY, sql)
    // Get response
    result := my.getResult(nil)
    res, ok := result.(*Result)
    if !ok {
        return nil, BAD_RESULT_ERROR
    }
    res.db = my
    if res.FieldCount == 0 {
        // This query was ended (OK result)
        my.unlock()
    } else {
        // This query can return rows
        res.db = my
        my.unreaded_rows = true
    }
    return
}

func (res *Result) GetTextRow() (row *TextRow, err os.Error) {
    if res.FieldCount == 0 {
        // There is no fields in result (OK result)
        return
    }
    defer res.db.unlockIfError(&err)
    defer catchOsError(&err)

    switch result := res.db.getResult(res).(type) {
    case *TextRow:
        row = result

    case *Result:
        // EOF result
        res.db.unreaded_rows = false
        res.db.unlock()

    default:
        err = BAD_RESULT_ERROR
    }
    return
}

func (res *Result) End() (err os.Error) {
    // Read all unreaded rows from server
    for err == nil && res.db.unreaded_rows {
        _, err = res.GetTextRow()
    }
    return
}

func (my *MySQL) Query(sql string) (rows []*TextRow, res *Result, err os.Error){
    res, err = my.Start(sql)
    if err != nil {
        return
    }
    // Read rows
    var row *TextRow
    for {
        row, err = res.GetTextRow()
        if err != nil || row == nil {
            break
        }
        rows = append(rows, row)
    }
    return
}

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
    my.sendCmd(COM_PING)
    // Get server response
    my.getResult(nil)

    return
}
