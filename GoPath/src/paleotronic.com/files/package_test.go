package files

import (
	"paleotronic.com/fmt"
	"testing"
	//    "paleotronic.com/fmt"
	"os"

	"bytes"

	"paleotronic.com/filerecord"
)

func TestPackagePackUnPack(t *testing.T) {

	zfr := &filerecord.FileRecord{
		Content: []byte(nil),
	}

	zip := NewZipProvider(zfr, "test.zip")

	nfr := &filerecord.FileRecord{
		FileName:    "foot.txt",
		Description: "foot things",
		Content:     []byte("A bunch of feet rained down from the sky in central Utah last week."),
	}

	zip.SetFileContent("", "foot.txt", nfr.Content)
	zip.SetFileContent("stuff/cribbage/junk", "otherjunk.txt", nfr.Content)

	for k, v := range zip.content {
		fmt.Printf("Pre-compression: %s -> %s/%s\n", k, v.FilePath, v.FileName)
	}

	fmt.Printf("Bytes of file data = %d\n", len(zip.source.Content))

	f, err := os.Create("test.zip")
	if err == nil {
		f.Write(zip.source.Content)
		f.Close()
	}

	if !zip.Exists("", "stuff") {
		t.Error("Directory 'stuff' should exist")
	}

	err = zip.readFiles()
	if err != nil {
		t.Error(err.Error())
	}

	gfr, err := zip.GetFileContent("stuff/cribbage/junk", "otherjunk.txt")
	if err != nil {
		t.Error(err.Error())
	}

	if bytes.Compare(gfr.Content, nfr.Content) != 0 {
		t.Error("Expected to get same content back")
	}

	for k, v := range zip.content {
		fmt.Printf("Extraction: %s -> %s/%s\n", k, v.FilePath, v.FileName)
	}

	err = zip.ChDir("stuff")
	if err != nil {
		t.Error("Failed to change dir: " + err.Error())
	}

	err = zip.ChDir("bagels")
	if err == nil {
		t.Error("Should not change dir to non-existent dir")
	}

	err = zip.ChDir("cribbage")
	if err != nil {
		t.Error("Failed to change dir: " + err.Error())
	}

	err = zip.ChDir("/")
	if err != nil {
		t.Error("Failed to change dir: " + err.Error())
	}

	dlist, flist, err := zip.DirFromBase("/", "*.txt")
	fmt.Println("DIRS", dlist)
	fmt.Println("FILES", flist)

}
