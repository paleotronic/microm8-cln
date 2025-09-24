package ducktape

import (
	"bytes"
	"compress/gzip"
	"encoding/ascii85"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log" //"paleotronic.com/fmt"
	"net"
	"paleotronic.com/encoding/ffpak" //"log"
	"strings"
	"time"
)

const A85 = false
const FF = true

type DuckTapeBundle struct {
	ID      string
	Payload []byte
	Binary  bool
	Channel string
	UUID    int64
}

func GZIPStream(in []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(in)
	w.Close()
	return b.Bytes()
}

func UnGZIPStream(in []byte) ([]byte, error) {
	b := bytes.NewBuffer(in)
	r, _ := gzip.NewReader(b)
	defer r.Close()
	out, e := ioutil.ReadAll(r)
	return out, e
}

func (d *DuckTapeBundle) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return errors.New("DuckTapeBundle: not enough data to unpack")
	}

	d.ID = string(data[0:3])
	if data[3] == byte('#') {
		d.Binary = true
	} else {
		d.Binary = false
	}

	data = data[4:]

	if d.Binary {

		if len(data) == 0 {
			d.Payload = make([]byte, 0)
			return nil
		}
		// otherwise try unpack it
		//tmp := make([]byte, len(data)) // at least this length
		var tmp []byte
		var err error
		var count int

		if FF {
			tmp = ffpak.FFUnpack(data)
		} else if A85 {
			// a85 here
			tmp = make([]byte, len(data))
			count, _, err = ascii85.Decode(tmp, data, true)
			tmp = tmp[:count]

		} else {
			tmp, err = base64.StdEncoding.DecodeString(string(data))
		}
		//////fmt.Printf("Decode: %s\n", string(data))
		//ndst, _, err := base64.Decode(tmp, data, true)
		d.Payload = tmp // just what was decoded successfully
		if err != nil {
			return errors.New("DuckTapeBundle: corrupt data found during unpack")
		}
		return nil
	} else {
		// Payload is a string
		d.Payload = data // copy as is
		return nil
	}
}

func (d DuckTapeBundle) MarshalBinaryUDP() ([]byte, error) {
	sbuffer := []byte(d.ID)

	if len(sbuffer) != 3 {
		return []byte(nil), errors.New("DuckTapeBundle: ID field must be a 3 character string got: " + d.ID)
	}

	if d.Binary {
		sbuffer = append(sbuffer, byte('#'))
		//chunk := make([]byte, ascii85.MaxEncodedLen(len(d.Payload)))

		var chunk []byte

		if FF {
			chunk = ffpak.FFPack(d.Payload)
		} else if A85 {
			chunk = make([]byte, ascii85.MaxEncodedLen(len(d.Payload)))
			count := ascii85.Encode(chunk, d.Payload)
			chunk = chunk[:count]
		} else {
			chunk = []byte(base64.StdEncoding.EncodeToString(d.Payload))
		}
		//////fmt.Printf("Encoded: %s\n", string(chunk))

		//c := ascii85.Encode(chunk, d.Payload)
		sbuffer = append(sbuffer, chunk...)
	} else {
		sbuffer = append(sbuffer, byte(' '))
		sbuffer = append(sbuffer, d.Payload...)
	}
	//sbuffer = append(sbuffer, byte(13), byte(10))
	return sbuffer, nil
}

func (d DuckTapeBundle) MarshalBinary() ([]byte, error) {
	sbuffer := []byte(d.ID)

	if len(sbuffer) != 3 {
		return []byte(nil), errors.New("DuckTapeBundle: ID field must be a 3 character string got: " + d.ID)
	}

	var chunk []byte

	if d.Binary {
		sbuffer = append(sbuffer, byte('#'))
		//chunk := make([]byte, ascii85.MaxEncodedLen(len(d.Payload)))

		if FF {
			chunk = ffpak.FFPack(d.Payload)
		} else if A85 {
			chunk = make([]byte, ascii85.MaxEncodedLen(len(d.Payload)))
			count := ascii85.Encode(chunk, d.Payload)
			chunk = chunk[:count]
		} else {
			chunk = []byte(base64.StdEncoding.EncodeToString(d.Payload))
		}
		//////fmt.Printf("Encoded: %s\n", string(chunk))

		//c := ascii85.Encode(chunk, d.Payload)
		sbuffer = append(sbuffer, chunk...)
	} else {
		sbuffer = append(sbuffer, byte(' '))
		sbuffer = append(sbuffer, d.Payload...)
	}
	sbuffer = append(sbuffer, byte(13), byte(10))
	return sbuffer, nil
}

func WriteLineWithTimeout(conn net.Conn, ms time.Duration, maxtries int, buffer []byte) ([]byte, error) {

	tries := 0
	//sz := 4096

	var err error
	var written int
	var total = len(buffer)
	//var count, countWOErr int
	for tries < maxtries && len(buffer) > 0 {
		// end := sz
		// if end > len(buffer) {
		// 	end = len(buffer)
		// }
		conn.SetWriteDeadline(time.Now().Add(ms))
		written, err = conn.Write(buffer)
		if written > 0 {
			// count += written
			// if err == nil {
			// 	countWOErr += written
			// }
			buffer = buffer[written:]
			//log.Printf("Written %d bytes total, %d without err, remaining %d", count, countWOErr, total-count)
		}
		if err != nil {
			log.Printf("Failed to write, retrying with %d remaining bytes of %d: %s", len(buffer), total, err)
			tries++
		} else {
			tries = 0
		}
	}

	if tries >= maxtries && len(buffer) > 0 {
		return buffer, fmt.Errorf("Failed after max tries exceeded: %s", err)
	}

	return buffer, err

}

// Utility function to read a line
func ReadLineWithTimeout(conn net.Conn, ms time.Duration, buffer []byte, sep []byte) ([]byte, []byte, error) {

	// rules... see if we have a line in the buffer supplied already
	// if not read more with timeout
	// if yes return line

	idx := bytes.Index(buffer, sep)

	if idx > -1 {
		// line in existing buffer
		str := buffer[0:idx]           // up to pos of sep
		buffer = buffer[idx+len(sep):] // chop off sep as well
		return str, buffer, nil        // OK
	} else {
		// no line in existing buffer - try to read more with timeout

		if ms == 0 {
			ms = 1000 // 1 seconds default
		}

		// establish our deadline for reading data
		//endtime := time.Now().Add(ms * time.Millisecond)

		tmp := make([]byte, 2048)
		//for time.Now().UnixNano() < endtime.UnixNano() {
		conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond)) // read timeout for this call so we can vary as needed
		bytesRead, err := conn.Read(tmp)
		//log.Printf("ReadLineWithTimeOut: read = %d, err = %v", bytesRead, err)
		if (err != nil) && (!strings.Contains(err.Error(), "timeout")) {
			// something bad happened pass it on
			return []byte(nil), buffer, err
		}
		if bytesRead > 0 {
			// got more data, let's munge it and try get a chunk of bytes again
			buffer = append(buffer, tmp[0:bytesRead]...)
			idx = bytes.Index(buffer, sep)
			if idx > -1 {
				// line in existing buffer
				str := buffer[0:idx]           // up to pos of sep
				buffer = buffer[idx+len(sep):] // chop off sep as well
				return str, buffer, nil
			}
		}

		//}
	}

	// remaining case - nothing read
	return []byte(nil), buffer, nil
}

func (b *DuckTapeBundle) String() string {
	//	return fmt.Sprintf("%s: [%s]", b.ID, string(b.Payload))
	return ""
}
