# Another MySQL connector for Go

This is a MySQL connector package created entirely in Go. It was created due to
lack of properly working connector, ready for my production application
(December 2010).

## Instaling

    $ git clone git://github.com/ziutek/mymysql
    $ cd mymysql && make install

## Testing

For testing you need test database and test user:

    mysql> create database test;
    mysql> grant all privileges on test.* to testuser@localhost;
    mysql> set password for testuser@localhost = password("TestPasswd9")

Set MySQL server IP address (or Unix socket name) in *mymy_test.go*. Next run
tests:

    $ gotest

## Interface

Main functions/methods:

    // Create new handler
    func New(proto, laddr, raddr, user, passwd string, db ...string) *MySQL

    // Connect to a server
    func (*MySQL) Connect() os.Error

    // Disconnect from a server
    func (*MySQL) Close() os.Error

    // Change database
    func (*MySQL) Use(dbname string) os.Error

    // Start new query sesion and send query to the server
    func (*MySQL) Start(sql string) (*Result, os.Error)

    // Get data row from a server. This method reads one row of result directly
    // from network connection.
    func (*Result) GetTextRow() (*TextRow, os.Error)

    // Read all unreaded rows form network connection and discard them
    func (*Result) End() os.Error

    // This call Start and next call GetTextRow once or more times. It read
    // all rows from connection and returns they as a slice.
    func (*MySQL) Query(sql string) ([]*TextRow, *Result, os.Error)

There are some mutations of *Start* and *Query*:

    func (*MySQL) Startv(a ...interface{}) (*Result, os.Error)
    func (*MySQL) Startf(format string, a ...interface{}) (*Result, os.Error)
    func (*MySQL) Queryv(a ...interface{}) ([]*TextRow, *Result, os.Error)
    func (*MySQL) Queryf(format string, a ...interface{}) ([]*TextRow, *Result, os.Error)

Data readed from a server are unmodified - they are character strings.
You can get data like in this example:

    rows, res, err := db.Query("select * from X")
    if err != nil {
        //...
    }
    for _, row := range rows {
        for _, col := range row.Data {
            if col == nil {
                // col has NULL value
            } else {
                // Do something with text in *col (type []byte)
            }
        }
        // You can get specific value from a row
        val1 := row.Data[1] // (type mymy.Nbin == *[]byte)

	// You can use it directly if conversion isn't needed
	os.Stdeout.Write(*val1)
	
        // You can get converted value
        number := row.Int(0)    // First value (type int, 0 if NULL)
        str    := row.Str(1)    // Second value (type string, "" if NULL)
        bignum := row.Uint64(2) // Thrid value (type uint64, 0 if NULL)
	
	// You may get value by column name
        val2 := row.Data[res.Map["FirstColumn"]] 
    }

If you do not want to load the entire result into memory you may use
Start / GetTextRow functions:

    res, err := db.Start("select * from X")
    if err != nil {
        //...
    }
    for {
        row, err := res.GetTextRow()
        if err != nil {
            //...
        }
        if row == nil {
            // No more rows
            break
        }

	// Print fields
        for _, field := range res.Fields {
		fmt.Print(field.Name, " ")
	}
	fmt.Println()

        // Print all rows
        for _, col := range row.Data {
            if col == nil {
                fmt.Print("<NULL>")
            } else {
                os.Stdout.Write(*col)
            }
            fmt.Print(" ")
        }
        fmt.Println()
    }

More examples are in *examples* directory.

## Thread safety

You can use this package in multithreading enviroment. All functions are thread
safe.

If one thread is calling *Query* method, other threads will be blocked if they
call *Query* or *Start* until *Query* return in first thread.

If one thread is calling *Start* method, other threads will be blocked if they
call *Query* or *Start* until all rows will be readed from a connection in first
thread.

## TODO

1. Prepared statements
2. Multiple results
3. More MySQL commands (if needed)
