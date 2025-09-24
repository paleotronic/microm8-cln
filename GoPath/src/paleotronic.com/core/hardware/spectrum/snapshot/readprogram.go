package snapshot

import "paleotronic.com/files"

func ReadProgram(filename string) (*Z80, error) {
	fp, err := files.ReadBytesViaProvider(files.GetPath(filename), files.GetFilename(filename))
	if err != nil {
		return nil, err
	}
	state := NewZ80FromData(fp.Content)
	return state, state.Load()
}
