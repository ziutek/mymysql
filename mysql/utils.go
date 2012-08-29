package mysql

import (
	"io"
)

// Calls Start and next calls GetRow as long as it reads all rows from the
// result. Next it returns all readed rows as the slice of rows.
func Query(c Conn, sql string, params ...interface{}) (rows []Row, res Result, err error) {
	res, err = c.Start(sql, params...)
	if err != nil {
		return
	}
	rows, err = GetRows(res)
	return
}

// Calls Start and next calls GetFirstRow
func QueryFirst(c Conn, sql string, params ...interface{}) (row Row, res Result, err error) {
	res, err = c.Start(sql, params...)
	if err != nil {
		return
	}
	row, err = GetFirstRow(res)
	return
}

// Calls Start and next calls GetLastRow
func QueryLast(c Conn, sql string, params ...interface{}) (row Row, res Result, err error) {
	res, err = c.Start(sql, params...)
	if err != nil {
		return
	}
	row, err = GetLastRow(res)
	return
}

// Calls Run and next call GetRow as long as it reads all rows from the
// result. Next it returns all readed rows as the slice of rows.
func Exec(s Stmt, params ...interface{}) (rows []Row, res Result, err error) {
	res, err = s.Run(params...)
	if err != nil {
		return
	}
	rows, err = GetRows(res)
	return
}

// Calls Run and next call GetFirstRow
func ExecFirst(s Stmt, params ...interface{}) (row Row, res Result, err error) {
	res, err = s.Run(params...)
	if err != nil {
		return
	}
	row, err = GetFirstRow(res)
	return
}

// Calls Run and next call GetLastRow
func ExecLast(s Stmt, params ...interface{}) (row Row, res Result, err error) {
	res, err = s.Run(params...)
	if err != nil {
		return
	}
	row, err = GetLastRow(res)
	return
}

// Calls r.MakeRow and next r.ScanRow. Doesn't return io.EOF error (returns nil
// row insted).
func GetRow(r Result) (Row, error) {
	row := r.MakeRow()
	err := r.ScanRow(row)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}
	return row, nil
}

// Reads all rows from result and returns them as slice.
func GetRows(r Result) (rows []Row, err error) {
	var row Row
	for {
		row, err = r.GetRow()
		if err != nil || row == nil {
			break
		}
		rows = append(rows, row)
	}
	return
}

// Returns last row and discard others
func GetLastRow(r Result) (Row, error) {
	row := r.MakeRow()
	err := r.ScanRow(row)
	for err == nil {
		err = r.ScanRow(row)
	}
	if err == io.EOF {
		return row, nil
	}
	return nil, err
}

// Read all unreaded rows and discard them. This function is useful if you
// don't want to use the remaining rows. It has an impact only on current
// result. If there is multi result query, you must use NextResult method and
// read/discard all rows in this result, before use other method that sends
// data to the server. You can't use this function if last GetRow returned nil.
func End(r Result) error {
	_, err := GetLastRow(r)
	return err
}

// Returns first row and discard others
func GetFirstRow(r Result) (row Row, err error) {
	row, err = r.GetRow()
	if err != nil && row != nil {
		err = r.End()
	}
	return
}
