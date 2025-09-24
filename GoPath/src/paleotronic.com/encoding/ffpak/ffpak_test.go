package ffpak

import "testing"
import "encoding/base64"


func BenchmarkBase64(b *testing.B) {

	str := "this is an input string\r\nand this is line 2"

	for i := 0; i < b.N; i ++ {
		es := base64.StdEncoding.EncodeToString( []byte(str) )
		
		_, _ = base64.StdEncoding.DecodeString( es )
	}

}


func BenchmarkFFPack(b *testing.B) {

	str := "this is an input string\r\nand this is line 2"

	for i := 0; i < b.N; i ++ {
		es := FFPack( []byte(str) )
		
		_ = FFUnpack( es )
	}

}
