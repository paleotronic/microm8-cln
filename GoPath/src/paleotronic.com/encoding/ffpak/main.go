package ffpak

func FFPack( in []byte ) []byte {

	out := make([]byte, 0)
	for _, b := range in {
		if b == 0xff || b < 32 {
			out = append(out, 0xff)
			out = append(out, b^0x80)
		} else {
			out = append(out, b)
		}
	}
	
	return out
}

func FFUnpack( in []byte ) []byte {
	out := make([]byte, len(in))
	i := 0
	oi := 0
	for i < len(in) {
		b := in[i]
		if b != 0xff {
			out[oi] = b
			oi++
			i++ 
		} else {
			if i == len(in)-1 {
				// last byte
				out[oi] = b
				oi++
				i++
			} else {
				i++ 
				out[oi] = in[i] ^ 0x80
				oi++
				i++
			}
		}
	}
	
	return out[:oi]
}

