package filerecord

import (
	"gopkg.in/mgo.v2/bson"
    "time"
)

type MsgComment struct {
	User    string
	Content string
    Created time.Time
}

type MsgAttachment struct {
	Name    string
	Content []byte
    Created time.Time
}

type MsgType int
const (
	  MT_EMAIL MsgType = iota
      MT_WALL
      MT_TOPIC
      MT_ATTACHMENT
      MT_TOPIC_REPLY
)

func (bt MsgType) String() string {
	 switch (bt) {
     case MT_EMAIL: return "email"
     case MT_WALL: return "wall"
     case MT_ATTACHMENT: return "attachment"
     case MT_TOPIC: return "bbs topic"
     case MT_TOPIC_REPLY: return "bbs reply"
     }
     return "Invalid"
}

type MsgState int
const (
	  MS_NEW MsgState = iota
      MS_READ
	  MS_ARCHIVED
)

func (bs MsgState) String() string {

	 switch (bs) {
     case MS_NEW: return "New"
     case MS_READ: return "Read"
     case MS_ARCHIVED: return "Archived"
     }

     return "Invalid"
}

type PercolMessage struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	ParentID    bson.ObjectId `bson:"parentid,omitempty"`
    State       MsgState
	Type        MsgType
	Summary     string
	Body        string
	BodyData    []byte
	Creator     string
    Created     time.Time
	Recipients  []string   // For "email"
	Comments    []bson.ObjectId
	Attachments []bson.ObjectId
}

func (this *PercolMessage) BSON() []byte {
	b, _ := bson.Marshal(this)
	return b
}

func (this *PercolMessage) UnBSON(data []byte) {
	_ = bson.Unmarshal(data, this)
}
