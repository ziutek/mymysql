package mysql

import (
    "testing"
)

type sio struct {
    in, out string
}

func checkRow(t *testing.T, examples []sio, conv func(string) interface{}) {
    row := make(Row, 1)
    for _, ex := range examples {
        row[0] = conv(ex.in)
        str := row.Str(0)
        if str != ex.out{
            t.Fatalf("Wrong conversion: '%s' != '%s'", str, ex.out)
        }
    }
}

var dates = []sio {
    sio{"2121-11-22",          "2121-11-22"},
    sio{"1234-56-78",          "1234-56-78"},
    sio{"0000-00-00",          "0000-00-00"},
    sio{" 1234-56-78  ",       "1234-56-78"},
    sio{"1234:56:78",          "NULL"},
    sio{"1234-56-78 00:00:00", "NULL"},
}

func TestConvDate(t *testing.T) {
    conv := func(str string) interface{} {
        return StrToDate(str)
    }
    checkRow(t, dates, conv)
}


var datetimes = []sio {
    sio{"2121-11-22 11:22:32",            "2121-11-22 11:22:32"},
    sio{"  1234-56-78  22:11:22 ",        "1234-56-78 22:11:22"},
    sio{"2000-11-11",                     "2000-11-11 00:00:00"},
    sio{"-2121-11-22 11:22:32",           "NULL"},
    sio{"0000-00-00 00:00:00",            "0000-00-00 00:00:00"},
    sio{"0000-00-00",                     "0000-00-00 00:00:00"},
    sio{"2000-11-22 11:11:11.000111222",  "2000-11-22 11:11:11.000111222"},
    sio{"2000-11-22 -11:11:11.000111222", "NULL"},
}

func TestConvDatetime(t *testing.T) {
    conv := func(str string) interface{} {
        return StrToDatetime(str)
    }
    checkRow(t, datetimes, conv)
}

var times = []sio {
    sio{"1:23:45",            "1:23:45"},
    sio{"-112:23:45",         "-112:23:45"},
    sio{"+112:23:45",         "112:23:45"},
    sio{"1:60:00",            "NULL"},
    sio{"1:00:60",            "NULL"},
    sio{"1:23:45.000111333",  "1:23:45.000111333"},
    sio{"-1:23:45.000111333", "-1:23:45.000111333"},
}

func TestConvTime(t *testing.T) {
    conv := func(str string) interface{} {
        return StrToTime(str)
    }
    checkRow(t, times, conv)
}
