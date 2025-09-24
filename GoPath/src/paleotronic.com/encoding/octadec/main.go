package octadec
import "paleotronic.com/fmt"
//import "os"
//import "io/ioutil"

const DICTBITS = 10
const DICTMAX = 1 << DICTBITS
 
// Compress a string to a list of output symbols.
func Compress(uncompressed []byte) []uint {
    // Build the dictionary.
    dictSize := 256
    dictionary := make(map[string]uint)
    usedcount  := make(map[string]int)
    for i := 0; i < 256; i++ {
        dictionary[string(i)] = uint(i)
    }
 
    w := ""
    result := make([]uint, 0)
    for _, c := range uncompressed {
        wc := w + string(c)
        if _, ok := dictionary[wc]; ok {
            w = wc
            usedcount[wc] += 1
        } else {
            result = append(result, dictionary[w])
            // Add wc to the dictionary.
            if dictSize < DICTMAX {
				dictionary[wc] = uint(dictSize)
				usedcount[wc] = 1
				dictSize++
			}
            w = string(c)
        }
    }
 
    // Output the code for w.
    if w != "" {
        result = append(result, dictionary[w])
    }
    
    //~ fmt.Printf("Resulting dict has %d entries...\n", len(dictionary))
    
    //~ for k, count := range usedcount {
		//~ if count > 2 {
			//~ fmt.Printf("%v, %d\n", []byte(k), count)
		//~ }
	//~ }
    
    return result
}
 
// Decompress a list of output ks to a string.
func Decompress(compressed []uint) []byte {
	
	if len(compressed) == 0 {
		return []byte(nil)
	}
	
    // Build the dictionary.
    dictSize := 256
    dictionary := make([][]byte, 256)
    for i := 0; i < 256; i++ {
        dictionary[uint(i)] = []byte{byte(i)}
    }
    w := []byte{byte(compressed[0])}
    result := w
    for _, k := range compressed[1:] {
        var entry []byte
        if int(k) < dictSize {
            entry = dictionary[k]
        } else if k == uint(dictSize) {
            entry = append( w, w[:1]... )
        } else {
            panic(fmt.Sprintf("Bad compressed k: %d", k))
        }
 
        result = append( result, entry... )
 
        // Add w+entry[0] to the dictionary.
        dictionary = append(dictionary, append( w, entry[:1]... ))
        //dictionary[uint(dictSize)] = append( w, entry[:1]... )
        dictSize++
 
        w = entry
    }
    
    return result
}
 
//~ func main() {
	
	//~ f, e := os.Open("../../octalyzer/out-pooyan.raw")
	//~ if e != nil {
		//~ panic(e)
	//~ }
	//~ data, e := ioutil.ReadAll(f)
	//~ f.Close()
	
	//~ data = data[30000:30345]
	
	//~ fmt.Printf("Input data length = %d bytes\n", len(data))
	
    //~ compressed := compress(data)
    
    //~ fmt.Printf("Compressed data length = %d bytes\n", (len(compressed) * DICTBITS) / 8)
    
    //~ fmt.Println(compressed)
    //~ decompressed := decompress(compressed)
    //~ fmt.Println("d =", len(decompressed))
//~ }
