package mysql

type ResUtils struct {
	Res Result
}

// Read all unreaded rows and discard them. This function is useful if you
// don't want to use the remaining rows. It has an impact only on current
// result. If there is multi result query, you must use NextResult method and
// read/discard all rows in this result, before use other method that sends
// data to the server.
func (ru ResUtils) End() (err error) {
	var row Row
	for {
		row, err = ru.Res.GetRow();
		if err != nil || row == nil {
			break
		}
	}
	return
}

// Reads all rows from result and returns them as slice.
func (ru ResUtils) GetRows() (rows []Row, err error) {
	var row Row
	for {
		row, err = ru.Res.GetRow()
		if err != nil || row == nil {
			break
		}
		rows = append(rows, row)
	}
	return
}
