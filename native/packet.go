package native

import (
	"bufio"
	"errors"
	"github.com/ziutek/mymysql/mysql"
	"io"
)

type pktReader struct {
	rd     *bufio.Reader
	seq    *byte
	remain int
	last   bool
	buf    [8]byte
	ibuf   [3]byte
}

func (my *Conn) newPktReader() *pktReader {
	return &pktReader{rd: my.rd, seq: &my.seq}
}

func (pr *pktReader) readHeader() {
	if pr.last {
		// No more packets
		panic(io.EOF)
	}
	// Read next packet header
	buf := pr.ibuf[:]
	_, err := io.ReadFull(pr.rd, buf)
	if err != nil {
		panic(err)
	}
	pr.remain = int(DecodeU24(buf))
	seq, err := pr.rd.ReadByte()
	if err != nil {
		panic(err)
	}
	// Chceck sequence number
	if *pr.seq != seq {
		panic(mysql.ErrSeq)
	}
	*pr.seq++
	// Last packet?
	pr.last = (pr.remain != 0xffffff)
}

/*func (pr *pktReader) Read(buf []byte) (num int, err error) {
	if len(buf) == 0 {
		return 0, nil
	}
	defer catchError(&err)

	if pr.remain == 0 {
		pr.readHeader()
	}
	// Reading data
	if len(buf) <= pr.remain {
		num, err = pr.rd.Read(buf)
	} else {
		num, err = pr.rd.Read(buf[0:pr.remain])
	}
	pr.remain -= num
	return
}*/

func (pr *pktReader) readFull(buf []byte) {
	for {
		if len(buf) == 0 {
			return
		}
		if pr.remain == 0 {
			pr.readHeader()
		}
		var (
			n   int
			err error
		)
		if len(buf) <= pr.remain {
			n, err = pr.rd.Read(buf)
		} else {
			n, err = pr.rd.Read(buf[0:pr.remain])
		}
		pr.remain -= n
		if err != nil {
			panic(err)
		}
		buf = buf[n:]
	}
}

func (pr *pktReader) readByte() byte {
	if pr.remain == 0 {
		pr.readHeader()
	}
	b, err := pr.rd.ReadByte()
	if err != nil {
		panic(err)
	}
	pr.remain--
	return b
}

func (pr *pktReader) readAll() (buf []byte) {
	buf = make([]byte, pr.remain)
	nn := 0
	for {
		pr.readFull(buf[nn:])
		if pr.last {
			break
		}
		// There is next packet to read
		new_buf := make([]byte, len(buf)+pr.remain)
		copy(new_buf[nn:], buf)
		nn += len(buf)
		buf = new_buf
	}
	return
}

var skipBuf [4069]byte

func (pr *pktReader) skipAll() {
	for {
		n := len(skipBuf)
		if n > pr.remain {
			n = pr.remain
		}
		pr.readFull(skipBuf[:n])
		if pr.last {
			break
		}
	}
}

func (pr *pktReader) skipN(n int) {
	for n != 0 {
		m := len(skipBuf)
		if m > n {
			m = n
		}
		pr.readFull(skipBuf[:m])
		n -= n
	}
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
		panic(mysql.ErrPktLong)
	}
}

type pktWriter struct {
	wr       *bufio.Writer
	seq      *byte
	remain   int
	to_write int
	last     bool
	buf      [13]byte
	ibuf     [3]byte
}

func (my *Conn) newPktWriter(to_write int) *pktWriter {
	return &pktWriter{wr: my.wr, seq: &my.seq, to_write: to_write}
}

/*func writePktHeader(wr io.Writer, seq byte, pay_len int) {
    writeU24(wr, uint32(pay_len))
    writeByte(wr, seq)
}*/

func (pw *pktWriter) writeHeader(l int) {
	buf := pw.ibuf[:]
	EncodeU24(buf, uint32(l))
	if _, err := pw.wr.Write(buf); err != nil {
		panic(err)
	}
	if err := pw.wr.WriteByte(*pw.seq); err != nil {
		panic(err)
	}
	// Update sequence number
	*pw.seq++
}

func (pw *pktWriter) write(buf []byte) {
	if len(buf) == 0 {
		return
	}
	var nn int
	for len(buf) != 0 {
		if pw.remain == 0 {
			if pw.to_write == 0 {
				panic(errors.New("too many data for write as packet"))
			}
			if pw.to_write >= 0xffffff {
				pw.remain = 0xffffff
			} else {
				pw.remain = pw.to_write
				pw.last = true
			}
			pw.to_write -= pw.remain
			pw.writeHeader(pw.remain)
		}
		nn = len(buf)
		if nn > pw.remain {
			nn = pw.remain
		}
		var err error
		nn, err = pw.wr.Write(buf[0:nn])
		pw.remain -= nn
		if err != nil {
			panic(err)
		}
		buf = buf[nn:]
	}
	if pw.remain+pw.to_write == 0 {
		if !pw.last {
			// Write  header for empty packet
			pw.writeHeader(0)
		}
		// Flush bufio buffers
		if err := pw.wr.Flush(); err != nil {
			panic(err)
		}
	}
	return
}

func (pw *pktWriter) writeByte(b byte) {
	pw.buf[0] = b
	pw.write(pw.buf[:1])
}

type writer interface {
	write([]byte)
}
