package native

import (
	"github.com/ziutek/mymysql/mysql"
	"log"
)

func (my *Conn) init() {
	my.seq = 0 // Reset sequence number, mainly for reconnect
	if my.Debug {
		log.Printf("[%2d ->] Init packet:", my.seq)
	}
	pr := my.newPktReader()

	my.info.prot_ver = pr.readByte()
	my.info.serv_ver = pr.readNTB()
	my.info.thr_id = pr.readU32()
	pr.readFull(my.info.scramble[0:8])
	pr.skipN(1)
	my.info.caps = pr.readU16()
	my.info.lang = pr.readByte()
	my.status = mysql.ConnStatus(pr.readU16())
	pr.skipN(13)
	if my.info.caps&_CLIENT_PROTOCOL_41 != 0 {
		pr.readFull(my.info.scramble[8:])
	}
	pr.skipAll() // Skip other information
	if my.Debug {
		log.Printf(tab8s+"ProtVer=%d, ServVer=\"%s\" Status=0x%x",
			my.info.prot_ver, my.info.serv_ver, my.status,
		)
	}
	if my.info.caps&_CLIENT_PROTOCOL_41 == 0 {
		panic(mysql.ErrOldProtocol)
	}
}

// return scramble password for auth switch
func (my *Conn) auth() []byte {
	if my.Debug {
		log.Printf("[%2d <-] Authentication packet", my.seq)
	}
	flags := uint32(
		_CLIENT_PROTOCOL_41 |
			_CLIENT_LONG_PASSWORD |
			_CLIENT_LONG_FLAG |
			_CLIENT_TRANSACTIONS |
			_CLIENT_SECURE_CONN |
			_CLIENT_LOCAL_FILES |
			_CLIENT_MULTI_STATEMENTS |
			_CLIENT_MULTI_RESULTS)
	// Reset flags not supported by server
	flags &= uint32(my.info.caps) | 0xffff0000
	var scrPasswd []byte
	switch my.plugin {
	case "caching_sha2_password":
		flags |= _CLIENT_PLUGIN_AUTH
		scrPasswd = encryptedSHA256Passwd(my.passwd, my.info.scramble[:])
	case "mysql_old_password":
		my.oldPasswd()
		return nil
	default:
		// mysql_native_password by default
		scrPasswd = encryptedPasswd(my.passwd, my.info.scramble[:])
	}

	// encode length of the auth plugin data
	var authRespLEIBuf [9]byte
	authRespLEI := appendLengthEncodedInteger(authRespLEIBuf[:0], uint64(len(scrPasswd)))
	if len(authRespLEI) > 1 {
		// if the length can not be written in 1 byte, it must be written as a
		// length encoded integer
		flags |= _CLIENT_PLUGIN_AUTH_LENENC_CLIENT_DATA
	}

	pay_len := 4 + 4 + 1 + 23 + len(my.user) + 1 + len(authRespLEI) + len(scrPasswd) + 21 + 1

	if len(my.dbname) > 0 {
		pay_len += len(my.dbname) + 1
		flags |= _CLIENT_CONNECT_WITH_DB
	}
	pw := my.newPktWriter(pay_len)
	pw.writeU32(flags)
	pw.writeU32(uint32(my.max_pkt_size))
	pw.writeByte(my.info.lang)   // Charset number
	pw.writeZeros(23)            // Filler
	pw.writeNTB([]byte(my.user)) // Username
	pw.writeBin(scrPasswd)       // Encrypted password

	// write database name
	if len(my.dbname) > 0 {
		pw.writeNTB([]byte(my.dbname))
	}

	// write plugin name
	if my.plugin != "" {
		pw.writeNTB([]byte(my.plugin))
	}
	return scrPasswd
}

func (my *Conn) authResponse(scrPasswd []byte) {
	// Read Result Packet
	authData, newPlugin := my.getAuthResult()

	// handle auth plugin switch, if requested
	if newPlugin != "" {
		my.plugin = newPlugin
		my.auth()

		// Read Result Packet
		authData, newPlugin = my.getAuthResult()

		// Do not allow to change the auth plugin more than once
		if newPlugin != "" {
			return
		}
	}

	switch my.plugin {

	// https://insidemysql.com/preparing-your-community-connector-for-mysql-8-part-2-sha256/
	case "caching_sha2_password":
		switch len(authData) {
		case 0:
			return // auth successful
		case 1:
			switch authData[0] {
			case 3: // cachingSha2PasswordFastAuthSuccess
				my.getResult(nil, nil)

			case 4: // cachingSha2PasswordPerformFullAuthentication
				// write plain text auth packet
				my.writeAuthSwitchPacket([]byte(scrPasswd))
			}
		}
	}
	return
}

// http://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::AuthSwitchResponse
func (my *Conn) writeAuthSwitchPacket(scrPasswd []byte) {
	pw := my.newPktWriter(len(scrPasswd) + 1)
	pw.writeBin(scrPasswd) // Encrypted password
	return
}

func (my *Conn) oldPasswd() {
	if my.Debug {
		log.Printf("[%2d <-] Password packet", my.seq)
	}
	scrPasswd := encryptedOldPassword(my.passwd, my.info.scramble[:])
	pw := my.newPktWriter(len(scrPasswd) + 1)
	pw.write(scrPasswd)
	pw.writeByte(0)
}
