package native

//import "log"

// _COM_QUIT, _COM_STATISTICS, _COM_PROCESS_INFO, _COM_DEBUG, _COM_PING:
func (my *Conn) sendCmd(cmd byte) {
	my.seq = 0
	pw := my.newPktWriter(1)
	writeByte(pw, cmd)
}

// _COM_QUERY, _COM_INIT_DB, _COM_CREATE_DB, _COM_DROP_DB, _COM_STMT_PREPARE:
func (my *Conn) sendCmdStr(cmd byte, s string) {
	my.seq = 0
	pw := my.newPktWriter(1 + len(s))
	writeByte(pw, cmd)
	writeString(pw, s)
}

// _COM_PROCESS_KILL, _COM_STMT_CLOSE, _COM_STMT_RESET:
func (my *Conn) sendCmdU32(cmd byte, u uint32) {
	my.seq = 0
	pw := my.newPktWriter(1 + 4)
	writeByte(pw, cmd)
	writeU32(pw, u)
}

func (my *Conn) sendLongData(stmtid uint32, pnum uint16, data []byte) {
	my.seq = 0
	pw := my.newPktWriter(1 + 4 + 2 + len(data))
	writeByte(pw, _COM_STMT_SEND_LONG_DATA)
	writeU32(pw, stmtid) // Statement ID
	writeU16(pw, pnum)   // Parameter number
	write(pw, data)      // payload
}

/*func (my *Conn) sendCmd(cmd byte, argv ...interface{}) {
	// Reset sequence number
	my.seq = 0
	// Write command
	switch cmd {
	case _COM_QUERY, _COM_INIT_DB, _COM_CREATE_DB, _COM_DROP_DB,
		_COM_STMT_PREPARE:
		pw := my.newPktWriter(1 + lenBS(argv[0]))
		writeByte(pw, cmd)
		writeBS(pw, argv[0])

	case _COM_STMT_SEND_LONG_DATA:
		pw := my.newPktWriter(1 + 4 + 2 + lenBS(argv[2]))
		writeByte(pw, cmd)
		writeU32(pw, argv[0].(uint32)) // Statement ID
		writeU16(pw, argv[1].(uint16)) // Parameter number
		writeBS(pw, argv[2])           // payload

	case _COM_QUIT, _COM_STATISTICS, _COM_PROCESS_INFO, _COM_DEBUG, _COM_PING:
		pw := my.newPktWriter(1)
		writeByte(pw, cmd)

	case _COM_FIELD_LIST:
		pay_len := 1 + lenBS(argv[0]) + 1
		if len(argv) > 1 {
			pay_len += lenBS(argv[1])
		}

		pw := my.newPktWriter(pay_len)
		writeByte(pw, cmd)
		writeNT(pw, argv[0])
		if len(argv) > 1 {
			writeBS(pw, argv[1])
		}

	case _COM_TABLE_DUMP:
		pw := my.newPktWriter(1 + lenLC(argv[0]) + lenLC(argv[1]))
		writeByte(pw, cmd)
		writeLC(pw, argv[0])
		writeLC(pw, argv[1])

	case _COM_REFRESH, _COM_SHUTDOWN:
		pw := my.newPktWriter(1 + 1)
		writeByte(pw, cmd)
		writeByte(pw, argv[0].(byte))

	case _COM_STMT_FETCH:
		pw := my.newPktWriter(1 + 4 + 4)
		writeByte(pw, cmd)
		writeU32(pw, argv[0].(uint32))
		writeU32(pw, argv[1].(uint32))

	case _COM_PROCESS_KILL, _COM_STMT_CLOSE, _COM_STMT_RESET:
		pw := my.newPktWriter(1 + 4)
		writeByte(pw, cmd)
		writeU32(pw, argv[0].(uint32))

	case _COM_SET_OPTION:
		pw := my.newPktWriter(1 + 2)
		writeByte(pw, cmd)
		writeU16(pw, argv[0].(uint16))

	case _COM_CHANGE_USER:
		pw := my.newPktWriter(
			1 + lenBS(argv[0]) + 1 + lenLC(argv[1]) + lenBS(argv[2]) + 1,
		)
		writeByte(pw, cmd)
		writeNT(pw, argv[0]) // User name
		writeLC(pw, argv[1]) // Scrambled password
		writeNT(pw, argv[2]) // Database name
		//writeU16(pw, argv[3]) // Character set number (since 5.1.23?)

	case _COM_BINLOG_DUMP:
		pay_len := 1 + 4 + 2 + 4
		if len(argv) > 3 {
			pay_len += lenBS(argv[3])
		}

		pw := my.newPktWriter(pay_len)
		writeByte(pw, cmd)
		writeU32(pw, argv[0].(uint32)) // Start position
		writeU16(pw, argv[1].(uint16)) // Flags
		writeU32(pw, argv[2].(uint32)) // Slave server id
		if len(argv) > 3 {
			writeBS(pw, argv[3])
		}

	// TODO: case COM_REGISTER_SLAVE:

	default:
		panic("Unknown code for MySQL command")
	}

	if my.Debug {
		log.Printf("[%2d <-] Command packet: Cmd=0x%x", my.seq-1, cmd)
	}
}*/
