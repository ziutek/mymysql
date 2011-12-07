package native

import (
	"bytes"
	"crypto/sha1"
	"github.com/ziutek/mymysql/mysql"
	"io"
)

func DecodeU16(buf []byte) uint16 {
	return uint16(buf[1])<<8 | uint16(buf[0])
}
func readU16(rd io.Reader) uint16 {
	buf := make([]byte, 2)
	readFull(rd, buf)
	return DecodeU16(buf)
}

func DecodeU24(buf []byte) uint32 {
	return (uint32(buf[2])<<8|uint32(buf[1]))<<8 | uint32(buf[0])
}
func readU24(rd io.Reader) uint32 {
	buf := make([]byte, 3)
	readFull(rd, buf)
	return DecodeU24(buf)
}

func DecodeU32(buf []byte) uint32 {
	return ((uint32(buf[3])<<8|uint32(buf[2]))<<8|
		uint32(buf[1]))<<8 | uint32(buf[0])
}
func readU32(rd io.Reader) uint32 {
	buf := make([]byte, 4)
	readFull(rd, buf)
	return DecodeU32(buf)
}

func DecodeU64(buf []byte) (rv uint64) {
	for ii, vv := range buf {
		rv |= uint64(vv) << uint(ii*8)
	}
	return
}
func readU64(rd io.Reader) (rv uint64) {
	buf := make([]byte, 8)
	readFull(rd, buf)
	return DecodeU64(buf)
}

func EncodeU16(val uint16) []byte {
	return []byte{byte(val), byte(val >> 8)}
}
func writeU16(wr io.Writer, val uint16) {
	write(wr, EncodeU16(val))
}

func EncodeU24(val uint32) []byte {
	return []byte{byte(val), byte(val >> 8), byte(val >> 16)}
}
func writeU24(wr io.Writer, val uint32) {
	write(wr, EncodeU24(val))
}

func EncodeU32(val uint32) []byte {
	return []byte{byte(val), byte(val >> 8), byte(val >> 16), byte(val >> 24)}
}
func writeU32(wr io.Writer, val uint32) {
	write(wr, EncodeU32(val))
}

func EncodeU64(val uint64) []byte {
	buf := make([]byte, 8)
	for ii := range buf {
		buf[ii] = byte(val >> uint(ii*8))
	}
	return buf
}
func writeU64(wr io.Writer, val uint64) {
	write(wr, EncodeU64(val))
}

func readNu64(rd io.Reader) *uint64 {
	bb := readByte(rd)
	var val uint64
	switch bb {
	case 251:
		return nil

	case 252:
		val = uint64(readU16(rd))

	case 253:
		val = uint64(readU24(rd))

	case 254:
		val = readU64(rd)

	default:
		val = uint64(bb)
	}
	return &val
}

func readNotNullU64(rd io.Reader) (val uint64) {
	nu := readNu64(rd)
	if nu == nil {
		panic(UNEXP_NULL_LCB_ERROR)
	}
	return *nu
}

func writeLCB(wr io.Writer, val uint64) {
	switch {
	case val <= 250:
		writeByte(wr, byte(val))

	case val <= 0xffff:
		writeByte(wr, 252)
		writeU16(wr, uint16(val))

	case val <= 0xffffff:
		writeByte(wr, 253)
		writeU24(wr, uint32(val))

	default:
		writeByte(wr, 254)
		writeU64(wr, val)
	}
}

func lenLCB(val uint64) int {
	switch {
	case val <= 250:
		return 1

	case val <= 0xffff:
		return 3

	case val <= 0xffffff:
		return 4
	}
	return 9
}

func writeNu64(wr io.Writer, nu *uint64) {
	if nu == nil {
		writeByte(wr, 251)
	} else {
		writeLCB(wr, *nu)
	}
}

func lenNu64(nu *uint64) int {
	if nu == nil {
		return 1
	}
	return lenLCB(*nu)
}

func readNbin(rd io.Reader) *[]byte {
	if nu := readNu64(rd); nu != nil {
		buf := make([]byte, *nu)
		readFull(rd, buf)
		return &buf
	}
	return nil
}

func readNotNullBin(rd io.Reader) []byte {
	nbuf := readNbin(rd)
	if nbuf == nil {
		panic(UNEXP_NULL_LCS_ERROR)
	}
	return *nbuf
}

func writeNbin(wr io.Writer, nbuf *[]byte) {
	if nbuf == nil {
		writeByte(wr, 251)
		return
	}
	writeLCB(wr, uint64(len(*nbuf)))
	write(wr, *nbuf)
}

func lenNbin(nbuf *[]byte) int {
	if nbuf == nil {
		return 1
	}
	return lenLCB(uint64(len(*nbuf))) + len(*nbuf)
}

func readNstr(rd io.Reader) (nstr *string) {
	if nbuf := readNbin(rd); nbuf != nil {
		str := string(*nbuf)
		nstr = &str
	}
	return
}

func readNotNullStr(rd io.Reader) (str string) {
	buf := readNotNullBin(rd)
	str = string(buf)
	return
}

func writeNstr(wr io.Writer, nstr *string) {
	if nstr == nil {
		writeByte(wr, 251)
		return
	}
	writeLCB(wr, uint64(len(*nstr)))
	writeString(wr, *nstr)
}

func lenNstr(nstr *string) int {
	if nstr == nil {
		return 1
	}
	return lenLCB(uint64(len(*nstr))) + len(*nstr)
}

func writeLC(wr io.Writer, v interface{}) {
	switch val := v.(type) {
	case []byte:
		writeNbin(wr, &val)
	case string:
		writeNstr(wr, &val)
	case *[]byte:
		writeNbin(wr, val)
	case *string:
		writeNstr(wr, val)
	default:
		panic("Unknown data type for write as lenght coded string")
	}
}

func lenLC(v interface{}) int {
	switch val := v.(type) {
	case []byte:
		return lenNbin(&val)
	case string:
		return lenNstr(&val)
	case *[]byte:
		return lenNbin(val)
	case *string:
		return lenNstr(val)
	}
	panic("Unknown data type for write as lenght coded string")
}

func readNTB(rd io.Reader) (buf []byte) {
	bb := new(bytes.Buffer)
	for {
		ch := readByte(rd)
		if ch == 0 {
			return bb.Bytes()
		}
		bb.WriteByte(ch)
	}
	return
}

func writeNTB(wr io.Writer, buf []byte) {
	write(wr, buf)
	writeByte(wr, 0)
}

func readNTS(rd io.Reader) (str string) {
	buf := readNTB(rd)
	str = string(buf)
	return
}

func writeNTS(wr io.Writer, str string) {
	writeNTB(wr, []byte(str))
}

func writeNT(wr io.Writer, v interface{}) {
	switch val := v.(type) {
	case []byte:
		writeNTB(wr, val)
	case string:
		writeNTS(wr, val)
	default:
		panic("Unknown type for write as null terminated data")
	}
}

func readNtime(rd io.Reader) *mysql.Time {
	dlen := readByte(rd)
	switch dlen {
	case 251:
		// Null
		return nil
	case 0:
		// 00:00:00
		return new(mysql.Time)
	case 5, 8, 12:
		// Properly time length
	default:
		panic(WRONG_DATE_LEN_ERROR)
	}
	buf := make([]byte, dlen)
	readFull(rd, buf)
	tt := int64(0)
	switch dlen {
	case 12:
		// Nanosecond part
		tt += int64(DecodeU32(buf[8:]))
		fallthrough
	case 8:
		// HH:MM:SS part
		tt += int64(int(buf[5])*3600+int(buf[6])*60+int(buf[7])) * 1e9
		fallthrough
	case 5:
		// Day part
		tt += int64(DecodeU32(buf[1:5])) * (24 * 3600 * 1e9)
		fallthrough
	}
	if buf[0] != 0 {
		tt = -tt
	}
	return (*mysql.Time)(&tt)
}

func readNotNullTime(rd io.Reader) mysql.Time {
	tt := readNtime(rd)
	if tt == nil {
		panic(UNEXP_NULL_DATE_ERROR)
	}
	return *tt
}

func EncodeTime(tt *mysql.Time) []byte {
	if tt == nil {
		return []byte{251}
	}
	buf := make([]byte, 13)
	ti := int64(*tt)
	if ti < 0 {
		buf[1] = 1
		ti = -ti
	}
	if ns := uint32(ti % 1e9); ns != 0 {
		copy(buf[9:13], EncodeU32(ns)) // nanosecond
		buf[0] += 4
	}
	ti /= 1e9
	if hms := int(ti % (24 * 3600)); buf[0] != 0 || hms != 0 {
		buf[8] = byte(hms % 60) // second
		hms /= 60
		buf[7] = byte(hms % 60) // minute
		buf[6] = byte(hms / 60) // hour
		buf[0] += 3
	}
	if day := uint32(ti / (24 * 3600)); buf[0] != 0 || day != 0 {
		copy(buf[2:6], EncodeU32(day)) // day
		buf[0] += 4
	}
	buf[0]++ // For sign byte
	buf = buf[0 : buf[0]+1]
	return buf
}

func writeNtime(wr io.Writer, tt *mysql.Time) {
	write(wr, EncodeTime(tt))
}

func lenNtime(tt *mysql.Time) int {
	if tt == nil || *tt == 0 {
		return 1
	}
	ti := int64(*tt)
	if ti%1e9 != 0 {
		return 13
	}
	ti /= 1e9
	if ti%(24*3600) != 0 {
		return 9
	}
	return 6
}

func readNdatetime(rd io.Reader) *mysql.Datetime {
	dlen := readByte(rd)
	switch dlen {
	case 251:
		// Null
		return nil
	case 0:
		// 0000-00-00
		return new(mysql.Datetime)
	case 4, 7, 11:
		// Properly datetime length
	default:
		panic(WRONG_DATE_LEN_ERROR)
	}

	buf := make([]byte, dlen)
	readFull(rd, buf)
	var dt mysql.Datetime
	switch dlen {
	case 11:
		// 2006-01-02 15:04:05.001004005
		dt.Nanosec = DecodeU32(buf[7:])
		fallthrough
	case 7:
		// 2006-01-02 15:04:05
		dt.Hour = buf[4]
		dt.Minute = buf[5]
		dt.Second = buf[6]
		fallthrough
	case 4:
		// 2006-01-02
		dt.Year = int16(DecodeU16(buf[0:2]))
		dt.Month = buf[2]
		dt.Day = buf[3]
	}
	return &dt
}

func readNotNullDatetime(rd io.Reader) (dt *mysql.Datetime) {
	dt = readNdatetime(rd)
	if dt == nil {
		panic(UNEXP_NULL_DATE_ERROR)
	}
	return
}

func EncodeDatetime(dt *mysql.Datetime) []byte {
	if dt == nil {
		return []byte{251}
	}
	buf := make([]byte, 12)
	switch {
	case dt.Nanosec != 0:
		copy(buf[7:12], EncodeU32(dt.Nanosec))
		buf[0] += 4
		fallthrough

	case dt.Second != 0 || dt.Minute != 0 || dt.Hour != 0:
		buf[7] = dt.Second
		buf[6] = dt.Minute
		buf[5] = dt.Hour
		buf[0] += 3
		fallthrough

	case dt.Day != 0 || dt.Month != 0 || dt.Year != 0:
		buf[4] = dt.Day
		buf[3] = dt.Month
		copy(buf[1:3], EncodeU16(uint16(dt.Year)))
		buf[0] += 4
	}
	buf = buf[0 : buf[0]+1]
	return buf
}

func writeNdatetime(wr io.Writer, dt *mysql.Datetime) {
	write(wr, EncodeDatetime(dt))
}

func lenNdatetime(dt *mysql.Datetime) int {
	switch {
	case dt == nil:
		return 1
	case dt.Nanosec != 0:
		return 12
	case dt.Second != 0 || dt.Minute != 0 || dt.Hour != 0:
		return 8
	case dt.Day != 0 || dt.Month != 0 || dt.Year != 0:
		return 5
	}
	return 1
}

func readNdate(rd io.Reader) *mysql.Date {
	dt := readNdatetime(rd)
	if dt == nil {
		return nil
	}
	return &mysql.Date{Year: dt.Year, Month: dt.Month, Day: dt.Day}
}

func readNotNullDate(rd io.Reader) (dt *mysql.Date) {
	dt = readNdate(rd)
	if dt == nil {
		panic(UNEXP_NULL_DATE_ERROR)
	}
	return
}

func EncodeDate(dd *mysql.Date) []byte {
	return EncodeDatetime(mysql.DateToDatetime(dd))
}

func writeNdate(wr io.Writer, dd *mysql.Date) {
	write(wr, EncodeDate(dd))
}

func lenNdate(dd *mysql.Date) int {
	return lenNdatetime(mysql.DateToDatetime(dd))
}

// Borrowed from GoMySQL
// SHA1(SHA1(SHA1(password)), scramble) XOR SHA1(password)
func (my *Conn) encryptedPasswd() (out []byte) {
	// Convert password to byte array
	passbytes := []byte(my.passwd)
	// stage1_hash = SHA1(password)
	// SHA1 encode
	crypt := sha1.New()
	crypt.Write(passbytes)
	stg1Hash := crypt.Sum(nil)
	// token = SHA1(SHA1(stage1_hash), scramble) XOR stage1_hash
	// SHA1 encode again
	crypt.Reset()
	crypt.Write(stg1Hash)
	stg2Hash := crypt.Sum(nil)
	// SHA1 2nd hash and scramble
	crypt.Reset()
	crypt.Write(my.info.scramble)
	crypt.Write(stg2Hash)
	stg3Hash := crypt.Sum(nil)
	// XOR with first hash
	out = make([]byte, len(my.info.scramble))
	for ii := range my.info.scramble {
		out[ii] = stg3Hash[ii] ^ stg1Hash[ii]
	}
	return
}

func escapeString(txt string) string {
	var (
		esc string
		buf bytes.Buffer
	)
	last := 0
	for ii, bb := range txt {
		switch bb {
		case 0:
			esc = `\0`
		case '\n':
			esc = `\n`
		case '\r':
			esc = `\r`
		case '\\':
			esc = `\\`
		case '\'':
			esc = `\'`
		case '"':
			esc = `\"`
		case '\032':
			esc = `\Z`
		default:
			continue
		}
		io.WriteString(&buf, txt[last:ii])
		io.WriteString(&buf, esc)
		last = ii + 1
	}
	io.WriteString(&buf, txt[last:])
	return buf.String()
}

func escapeQuotes(txt string) string {
	var buf bytes.Buffer
	last := 0
	for ii, bb := range txt {
		if bb == '\'' {
			io.WriteString(&buf, txt[last:ii])
			io.WriteString(&buf, `''`)
			last = ii + 1
		}
	}
	io.WriteString(&buf, txt[last:])
	return buf.String()
}
