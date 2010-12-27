package mymy

import "log"

func (my *MySQL) sendCmd(cmd byte, argv ...interface{}) {
    // Reset sequence number
    my.seq = 0
    // Write command
    switch cmd {
    case COM_QUERY, COM_INIT_DB, COM_CREATE_DB, COM_DROP_DB, COM_STMT_PREPARE:
        my.writeHead(1 + lenBS(argv[0]))
        writeByte(my.wr, cmd)
        writeBS(my.wr, argv[0])
        flush(my.wr)

    // TODO: case COM_STMT_EXECUTE:

    case COM_STMT_SEND_LONG_DATA:
        my.writeHead(1 + 4 + 2 + lenBS(argv[2]))
        writeByte(my.wr, cmd)
        writeU32(my.wr, argv[0].(uint32))
        writeU16(my.wr, argv[1].(uint16))
        writeBS(my.wr, argv[2])

    case COM_QUIT, COM_STATISTICS, COM_PROCESS_INFO, COM_DEBUG, COM_PING:
        my.writeHead(1)
        writeByte(my.wr, cmd)

    case COM_FIELD_LIST:
        pay_len := 1 + lenBS(argv[0])+1
        if len(argv) > 1 {
            pay_len += lenBS(argv[1])
        }

        my.writeHead(pay_len)
        writeByte(my.wr, cmd)
        writeNT(my.wr, argv[0])
        if len(argv) > 1 {
            writeBS(my.wr, argv[1])
        }

    case COM_TABLE_DUMP:
        my.writeHead(1 + lenLC(argv[0]) + lenLC(argv[1]))
        writeByte(my.wr, cmd)
        writeLC(my.wr, argv[0])
        writeLC(my.wr, argv[1])

    case COM_REFRESH, COM_SHUTDOWN :
        my.writeHead(1 + 1)
        writeByte(my.wr, cmd)
        writeByte(my.wr, argv[0].(byte))

    case COM_STMT_FETCH:
        my.writeHead(1 + 4 + 4)
        writeByte(my.wr, cmd)
        writeU32(my.wr, argv[0].(uint32))
        writeU32(my.wr, argv[1].(uint32))

    case COM_PROCESS_KILL, COM_STMT_CLOSE, COM_STMT_RESET:
        my.writeHead(1 + 4)
        writeByte(my.wr, cmd)
        writeU32(my.wr, argv[0].(uint32))

    case COM_SET_OPTION:
        my.writeHead(1 + 2)
        writeByte(my.wr, cmd)
        writeU16(my.wr, argv[0].(uint16))

    case COM_CHANGE_USER:
        my.writeHead(1 + lenBS(argv[0])+1 + lenLC(argv[1]) + lenBS(argv[2])+1)
        writeByte(my.wr, cmd)
        writeNT(my.wr, argv[0]) // User name
        writeLC(my.wr, argv[1]) // Scrambled password
        writeNT(my.wr, argv[2]) // Database name
        //writeU16(my.wr, argv[3]) // Character set number (since 5.1.23?)

    case COM_BINLOG_DUMP:
        pay_len := 1 + 4 + 2 + 4
        if len(argv) > 3 {
            pay_len += lenBS(argv[3])
        }

        my.writeHead(pay_len)
        writeByte(my.wr, cmd)
        writeU32(my.wr, argv[0].(uint32)) // Start position
        writeU16(my.wr, argv[1].(uint16)) // Flags
        writeU32(my.wr, argv[2].(uint32)) // Slave server id
        if len(argv) > 3 {
            writeBS(my.wr, argv[3])
        }

    // TODO: case COM_REGISTER_SLAVE:

    default:
        panic("Unknown code for MySQL command")
    }
    // Send packet
    flush(my.wr)

    if my.Debug {
        log.Printf("[%d <-] Command packet: Cmd=0x%x", my.seq - 1, cmd) 
    }
}
