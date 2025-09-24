package files

import "paleotronic.com/fmt"
import "paleotronic.com/core/settings"
import "log"

func GetNextAutosaveFilename(filename string) string {
	p := GetPath(filename)
	f := GetFilename(filename)
	ext := GetExt(f)
	base := f[:len(f)-len(ext)-1]

	i := 1
	name := fmt.Sprintf("%s_%d.%s", base, i, ext)
	for ExistsViaProvider(p, name) {
		i++
		name = fmt.Sprintf("%s_%d.%s", base, i, ext)
	}

	return p + "/" + name
}

func AutoSave(slotid int, data []byte) error {
	log.Printf("[autosave] called with target %s", settings.AutosaveFilename[slotid])
	if settings.AutosaveFilename[slotid] == "" {
		return nil
	}
	current := settings.AutosaveFilename[slotid]
	backup := GetNextAutosaveFilename(current)

	if ExistsViaProvider(GetPath(current), GetFilename(current)) {
		fp, err := ReadBytesViaProvider(GetPath(current), GetFilename(current))
		if err != nil {
			log.Printf("[autosave] error reading old file: %v", err)
			return err
		}
		err = WriteBytesViaProvider(GetPath(backup), GetFilename(backup), fp.Content)
		if err != nil {
			log.Printf("[autosave] error writing old file: %v", err)
			return err
		}
		log.Printf("[autosave] backed up %s to %s", current, backup)
	}

	log.Printf("[autosave] saving %d bytes to %s", len(data), current)
	return WriteBytesViaProvider(GetPath(current), GetFilename(current), data)
}
