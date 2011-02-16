package mymysql

import (
    "log"
    "strconv"
    "bytes"
    "fmt"
    "os"
    "math"
)

type Field struct {
    Catalog  string
    Db       string
    Table    string
    OrgTable string
    Name     string
    OrgName  string
    DispLen  uint32
//  Charset  uint16
    Flags    uint16
    Type     byte
    Scale    byte
}

type Result struct {
    db     *MySQL
    binary bool // Binary result expected

    FieldCount int
    Fields     []*Field       // Fields table
    Map        map[string]int // Maps field name to column number

    Message       []byte
    AffectedRows  uint64

    // Primary key value (useful for AUTO_INCREMENT primary keys)
    InsertId      uint64

    // Number of warinigs during command execution
    // You can use the SHOW WARNINGS query for details.
    WarningCount  int

    // MySQL server status immediately after the query execution
    Status        uint16
}

// Result row. Data field is a slice that contains values for any column of
// received row.
//
// If row is a result of ordinary text query, an element of Data field can be
// []byte slice, contained result text or nil if NULL is returned.
//
// If it is result of prepared statement execution, an element of Data field can
// be: intXX, uintXX, floatXX, []byte, *Date, *Datetime, Time or nil
type Row struct {
    Data []interface{}
}

// Get the nn-th value and return it as []byte ([]byte{} if NULL)
func (tr *Row) Bin(nn int) (bin []byte) {
    switch data := tr.Data[nn].(type) {
    case nil:
        // bin = []byte{}
    case []byte:
        bin = data
    default:
        buf := new(bytes.Buffer)
        fmt.Fprint(buf, data)
        bin = buf.Bytes()
    }
    return
}

// Get the nn-th value and return it as string ("" if NULL)
func (tr *Row) Str(nn int) (str string) {
    switch data := tr.Data[nn].(type) {
    case nil:
        // str = ""
    case []byte:
        str = string(data)
    default:
        str = fmt.Sprint(data)
    }
    return
}

// Get the nn-th value and return it as int (0 if NULL). Return error if
// conversion is impossible.
func (tr *Row) IntErr(nn int) (val int, err os.Error) {
    switch data := tr.Data[nn].(type) {
    case nil:
        val = 0
    case int32:
        val = int(data)
    case int16:
        val = int(data)
    case uint16:
        val = int(data)
    case int8:
        val = int(data)
    case uint8:
        val = int(data)
    case []byte:
        val, err = strconv.Atoi(string(data))
    case uint32:
        if _SIZE_OF_INT > 4 {
            val = int(data)
        } else {
            err = &strconv.NumError{fmt.Sprint(data), os.ERANGE}
        }
    case int64:
        if _SIZE_OF_INT > 4 {
            val = int(data)
        } else {
            err = &strconv.NumError{fmt.Sprint(data), os.ERANGE}
        }
    case uint64:
        err = &strconv.NumError{fmt.Sprint(data), os.ERANGE}
    default:
        err = &strconv.NumError{fmt.Sprint(data), os.EINVAL}
    }
    return
}

// Get the nn-th value and return it as int (0 if NULL). Panic if conversion is
// impossible.
func (tr *Row) MustInt(nn int) (val int) {
    val, err := tr.IntErr(nn)
    if err != nil {
        panic(err)
    }
    return
}

// Get the nn-th value and return it as int. Return 0 if value is NULL or
// conversion is impossible.
func (tr *Row) Int(nn int) (val int) {
    val, _ = tr.IntErr(nn)
    return
}

// Get the nn-th value and return it as uint (0 if NULL). Return error if
// conversion is impossible.
func (tr *Row) UintErr(nn int) (val uint, err os.Error) {
    switch data := tr.Data[nn].(type) {
    case uint32:
        val = uint(data)
    case uint16:
        val = uint(data)
    case uint8:
        val = uint(data)
    case []byte:
        val, err = strconv.Atoui(string(data))
    case uint64:
        if _SIZE_OF_INT > 4 {
            val = uint(data)
        } else {
            err = &strconv.NumError{fmt.Sprint(data), os.ERANGE}
        }
    case int8, int16, int32, int64:
        err = &strconv.NumError{fmt.Sprint(data), os.ERANGE}
    default:
        err = &strconv.NumError{fmt.Sprint(data), os.EINVAL}
    }
    return
}

// Get the nn-th value and return it as uint (0 if NULL). Panic if conversion is
// impossible.
func (tr *Row) MustUint(nn int) (val uint) {
    val, err := tr.UintErr(nn)
    if err != nil {
        panic(err)
    }
    return
}

// Get the nn-th value and return it as uint. Return 0 if value is NULL or
// conversion is impossible.
func (tr *Row) Uint(nn int) (val uint) {
    val, _ = tr.UintErr(nn)
    return
}

// Get the nn-th value and return it as Date (0000-00-00 if NULL). Return error
// if conversion is impossible.
func (tr *Row) DateErr(nn int) (val *Date, err os.Error) {
    switch data := tr.Data[nn].(type) {
    case nil:
        val = new(Date)
    case *Date:
        val = data
    case []byte:
        val = StrToDate(string(data))
    }
    if val == nil {
        err = os.NewError(
            fmt.Sprintf("Can't convert `%v` to Date", tr.Data[nn]),
        )
    }
    return
}

// It is like DateErr but panics if conversion is impossible.
func (tr *Row) MustDate(nn int) (val *Date) {
    val, err := tr.DateErr(nn)
    if err != nil {
        panic(err)
    }
    return
}

// It is like DateErr but return 0000-00-00 if conversion is impossible.
func (tr *Row) Date(nn int) (val *Date) {
    val, _ = tr.DateErr(nn)
    if val == nil {
        val = new(Date)
    }
    return
}

// Get the nn-th value and return it as Datetime (0000-00-00 00:00:00 if NULL).
// Return error if conversion is impossible. It can convert Date to Datetime.
func (tr *Row) DatetimeErr(nn int) (val *Datetime, err os.Error) {
    switch data := tr.Data[nn].(type) {
    case nil:
        val = new(Datetime)
    case *Datetime:
        val = data
    case *Date:
        val = DateToDatetime(data)
    case []byte:
        val = StrToDatetime(string(data))
    }
    if val == nil {
        err = os.NewError(
            fmt.Sprintf("Can't convert `%v` to Datetime", tr.Data[nn]),
        )
    }
    return
}

// As DatetimeErr but panics if conversion is impossible.
func (tr *Row) MustDatetime(nn int) (val *Datetime) {
    val, err := tr.DatetimeErr(nn)
    if err != nil {
        panic(err)
    }
    return
}

// It is like DatetimeErr but return 0000-00-00 00:00:00 if conversion is
// impossible.
func (tr *Row) Datetime(nn int) (val *Datetime) {
    val, _ = tr.DatetimeErr(nn)
    if val == nil {
        val = new(Datetime)
    }
    return
}

// Get the nn-th value and return it as Time (0:00:00 if NULL). Return error
// if conversion is impossible.
func (tr *Row) TimeErr(nn int) (val Time, err os.Error) {
    var tp *Time
    switch data := tr.Data[nn].(type) {
    case nil:
        return
    case Time:
        val = data
        return
    case []byte:
        tp = StrToTime(string(data))
    }
    if tp == nil {
        err = os.NewError(
            fmt.Sprintf("Can't convert `%v` to Time", tr.Data[nn]),
        )
        return
    }
    val = *tp
    return
}

// It is like TimeErr but panics if conversion is impossible.
func (tr *Row) MustTime(nn int) (val Time) {
    val, err := tr.TimeErr(nn)
    if err != nil {
        panic(err)
    }
    return
}

// It is like TimeErr but return 0:00:00 if conversion is impossible.
func (tr *Row) Time(nn int) (val Time) {
    val, _ = tr.TimeErr(nn)
    return
}
func (my *MySQL) getResult(res *Result) interface{} {
loop:
    pr   := my.newPktReader() // New reader for next packet
    pkt0 := readByte(pr)

    if pkt0 == 255 {
        // Error packet
        my.getErrorPacket(pr)
    }

    if res == nil {
        switch {
        case pkt0 == 0:
            // OK packet
            return my.getOkPacket(pr)

        case pkt0 > 0 && pkt0 < 251:
            // Result set header packet
            res = my.getResSetHeadPacket(pr)
            // Read next packet
            goto loop
        }
    } else {
        switch {
        case pkt0 == 254:
            // EOF packet
            res.WarningCount, res.Status = my.getEofPacket(pr)
            my.Status = res.Status
            return res

        case pkt0 > 0 && pkt0 < 251 && res.FieldCount < len(res.Fields):
            // Field packet
            field := my.getFieldPacket(pr)
            res.Fields[res.FieldCount] = field
            res.Map[field.Name] = res.FieldCount
            // Increment field count
            res.FieldCount++
            // Read next packet
            goto loop

        case pkt0 < 254  && res.FieldCount == len(res.Fields):
            // Row Data Packet
            if res.binary {
                return my.getBinRowPacket(pr, res)
            } else {
                return my.getTextRowPacket(pr, res)
            }
        }
    }
    panic(UNK_RESULT_PKT_ERROR)
}

func (my *MySQL) getOkPacket(pr *pktReader) (res *Result) {
    if my.Debug {
        log.Printf("[%2d ->] OK packet:", my.seq - 1)
    }
    res = new(Result)
    // First byte was readed by getResult
    res.db           = my
    res.AffectedRows = readNotNullU64(pr)
    res.InsertId     = readNotNullU64(pr)
    res.Status       = readU16(pr)
    my.Status        = res.Status
    res.WarningCount = int(readU16(pr))
    res.Message      = pr.readAll()
    pr.checkEof()

    if my.Debug {
        log.Printf(tab8s + "AffectedRows=%d InsertId=0x%x Status=0x%x " +
            "WarningCount=%d Message=\"%s\"", res.AffectedRows, res.InsertId,
            res.Status, res.WarningCount, res.Message,
        )
    }
    return
}

func (my *MySQL) getErrorPacket(pr *pktReader) {
    if my.Debug {
        log.Printf("[%2d ->] Error packet:", my.seq - 1)
    }
    var err Error
    err.Code = readU16(pr)
    if readByte(pr) != '#' {
        panic(PKT_ERROR)
    }
    read(pr, 5)
    err.Msg = pr.readAll()
    pr.checkEof()

    if my.Debug {
        log.Printf(tab8s + "code=0x%x msg=\"%s\"", err.Code, err.Msg)
    }
    panic(&err)
}


func (my *MySQL) getEofPacket(pr *pktReader) (warn_count int, status uint16) {
    if my.Debug {
        log.Printf("[%2d ->] EOF packet:", my.seq - 1)
    }
    warn_count = int(readU16(pr))
    status     = readU16(pr)
    pr.checkEof()

    if my.Debug {
        log.Printf(tab8s + "WarningCount=%d Status=0x%x", warn_count, status)
    }
    return
}

func (my *MySQL) getResSetHeadPacket(pr *pktReader) (res *Result) {
    if my.Debug {
        log.Printf("[%2d ->] Result set header packet:", my.seq - 1)
    }
    pr.unreadByte()

    field_count := int(readNotNullU64(pr))
    pr.checkEof()

    res = &Result {
        db:     my,
        Fields: make([]*Field, field_count),
        Map:    make(map[string]int),
    }

    if my.Debug {
        log.Printf(tab8s + "FieldCount=%d", field_count)
    }
    return
}

func (my *MySQL) getFieldPacket(pr *pktReader) (field *Field) {
    if my.Debug {
        log.Printf("[%2d ->] Field packet:", my.seq - 1)
    }
    pr.unreadByte()

    field = new(Field)
    field.Catalog  = readNotNullStr(pr)
    field.Db       = readNotNullStr(pr)
    field.Table    = readNotNullStr(pr)
    field.OrgTable = readNotNullStr(pr)
    field.Name     = readNotNullStr(pr)
    field.OrgName  = readNotNullStr(pr)
    read(pr, 1 + 2)
    //field.Charset= readU16(pr)
    field.DispLen  = readU32(pr)
    field.Type     = readByte(pr)
    field.Flags    = readU16(pr)
    field.Scale    = readByte(pr)
    read(pr, 2)
    pr.checkEof()

    if my.Debug {
        log.Printf(tab8s + "Name=\"%s\" Type=0x%x", field.Name, field.Type)
    }
    return
}

func (my *MySQL) getTextRowPacket(pr *pktReader, res *Result) *Row {
    if my.Debug {
        log.Printf("[%2d ->] Text row data packet", my.seq - 1)
    }
    pr.unreadByte()

    row := Row{Data: make([]interface{}, res.FieldCount)}
    for ii := 0; ii < res.FieldCount; ii++ {
        nbin := readNbin(pr)
        if nbin == nil {
            row.Data[ii] = nil
        } else {
            row.Data[ii] = *nbin
        }
    }
    pr.checkEof()

    return &row
}

func (my *MySQL) getBinRowPacket(pr *pktReader, res *Result) *Row {
    if my.Debug {
        log.Printf("[%2d ->] Binary row data packet", my.seq - 1)
    }
    // First byte was readed by getResult

    null_bitmap := make([]byte, (res.FieldCount + 7 + 2) >> 3)
    readFull(pr, null_bitmap)

    row := Row{Data: make([]interface{}, res.FieldCount)}
    for ii, field := range res.Fields {
        null_byte := (ii + 2) >> 3
        null_mask := byte(1) << uint(2 + ii - (null_byte << 3))
        if null_bitmap[null_byte] & null_mask != 0 {
            // Null field
            row.Data[ii] = nil
            continue
        }
        typ := field.Type
        unsigned := (field.Flags & _FLAG_UNSIGNED) != 0
        switch typ {
        case MYSQL_TYPE_TINY:
            if unsigned {
                row.Data[ii] = readByte(pr)
            } else {
                row.Data[ii] = int8(readByte(pr))
            }

        case MYSQL_TYPE_SHORT:
            if unsigned {
                row.Data[ii] = readU16(pr)
            } else {
                row.Data[ii] = int16(readU16(pr))
            }

        case MYSQL_TYPE_LONG:
            if unsigned {
                row.Data[ii] = readU32(pr)
            } else {
                row.Data[ii] = int32(readU32(pr))
            }

        case MYSQL_TYPE_LONGLONG:
            if unsigned {
                row.Data[ii] = readU64(pr)
            } else {
                row.Data[ii] = int64(readU64(pr))
            }

        case MYSQL_TYPE_INT24:
            if unsigned {
                row.Data[ii] = readU24(pr)
            } else {
                row.Data[ii] = int32(readU24(pr))
            }

        case MYSQL_TYPE_FLOAT:
            row.Data[ii] = math.Float32frombits(readU32(pr))

        case MYSQL_TYPE_DOUBLE:
            row.Data[ii] = math.Float64frombits(readU64(pr))

        case MYSQL_TYPE_STRING, MYSQL_TYPE_VAR_STRING, MYSQL_TYPE_DECIMAL,
                MYSQL_TYPE_VARCHAR, MYSQL_TYPE_BIT, MYSQL_TYPE_BLOB,
                MYSQL_TYPE_TINY_BLOB, MYSQL_TYPE_MEDIUM_BLOB,
                MYSQL_TYPE_LONG_BLOB, MYSQL_TYPE_SET, MYSQL_TYPE_ENUM:
            row.Data[ii] = readNotNullBin(pr)

        case MYSQL_TYPE_DATE:
            row.Data[ii] = readNotNullDate(pr)

        case MYSQL_TYPE_DATETIME, MYSQL_TYPE_TIMESTAMP:
            row.Data[ii] = readNotNullDatetime(pr)

        case MYSQL_TYPE_TIME:
            row.Data[ii] = readNotNullTime(pr)

        // TODO:
        // MYSQL_TYPE_NEWDATE, MYSQL_TYPE_NEWDECIMAL, MYSQL_TYPE_GEOMETRY      

        default:
            panic(UNK_MYSQL_TYPE_ERROR)
        }
    }
    return &row
}
