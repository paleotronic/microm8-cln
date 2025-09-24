package vduproto

import (
	"errors"
	"paleotronic.com/core/types"
)

type AssetType byte

const (
	AT_Audio_WAV    AssetType = 1 + iota
	AT_Audio_OGG    AssetType = 1 + iota
	AT_Image_PNG    AssetType = 1 + iota
	AT_Image_PNG_43 AssetType = 1 + iota
	AT_Music_8T     AssetType = 1 + iota
)

type AssetQuery struct {
	MD5 [16]byte
}

func (aq *AssetQuery) MarshalBinary() ([]byte, error) {
	b := make([]byte, 17)
	b[0] = types.MtAssetQuery
	for i := 0; i < 16; i++ {
		b[1+i] = aq.MD5[i]
	}
	return b, nil
}

func (aq *AssetQuery) UnmarshalBinary(data []byte) error {

	if len(data) != 17 {
		return errors.New("Incorrect length")
	}

	if data[0] != types.MtAssetQuery {
		return errors.New("Incorrect type")
	}

	for i := 0; i < 16; i++ {
		aq.MD5[i] = data[1+i]
	}

	return nil
}

type AssetBlock struct {
	MD5      [16]byte
	Sequence uint16
	Total    uint16
	Data     []byte
}

func (ab *AssetBlock) MarshalBinary() ([]byte, error) {

	b := make([]byte, 17)
	b[0] = types.MtAssetBlock
	for i := 0; i < 16; i++ {
		b[1+i] = ab.MD5[i]
	}

	// Sequence no
	b = append(b, byte(ab.Sequence%256), byte(ab.Sequence/256))

	// Total
	b = append(b, byte(ab.Total%256), byte(ab.Total/256))

	// Data
	b = append(b, ab.Data...)

	return b, nil

}

func (ab *AssetBlock) UnmarshalBinary(data []byte) error {

	if len(data) < 21 {
		return errors.New("Incorrect length")
	}

	if data[0] != types.MtAssetBlock {
		return errors.New("Incorrect type")
	}

	// MD5
	for i := 0; i < 16; i++ {
		ab.MD5[i] = data[1+i]
	}

	// Sequence
	ab.Sequence = 256*uint16(data[18]) + uint16(data[17])

	// Total
	ab.Total = 256*uint16(data[20]) + uint16(data[19])

	// Data
	ab.Data = data[21:]

	return nil
}

// action
type AssetAction struct {
	MD5    [16]byte
	Action AssetType
}

func (aq *AssetAction) MarshalBinary() ([]byte, error) {
	b := make([]byte, 18)
	b[0] = types.MtAssetQuery
	for i := 0; i < 16; i++ {
		b[1+i] = aq.MD5[i]
	}
	b[17] = byte(aq.Action)
	return b, nil
}

func (aq *AssetAction) UnmarshalBinary(data []byte) error {

	if len(data) != 18 {
		return errors.New("Incorrect length")
	}

	if data[0] != types.MtAssetQuery {
		return errors.New("Incorrect type")
	}

	for i := 0; i < 16; i++ {
		aq.MD5[i] = data[1+i]
	}

	aq.Action = AssetType(data[17])

	return nil
}
