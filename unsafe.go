package mymy

import (
    "io"
    "unsafe"
)

func (val *Value) Len() int {
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

    case MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME:
        if val.is_ptr {
            return lenNdatetime(*(**Datetime)(ptr))
        }
        return lenNdatetime((*Datetime)(ptr))
    }
    // MYSQL_TYPE_VAR_STRING, MYSQL_TYPE_BLOB and type of Raw value
    if val.is_ptr {
        return lenNbin(*(**[]byte)(ptr))
    }
    return lenNbin((*[]byte)(ptr))
}

func writeValue(wr io.Writer, val *Value) {
    ptr := unsafe.Pointer(val.addr)
    if val.raw || val.typ == MYSQL_TYPE_VAR_STRING ||
            val.typ == MYSQL_TYPE_BLOB {
        if val.is_ptr {
            if vp := *(**[]byte)(ptr); vp != nil {
                writeNbin(wr, vp)
            }
        } else {
            writeNbin(wr, (*[]byte)(ptr))
        }
        return
    }
    // We don't need unsigned bit
    switch val.typ & ^MYSQL_UNSIGNED_MASK {
    case MYSQL_TYPE_NULL:
        // Don't write null values

    case MYSQL_TYPE_STRING:
        if val.is_ptr {
            if vp := *(**string)(ptr); vp != nil {
                writeNstr(wr, vp)
            }
        } else {
            writeNstr(wr, (*string)(ptr))
        }

    case MYSQL_TYPE_LONG, MYSQL_TYPE_FLOAT:
        if val.is_ptr {
            if vp := *(**uint32)(ptr); vp != nil {
                writeU32(wr, *vp)
            }
        } else {
            writeU32(wr, *(*uint32)(ptr))
        }

    case MYSQL_TYPE_SHORT:
        if val.is_ptr {
            if vp := *(**uint16)(ptr); vp != nil {
                writeU16(wr, *vp)
            }
        } else {
            writeU16(wr, *(*uint16)(ptr))
        }

    case MYSQL_TYPE_TINY:
        if val.is_ptr {
            if vp := *(**byte)(ptr); vp != nil {
                writeByte(wr, *vp)
            }
        } else {
            writeByte(wr, *(*byte)(ptr))
        }

    case MYSQL_TYPE_LONGLONG, MYSQL_TYPE_DOUBLE:
        if val.is_ptr {
            if vp := *(**uint64)(ptr); vp != nil {
                writeU64(wr, *vp)
            }
        } else {
            writeU64(wr, *(*uint64)(ptr))
        }

    case MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME:
        if val.is_ptr {
            if vp := *(**Datetime)(ptr); vp != nil {
                writeNdatetime(wr, vp)
            }
        } else {
            writeNdatetime(wr, (*Datetime)(ptr))
        }

    default:
        panic(BIND_UNK_TYPE)
    }
    return
}
