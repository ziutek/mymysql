package mymy

import (
    "io"
    "crypto/sha1"
    "bytes"
)

func readU16(rd io.Reader) uint16 {
    buf := make([]byte, 2)
    readFull(rd, buf)
    return uint16(buf[1]) << 8 | uint16(buf[0])
}

func readU24(rd io.Reader) uint32 {
    buf := make([]byte, 3)
    readFull(rd, buf)
    return (uint32(buf[2]) << 8 | uint32(buf[1])) << 8 | uint32(buf[0])
}

func readU32(rd io.Reader) uint32 {
    buf := make([]byte, 4)
    readFull(rd, buf)
    return ((uint32(buf[3]) << 8 | uint32(buf[2])) << 8 |
        uint32(buf[1])) << 8 | uint32(buf[0])
}

func readU64(rd io.Reader) (rv uint64) {
    buf := make([]byte, 8)
    readFull(rd, buf)
    for ii, vv := range buf {
        rv |= uint64(vv) << uint(ii * 8)
    }
    return
}

func writeU16(wr io.Writer, val uint16) {
    write(wr, []byte{byte(val), byte(val >> 8)})
}

func writeU24(wr io.Writer, val uint32) {
    write(wr, []byte{byte(val), byte(val >> 8), byte(val >> 16)})
}

func writeU32(wr io.Writer, val uint32) {
    write(wr,
        []byte{byte(val), byte(val >> 8), byte(val >> 16), byte(val >> 24)},
    )
}

func writeU64(wr io.Writer, val uint64) {
    buf := make([]byte, 8)
    for ii := range buf {
        buf[ii] = byte(val >> uint(ii * 8))
    }
    write(wr, buf)
}

type Nu64 *uint64

func readNu64(rd io.Reader) Nu64 {
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

func readNotNullNu64(rd io.Reader) (val uint64) {
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

func writeNu64(wr io.Writer, nu Nu64) {
    if nu == nil {
        writeByte(wr, 251)
    } else {
        writeLCB(wr, *nu)
    }
}

func lenNu64(nu Nu64) int {
    if nu == nil {
        return 1
    }
    return lenLCB(*nu)
}

type Nbin *[]byte

func readNbin(rd io.Reader) Nbin {
    if nu := readNu64(rd); nu != nil {
        buf := make([]byte, *nu)
        readFull(rd, buf)
        return &buf
    }
    return nil
}

func readNotNullNbin(rd io.Reader) []byte {
    nbuf := readNbin(rd)
    if nbuf == nil {
        panic(UNEXP_NULL_LCS_ERROR)
    }
    return *nbuf
}

func writeNbin(wr io.Writer, nbuf Nbin) {
    if nbuf == nil {
        writeByte(wr, 251)
        return
    }
    writeLCB(wr, uint64(len(*nbuf)))
    write(wr, *nbuf)
}

func lenNbin(nbuf Nbin) int {
    if nbuf == nil {
        return 1
    }
    return lenLCB(uint64(len(*nbuf))) + len(*nbuf)
}

type Nstr *string

func readNstr(rd io.Reader) (nstr Nstr) {
    if nbuf := readNbin(rd); nbuf != nil {
        str := string(*nbuf)
        nstr = &str
    }
    return
}

func readNotNullNstr(rd io.Reader) (str string) {
    buf := readNotNullNbin(rd)
    str = string(buf)
    return
}

func writeNstr(wr io.Writer, nstr Nstr) {
    if nstr == nil {
        writeByte(wr, 251)
        return
    }
    writeLCB(wr, uint64(len(*nstr)))
    writeString(wr, *nstr)
}

func lenNstr(nstr Nstr) int {
    if nstr == nil {
        return 1
    }
    return lenLCB(uint64(len(*nstr))) + len(*nstr)
}

func writeLC(wr io.Writer, v interface{}) {
    switch val := v.(type) {
    case Nbin:   writeNbin(wr, val)
    case []byte: writeNbin(wr, &val)
    case Nstr:   writeNstr(wr, val)
    case string: writeNstr(wr, &val)
    default: panic("Unknown data type for write as lenght coded string")
    }
}

func lenLC(v interface{}) int {
    switch val := v.(type) {
    case Nbin:   return lenNbin(val)
    case []byte: return lenNbin(&val)
    case Nstr:   return lenNstr(val)
    case string: return lenNstr(&val)
    }
    panic("Unknown data type for write as lenght coded string")
}

func readNTB(rd io.Reader) (buf []byte) {
    bb := new(bytes.Buffer)
    ch := make([]byte, 1)
    for {
        if _, err := rd.Read(ch); err != nil {
            panic(err)
        }
        if ch[0] == 0 {
            buf = bb.Bytes()
            break
        }
        bb.WriteByte(ch[0])
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
    case []byte: writeNTB(wr, val)
    case string: writeNTS(wr, val)
    default: panic("Unknown type for write as null terminated data")
    }
}

// Borrowed from GoMySQL
// SHA1(SHA1(SHA1(password)), scramble) XOR SHA1(password)
func (my *MySQL) encryptedPasswd() (out []byte) {
    // Convert password to byte array
    passbytes := []byte(my.passwd)
    // stage1_hash = SHA1(password)
    // SHA1 encode
    crypt := sha1.New()
    crypt.Write(passbytes)
    stg1Hash := crypt.Sum()
    // token = SHA1(SHA1(stage1_hash), scramble) XOR stage1_hash
    // SHA1 encode again
    crypt.Reset()
    crypt.Write(stg1Hash)
    stg2Hash := crypt.Sum()
    // SHA1 2nd hash and scramble
    crypt.Reset()
    crypt.Write(my.info.scramble)
    crypt.Write(stg2Hash)
    stg3Hash := crypt.Sum()
    // XOR with first hash
    out = make([]byte, len(my.info.scramble))
    for ii := range my.info.scramble {
        out[ii] = stg3Hash[ii] ^ stg1Hash[ii]
    }
    return
}
