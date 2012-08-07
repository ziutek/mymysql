package native

import (
	"crypto/sha1"
)

// Borrowed from GoMySQL
// SHA1(SHA1(SHA1(password)), scramble) XOR SHA1(password)
func encryptedPasswd(password string, scramble []byte) (out []byte) {
	if len(password) == 0 {
		return
	}
	// stage1_hash = SHA1(password)
	// SHA1 encode
	crypt := sha1.New()
	crypt.Write([]byte(password))
	stg1Hash := crypt.Sum(nil)
	// token = SHA1(SHA1(stage1_hash), scramble) XOR stage1_hash
	// SHA1 encode again
	crypt.Reset()
	crypt.Write(stg1Hash)
	stg2Hash := crypt.Sum(nil)
	// SHA1 2nd hash and scramble
	crypt.Reset()
	crypt.Write(scramble)
	crypt.Write(stg2Hash)
	stg3Hash := crypt.Sum(nil)
	// XOR with first hash
	out = make([]byte, len(scramble))
	for ii := range scramble {
		out[ii] = stg3Hash[ii] ^ stg1Hash[ii]
	}
	return
}

// libmysql/password.c hash_password function translated to Go
func hash_password(password []byte) (result [2]uint32) {
	var nr, add, nr2, tmp uint32
	nr, add, nr2 = 1345345333, 7, 0x12345671

	for _, c := range password {
		if c == ' ' || c == '\t' {
			continue // skip space in password
		}

		tmp = uint32(c)
		nr ^= (((nr & 63) + add) * tmp) + (nr << 8)
		nr2 += (nr2 << 8) ^ nr
		add += tmp
	}

	result[0] = nr & ((1 << 31) - 1) // Don't use sign bit (str2int)
	result[1] = nr2 & ((1 << 31) - 1)
	return
}

func encryptedOldPassword(password string, scramble []byte) (out []byte) {
	if len(password) == 0 {
		return
	}
	return
}
