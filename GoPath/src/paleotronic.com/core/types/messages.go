package types

type MessageType byte

const (
	MtScreenPositionEvent byte = 1 + iota
	MtKeyPressEvent       byte = 1 + iota
	MtVideoColor          byte = 1 + iota
	MtThinScreen          byte = 1 + iota
	MtScreenMemoryEvent   byte = 1 + iota
	MtVideoMode           byte = 1 + iota
	MtEmptyClientRequest  byte = 1 + iota
	MtStringOutEvent      byte = 1 + iota
	MtPaddleButtonEvent   byte = 1 + iota
	MtPaddleValueEvent    byte = 1 + iota
	MtPaddleModifyEvent   byte = 1 + iota
	MtSpeakerClick        byte = 1 + iota
	MtScanLineEvent       byte = 1 + iota
	MtMemoryEvent         byte = 1 + iota
	MtCallEvent           byte = 1 + iota
	MtBGColorEvent        byte = 1 + iota
	MtAssetQuery          byte = 1 + iota
	MtAssetRequest        byte = 1 + iota
	MtAssetBlock          byte = 1 + iota
	MtRestalgiaCommand    byte = 1 + iota
	MtConnectCommand      byte = 1 + iota
	MtSwitchControls      byte = 1 + iota
	MtLayerSpec           byte = 1 + iota
	MtLayerBundle         byte = 1 + iota
)

const (
	RestalgiaRedefineInstrument byte = 1 + iota
	RestalgiaPlayNoteStream     byte = 1 + iota
	RestalgiaPlaySongFile       byte = 1 + iota
	RestalgiaSoundEffect        byte = 1 + iota
)
