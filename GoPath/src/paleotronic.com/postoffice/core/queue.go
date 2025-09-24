package postoffice

import (
	"bytes"
	"errors"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"paleotronic.com/log"
	"os"
	"sync"
	"time"
)

type MQAddress struct {
	Hostname string
	Port     int
}

type QueueMap map[string]*QueueDefinition

type QueueEntry struct {
	Payload     []byte
	Expiry      int64
	Tag         string
	ID          int64
	SecretID    string
	LeasedUntil int64
}

type QueueDefinition struct {
	Limit         int
	Mode          string
	UniqueID      int64
	Timeout       int
	Endpoint      string
	Entries       []QueueEntry
	Error         string
	UniquePayload bool
	// mutex
	mu sync.Mutex
}

func ReadMQClientConfig(filename string) *MQAddress {

	log.Printf("Loading MQ state")

	var config MQAddress = MQAddress{Hostname: "localhost", Port: 9001}

	// does file exist
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Load client config failed due to: %v", err.Error())
		os.Exit(2)
	}

	erru := yaml.Unmarshal(raw, &config)

	if erru != nil {
		log.Fatal("Load client config failed due to corrupt queue config")
		os.Exit(1)
	}

	return &config
}

func WriteMQClientConfig(filename string, config *MQAddress) {
	log.Printf("Saving MQ Client config")

	raw, merr := yaml.Marshal(&config)
	if merr != nil {
		log.Fatal("Config save failed: %v", merr.Error())
		os.Exit(2)
	}

	err := ioutil.WriteFile(filename, raw, 0644)
	//log.Printf("YAML: %v", string(raw))
	if err != nil {
		log.Fatal("Config save failed: %v", err.Error())
		os.Exit(1)
	}
}

// put message into queue
func (qd *QueueDefinition) Put(payload []byte, expiry int64, tag string, sid string) (int64, error) {

	now := time.Now().UnixNano()

	if expiry < now {
		// set expiry from the QueueDefinition
		expiry = now + (int64(qd.Timeout) * 1000000000)
	}

	// purge expired entries if needed
	qd.Purge()

	if len(qd.Entries) >= qd.Limit {
		return -1, errors.New("Queue is currently at Limit")
	}

	var pidx int = -1

	// queue can take new entry
	if qd.UniquePayload {
		// if there is an entry with the current payload we need to remove it..
		//log.Printf("Replacing matched payload if existing - %v", string(payload) )
		pidx = qd.GetPayloadIndex(payload)
	}

	qd.mu.Lock()
	defer qd.mu.Unlock() // always unlock mutexes

	qd.UniqueID = qd.UniqueID + 1
	id := qd.UniqueID

	qe := QueueEntry{ID: id, Payload: payload, Tag: tag, Expiry: expiry, SecretID: sid}

	if pidx == -1 {
		qd.Entries = append(qd.Entries, qe)
		//log.Printf("Added message to queue - %v", qe)
	} else {
		qd.Entries[pidx] = qe
		//log.Printf("Updated message in queue - %v", qe )
	}

	return id, nil
}

// return position in queue / size of queue
func (qd *QueueDefinition) Query(payload []byte, expiry int64, tag string, sid string) (int64, int64, []string, error) {

	// purge expired entries if needed
	qd.Purge()

	qd.mu.Lock()
	defer qd.mu.Unlock() // always unlock mutexes

	var tagmap map[string]int
	tagmap = make(map[string]int)
	var taglist []string

	// find current payload in queue
	var idx int64 = -1
	for i := range qd.Entries {
		qe := qd.Entries[i]
		if bytes.Compare(payload, qe.Payload) == 0 {
			idx = int64(i)
		}
		if qe.Tag != "" {
			// handle tag map
			tagmap[qe.Tag] = 1
		}
	}

	// build taglist
	for k, _ := range tagmap {
		taglist = append(taglist, k)
	}

	return idx, int64(len(qd.Entries)), taglist, nil
}

func (qd *QueueDefinition) Delete(id int64, sid string) {
	// purges expired entries from the queue
	qd.mu.Lock()
	defer qd.mu.Unlock()

	if len(qd.Entries) == 0 {
		return
	}

	var newlist []QueueEntry
	for e := range qd.Entries {
		if qd.Entries[e].ID == id || (qd.Entries[e].SecretID == sid && sid != "") {
			//log.Printf("Deleting item from queue - %v", e)
		} else {
			newlist = append(newlist, qd.Entries[e])
		}
	}

	qd.Entries = newlist
	return
}

func (qd *QueueDefinition) Purge() {
	qd.mu.Lock()
	defer qd.mu.Unlock()
	// purges expired entries from the queue
	if len(qd.Entries) == 0 {
		return
	}

	var newlist []QueueEntry
	for e := range qd.Entries {
		if qd.Entries[e].Expiry != 0 && qd.Entries[e].Expiry < time.Now().UnixNano() {
			log.Printf("Purging expired item from queue - %v", e)
		} else {
			newlist = append(newlist, qd.Entries[e])
		}
	}

	qd.Entries = newlist
	return
}

func Contains(list []string, elem string) bool {
	for _, t := range list {
		if t == elem {
			return true
		}
	}
	return false
}

func (qd *QueueDefinition) ClearTags(tags []string) error {
	qd.mu.Lock()
	defer qd.mu.Unlock()
	// purges expired entries from the queue
	if len(qd.Entries) == 0 {
		return nil
	}

	for e := range qd.Entries {
		if qd.Entries[e].Tag != "" && Contains(tags, qd.Entries[e].Tag) == false {
			log.Printf("Clearing non-existent tag [%v]", qd.Entries[e].Tag)
			qd.Entries[e].Tag = ""
		}
	}

	return nil
}

func (qd *QueueDefinition) GetPayloadIndex(payload []byte) int {
	qd.mu.Lock()
	defer qd.mu.Unlock()
	// purges expired entries from the queue
	if len(qd.Entries) == 0 {
		return -1
	}

	for e := range qd.Entries {
		if bytes.Compare(payload, qd.Entries[e].Payload) == 0 {
			return e
		}
	}

	return -1
}

func (qd *QueueDefinition) Flush() {
	qd.mu.Lock()
	defer qd.mu.Unlock()
	// purges expired entries from the queue
	if len(qd.Entries) == 0 {
		return
	}

	var newlist []QueueEntry
	for e := range qd.Entries {
		log.Printf("Flushing item from queue - %v", e)
	}

	qd.Entries = newlist
	return
}

// Find first task with optional tag that isn't leased and  Lease it
func (qd *QueueDefinition) Lease(tag string, duration int, max int) ([]QueueEntry, error) {

	now := time.Now().UnixNano()
	var list []QueueEntry
	var count int

	qd.mu.Lock()
	defer qd.mu.Unlock()

	if len(qd.Entries) == 0 {
		return list, nil
	}

	for idx := range qd.Entries {
		qe := qd.Entries[idx]
		if (tag == "" || tag == qe.Tag) || (tag == "*" && qe.Tag == "") {
			//log.Printf("LEASED UNTIL = %d, NOW = %d", qe.LeasedUntil, now)
			if qe.LeasedUntil < now && count < max {
				// lease task and return it
				qe.LeasedUntil = time.Now().UnixNano() + int64(1000000000*duration)
				qd.Entries[idx].LeasedUntil = qe.LeasedUntil
				list = append(list, qe)
				count = count + 1
			}
		}
	}

	return list, nil
}

func (qd *QueueDefinition) Allocate(taglist []string, index int) (int, error) {

	if index >= len(taglist) {
		index = 0
	}

	qd.mu.Lock()
	defer qd.mu.Unlock()

	if len(qd.Entries) == 0 {
		return index, nil
	}

	for idx := range qd.Entries {
		qe := qd.Entries[idx]

		if qe.Tag == "" {
			// allocate to next tag
			qe.Tag = taglist[index]
			qd.Entries[idx].Tag = taglist[index]
			log.Printf("Allocating request %v to tag %v", qe.ID, qe.Tag)

			index = index + 1
			if index >= len(taglist) {
				index = 0
			}

		}
	}

	return index, nil
}
