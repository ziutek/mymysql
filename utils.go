package mysql


type ConnUtils struct {
	Con Conn
}

// This call Start and next call GetRow as long as it reads all rows from the
// result. Next it returns all readed rows as the slice of rows.
func (cu ConnUtils) Query(sql string, params ...interface{}) (rows []Row, res Result, err error) {
	res, err = cu.Con.Start(sql, params...)
	if err != nil {
		return
	}
	// Read rows
	var row Row
	for {
		row, err = res.GetRow()
		if err != nil || row == nil {
			break
		}
		rows = append(rows, row)
	}
	return
}


type StmtUtils struct {
	Stm Stmt
}

func (su StmtUtils) Exec(params ...interface{}) (rows []Row, res Result, err error) {

	res, err = su.Stm.Run(params...)
	if err != nil {
		return
	}
	// Read rows
	var row Row
	for {
		row, err = res.GetRow()
		if err != nil || row == nil {
			break
		}
		rows = append(rows, row)
	}
	return
}


type ResultUtils struct {
	Res Result
}

// Read all unreaded rows and discard them. This function is useful if you
// don't want to use the remaining rows. It has an impact only on current
// result. If there is multi result query, you must use NextResult method and
// read/discard all rows in this result, before use other method that sends
// data to the server.
func (ru ResultUtils) End() (err error) {
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
func (ru ResultUtils) GetRows() (rows []Row, err error) {
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
