package mymysql

import (
    "log"
)

type Statement struct {
    db  *MySQL
    id  uint32
    sql string // For reprepare during reconnect

    params []*paramValue // Parameters binding
    rebind bool

    Fields []*Field
    Map     map[string]int // Maps field name to column number

    FieldCount   int
    ParamCount   int
    WarningCount int
    Status       uint16
}

func (stmt *Statement) sendCmdExec() {
    // Calculate packet length and NULL bitmap
    null_bitmap := make([]byte, (stmt.ParamCount + 7) >> 3)
    pkt_len := 1 + 4 + 1 + 4 + 1 + len(null_bitmap)
    for ii, param := range stmt.params {
        par_len := param.Len()
        pkt_len += par_len
        if par_len == 0 {
            null_byte := ii >> 3
            null_mask := byte(1) << uint(ii - (null_byte << 3))
            null_bitmap[null_byte] |= null_mask
        }
    }
    if stmt.rebind {
        pkt_len += stmt.ParamCount * 2
    }
    // Reset sequence number
    stmt.db.seq = 0
    // Packet sending
    pw := stmt.db.newPktWriter(pkt_len)
    writeByte(pw, _COM_STMT_EXECUTE)
    writeU32(pw, stmt.id)
    writeByte(pw, 0) // flags = CURSOR_TYPE_NO_CURSOR
    writeU32(pw, 1)  // iteration_count
    write(pw, null_bitmap)
    if stmt.rebind {
        writeByte(pw, 1)
        // Types
        for _, param := range stmt.params {
            writeU16(pw, param.typ)
        }
    } else {
        writeByte(pw, 0)
    }
    // Values
    for _, param := range stmt.params {
        writeValue(pw, param)
    }
    // Mark that we sended information about binded types
    stmt.rebind = false

    if stmt.db.Debug {
        log.Printf("[%2d <-] Exec command packet: len=%d",
            stmt.db.seq - 1, pkt_len)
    }
}

func (my *MySQL) getPrepareResult(stmt *Statement) interface{} {
loop:
    pr   := my.newPktReader() // New reader for next packet
    pkt0 := readByte(pr)

    //log.Println("pkt0:", pkt0, "stmt:", stmt)

    if pkt0 == 255 {
        // Error packet
        my.getErrorPacket(pr)
    }

    if stmt == nil {
        if pkt0 == 0 {
            // OK packet
            return my.getPrepareOkPacket(pr)
        }
    } else {
        unreaded_params := (stmt.ParamCount < len(stmt.params))
        switch {
        case pkt0 == 254:
            // EOF packet
            stmt.WarningCount, stmt.Status = my.getEofPacket(pr)
            stmt.db.Status = stmt.Status
            return stmt

        case pkt0 > 0 && pkt0 < 251 && (stmt.FieldCount < len(stmt.Fields) ||
                unreaded_params):
            // Field packet
            if unreaded_params {
                // Read and ignore parameter field. Sentence from MySQL source:
                /* skip parameters data: we don't support it yet */
                my.getFieldPacket(pr)
                // Increment field count
                stmt.ParamCount++
            } else {
                field := my.getFieldPacket(pr)
                stmt.Fields[stmt.FieldCount] = field
                stmt.Map[field.Name] = stmt.FieldCount
                // Increment field count
                stmt.FieldCount++
            }
            // Read next packet
            goto loop
        }
    }
    panic(UNK_RESULT_PKT_ERROR)
}

func (my *MySQL) getPrepareOkPacket(pr *pktReader) (stmt *Statement) {
    if my.Debug {
        log.Printf("[%2d ->] Perpared OK packet:", my.seq - 1)
    }

    stmt = new(Statement)
    // First byte was readed by getPrepRes
    stmt.db     = my
    stmt.id     = readU32(pr)
    stmt.Fields = make([]*Field, int(readU16(pr))) // FieldCount
    stmt.params = make([]*paramValue, int(readU16(pr))) // ParamCount
    read(pr, 1)
    stmt.WarningCount = int(readU16(pr))
    pr.checkEof()

    // Make field map if fields exists.
    if len(stmt.Fields) > 0 {
        stmt.Map = make(map[string]int)
    }
    if my.Debug {
        log.Printf(tab8s + "ID=0x%x ParamCount=%d FieldsCount=%d WarnCount=%d",
            stmt.id, len(stmt.params), len(stmt.Fields), stmt.WarningCount,
        )
    }
    return
}


