package buffer

import "testing"


func BenchmarkChannel(b *testing.B) {

	str := "this is an input string\r\nand this is line 2"

	ch := make(chan string, 10)

	for i := 0; i < b.N; i ++ {

		ch <- str
		_ = <- ch

	}

}


func BenchmarkRingBuffer(b *testing.B) {

	str := "this is an input string\r\nand this is line 2"

	rb := NewRingBuffer( 10, false )

	for i := 0; i < b.N; i ++ {
		
		_ = rb.Push(str)
		_, _ = rb.Pop()
		
	}

}
