package mymy

import (
    "reflect"
    "fmt"
)

type paramValue struct {
    typ    uint16
    addr   uintptr
    is_ptr bool
    raw    bool
    length int  // >=0 - length ov value, <0 - unknown length
}

type Datetime struct {
    Year  int16
    Month, Day, Hour, Minute, Second uint8
    Nanosec uint32
}
func (dt *Datetime) String() string {
    switch {
    case dt == nil:
        return "NULL"
    case dt.Nanosec != 0:
        return fmt.Sprintf(
            "%04d-%02d-%02d %02d:%02d:%02d.%09d", dt.Year, dt.Month, dt.Day,
            dt.Hour, dt.Minute, dt.Second,dt.Nanosec,
        )
    case dt.Second != 0 || dt.Minute != 0 || dt.Hour != 0:
        return fmt.Sprintf(
            "%04d-%02d-%02d %02d:%02d:%02d", dt.Year, dt.Month, dt.Day,
            dt.Hour, dt.Minute, dt.Second,
        )
    case dt.Day != 0 || dt.Month != 0 || dt.Year != 0:
        return fmt.Sprintf( "%04d-%02d-%02d", dt.Year, dt.Month, dt.Day)
    }
   return "0000-00-00"
}

type Timestamp Datetime
func (ts *Timestamp) String() string {
    return (*Datetime)(ts).String()
}

type Blob []byte

type Raw struct {
    Typ uint16
    Val *[]byte
}

var (
    reflectBlobType = reflect.Typeof(Blob{})
    reflectDatetimeType = reflect.Typeof(Datetime{})
    reflectTimestampType = reflect.Typeof(Timestamp{})
    reflectRawType = reflect.Typeof(Raw{})
)

func bindValue(val reflect.Value) (out *paramValue) {
    if val == nil {
        return &paramValue{typ: MYSQL_TYPE_NULL}
    }

    out = &paramValue{addr: val.Addr()}
    typ := val.Type()

    // Dereference type
    if tp, ok := typ.(*reflect.PtrType); ok {
        typ = tp.Elem()
        out.is_ptr = true
    }

    // Obtain value type
    switch tt := typ.(type) {
    case *reflect.StringType:
        out.typ    = MYSQL_TYPE_STRING
        out.length = -1
        return

    case *reflect.IntType:
        out.typ, out.length = mysqlIntType(tt.Kind())
        return

    case *reflect.UintType:
        out.typ, out.length = mysqlIntType(tt.Kind())
        out.typ |= MYSQL_UNSIGNED_MASK
        return

    case *reflect.FloatType:
        out.typ, out.length = mysqlFloatType(tt.Kind())
        return

    case *reflect.SliceType:
        out.length = -1
        if tt == reflectBlobType {
            out.typ = MYSQL_TYPE_BLOB
            return
        }
        if it, ok := tt.Elem().(*reflect.UintType); ok &&
                it.Kind() == reflect.Uint8 {
            out.typ = MYSQL_TYPE_VAR_STRING
            return
        }

    case *reflect.StructType:
        out.length = -1
        if tt == reflectDatetimeType {
            out.typ = MYSQL_TYPE_DATETIME
            return
        }
        if tt == reflectTimestampType {
            out.typ = MYSQL_TYPE_TIMESTAMP
            return
        }
        if tt == reflectRawType {
            rv := val.(*reflect.StructValue)
            out.typ = uint16(rv.FieldByName("Typ").(*reflect.UintValue).Get())
            out.addr = rv.FieldByName("Val").(*reflect.PtrValue).Get()
            out.is_ptr = true
            out.raw = true
            return
        }
    }
    panic(BIND_UNK_TYPE)
}

func mysqlIntType(kind reflect.Kind) (uint16, int) {
    switch kind {
    case reflect.Int, reflect.Uint:
        return _INT_TYPE, _SIZE_OF_INT

    case reflect.Int8, reflect.Uint8:
        return MYSQL_TYPE_TINY, 1

    case reflect.Int16, reflect.Uint16:
        return MYSQL_TYPE_SHORT, 2

    case reflect.Int32, reflect.Uint32:
        return MYSQL_TYPE_LONG, 4

    case reflect.Int64, reflect.Uint64:
        return MYSQL_TYPE_LONGLONG, 8
    }
    panic("unknown int kind")
}

func mysqlFloatType(kind reflect.Kind) (uint16, int) {
    switch kind {
    case reflect.Float:
        return _FLOAT_TYPE, _SIZE_OF_FLOAT

    case reflect.Float32:
        return MYSQL_TYPE_FLOAT, 4

    case reflect.Float64:
        return MYSQL_TYPE_DOUBLE, 8
    }
    panic("unknown float kind")
}
