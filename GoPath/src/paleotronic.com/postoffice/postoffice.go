package main

import (
	"errors"
	"paleotronic.com/fmt"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"paleotronic.com/log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"paleotronic.com/postoffice/core"
	"strconv"
	"syscall"
	"time"
)

type MQ struct{}

const (
	CLIENTCONFIG = "mqclient.yaml"
)

var (
	queues postoffice.QueueMap
	PORT   string = ":9002"
)

func (t *MQ) Flush(args *postoffice.PurgeArgs, reply *postoffice.MQResponse) error {

	s := time.Now().UnixNano()

	// validate queue exists
	qd, ok := queues[args.Queue]
	if !ok {
		return errors.New("Queue does not exist [" + args.Queue + "]")
	}

	// ok queue does exist
	qd.Flush()

	//reply.UID = 0
	reply.StatusMessage = "FLUSH OK"
	reply.Nanos = time.Now().UnixNano() - s

	return nil
}

func (t *MQ) Purge(args *postoffice.PurgeArgs, reply *postoffice.MQResponse) error {

	s := time.Now().UnixNano()

	// validate queue exists
	qd, ok := queues[args.Queue]
	if !ok {
		return errors.New("Queue does not exist [" + args.Queue + "]")
	}

	// ok queue does exist
	qd.Purge()

	reply.StatusMessage = "PURGE OK"
	reply.Nanos = time.Now().UnixNano() - s

	return nil
}

func (t *MQ) Delete(args *postoffice.DeleteArgs, reply *postoffice.MQResponse) error {

	s := time.Now().UnixNano()

	// validate queue exists
	qd, ok := queues[args.Queue]
	if !ok {
		return errors.New("Queue does not exist [" + args.Queue + "]")
	}

	qd.Delete(args.RequestID, args.SecretID)
	//reply.UID = args.RequestID
	reply.StatusMessage = "DELETE OK"

	reply.Nanos = time.Now().UnixNano() - s

	return nil
}

func (t *MQ) Query(args *postoffice.QueryArgs, reply *postoffice.MQResponse) error {

	s := time.Now().UnixNano()

	// validate the request
	if len(args.Payload) == 0 {
		return errors.New("No payload specified for QUERY request")
	}

	// validate queue exists
	qd, ok := queues[args.Queue]
	if !ok {
		return errors.New("Queue does not exist [" + args.Queue + "]")
	}

	// ok queue does exist
	idx, count, taglist, response := qd.Query(args.Payload, args.Expiry, args.Tag, args.SecretID)

	reply.StatusMessage = fmt.Sprintf("%d:%d", idx, count)
	reply.Count = count
	reply.Index = idx
	reply.DistinctTags = taglist
	reply.Nanos = time.Now().UnixNano() - s

	return response
}

func (t *MQ) Put(args *postoffice.PutArgs, reply *postoffice.MQResponse) error {

	s := time.Now().UnixNano()

	// validate the request
	if len(args.Payload) == 0 {
		return errors.New("No payload specified for PUT request")
	}

	// validate queue exists
	qd, ok := queues[args.Queue]
	if !ok {
		return errors.New("Queue does not exist [" + args.Queue + "]")
	}

	// ok queue does exist
	id, response := qd.Put(args.Payload, args.Expiry, args.Tag, args.SecretID)

	reply.StatusMessage = fmt.Sprintf("PUT %v", id)

	reply.Nanos = time.Now().UnixNano() - s

	return response
}

func (t *MQ) Lease(args *postoffice.LeaseArgs, reply *postoffice.MQResponse) error {

	s := time.Now().UnixNano()

	// validate queue exists
	qd, ok := queues[args.Queue]
	if !ok {
		return errors.New("Queue does not exist [" + args.Queue + "]")
	}

	// handle default
	if args.Max == 0 {
		args.Max = 1
	}

	qe, err := qd.Lease(args.Tag, args.Time, args.Max)
	if err != nil {
		return err
	}

	// yaml encode data
	reply.Data = qe
	reply.StatusMessage = "LEASE OK"

	reply.Nanos = time.Now().UnixNano() - s

	return nil
}

func (t *MQ) ClearTags(args *postoffice.ClearArgs, reply *postoffice.MQResponse) error {

	s := time.Now().UnixNano()

	// validate queue exists
	qd, ok := queues[args.Queue]
	if !ok {
		return errors.New("Queue does not exist [" + args.Queue + "]")
	}

	err := qd.ClearTags(args.TagList)

	reply.Nanos = time.Now().UnixNano() - s

	return err
}

func (t *MQ) Allocate(args *postoffice.AllocateArgs, reply *postoffice.MQResponse) error {

	s := time.Now().UnixNano()

	// validate queue exists
	qd, ok := queues[args.Queue]
	if !ok {
		return errors.New("Queue does not exist [" + args.Queue + "]")
	}

	idx, err := qd.Allocate(args.TagList, args.StartIndex)

	reply.Index = int64(idx)

	reply.Nanos = time.Now().UnixNano() - s

	return err
}

func saveConfig(filename string) {
	log.Printf("Saving MQ state")

	raw, merr := yaml.Marshal(&queues)
	if merr != nil {
		log.Fatal("MQ Shutdown failed due to: %v", merr.Error())
		os.Exit(2)
	}

	err := ioutil.WriteFile(filename, raw, 0644)
	//log.Printf("YAML: %v", string(raw))
	if err != nil {
		log.Fatal("MQ Shutdown failed due to: %v", err.Error())
		os.Exit(1)
	}

}

func loadConfig(filename string) {

	log.Printf("Loading MQ state")

	queues = make(postoffice.QueueMap)

	// does file exist
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("MQ Startup failed due to: %v", err.Error())
		os.Exit(2)
	}

	erru := yaml.Unmarshal(raw, &queues)

	if erru != nil {
		log.Fatal("MQ Startup failed due to corrupt queue defintions")
		os.Exit(1)
	}

	for k := range queues {
		log.Printf("Queue defined %v - %v entries in queue", k, len(queues[k].Entries))
		var qd *postoffice.QueueDefinition = queues[k]
		qd.Purge()
		//qd.Put( []byte("Frogs"), time.Now().UnixNano() + 3000000000, "frog-tag" )
		//qd.Entries = append( qd.Entries, QueueEntry{ Tag: "sample", Expiry: time.Now().UnixNano()+10000000, Payload: []byte("hello world") } )
	}
}

// this process expires messages from all queues
func PurgeExpiredMessages() {
	for {

		for qidx := range queues {

			qd := queues[qidx]
			if qd.Timeout > 0 && len(qd.Entries) > 0 {
				qd.Purge()
			}
		}

		time.Sleep(1000000000)

	}
}

func InitMQAddress() {

	config := postoffice.ReadMQClientConfig(CLIENTCONFIG)

	PORT = ":" + strconv.FormatInt(int64(config.Port), 10)

	log.Printf("Using MQ Address: %v", PORT)
}

func main() {

	hn, _ := os.Hostname()
	port := 9001
	postoffice.WriteMQClientConfig("mqclient.yaml", &postoffice.MQAddress{Hostname: hn, Port: port})

	// load config
	loadConfig("mqstate.yaml")

	// Term handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			////fmt.Println()
			log.Printf("MQ Service (postoffice) shutting down on %v", sig)
			saveConfig("mqstate.yaml")
			os.Exit(1)
		}
	}()

	InitMQAddress()

	mq := new(MQ)
	rpc.Register(mq)
	rpc.HandleHTTP()
	listener, e := net.Listen("tcp", PORT)
	if e != nil {
		log.Fatal("MQ Service listen error:", e)
	}
	log.Println("MQ Service (postoffice) listening on", PORT)

	// start purge process
	go PurgeExpiredMessages()

	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("Error accepting MQ connection: " + err.Error())
		} else {
			log.Printf("New MQ connection established: %v", conn.RemoteAddr())
			go rpc.ServeConn(conn)
		}
	}
}
