package filerecord

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type BugComment struct {
	CommentId int
	User      string
	Content   string
	Created   time.Time
}

type BugAttachment struct {
	AttachmentId int
	Name         string
	Content      []byte
	Created      time.Time
}

type BugType int

const (
	BT_BUG BugType = iota
	BT_FEATURE
)

func (bt BugType) String() string {
	switch bt {
	case BT_BUG:
		return "Bug"
	case BT_FEATURE:
		return "Feature"
	}
	return "Invalid"
}

type BugState int

const (
	BS_NEW BugState = iota
	BS_OPEN
	BS_INPROGRESS
	BS_WAITING
	BS_RETEST
	BS_FIXED
	BS_CLOSED
)

func (bs BugState) String() string {

	switch bs {
	case BS_NEW:
		return "New"
	case BS_OPEN:
		return "Open"
	case BS_INPROGRESS:
		return "Progress"
	case BS_WAITING:
		return "Waiting"
	case BS_RETEST:
		return "Retest"
	case BS_FIXED:
		return "Fixed"
	case BS_CLOSED:
		return "Closed"
	}

	return "Invalid"
}

type BugReport struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	DefectID    int64
	State       BugState
	Type        BugType
	Summary     string
	Body        string
	Creator     string
	Created     time.Time
	Filepath    string
	Filename    string
	Assigned    string
	Comments    []BugComment
	Attachments []BugAttachment
}

func (this *BugReport) BSON() []byte {
	b, _ := bson.Marshal(this)
	return b
}

func (this *BugReport) UnBSON(data []byte) {
	_ = bson.Unmarshal(data, this)
}

func init() {
}
