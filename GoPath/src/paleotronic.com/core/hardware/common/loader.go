package common

import (
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/assets"
)

func LoadData(name string, start int, count int) ([]uint64, error) {

	data, e := assets.Asset(name)
	if e != nil {
		log.Printf("Error loading rom '%s': %s", name, e.Error())
		return []uint64(nil), e
	}

	if count == 0 {
		count = len(data) - start
	}

	rawdata := make([]uint64, count)
	for i := start; i < start+count; i++ {
		rawdata[i-start] = uint64(data[i])
	}

	return rawdata, nil

}
