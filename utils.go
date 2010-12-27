package mymy

import (
    "log"
    "os"
)

func (my *MySQL) lock() {
    my.mutex.Lock()
}

func (my *MySQL) unlock() {
    my.mutex.Unlock()
}

func (my *MySQL) writeHead(pay_len int) {
    // Write header
    writeU24(my.wr, uint32(pay_len))
    writeByte(my.wr, my.seq)
    // Update sequence number
    my.seq++
}

/*func (my *MySQL) sendPkt(payload []byte) {
    my.writeHead(len(payload))
    write(my.wr, payload)
    flush(my.wr)
}*/

func (my *MySQL) init() {
    pr := newPktReader(my.rd, &my.seq)
    my.info.scramble = make([]byte, 20)

    my.info.prot_ver = readByte(pr)
    my.info.serv_ver = readNTS(pr)
    my.info.thr_id   = readU32(pr)
    readFull(pr, my.info.scramble[0:8])
    read(pr, 1)
    my.info.caps     = readU16(pr)
    my.info.lang     = readByte(pr)
    status          := readU16(pr)
    read(pr, 13)
    readFull(pr, my.info.scramble[8:])
    read(pr, 1)
    pr.checkEof()

    if my.Debug {
        log.Printf(
            "[%d ->] Init packet: ProtVer=%d, ServVer=\"%s\" Status=0x%x",
            my.seq - 1, my.info.prot_ver, my.info.serv_ver, status,
        )
    }
}

func (my *MySQL) auth() {
    pay_len := 4 + 4 + 1 + 23 + len(my.user)+1 + 1+len(my.info.scramble)
    flags := uint32(
        CLIENT_PROTOCOL_41 |
        CLIENT_LONG_PASSWORD |
        CLIENT_SECURE_CONN |
        CLIENT_TRANSACTIONS,
    )
    if len(my.dbname) > 0 {
        pay_len += len(my.dbname)+1
        flags |= CLIENT_CONNECT_WITH_DB
    }
    encr_passwd := my.encryptedPasswd()

    my.writeHead(pay_len)
    writeU32(my.wr, flags)
    writeU32(my.wr, uint32(1 << 24)) // Max packet size
    writeByte(my.wr, my.info.lang)   // Charset number
    write(my.wr, make([]byte, 23))   // Filler
    writeNTS(my.wr, my.user)         // Username
    writeNbin(my.wr, &encr_passwd)   // Encrypted password
    if len(my.dbname) > 0 {
        writeNTS(my.wr, my.dbname)
    }
    flush(my.wr)

    if my.Debug {
        log.Printf("[%d <-] Authentication packet", my.seq)
    }
    return
}

func (my *MySQL) unlockIfError(err *os.Error) {
    if *err != nil {
        my.unlock()
    }
}
