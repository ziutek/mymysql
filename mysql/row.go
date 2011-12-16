package mysql

import (
	"errors"
	"strconv"
	"bytes"
	"fmt"
	"os"
)

// Result row - contains values for any column of received row.
//
// If row is a result of ordinary text query, its element can be
// []byte slice, contained result text or nil if NULL is returned.
//
// If it is result of prepared statement execution, its element field can
// be: intXX, uintXX, floatXX, []byte, *Date, *Datetime, Time or nil
type Row []interface{}

// Get the nn-th value and return it as []byte ([]byte{} if NULL)
func (tr Row) Bin(nn int) (bin []byte) {
	switch data := tr[nn].(type) {
	case nil:
		// bin = []byte{}
	case []byte:
		bin = data
	default:
		buf := new(bytes.Buffer)
		fmt.Fprint(buf, data)
		bin = buf.Bytes()
	}
	return
}

// Get the nn-th value and return it as string ("" if NULL)
func (tr Row) Str(nn int) (str string) {
	switch data := tr[nn].(type) {
	case nil:
		// str = ""
	case []byte:
		str = string(data)
	default:
		str = fmt.Sprint(data)
	}
	return
}

// Get the nn-th value and return it as int (0 if NULL). Return error if
// conversion is impossible.
func (tr Row) IntErr(nn int) (val int, err error) {
	fn := "IntErr"
	switch data := tr[nn].(type) {
	case nil:
		val = 0
	case int32:
		val = int(data)
	case int16:
		val = int(data)
	case uint16:
		val = int(data)
	case int8:
		val = int(data)
	case uint8:
		val = int(data)
	case []byte:
		val, err = strconv.Atoi(string(data))
	case uint32:
		if strconv.IntSize > 32 {
			val = int(data)
		} else {
			err = &strconv.NumError{fn, fmt.Sprint(data), os.ERANGE}
		}
	case int64:
		if strconv.IntSize > 32 {
			val = int(data)
		} else {
			err = &strconv.NumError{fn, fmt.Sprint(data), os.ERANGE}
		}
	case uint64:
		err = &strconv.NumError{fn, fmt.Sprint(data), os.ERANGE}
	default:
		err = &strconv.NumError{fn, fmt.Sprint(data), os.EINVAL}
	}
	return
}

// Get the nn-th value and return it as int (0 if NULL). Panic if conversion is
// impossible.
func (tr Row) MustInt(nn int) (val int) {
	val, err := tr.IntErr(nn)
	if err != nil {
		panic(err)
	}
	return
}

// Get the nn-th value and return it as int. Return 0 if value is NULL or
// conversion is impossible.
func (tr Row) Int(nn int) (val int) {
	val, _ = tr.IntErr(nn)
	return
}

// Get the nn-th value and return it as uint (0 if NULL). Return error if
// conversion is impossible.
func (tr Row) UintErr(nn int) (val uint, err error) {
	fn := "UintErr"
	switch data := tr[nn].(type) {
	case uint32:
		val = uint(data)
	case uint16:
		val = uint(data)
	case uint8:
		val = uint(data)
	case []byte:
		var v uint64
		v, err = strconv.ParseUint(string(data), 0, 0)
		val = uint(v)
	case uint64:
		if strconv.IntSize > 32 {
			val = uint(data)
		} else {
			err = &strconv.NumError{fn, fmt.Sprint(data), os.ERANGE}
		}
	case int8, int16, int32, int64:
		err = &strconv.NumError{fn, fmt.Sprint(data), os.ERANGE}
	default:
		err = &strconv.NumError{fn, fmt.Sprint(data), os.EINVAL}
	}
	return
}

// Get the nn-th value and return it as uint (0 if NULL). Panic if conversion is
// impossible.
func (tr Row) MustUint(nn int) (val uint) {
	val, err := tr.UintErr(nn)
	if err != nil {
		panic(err)
	}
	return
}

// Get the nn-th value and return it as uint. Return 0 if value is NULL or
// conversion is impossible.
func (tr Row) Uint(nn int) (val uint) {
	val, _ = tr.UintErr(nn)
	return
}

// Get the nn-th value and return it as Date (0000-00-00 if NULL). Return error
// if conversion is impossible.
func (tr Row) DateErr(nn int) (val *Date, err error) {
	switch data := tr[nn].(type) {
	case nil:
		val = new(Date)
	case *Date:
		val = data
	case []byte:
		val = StrToDate(string(data))
	}
	if val == nil {
		err = errors.New(
			fmt.Sprintf("Can't convert `%v` to Date", tr[nn]),
		)
	}
	return
}

// It is like DateErr but panics if conversion is impossible.
func (tr Row) MustDate(nn int) (val *Date) {
	val, err := tr.DateErr(nn)
	if err != nil {
		panic(err)
	}
	return
}

// It is like DateErr but return 0000-00-00 if conversion is impossible.
func (tr Row) Date(nn int) (val *Date) {
	val, _ = tr.DateErr(nn)
	if val == nil {
		val = new(Date)
	}
	return
}

// Get the nn-th value and return it as Datetime (0000-00-00 00:00:00 if NULL).
// Return error if conversion is impossible. It can convert Date to Datetime.
func (tr Row) DatetimeErr(nn int) (val *Datetime, err error) {
	switch data := tr[nn].(type) {
	case nil:
		val = new(Datetime)
	case *Datetime:
		val = data
	case *Date:
		val = DateToDatetime(data)
	case []byte:
		val = StrToDatetime(string(data))
	}
	if val == nil {
		err = errors.New(
			fmt.Sprintf("Can't convert `%v` to Datetime", tr[nn]),
		)
	}
	return
}

// As DatetimeErr but panics if conversion is impossible.
func (tr Row) MustDatetime(nn int) (val *Datetime) {
	val, err := tr.DatetimeErr(nn)
	if err != nil {
		panic(err)
	}
	return
}

// It is like DatetimeErr but return 0000-00-00 00:00:00 if conversion is
// impossible.
func (tr Row) Datetime(nn int) (val *Datetime) {
	val, _ = tr.DatetimeErr(nn)
	if val == nil {
		val = new(Datetime)
	}
	return
}

// Get the nn-th value and return it as Time (0:00:00 if NULL). Return error
// if conversion is impossible.
func (tr Row) TimeErr(nn int) (val Time, err error) {
	var tp *Time
	switch data := tr[nn].(type) {
	case nil:
		return
	case Time:
		val = data
		return
	case []byte:
		tp = StrToTime(string(data))
	}
	if tp == nil {
		err = errors.New(
			fmt.Sprintf("Can't convert `%v` to Time", tr[nn]),
		)
		return
	}
	val = *tp
	return
}

// It is like TimeErr but panics if conversion is impossible.
func (tr Row) MustTime(nn int) (val Time) {
	val, err := tr.TimeErr(nn)
	if err != nil {
		panic(err)
	}
	return
}

// It is like TimeErr but return 0:00:00 if conversion is impossible.
func (tr Row) Time(nn int) (val Time) {
	val, _ = tr.TimeErr(nn)
	return
}


