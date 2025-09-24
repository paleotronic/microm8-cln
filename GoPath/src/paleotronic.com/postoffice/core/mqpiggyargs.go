package postoffice

type QueryArgs struct {
	Tag           string
	Queue         string
	Payload       []byte
	Expiry        int64  // zero = no expiry - be careful
	ReturnAddress string // message queue for any response
	SecretID      string
}

type PutArgs struct {
	Tag           string
	Queue         string
	Payload       []byte
	Expiry        int64  // zero = no expiry - be careful
	ReturnAddress string // message queue for any response
	SecretID      string
}

type PurgeArgs struct {
	Queue string
}

type DeleteArgs struct {
	Queue     string
	RequestID int64
	SecretID  string
}

type LeaseArgs struct {
	Tag   string
	Queue string
	Time  int
	Max   int
}

type ClearArgs struct {
	TagList []string
	Queue   string
}

type AllocateArgs struct {
	TagList    []string
	Queue      string
	StartIndex int
}

type MQResponse struct {
	Status        int
	StatusMessage string
	Data          []QueueEntry
	Index         int64
	Count         int64
	DistinctTags  []string // non blank tags used in stuff - gotten from query
	Nanos         int64    // processing time
}
