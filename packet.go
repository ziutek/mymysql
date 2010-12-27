package mymy

import (
    "os"
    "bufio"
)

type pktReader struct {
    rd     *bufio.Reader
    seq    *byte
    remain int
    last   bool
}

func newPktReader(rd *bufio.Reader, seq *byte) *pktReader {
    return &pktReader{rd: rd, seq: seq}
}

func (pr *pktReader) Read(buf []byte) (num int, err os.Error) {
    defer catchOsError(&err)

    if len(buf) == 0 {
        return 0, nil
    }
    if pr.remain == 0 {
        // No data to read from current packet
        if pr.last {
            // No more packets
            return 0, os.EOF
        }
        // Read next packet header
        pr.remain = int(readU24(pr.rd))
        seq      := readByte(pr.rd)
        // Chceck sequence number
        if *pr.seq != seq {
            return 0, SEQ_ERROR
        }
        *pr.seq++
        // Last packet?
        pr.last = (pr.remain != 0xffffff)
    }
    // Reading data
    if len(buf) <= pr.remain {
        num, err = pr.rd.Read(buf)
    } else {
        num, err = pr.rd.Read(buf[0:pr.remain])
    }
    pr.remain -= num
    return
}

func (pr *pktReader) readAll() (buf []byte) {
    buf = make([]byte, pr.remain)
    nn := 0
    for {
        readFull(pr, buf[nn:])
        if pr.last {
            break
        }
        // There is next packet to read
        new_buf := make([]byte, len(buf) + pr.remain)
        copy(new_buf[nn:], buf)
        nn += len(buf)
        buf = new_buf
    }
    return
}

func (pr *pktReader) unreadByte() {
    if err := pr.rd.UnreadByte(); err != nil {
        panic(err)
    }
    pr.remain++
}

func (pr *pktReader) eof() bool {
    return pr.remain == 0 && pr.last
}

func (pr *pktReader) checkEof() {
    if !pr.eof() {
        panic(PKT_LONG_ERROR)
    }
}
