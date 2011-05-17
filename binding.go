package mymysql

import (
    "reflect"
    "fmt"
)

type Datetime struct {
    Year  int16
    Month, Day, Hour, Minute, Second uint8
    Nanosec uint32
}
func (dt *Datetime) String() string {
    if dt == nil {
        return "NULL"
    }
    if dt.Nanosec != 0 {
        return fmt.Sprintf(
            "%04d-%02d-%02d %02d:%02d:%02d.%09d",
            dt.Year, dt.Month, dt.Day, dt.Hour, dt.Minute, dt.Second,
            dt.Nanosec,
        )
    }
    return fmt.Sprintf(
        "%04d-%02d-%02d %02d:%02d:%02d",
        dt.Year, dt.Month, dt.Day, dt.Hour, dt.Minute, dt.Second,
    )
}

type Date struct {
    Year  int16
    Month, Day uint8
}
func (dd *Date) String() string {
    if dd == nil {
        return "NULL"
    }
    return fmt.Sprintf("%04d-%02d-%02d", dd.Year, dd.Month, dd.Day)
}

type Timestamp Datetime
func (ts *Timestamp) String() string {
    return (*Datetime)(ts).String()
}

// MySQL TIME in nanoseconds. Note that MySQL doesn't store fractional part
// of second but it is permitted for temporal values.
type Time int64
func (tt *Time) String() string {
    if tt == nil {
        return "NULL"
    }
    ti := int64(*tt)
    sign := 1
    if ti < 0 {
        sign = -1
        ti = -ti
    }
    ns := int(ti % 1e9)
    ti /= 1e9
    sec := int(ti % 60)
    ti /= 60
    min := int(ti % 60)
    hour := int(ti / 60) * sign
    if ns == 0 {
        return fmt.Sprintf("%d:%02d:%02d", hour, min, sec)
    }
    return fmt.Sprintf("%d:%02d:%02d.%09d", hour, min, sec, ns)
}

type Blob []byte

type Raw struct {
    Typ uint16
    Val *[]byte
}

var (
    reflectBlobType = reflect.TypeOf(Blob{})
    reflectDatetimeType = reflect.TypeOf(Datetime{})
    reflectDateType = reflect.TypeOf(Date{})
    reflectTimestampType = reflect.TypeOf(Timestamp{})
    reflectTimeType = reflect.TypeOf(Time(0))
    reflectRawType = reflect.TypeOf(Raw{})
)

// val should be an addressable value
func bindValue(val reflect.Value) (out *paramValue) {
    if !val.IsValid() {
        return &paramValue{typ: MYSQL_TYPE_NULL}
    }
    // We allways return an unsafe pointer to pointer to value, so create it
    typ := val.Type()
    if typ.Kind() == reflect.Ptr {
        // We have addressable pointer
        out = &paramValue{addr: unsafePointer(val.UnsafeAddr())}
        // Dereference pointer for next operation on its value
        typ = typ.Elem()
        val = val.Elem()
    } else {
        // We have addressable value. Create a pointer to it
        pv := val.Addr()
        // This pointer is unaddressable so copy it and return an address
        ppv := reflect.New(pv.Type())
        ppv.Elem().Set(pv)
        out = &paramValue{addr: unsafePointer(ppv.Pointer())}
    }

    // Obtain value type
    switch typ.Kind() {
    case reflect.String:
        out.typ    = MYSQL_TYPE_STRING
        out.length = -1
        return

    case reflect.Int:
        out.typ = _INT_TYPE
        out.length = _SIZE_OF_INT
        return

    case reflect.Int8:
        out.typ = MYSQL_TYPE_TINY
        out.length = 1
        return

    case reflect.Int16:
        out.typ = MYSQL_TYPE_SHORT
        out.length = 2
        return

    case reflect.Int32:
        out.typ = MYSQL_TYPE_LONG
        out.length = 4
        return

    case reflect.Int64:
        if typ == reflectTimeType {
            out.typ = MYSQL_TYPE_TIME
            out.length = -1
            return
        }
        out.typ = MYSQL_TYPE_LONGLONG
        out.length = 8
        return

    case reflect.Uint:
        out.typ = _INT_TYPE | MYSQL_UNSIGNED_MASK
        out.length = _SIZE_OF_INT
        return

    case reflect.Uint8:
        out.typ = MYSQL_TYPE_TINY | MYSQL_UNSIGNED_MASK
        out.length = 1
        return

    case reflect.Uint16:
        out.typ = MYSQL_TYPE_SHORT | MYSQL_UNSIGNED_MASK
        out.length = 2
        return

    case reflect.Uint32:
        out.typ = MYSQL_TYPE_LONG | MYSQL_UNSIGNED_MASK
        out.length = 4
        return

    case reflect.Uint64:
        if typ == reflectTimeType {
            out.typ = MYSQL_TYPE_TIME
            out.length = -1
            return
        }
        out.typ = MYSQL_TYPE_LONGLONG | MYSQL_UNSIGNED_MASK
        out.length = 8
        return

    case reflect.Float32:
        out.typ = MYSQL_TYPE_FLOAT
        out.length = 4
        return

    case reflect.Float64:
        out.typ = MYSQL_TYPE_DOUBLE
        out.length = 8
        return

    case reflect.Slice:
        out.length = -1
        if typ == reflectBlobType {
            out.typ = MYSQL_TYPE_BLOB
            return
        }
        if typ.Elem().Kind() == reflect.Uint8 {
            out.typ = MYSQL_TYPE_VAR_STRING
            return
        }

    case reflect.Struct:
        out.length = -1
        if typ == reflectDatetimeType {
            out.typ = MYSQL_TYPE_DATETIME
            return
        }
        if typ == reflectDateType {
            out.typ = MYSQL_TYPE_DATE
            return
        }
        if typ == reflectTimestampType {
            out.typ = MYSQL_TYPE_TIMESTAMP
            return
        }
        if typ == reflectRawType {
            out.typ = val.FieldByName("Typ").Interface().(uint16)
            out.addr = unsafePointer(
                val.FieldByName("Val").Pointer(),
            )
            out.raw = true
            return
        }
    }
    panic(BIND_UNK_TYPE)
}
