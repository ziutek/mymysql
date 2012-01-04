package native

import (
	"github.com/ziutek/mymysql/mysql"
	"log"
	"math"
)

type Result struct {
	my     *Conn
	binary bool // Binary result expected

	field_count int
	fields      []*mysql.Field // Fields table
	fc_map      map[string]int // Maps field name to column number

	message       []byte
	affected_rows uint64

	// Primary key value (useful for AUTO_INCREMENT primary keys)
	insert_id uint64

	// Number of warinigs during command execution
	// You can use the SHOW WARNINGS query for details.
	warning_count int

	// MySQL server status immediately after the query execution
	status uint16

	// Have we received the end of Field Packets EOF?
	field_eof bool

	// Have we received the end of Data Packets EOF?
	row_data_eof bool
}

func (res *Result) Fields() []*mysql.Field {
	return res.fields
}

// Returns index for given name or -1 if field of that name doesn't exist
func (res *Result) Map(field_name string) int {
	if fi, ok := res.fc_map[field_name]; ok {
		return fi
	}
	return -1
}

func (res *Result) Message() string {
	return string(res.message)
}

func (res *Result) AffectedRows() uint64 {
	return res.affected_rows
}

func (res *Result) InsertId() uint64 {
	return res.insert_id
}

func (res *Result) WarnCount() int {
	return res.warning_count
}

func (my *Conn) getResult(res *Result) interface{} {

	// No-Op if done reading rows
	// This avoids a case where readByte() blocks
	// indefintely
	if res != nil && res.row_data_eof {
		return res
	}

loop:
	pr := my.newPktReader() // New reader for next packet
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
			res.warning_count, res.status = my.getEofPacket(pr)
			my.status = res.status
			if res.field_eof {
				res.row_data_eof = true
			} else {
				res.field_eof = true
			}
			return res

		case pkt0 > 0 && pkt0 < 251 && res.field_count < len(res.fields):
			// Field packet
			field := my.getFieldPacket(pr)
			res.fields[res.field_count] = field
			res.fc_map[field.Name] = res.field_count
			// Increment field count
			res.field_count++
			// Read next packet
			goto loop

		case pkt0 < 254 && res.field_count == len(res.fields):
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

func (my *Conn) getOkPacket(pr *pktReader) (res *Result) {
	if my.Debug {
		log.Printf("[%2d ->] OK packet:", my.seq-1)
	}
	res = new(Result)
	res.my = my
	// First byte was readed by getResult
	res.affected_rows = readNotNullU64(pr)
	res.insert_id = readNotNullU64(pr)
	res.status = readU16(pr)
	my.status = res.status
	res.warning_count = int(readU16(pr))
	res.message = pr.readAll()
	pr.checkEof()

	if my.Debug {
		log.Printf(tab8s+"AffectedRows=%d InsertId=0x%x Status=0x%x "+
			"WarningCount=%d Message=\"%s\"", res.affected_rows, res.insert_id,
			res.status, res.warning_count, res.message,
		)
	}
	return
}

func (my *Conn) getErrorPacket(pr *pktReader) {
	if my.Debug {
		log.Printf("[%2d ->] Error packet:", my.seq-1)
	}
	var err mysql.Error
	err.Code = readU16(pr)
	if readByte(pr) != '#' {
		panic(PKT_ERROR)
	}
	read(pr, 5)
	err.Msg = pr.readAll()
	pr.checkEof()

	if my.Debug {
		log.Printf(tab8s+"code=0x%x msg=\"%s\"", err.Code, err.Msg)
	}
	panic(&err)
}

func (my *Conn) getEofPacket(pr *pktReader) (warn_count int, status uint16) {
	if my.Debug {
		log.Printf("[%2d ->] EOF packet:", my.seq-1)
	}
	warn_count = int(readU16(pr))
	status = readU16(pr)
	pr.checkEof()

	if my.Debug {
		log.Printf(tab8s+"WarningCount=%d Status=0x%x", warn_count, status)
	}
	return
}

func (my *Conn) getResSetHeadPacket(pr *pktReader) (res *Result) {
	if my.Debug {
		log.Printf("[%2d ->] Result set header packet:", my.seq-1)
	}
	pr.unreadByte()

	field_count := int(readNotNullU64(pr))
	pr.checkEof()

	res = &Result{
		my:           my,
		fields:       make([]*mysql.Field, field_count),
		fc_map:       make(map[string]int),
		field_eof:    false,
		row_data_eof: false,
	}

	if my.Debug {
		log.Printf(tab8s+"FieldCount=%d", field_count)
	}
	return
}

func (my *Conn) getFieldPacket(pr *pktReader) (field *mysql.Field) {
	if my.Debug {
		log.Printf("[%2d ->] Field packet:", my.seq-1)
	}
	pr.unreadByte()

	field = new(mysql.Field)
	field.Catalog = readNotNullStr(pr)
	field.Db = readNotNullStr(pr)
	field.Table = readNotNullStr(pr)
	field.OrgTable = readNotNullStr(pr)
	field.Name = readNotNullStr(pr)
	field.OrgName = readNotNullStr(pr)
	read(pr, 1+2)
	//field.Charset= readU16(pr)
	field.DispLen = readU32(pr)
	field.Type = readByte(pr)
	field.Flags = readU16(pr)
	field.Scale = readByte(pr)
	read(pr, 2)
	pr.checkEof()

	if my.Debug {
		log.Printf(tab8s+"Name=\"%s\" Type=0x%x", field.Name, field.Type)
	}
	return
}

func (my *Conn) getTextRowPacket(pr *pktReader, res *Result) mysql.Row {
	if my.Debug {
		log.Printf("[%2d ->] Text row data packet", my.seq-1)
	}
	pr.unreadByte()

	row := make(mysql.Row, res.field_count)
	for ii := 0; ii < res.field_count; ii++ {
		nbin := readNbin(pr)
		if nbin == nil {
			row[ii] = nil
		} else {
			row[ii] = *nbin
		}
	}
	pr.checkEof()

	return row
}

func (my *Conn) getBinRowPacket(pr *pktReader, res *Result) mysql.Row {
	if my.Debug {
		log.Printf("[%2d ->] Binary row data packet", my.seq-1)
	}
	// First byte was readed by getResult

	null_bitmap := make([]byte, (res.field_count+7+2)>>3)
	readFull(pr, null_bitmap)

	row := make(mysql.Row, res.field_count)
	for ii, field := range res.fields {
		null_byte := (ii + 2) >> 3
		null_mask := byte(1) << uint(2+ii-(null_byte<<3))
		if null_bitmap[null_byte]&null_mask != 0 {
			// Null field
			row[ii] = nil
			continue
		}
		typ := field.Type
		unsigned := (field.Flags & _FLAG_UNSIGNED) != 0
		switch typ {
		case MYSQL_TYPE_TINY:
			if unsigned {
				row[ii] = readByte(pr)
			} else {
				row[ii] = int8(readByte(pr))
			}

		case MYSQL_TYPE_SHORT:
			if unsigned {
				row[ii] = readU16(pr)
			} else {
				row[ii] = int16(readU16(pr))
			}

		case MYSQL_TYPE_LONG:
			if unsigned {
				row[ii] = readU32(pr)
			} else {
				row[ii] = int32(readU32(pr))
			}

		case MYSQL_TYPE_LONGLONG:
			if unsigned {
				row[ii] = readU64(pr)
			} else {
				row[ii] = int64(readU64(pr))
			}

		case MYSQL_TYPE_INT24:
			if unsigned {
				row[ii] = readU24(pr)
			} else {
				row[ii] = int32(readU24(pr))
			}

		case MYSQL_TYPE_FLOAT:
			row[ii] = math.Float32frombits(readU32(pr))

		case MYSQL_TYPE_DOUBLE:
			row[ii] = math.Float64frombits(readU64(pr))

		case MYSQL_TYPE_STRING, MYSQL_TYPE_VAR_STRING, MYSQL_TYPE_DECIMAL,
			MYSQL_TYPE_VARCHAR, MYSQL_TYPE_BIT, MYSQL_TYPE_BLOB,
			MYSQL_TYPE_TINY_BLOB, MYSQL_TYPE_MEDIUM_BLOB,
			MYSQL_TYPE_LONG_BLOB, MYSQL_TYPE_SET, MYSQL_TYPE_ENUM:
			row[ii] = readNotNullBin(pr)

		case MYSQL_TYPE_DATE:
			row[ii] = readNotNullDate(pr)

		case MYSQL_TYPE_DATETIME, MYSQL_TYPE_TIMESTAMP:
			row[ii] = readNotNullDatetime(pr)

		case MYSQL_TYPE_TIME:
			row[ii] = readNotNullTime(pr)

		// TODO:
		// MYSQL_TYPE_NEWDATE, MYSQL_TYPE_NEWDECIMAL, MYSQL_TYPE_GEOMETRY      

		default:
			panic(UNK_MYSQL_TYPE_ERROR)
		}
	}
	return row
}
