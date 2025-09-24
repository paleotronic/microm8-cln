package crypt

func YogoCrypt( input []byte, key []byte, shiftiness int ) []byte{

	 // XOR cipher, but key repeats are bit shifted by amount cycle through
     
     output := make( []byte, len(input) )
     
     var shKeyByte byte
     var shiftCount int
	 for i, inByte := range input {
     	 shiftCount = (i+1) % shiftiness
     	 shKeyByte = key[i % len(key)]
        // shKeyByte  = rawKeyByte
         for z:=0; z<shiftCount; z++ {
         	 shKeyByte = (shKeyByte << 1) | ((shKeyByte & 128) >> 7)
         }
         output[i] = (inByte ^ shKeyByte)
     }
     
     return output
}

