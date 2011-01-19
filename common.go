package mymy

import (
    "os"
    "io"
    "time"
    "strings"
    "strconv"
)

var tab8s = "        "

func readFull(rd io.Reader, buf []byte) {
    for nn := 0; nn < len(buf); {
        kk, err := rd.Read(buf[nn:])
        nn += kk
        if err != nil {
            if err == os.EOF {
                err = io.ErrUnexpectedEOF
            }
            panic(err)
        }
    }
}

func read(rd io.Reader, nn int) (buf []byte) {
    buf = make([]byte, nn)
    readFull(rd, buf)
    return
}

func readByte(rd io.Reader) byte {
    buf := make([]byte, 1)
    if _, err := rd.Read(buf); err != nil {
        if err == os.EOF {
            err = io.ErrUnexpectedEOF
       }
       panic(err)
    }
    return buf[0]
}

func write(wr io.Writer, buf []byte) {
    if _, err := wr.Write(buf); err != nil {
        panic(err)
    }
}

func writeByte(wr io.Writer, ch byte) {
    write(wr, []byte{ch})
}

func writeString(wr io.Writer, str string) {
    write(wr, []byte(str))
}

func writeBS(wr io.Writer, bs interface{}) {
    switch buf := bs.(type) {
    case string:
        writeString(wr, buf)
    case []byte:
        write(wr, buf)
    default:
        panic("Can't write: argument isn't a string nor []byte")
    }
}

func lenBS(bs interface{}) int {
    switch buf := bs.(type) {
    case string:
        return len(buf)
    case []byte:
        return len(buf)
    }
    panic("Can't get length: argument isn't a string nor []byte")
}

func catchOsError(err *os.Error) {
    if pv := recover(); pv != nil {
        if er, ok := pv.(os.Error); ok {
            *err = er
        } else {
            panic(pv)
        }
    }
}

// True if datetime is 0000-00-00 00:00:00
func IsDatetimeZero(dt *Datetime) bool {
    return dt.Day==0 && dt.Month==0 && dt.Year==0 && dt.Hour==0 &&
        dt.Minute == 0 && dt.Second == 0 && dt.Nanosec == 0
}

// Convert *time.Time to *Datetime. Return nil if tt is nil
func TimeToDatetime(tt *time.Time) *Datetime {
    if tt == nil {
        return nil
    }
    return &Datetime {
        Year:   int16(tt.Year),
        Month:  uint8(tt.Month),
        Day:    uint8(tt.Day),
        Hour:   uint8(tt.Hour),
        Minute: uint8(tt.Minute),
        Second: uint8(tt.Second),
    }
}

// Convert *Date to *Datetime. Return nil if dd is nil
func DateToDatetime(dd *Date) *Datetime {
    if dd == nil {
        return nil
    }
    return &Datetime {
        Year:  dd.Year,
        Month: dd.Month,
        Day:   dd.Day,
    }
}

// True if date is 0000-00-00
func IsDateZero(dd *Date) bool {
    return dd.Day==0 && dd.Month==0 && dd.Year==0
}

// Convert string date in format YYYY-MM-DD to Date.
// Leading and trailing spaces are ignored. If format is invalid returns nil.
func StrToDate(str string) (dd *Date) {
    str = strings.TrimSpace(str)
    if len(str) != 10 || str[4] != '-' || str[7] != '-' {
        return nil
    }
    dd = new(Date)
    var (
        ii int
        ok os.Error
    )
    if ii, ok = strconv.Atoi(str[0:4]); ok != nil { return nil }
    dd.Year = int16(ii)
    if ii, ok = strconv.Atoi(str[5:7]); ok != nil { return nil }
    dd.Month = uint8(ii)
    if ii, ok = strconv.Atoi(str[8:10]); ok != nil { return nil }
    dd.Day = uint8(ii)
    return
}

// Convert string datetime in format YYYY-MM-DD[ HH:MM:SS] to Datetime.
// Leading and trailing spaces are ignored. If format is invalid returns nil.
func StrToDatetime(str string) (dt *Datetime) {
    str = strings.TrimSpace(str)
    if len(str) >= 10 {
        dt = DateToDatetime(StrToDate(str[0:10]))
    }
    if len(str) == 10 || dt == nil {
        return
    }
    str = strings.TrimSpace(str[10:])
    if len(str) != 8 || str[2] != ':' || str[5] != ':' {
        return nil
    }
    var (
        ii int
        ok os.Error
    )
    if ii, ok = strconv.Atoi(str[0:2]); ok != nil { return nil }
    dt.Hour = uint8(ii)
    if ii, ok = strconv.Atoi(str[3:5]); ok != nil { return nil }
    dt.Minute = uint8(ii)
    if ii, ok = strconv.Atoi(str[6:8]); ok != nil { return nil }
    dt.Second = uint8(ii)
    return
}
