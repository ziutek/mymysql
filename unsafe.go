package mymy

import (
    "io"
    "unsafe"
)

type paramValue struct {
    typ    uint16
    addr   unsafe.Pointer
    is_ptr bool
    raw    bool
    length int  // >=0 - length of value, <0 - unknown length
}

func unsafePointer(addr uintptr) unsafe.Pointer {
    return unsafe.Pointer(addr)
}

func (val *paramValue) Len() int {
    ptr := unsafe.Pointer(val.addr)
    if val.is_ptr && *(*unsafe.Pointer)(ptr) == nil {
            // NULL value
            return 0
    }
    if val.length >= 0 {
        return val.length
    }

    switch val.typ {
    case MYSQL_TYPE_STRING:
        if val.is_ptr {
            return lenNstr(*(**string)(ptr))
        }
        return lenNstr((*string)(ptr))

    case MYSQL_TYPE_DATE:
        if val.is_ptr {
            return lenNdate(*(**Date)(ptr))
        }
        return lenNdate((*Date)(ptr))

    case MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME:
        if val.is_ptr {
            return lenNdatetime(*(**Datetime)(ptr))
        }
        return lenNdatetime((*Datetime)(ptr))

    case MYSQL_TYPE_TIME:
        if val.is_ptr {
            return lenNtime(*(**Time)(ptr))
        }
        return lenNtime((*Time)(ptr))
    }
    // MYSQL_TYPE_VAR_STRING, MYSQL_TYPE_BLOB and type of Raw value
    if val.is_ptr {
        return lenNbin(*(**[]byte)(ptr))
    }
    return lenNbin((*[]byte)(ptr))
}

func writeValue(wr io.Writer, val *paramValue) {
    if val.raw || val.typ == MYSQL_TYPE_VAR_STRING ||
            val.typ == MYSQL_TYPE_BLOB {
        if val.is_ptr {
            if vp := *(**[]byte)(val.addr); vp != nil {
                writeNbin(wr, vp)
            }
        } else {
            writeNbin(wr, (*[]byte)(val.addr))
        }
        return
    }
    // We don't need unsigned bit
    switch val.typ & ^MYSQL_UNSIGNED_MASK {
    case MYSQL_TYPE_NULL:
        // Don't write null values

    case MYSQL_TYPE_STRING:
        if val.is_ptr {
            if vp := *(**string)(val.addr); vp != nil {
                writeNstr(wr, vp)
            }
        } else {
            writeNstr(wr, (*string)(val.addr))
        }

    case MYSQL_TYPE_LONG, MYSQL_TYPE_FLOAT:
        if val.is_ptr {
            if vp := *(**uint32)(val.addr); vp != nil {
                writeU32(wr, *vp)
            }
        } else {
            writeU32(wr, *(*uint32)(val.addr))
        }

    case MYSQL_TYPE_SHORT:
        if val.is_ptr {
            if vp := *(**uint16)(val.addr); vp != nil {
                writeU16(wr, *vp)
            }
        } else {
            writeU16(wr, *(*uint16)(val.addr))
        }

    case MYSQL_TYPE_TINY:
        if val.is_ptr {
            if vp := *(**byte)(val.addr); vp != nil {
                writeByte(wr, *vp)
            }
        } else {
            writeByte(wr, *(*byte)(val.addr))
        }

    case MYSQL_TYPE_LONGLONG, MYSQL_TYPE_DOUBLE:
        if val.is_ptr {
            if vp := *(**uint64)(val.addr); vp != nil {
                writeU64(wr, *vp)
            }
        } else {
            writeU64(wr, *(*uint64)(val.addr))
        }

    case MYSQL_TYPE_DATE:
        if val.is_ptr {
            if vp := *(**Date)(val.addr); vp != nil {
                writeNdate(wr, vp)
            }
        } else {
            writeNdate(wr, (*Date)(val.addr))
        }

    case MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME:
        if val.is_ptr {
            if vp := *(**Datetime)(val.addr); vp != nil {
                writeNdatetime(wr, vp)
            }
        } else {
            writeNdatetime(wr, (*Datetime)(val.addr))
        }

    case MYSQL_TYPE_TIME:
        if val.is_ptr {
            if vp := *(**Time)(val.addr); vp != nil {
                writeNtime(wr, vp)
            }
        } else {
            writeNtime(wr, (*Time)(val.addr))
        }

    default:
        panic(BIND_UNK_TYPE)
    }
    return
}
