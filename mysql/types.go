package mysql

import (
	"strings"
	"strconv"
	"time"
    "fmt"
	"reflect"
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
	tt := StrToTime(str[10:])
	if tt == nil || *tt < 0 {
		return nil
	}
	ti := *tt
	dt.Nanosec = uint32(ti % 1e9)
	ti /= 1e9
	dt.Second = uint8(ti % 60)
	ti /= 60
	dt.Minute = uint8(ti % 60)
	ti /= 60
	if ti > 23 {
		return nil
	}
	dt.Hour = uint8(ti)
	return
}

// Convert *Date to *Datetime. Return nil if dd is nil
func DateToDatetime(dd *Date) *Datetime {
	if dd == nil {
		return nil
	}
	return &Datetime{
		Year:  dd.Year,
		Month: dd.Month,
		Day:   dd.Day,
	}
}


// True if datetime is 0000-00-00 00:00:00
func IsDatetimeZero(dt *Datetime) bool {
	return dt.Day == 0 && dt.Month == 0 && dt.Year == 0 && dt.Hour == 0 &&
		dt.Minute == 0 && dt.Second == 0 && dt.Nanosec == 0
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
		ok error
	)
	if ii, ok = strconv.Atoi(str[0:4]); ok != nil {
		return nil
	}
	dd.Year = int16(ii)
	if ii, ok = strconv.Atoi(str[5:7]); ok != nil {
		return nil
	}
	dd.Month = uint8(ii)
	if ii, ok = strconv.Atoi(str[8:10]); ok != nil {
		return nil
	}
	dd.Day = uint8(ii)
	return
}

// True if date is 0000-00-00
func IsDateZero(dd *Date) bool {
	return dd.Day == 0 && dd.Month == 0 && dd.Year == 0
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

// Convert string time in format [+-]H+:MM:SS[.UUUUUUUUU] to Time.
// Leading and trailing spaces are ignored. If format is invalid returns nil.
func StrToTime(str string) *Time {
	str = strings.TrimSpace(str)
	// Check sign
	sign := Time(1)
	switch str[0] {
	case '-':
		sign = -1
		fallthrough
	case '+':
		str = str[1:]
	}
	var (
		ii int
		ok error
		tt Time
	)
	// Find houre
	if nn := strings.IndexRune(str, ':'); nn != -1 {
		if ii, ok = strconv.Atoi(str[0:nn]); ok != nil {
			return nil
		}
		tt = Time(ii * 3600)
		str = str[nn+1:]
	} else {
		return nil
	}
	if len(str) != 5 && len(str) != 15 || str[2] != ':' {
		return nil
	}
	if ii, ok = strconv.Atoi(str[0:2]); ok != nil || ii > 59 {
		return nil
	}
	tt += Time(ii * 60)
	if ii, ok = strconv.Atoi(str[3:5]); ok != nil || ii > 59 {
		return nil
	}
	tt += Time(ii)
	tt *= 1e9
	if len(str) == 15 {
		if str[5] != '.' {
			return nil
		}
		if ii, ok = strconv.Atoi(str[6:15]); ok != nil {
			return nil
		}
		tt += Time(ii)
	}
	tt *= sign
	return &tt
}

// Convert time.Time to *Datetime. Return nil if tt is zero
func TimeToDatetime(tt time.Time) *Datetime {
	if tt.IsZero()  {
		return nil
	}
	return &Datetime{
		Year:   int16(tt.Year()),
		Month:  uint8(tt.Month()),
		Day:    uint8(tt.Day()),
		Hour:   uint8(tt.Hour()),
		Minute: uint8(tt.Minute()),
		Second: uint8(tt.Second()),
	}
}


type Blob []byte

type Raw struct {
    Typ uint16
    Val *[]byte
}

var (
	BlobType = reflect.TypeOf(Blob{})
	DatetimeType = reflect.TypeOf(Datetime{})
	DateType = reflect.TypeOf(Date{})
	TimestampType = reflect.TypeOf(Timestamp{})
	TimeType = reflect.TypeOf(Time(0))
	RawType = reflect.TypeOf(Raw{})
)
