package update

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"paleotronic.com/log"

	"github.com/ulikunitz/xz"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/files"
	"paleotronic.com/fmt"
)

const UPDATESERVER = "http://update.paleotronic.com:6522/"
const DOWNLOAD_TEMPFILE = "download.part"

var (
	CHANNEL                   = "stable"
	PLATFORM           string = runtime.GOOS + "/" + runtime.GOARCH
	PERCOL8_VERSION    string = "0.0.1alpha"
	VERSION_CHECK_URL  string = UPDATESERVER + CHANNEL + "/" + PLATFORM + "/current/version"
	CHECKSUM_CHECK_URL string = UPDATESERVER + CHANNEL + "/" + PLATFORM + "/current/checksum.sha256"
	APP_DOWNLOAD_URL   string = UPDATESERVER + CHANNEL + "/" + PLATFORM + "/current/package.xz"
	//
	appFullBinary       string = settings.GetBinaryFile()
	appFullPath         string = settings.GetBinaryPath()
	appDownloadTempfile string = appFullPath + string(os.PathSeparator) + DOWNLOAD_TEMPFILE
	appArchiveBinary    string = appFullBinary + "." + GetBuildNumber()
)

func SetupChannels() { // foo
	CHECKSUM_CHECK_URL = UPDATESERVER + CHANNEL + "/" + PLATFORM + "/current/checksum.sha256"
	VERSION_CHECK_URL = UPDATESERVER + CHANNEL + "/" + PLATFORM + "/current/version"
	APP_DOWNLOAD_URL = UPDATESERVER + CHANNEL + "/" + PLATFORM + "/current/package.xz"
}

func init() {
	SetupChannels()
}

func GetVersion() string {
	return PERCOL8_VERSION
}

func GetBuildNumber() string {
	return PERCOL8_BUILD
}

func GetBuildHash() string {
	return PERCOL8_GITHASH
}

func GetBuildDate() string {
	return PERCOL8_DATE
}

func GetHumanVersion() string {
	return fmt.Sprintf("Octalyzer version %s, build %s (built on %s (UTC) from git revision %s)",
		GetVersion(), GetBuildNumber(), GetBuildDate(), GetBuildHash(),
	)
}

func CheckVersion() string {

	if settings.SystemType == "nox" {
		return GetBuildNumber()
	}

	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	response, err := client.Get(VERSION_CHECK_URL)
	//fmt.RPrintln("Response", response.Status)

	if err != nil {
		//fmt.Printf("%s", err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			//fmt.Printf("%s", err)
		}
		return strings.Trim(string(contents), " \r\n")
	}
	return GetBuildNumber()
}

func GetChecksum() string {

	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	response, err := client.Get(CHECKSUM_CHECK_URL)
	//fmt.RPrintln("Response", response.Status)

	if err != nil {
		//fmt.Printf("%s", err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			//fmt.Printf("%s", err)
		}

		parts := strings.Fields(string(contents))

		return parts[0]
	}
	return ""
}

func DownloadVersion(txt *types.TextBuffer) (string, error) {

	if settings.SystemType == "nox" {
		return "Ok", nil
	}

	log.Printf("Current binary: %s", appFullBinary)
	log.Printf("Tempfile      : %s", appDownloadTempfile)
	log.Printf("Archive binary: %s", appArchiveBinary)

	fmt.Println("Downloading:", APP_DOWNLOAD_URL)

	//timeout := time.Duration(60 * time.Second)
	client := http.Client{
	//Timeout: timeout,
	}

	response, err := client.Get(APP_DOWNLOAD_URL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close() //

	f, err := os.Create(appDownloadTempfile)

	if err != nil {
		return "", err
	}

	written := 0
	size := int(response.ContentLength)

	var buf bytes.Buffer

	txt.ClearScreen()

	d := make(chan []byte)
	t := time.NewTicker(time.Second)

	go func() {
		data := make([]byte, 4096)
		count, err := response.Body.Read(data)
		for count > 0 {

			// got count bytes
			d <- data[0:count]
			if err != nil && err.Error() != "EOF" {
				return
			}
			written += count

			data = make([]byte, 4096)
			count, err = response.Body.Read(data)
			//fmt.RPrintf("count = %d, err = %v\n", count, err)
		}
	}()

	idle := 0
	for buf.Len() < size {
		select {
		case _ = <-t.C:
			idle++
			if idle > 60 {
				return "Failed: timeout", errors.New("Timeout exceeded")
			}
		case chunk := <-d:
			buf.Write(chunk)
			idle = 0 // reset timer
			ratio := float32(written) / float32(size)
			if ratio > 1 {
				ratio = 1
			}
			percent := int(ratio * 100)

			txt.GotoXY(0, 10)
			txt.FGColor = 15
			txt.BGColor = 0
			txt.Printf("Downloading update: %d%%", percent)

			txt.GotoXY(0, 12)
			num := int(ratio * 40)
			txt.FGColor = 6
			for i := 0; i < num; i++ {
				txt.Put(rune(1129))
			}
			txt.FGColor = 8
			for i := 0; i < (40 - num); i++ {
				txt.Put(rune(1129))
			}

			txt.FullRefresh()
		}
	}

	//txt.PutStr("\r\n")
	txt.GotoXY(0, 14)

	txt.FGColor = 15

	// filled buffer now unpack
	r, err := xz.NewReader(&buf)
	if err != nil {

		txt.ClearScreen()
		txt.PutStr("Decompression failed")
		time.Sleep(5 * time.Second)
		txt.ClearScreen()

		return "Failed", err
	}
	if _, err = io.Copy(f, r); err != nil {

		txt.ClearScreen()
		txt.PutStr("Decompression failed")
		time.Sleep(5 * time.Second)
		txt.ClearScreen()

		return "Failed", err
	}
	f.Close()

	// See if we can verify the checksum
	checksum := GetChecksum()
	if checksum != "" {
		txt.GotoXY(0, 14)
		txt.FGColor = 15
		txt.PutStr("Verifying checksum... ")
		chunk, _ := ioutil.ReadFile(appDownloadTempfile)
		tmp := sha256.Sum256(chunk)
		cksum := hex.EncodeToString(tmp[:])

		if cksum != checksum {
			txt.FGColor = 1
			txt.PutStr("BAD\r\n")
			txt.PutStr("Expected:\r\n" + checksum + "\r\n")
			txt.PutStr("Received:\r\n" + cksum + "\r\n")
			time.Sleep(5 * time.Second)
			return "Failed", nil
		}
		txt.FGColor = 4
		txt.PutStr("OK\r\n")
		time.Sleep(1 * time.Second)

	}

	txt.FGColor = 15
	txt.PutStr("Archiving old version... ")

	// rename existing
	err = os.Rename(appFullBinary, appArchiveBinary)
	if err != nil {
		log.Printf("Rename failed: %v", err)
		return "Failed", nil
	}
	txt.FGColor = 4
	txt.PutStr("OK!\r\n")

	txt.FGColor = 15
	txt.PutStr("Updating... ")

	err = os.Rename(appDownloadTempfile, appFullBinary)
	if err != nil {
		log.Printf("Rename of new failed: %v", err)
		return "Failed", nil
	}

	txt.FGColor = 4
	txt.PutStr("OK!\r\n")

	files.PurgeCache()

	os.Chmod(appFullBinary, 0755)

	cmd := exec.Command(appFullBinary)
	cmd.Args = settings.Args
	err = cmd.Start()
	if err != nil {
		log.Printf("Execution failed: %v", err)
	}
	time.Sleep(1 * time.Second)
	if err != nil {
		fmt.Println(err)
		return "Failed", nil
	}
	os.Exit(0)

	return "Ok", nil

}

func CheckAndDownload(txt *types.TextBuffer) {

	time.Sleep(500 * time.Millisecond)

	version := CheckVersion()
	current := GetBuildNumber()

	fmt.Printf("ONLINE VERSION = [%s]\n", version)
	fmt.Printf("LOCAL  VERSION = [%s]\n", current)

	if version > current {
		//fmt.Printf("A new build (%s) is available...\n", version)
		fn, _ := DownloadVersion(txt)
		txt.Printf("Download %s\r\n", fn)
	} else {
		//fmt.Printf("Current build is newest.\r\n")
	}

	time.Sleep(1000 * time.Millisecond)
}

func splitPath(filename string) (string, string) {
	parts := strings.Split(filename, string(os.PathSeparator))
	if len(parts) == 1 {
		return "", parts[0]
	}
	return strings.Join(parts[0:len(parts)-1], string(os.PathSeparator)), parts[len(parts)-1]
}

func CheckFilename() {
	full := os.Args[0]
	p, current := splitPath(full)

	var basename = "microm8"

	if settings.SystemType == "nox" {
		basename = "noxarchaist"
	}

	//fmt.RPrintf("arg0 = %s\n", current)
	if !strings.HasPrefix(strings.ToLower(current), basename) {
		appname := p + string(os.PathSeparator) + basename
		if runtime.GOOS == "windows" {
			appname += ".exe"
		}
		_ = os.Rename(full, appname)

		os.Chmod(appname, 0755)

		cmd := exec.Command(appname)
		//cmd.Args = settings.Args
		err := cmd.Start()
		time.Sleep(1 * time.Second)
		if err != nil {
			fmt.Println(err)
			return
		}
		os.Exit(0)
	}
}
