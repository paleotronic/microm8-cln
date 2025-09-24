package s8webclient

import (
	"crypto/md5"
	"encoding/json"
	"errors" //	"paleotronic.com/fmt"
	"strings"
	"time"

	"paleotronic.com/ducktape/client"
	"paleotronic.com/filerecord"
	"paleotronic.com/log"
)

// CacheCustomFile sends a request for updates to an existing file
func (c *Client) CacheCustomFile(req string, additional string, filepath string, filename string, data []byte) (filerecord.FileRecord, error) {

	fullpath := filepath + "/" + filename

	if len(fullpath) > 1 && rune(fullpath[0]) == '/' {
		fullpath = fullpath[1:]
	}

	var err error

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return filerecord.FileRecord{}, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	md5 := md5.Sum(data)

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, md5[0:]...)
	r = append(r, []byte(fullpath)...)
	if additional != "" {
		r = append(r, []byte(":"+additional)...)
	}
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	bb := &filerecord.FileRecord{}
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in CacheCustomFile() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "COK" {
			err = nil
			bb.UnJSON(msg.Payload)
			bb.Content = data
			log.Printf("(ok) file data: %+v", bb)
		} else if msg.ID == "FIL" {
			// Login OK
			err = nil
			bb.UnJSON(msg.Payload)
			// this is JSON, including meta data
			log.Printf("(not ok) file data: %+v", bb)
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return *bb, err
}

func (c *Client) ValidateCacheCustomFile(req string, additional string, current *filerecord.FileRecord) (*filerecord.FileRecord, error) {

	fullpath := current.FilePath + "/" + current.FileName

	if len(fullpath) > 1 && rune(fullpath[0]) == '/' {
		fullpath = fullpath[1:]
	}

	var err error

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return &filerecord.FileRecord{}, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, current.Checksum[0:]...)
	r = append(r, []byte(fullpath)...)
	if additional != "" {
		r = append(r, []byte(":"+additional)...)
	}
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	bb := &filerecord.FileRecord{}
	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in CacheCustomFile() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "COK" {
			err = nil
			bb = current
		} else if msg.ID == "FIL" {
			// Login OK
			err = nil
			bb.UnJSON(msg.Payload)
			log.Printf("size = %d", len(bb.Content))
			// this is JSON, including meta data

		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return bb, err
}

// ExistsCustomFile returns true if a file exists
func (c *Client) ExistsCustomFile(req string, additional string, filepath string, filename string) (bool, error) {

	fullpath := strings.Trim(filepath, "/") + "/" + filename

	//log2.Printf("NETWORK EXISTS CHECK FOR %s", fullpath)

	if len(fullpath) > 1 && rune(fullpath[0]) == '/' {
		fullpath = fullpath[1:]
	}

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return false, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	md5 := md5.Sum([]byte("frog"))

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, md5[0:]...)
	r = append(r, []byte(fullpath)...)
	if additional != "" {
		r = append(r, []byte(":"+additional)...)
	}
	c.c.SendMessage(req, r, true)

	log.Printf("%s: %v\n", req, string(r))

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		return false, errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in ExistsCustomFile() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "EOK" {
			return true, nil
		} else if msg.ID == "ERR" {
			return false, nil
		}
	}

	return false, nil
}

// LockCustomFile attempts to lock a file for update
func (c *Client) LockCustomFile(req string, additional string, filepath string, filename string) error {

	fullpath := filepath + "/" + filename

	if len(fullpath) > 1 && rune(fullpath[0]) == '/' {
		fullpath = fullpath[1:]
	}

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	md5 := md5.Sum([]byte("frog"))

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, md5[0:]...)
	r = append(r, []byte(fullpath)...)
	if additional != "" {
		r = append(r, []byte(":"+additional)...)
	}
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		return errors.New("timeout")
	case msg := <-c.c.Incoming:
		//			//fmt.Printf("in ExistsCustomFile() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "LOK" {
			return nil
		} else if msg.ID == "ERR" {
			return errors.New("LOCK FAILED")
		}
	}

	return nil
}

// LockUserFile locks a user space file
func (c *Client) LockUserFile(filepath string, filename string) error {

	return c.LockCustomFile("ULF", "", filepath, filename)

}

// LockSystemFile locks a system file
func (c *Client) LockSystemFile(filepath string, filename string) error {

	return c.LockCustomFile("SLF", "", filepath, filename)

}

// LockLegacyFile locks a file under the software space
func (c *Client) LockLegacyFile(filepath string, filename string) error {

	return c.LockCustomFile("LLF", "", filepath, filename)

}

// LockProjectFile locks a file under a project space
func (c *Client) LockProjectFile(project string, filepath string, filename string) error {

	return c.LockCustomFile("PLF", project, filepath, filename)

}

// LockRemIntFile locks a file owned by a remote interpreter (deprecated)
func (c *Client) LockRemIntFile(project string, filepath string, filename string) error {

	return c.LockCustomFile("ILF", project, filepath, filename)

}

//

// CacheUserFile puts a user file in the network cache
func (c *Client) CacheUserFile(filepath string, filename string, data []byte) (filerecord.FileRecord, error) {

	return c.CacheCustomFile("UCF", "", filepath, filename, data)

}

// CacheSystemFile puts a system file in the network cache
func (c *Client) CacheSystemFile(filepath string, filename string, data []byte) (filerecord.FileRecord, error) {

	return c.CacheCustomFile("SCF", "", filepath, filename, data)

}

// CacheLegacyFile puts a software file in the network cache
func (c *Client) CacheLegacyFile(filepath string, filename string, data []byte) (filerecord.FileRecord, error) {

	return c.CacheCustomFile("LCF", "", filepath, filename, data)

}

// CacheProjectFile puts a project file in the network cache
func (c *Client) CacheProjectFile(project string, filepath string, filename string, data []byte) (filerecord.FileRecord, error) {

	return c.CacheCustomFile("PCF", project, filepath, filename, data)

}

// CacheRemIntFile puts remote file in the network cache
func (c *Client) CacheRemIntFile(project string, filepath string, filename string, data []byte) (filerecord.FileRecord, error) {

	return c.CacheCustomFile("ICF", project, filepath, filename, data)

}

// ExistsProjectFile returns true if the file exists
func (c *Client) ExistsProjectFile(project string, filepath string, filename string) (bool, error) {

	return c.ExistsCustomFile("PEF", project, filepath, filename)

}

// ExistsRemIntFile returns true if a file exists
func (c *Client) ExistsRemIntFile(project string, filepath string, filename string) (bool, error) {

	return c.ExistsCustomFile("IEF", project, filepath, filename)

}

// ExistsSystemFile returns true if a file exists
func (c *Client) ExistsSystemFile(filepath string, filename string) (bool, error) {

	return c.ExistsCustomFile("SEF", "", filepath, filename)

}

// ExistsLegacyFile returns true if a file exists
func (c *Client) ExistsLegacyFile(filepath string, filename string) (bool, error) {

	return c.ExistsCustomFile("LEF", "", filepath, filename)

}

// ExistsUserFile returns true if a file exists
func (c *Client) ExistsUserFile(filepath string, filename string) (bool, error) {

	return c.ExistsCustomFile("UEF", "", filepath, filename)

}

// CreateCustomDir creates a directory
func (c *Client) CreateCustomDir(req string, additional string, filepath string, filename string) error {

	fullpath := filepath + "/" + filename

	if len(fullpath) > 1 && rune(fullpath[0]) == '/' {
		fullpath = fullpath[1:]
	}

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, []byte(fullpath)...)
	r = append(r, []byte(":"+additional)...)

	log.Printf("Send %s: [%s]\n", req, string(r))

	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		return errors.New("timeout")
	case msg := <-c.c.Incoming:
		//			//fmt.Printf("in ExistsCustomFile() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "SOK" {
			return nil
		} else if msg.ID == "ERR" {
			return errors.New("MKDIR FAILED")
		}
	}

	return nil
}

// CreateUserDir creates a directory
func (c *Client) CreateUserDir(filepath string, filename string) error {

	return c.CreateCustomDir("UCD", "", filepath, filename)

}

// CreateSystemDir creates a directory
func (c *Client) CreateSystemDir(filepath string, filename string) error {

	return c.CreateCustomDir("SCD", "", filepath, filename)

}

// CreateLegacyDir creates a directory
func (c *Client) CreateLegacyDir(filepath string, filename string) error {

	return c.CreateCustomDir("LCD", "", filepath, filename)

}

// CreateAProjectDir creates a project directory
func (c *Client) CreateAProjectDir(project string, filepath string, filename string) error {

	return c.CreateCustomDir("PCD", project, filepath, filename)

}

// DeleteCustomFile deletes a file
func (c *Client) DeleteCustomFile(req string, additional string, filepath string, filename string) error {

	fullpath := filepath + "/" + filename

	if len(fullpath) > 1 && rune(fullpath[0]) == '/' {
		fullpath = fullpath[1:]
	}

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, []byte(fullpath)...)
	r = append(r, []byte(":"+additional)...)

	log.Printf("Send %s: [%s]\n", req, string(r))

	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		return errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in DeleteCustomFile() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "DOK" {
			return nil
		} else if msg.ID == "ERR" {
			return errors.New("DELETE FAILED")
		}
	}

	return nil
}

// DeleteUserFile removes a user file
func (c *Client) DeleteUserFile(filepath string, filename string) error {

	return c.DeleteCustomFile("UDF", "", filepath, filename)

}

// DeleteSystemFile removes a system file
func (c *Client) DeleteSystemFile(filepath string, filename string) error {

	return c.DeleteCustomFile("SDF", "", filepath, filename)

}

// DeleteLegacyFile removes a legacy file.
func (c *Client) DeleteLegacyFile(filepath string, filename string) error {

	return c.DeleteCustomFile("LDF", "", filepath, filename)

}

// DeleteProjectFile removes a project file
func (c *Client) DeleteProjectFile(project string, filepath string, filename string) error {

	return c.DeleteCustomFile("PDF", project, filepath, filename)

}

// ------------------------------

// ShareCustomFile sends a request for updates to an existing file
func (c *Client) ShareCustomFile(req string, additional string, filepath string, filename string) (string, string, bool, error) {

	fullpath := filepath + "/" + filename

	if len(fullpath) > 1 && rune(fullpath[0]) == '/' {
		fullpath = fullpath[1:]
	}

	var err error

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return "", "", false, err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	md5 := [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, md5[0:]...)
	r = append(r, []byte(fullpath)...)
	if additional != "" {
		r = append(r, []byte(":"+additional)...)
	}
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)

	var host string
	var port string
	var created bool

	select {
	case _ = <-tochan:
		err = errors.New("timeout")
	case msg := <-c.c.Incoming:
		////fmt.Printf("in CacheCustomFile() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "EXE" {
			parts := strings.Split(string(msg.Payload), ":")
			host = parts[0]
			port = ":" + parts[1]
			created = (parts[2] != "0")
		} else if msg.ID == "ERR" {
			err = errors.New("i/o error")
		}
	}

	return host, port, created, err
}

// ShareUserFile shares to remote
func (c *Client) ShareUserFile(filepath string, filename string) (string, string, bool, error) {

	return c.ShareCustomFile("USH", "", filepath, filename)

}

// ShareSystemFile shares to remote
func (c *Client) ShareSystemFile(filepath string, filename string) (string, string, bool, error) {

	return c.ShareCustomFile("SSH", "", filepath, filename)

}

// ShareLegacyFile shares to remote
func (c *Client) ShareLegacyFile(filepath string, filename string) (string, string, bool, error) {

	return c.ShareCustomFile("LSH", "", filepath, filename)

}

// ShareProjectFile shares to remote
func (c *Client) ShareProjectFile(project string, filepath string, filename string) (string, string, bool, error) {

	return c.ShareCustomFile("PSH", project, filepath, filename)

}

// MetaDataCustomFile sets metadata
func (c *Client) MetaDataCustomFile(req string, additional string, filepath string, filename string, meta map[string]string) error {

	fullpath := filepath + "/" + filename

	if len(fullpath) > 1 && rune(fullpath[0]) == '/' {
		fullpath = fullpath[1:]
	}

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	fr := filerecord.FileRecord{}
	fr.MetaData = meta
	fr.FileName = fullpath
	fr.Owner = additional

	md5, _ := json.Marshal(fr)

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, md5[0:]...)
	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		return errors.New("timeout")
	case msg := <-c.c.Incoming:
		//			//fmt.Printf("in ExistsCustomFile() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "MOK" {
			return nil
		} else if msg.ID == "ERR" {
			return errors.New("META UPDATE FAILED")
		}
	}

	return nil
}

// MetaDataUserFile sets metadata
func (c *Client) MetaDataUserFile(filepath string, filename string, meta map[string]string) error {

	return c.MetaDataCustomFile("UMD", "", filepath, filename, meta)

}

// MetaDataSystemFile sets metadata
func (c *Client) MetaDataSystemFile(filepath string, filename string, meta map[string]string) error {

	return c.MetaDataCustomFile("SMD", "", filepath, filename, meta)

}

// MetaDataLegacyFile sets metadata
func (c *Client) MetaDataLegacyFile(filepath string, filename string, meta map[string]string) error {

	return c.MetaDataCustomFile("LMD", "", filepath, filename, meta)

}

// MetaDataProjectFile sets metadata
func (c *Client) MetaDataProjectFile(project string, filepath string, filename string, meta map[string]string) error {

	return c.MetaDataCustomFile("PMD", project, filepath, filename, meta)

}

// MetaDataRemIntFile sets metadata
func (c *Client) MetaDataRemIntFile(project string, filepath string, filename string, meta map[string]string) error {

	return c.MetaDataCustomFile("IMD", project, filepath, filename, meta)

}

// =====================================================================================================
// ValidateCacheXXX functions ::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
// =====================================================================================================

// CacheUserFile puts a user file in the network cache
func (c *Client) ValidateCacheUserFile(current *filerecord.FileRecord) (*filerecord.FileRecord, error) {

	return c.ValidateCacheCustomFile("UCF", "", current)

}

// CacheSystemFile puts a system file in the network cache
func (c *Client) ValidateCacheSystemFile(current *filerecord.FileRecord) (*filerecord.FileRecord, error) {

	return c.ValidateCacheCustomFile("SCF", "", current)

}

// CacheLegacyFile puts a software file in the network cache
func (c *Client) ValidateCacheLegacyFile(current *filerecord.FileRecord) (*filerecord.FileRecord, error) {

	return c.ValidateCacheCustomFile("LCF", "", current)

}

// CacheProjectFile puts a project file in the network cache
func (c *Client) ValidateCacheProjectFile(project string, current *filerecord.FileRecord) (*filerecord.FileRecord, error) {

	return c.ValidateCacheCustomFile("PCF", project, current)

}

// DeleteCustomFile deletes a file
func (c *Client) RenameCustomFile(req string, additional string, filepath string, filename string, newfilename string) error {

	fullpath := filepath + "/" + filename

	if len(fullpath) > 1 && rune(fullpath[0]) == '/' {
		fullpath = fullpath[1:]
	}

	if c.c == nil {
		c.Username = "system"
		c.c = client.NewDuckTapeClient(c.Hostname, c.Port, c.Username, "tcp")
		err := c.c.Connect()
		if err != nil {
			//			//fmt.Println("Failed connect")
			return err
		}
		//		//fmt.Println("Connected")
	}

	if c.Session == "" {
		c.Session = "12345678123456781234567812345678"
	}

	// Now do the connection
	r := []byte(c.Session)
	r = append(r, []byte(fullpath)...)
	r = append(r, []byte(":"+additional)...)
	r = append(r, []byte(":"+newfilename)...)

	log.Printf("Send %s: [%s]\n", req, string(r))

	c.c.SendMessage(req, r, true)

	// get response
	tochan := time.After(time.Second * 20)
	select {
	case _ = <-tochan:
		return errors.New("timeout")
	case msg := <-c.c.Incoming:
		//fmt.Printf("in DeleteCustomFile() %s, %s\n", msg.ID, string(msg.Payload))
		if msg.ID == "DOK" {
			return nil
		} else if msg.ID == "ERR" {
			return errors.New("DELETE FAILED")
		}
	}

	return nil
}

// DeleteUserFile removes a user file
func (c *Client) RenameUserFile(filepath string, filename string, newfilename string) error {

	return c.RenameCustomFile("RFU", "", filepath, filename, newfilename)

}

// DeleteSystemFile removes a system file
func (c *Client) RenameSystemFile(filepath string, filename string, newfilename string) error {

	return c.RenameCustomFile("RFS", "", filepath, filename, newfilename)

}

// DeleteLegacyFile removes a legacy file.
func (c *Client) RenameLegacyFile(filepath string, filename string, newfilename string) error {

	return c.RenameCustomFile("RFL", "", filepath, filename, newfilename)

}

// DeleteProjectFile removes a project file
func (c *Client) RenameProjectFile(project string, filepath string, filename string, newfilename string) error {

	return c.RenameCustomFile("RFP", project, filepath, filename, newfilename)

}

// ------------------------------
