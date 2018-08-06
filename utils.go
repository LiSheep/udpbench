package udpbench

import (
	"encoding/binary"
	"time"
	"fmt"
)

/* encode 8 bits unsigned int */
func Encode8u(p []byte, c byte) []byte {
	p[0] = c
	return p[1:]
}

/* decode 8 bits unsigned int */
func Decode8u(p []byte, c *byte) []byte {
	*c = p[0]
	return p[1:]
}

/* encode 16 bits unsigned int (lsb) */
func Encode16u(p []byte, w uint16) []byte {
	binary.LittleEndian.PutUint16(p, w)
	return p[2:]
}

/* decode 16 bits unsigned int (lsb) */
func Decode16u(p []byte, w *uint16) []byte {
	*w = binary.LittleEndian.Uint16(p)
	return p[2:]
}

/* encode 32 bits unsigned int (lsb) */
func Encode32u(p []byte, l uint32) []byte {
	binary.LittleEndian.PutUint32(p, l)
	return p[4:]
}

/* decode 32 bits unsigned int (lsb) */
func Decode32u(p []byte, l *uint32) []byte {
	*l = binary.LittleEndian.Uint32(p)
	return p[4:]
}

func Check_package(data []byte, real_len int) (ok bool, id, ts, data_len uint32) {
	var key uint16
	data = Decode16u(data, &key)
	if key == 22 {
		ok = true
	} else {
		fmt.Println("check package: key fail")
		ok = false
		return
	}
	data = Decode32u(data, &id)
	data = Decode32u(data, &ts)
	data = Decode32u(data, &data_len)
	if (real_len - 14) != int(data_len) {
		fmt.Println("check package: head error")
		ok = false
		return
	}
	var i uint32
	for i = 0; i < data_len; i++ {
		if data[i] != 'a' {
			ok = false
			fmt.Println("check package: data error")
			break
		}
	}
	return
}

func Encode_package(id uint32, data []byte) []byte {
	// head: 22(u16) + id(u32) + ts(u32) + data_len(u32)
	snd_data := make([]byte, len(data) + 14)
	tmp := Encode16u(snd_data, 22)
	tmp = Encode32u(tmp, id)
	tmp = Encode32u(tmp, Iclock())
	tmp = Encode32u(tmp, uint32(len(data)))
	copy(tmp, data)
	return snd_data
}

func Iclock() uint32 {
	return uint32(time.Now().UnixNano() / int64(time.Millisecond))
}
