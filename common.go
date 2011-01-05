package mymy

import (
    "os"
    "io"
    "time"
)

var tab8s = "        "

func readFull(rd io.Reader, buf []byte) {
    for nn := 0; nn < len(buf); {
        kk, err := rd.Read(buf[nn:])
        nn += kk
        if err != nil {
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

/*func flush(wr *bufio.Writer) {
    if err := wr.Flush(); err != nil {
        panic(err)
    }
}*/

func catchOsError(err *os.Error) {
    if pv := recover(); pv != nil {
        if er, ok := pv.(os.Error); ok {
            *err = er
        } else {
            panic(pv)
        }
    }
}

func IsDatetimeZero(dt *Datetime) bool {
    return dt.Day==0 && dt.Month==0 && dt.Year==0 && dt.Hour==0 &&
        dt.Minute == 0 && dt.Second == 0 && dt.Nanosec == 0
}

func TimeToDatetime(tt *time.Time) *Datetime {
    return &Datetime {
        Year:   int16(tt.Year),
        Month:  uint8(tt.Month),
        Day:    uint8(tt.Day),
        Hour:   uint8(tt.Hour),
        Minute: uint8(tt.Minute),
        Second: uint8(tt.Second),
    }
}

func TimeToTimestamp(tt *time.Time) *Timestamp {
    return (*Timestamp)(TimeToDatetime(tt))
}
