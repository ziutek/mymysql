package mymy

import (
    "log"
    "strconv"
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
    db *MySQL

    Message       []byte
    AffectedRows  uint64
    InsertId      uint64
    WarningCount  uint16
    Status        uint16

    FieldCount int
    Fields     []*Field       // Fields table
    Map        map[string]int // Maps field name to column number
}

type TextRow struct {
    Data []Nbin
}

func (tr *TextRow) Str(nn int) string {
    if tr.Data[nn] == nil {
        return ""
    }
    return string(*tr.Data[nn])
}

func (tr *TextRow) Int(nn int) (v int) {
    v, _ = strconv.Atoi(tr.Str(nn))
    return
}

func (tr *TextRow) Uint(nn int) (v uint) {
    v, _ = strconv.Atoui(tr.Str(nn))
    return
}

func (tr *TextRow) Int64(nn int) (v int64) {
    v, _ = strconv.Atoi64(tr.Str(nn))
    return
}

func (tr *TextRow) Uint64(nn int) (v uint64) {
    v, _ = strconv.Atoui64(tr.Str(nn))
    return
}

func (my *MySQL) getResult(res *Result) interface{} {
loop:
    pr    := newPktReader(my.rd, &my.seq) // New reader for next packet
    pkt0  := readByte(pr)
    re_fi := (pkt0 > 0 && pkt0 < 251)
    switch {
    case pkt0 == 0 && res == nil:
        // OK packet
        res = new(Result)
        res.AffectedRows = readNotNullNu64(pr)
        res.InsertId     = readNotNullNu64(pr)
        res.Status       = readU16(pr)
        res.WarningCount = readU16(pr)
        res.Message      = pr.readAll()
        pr.checkEof()

        if my.Debug {
            log.Printf(
                "[%d ->] OK packet: AffectedRows=%d InsertId=0x%x " +
                "Status=0x%x WarningCount=%d Message=\"%s\"",
                my.seq - 1, res.AffectedRows, res.InsertId, res.Status,
                res.WarningCount, res.Message,
            )
        }
        return res

    case pkt0 == 254 && res != nil:
        // EOF packet
        res.WarningCount = readU16(pr)
        res.Status       = readU16(pr)
        pr.checkEof()

        if my.Debug {
            log.Printf(
                "[%d ->] EOF packet: WarningCount=%d Status=0x%x",
                my.seq - 1, res.WarningCount, res.Status,
            )
        }
        return res

    case pkt0 == 255:
        // Error packet
        var err Error
        err.code = readU16(pr)
        if readByte(pr) != '#' {
            panic(PKT_ERROR)
        }
        read(pr, 5)
        err.msg = pr.readAll()
        pr.checkEof()

        if my.Debug {
            log.Printf(
                "[%d ->] Error packet: code=0x%x msg=\"%s\"",
                my.seq - 1, err.code, err.msg,
            )
        }
        panic(&err)

    case re_fi && res == nil:
        // Result set header packet
        pr.unreadByte()

        field_count := int(readNotNullNu64(pr))
        pr.checkEof()

        res = &Result {
            Fields: make([]*Field, field_count),
            Map:    make(map[string]int),
        }

        if my.Debug {
            log.Printf("[%d ->] Result set header packet: Fields=%d",
                my.seq - 1, field_count)
        }
        // Read next packet
        goto loop

    case re_fi && res != nil && res.FieldCount < len(res.Fields):
        // Field packet
        pr.unreadByte()

        field := new(Field)

        field.Catalog  = readNotNullNstr(pr)
        field.Db       = readNotNullNstr(pr)
        field.Table    = readNotNullNstr(pr)
        field.OrgTable = readNotNullNstr(pr)
        field.Name     = readNotNullNstr(pr)
        field.OrgName  = readNotNullNstr(pr)
        read(pr, 1 + 2)
        //field.Charset= readU16(pr)
        field.DispLen  = readU32(pr)
        field.Type     = readByte(pr)
        field.Flags    = readU16(pr)
        field.Scale    = readByte(pr)
        read(pr, 2)
        pr.checkEof()

        res.Fields[res.FieldCount] = field
        // Add name and id to a Map
        res.Map[field.Name] = res.FieldCount

        if my.Debug {
            log.Printf(
                "[%d ->] Field packet %d: Name=\"%s\" Type=0x%x", my.seq - 1,
                res.FieldCount, res.Fields[res.FieldCount].Name,
                res.Fields[res.FieldCount].Type,
            )
        }
        // Increment field count
        res.FieldCount++
        // Read next packet
        goto loop

    case pkt0 < 254 && res != nil && res.FieldCount == len(res.Fields):
        // Row Data Packet
        pr.unreadByte()

        row := TextRow{Data: make([]Nbin, res.FieldCount)}
        for ii := 0; ii < res.FieldCount; ii++ {
            row.Data[ii] = readNbin(pr)
        }
        pr.checkEof()
        if my.Debug {
            log.Printf("[%d ->] Row packet", my.seq - 1)
        }
        return &row
    }
    panic(UNK_RESULT_PKT_ERROR)
}
