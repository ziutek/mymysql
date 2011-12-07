package native

import (
	"github.com/ziutek/mymysql/mysql"
	"io"
	"unsafe"
)

type paramValue struct {
	typ    uint16
	addr   unsafe.Pointer
	raw    bool
	length int // >=0 - length of value, <0 - unknown length
}

func (pv *paramValue) SetAddr(addr uintptr) {
	pv.addr = unsafe.Pointer(addr)
}

func (val *paramValue) Len() int {
	if val.addr == nil {
		// Invalid Value was binded
		return 0
	}
	// val.addr always points to the pointer - lets dereference it
	ptr := *(*unsafe.Pointer)(val.addr)
	if ptr == nil {
		// Binded Ptr Value is nil
		return 0
	}

	if val.length >= 0 {
		return val.length
	}

	switch val.typ {
	case MYSQL_TYPE_STRING:
		return lenNstr((*string)(ptr))

	case MYSQL_TYPE_DATE:
		return lenNdate((*mysql.Date)(ptr))

	case MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME:
		return lenNdatetime((*mysql.Datetime)(ptr))

	case MYSQL_TYPE_TIME:
		return lenNtime((*mysql.Time)(ptr))
	}
	// MYSQL_TYPE_VAR_STRING, MYSQL_TYPE_BLOB and type of Raw value
	return lenNbin((*[]byte)(ptr))
}

func writeValue(wr io.Writer, val *paramValue) {
	if val.addr == nil {
		// Invalid Value was binded
		return
	}
	// val.addr always points to the pointer - lets dereference it
	ptr := *(*unsafe.Pointer)(val.addr)
	if ptr == nil {
		// Binded Ptr Value is nil
		return
	}

	if val.raw || val.typ == MYSQL_TYPE_VAR_STRING ||
		val.typ == MYSQL_TYPE_BLOB {
		writeNbin(wr, (*[]byte)(ptr))
		return
	}
	// We don't need unsigned bit to check type
	switch val.typ & ^MYSQL_UNSIGNED_MASK {
	case MYSQL_TYPE_NULL:
		// Don't write null values

	case MYSQL_TYPE_STRING:
		writeNstr(wr, (*string)(ptr))

	case MYSQL_TYPE_LONG, MYSQL_TYPE_FLOAT:
		writeU32(wr, *(*uint32)(ptr))

	case MYSQL_TYPE_SHORT:
		writeU16(wr, *(*uint16)(ptr))

	case MYSQL_TYPE_TINY:
		writeByte(wr, *(*byte)(ptr))

	case MYSQL_TYPE_LONGLONG, MYSQL_TYPE_DOUBLE:
		writeU64(wr, *(*uint64)(ptr))

	case MYSQL_TYPE_DATE:
		writeNdate(wr, (*mysql.Date)(ptr))

	case MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME:
		writeNdatetime(wr, (*mysql.Datetime)(ptr))

	case MYSQL_TYPE_TIME:
		writeNtime(wr, (*mysql.Time)(ptr))

	default:
		panic(BIND_UNK_TYPE)
	}
	return
}
