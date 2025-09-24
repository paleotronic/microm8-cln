package forumtool

import (
	"paleotronic.com/server/forumapi"
)

type byThread []*forumapi.MessageDetails

func (s byThread) Len() int {
	return len(s)
}

func (s byThread) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byThread) Less(i, j int) bool {
	a := s[i]
	b := s[j]
	var akey, bkey int64

	if a.ParentId != 0 {
		akey = int64(a.ParentId)*1000000 + int64(a.MessageId)
	} else {
		akey = int64(a.MessageId) * 1000000
	}

	if b.ParentId != 0 {
		bkey = int64(b.ParentId)*1000000 + int64(b.MessageId)
	} else {
		bkey = int64(b.MessageId) * 1000000
	}

	return akey < bkey
}

type byDateAsc []*forumapi.MessageDetails

func (s byDateAsc) Len() int {
	return len(s)
}

func (s byDateAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byDateAsc) Less(i, j int) bool {
	a := s[i]
	b := s[j]

	return a.Created < b.Created
}

type byDateDesc []*forumapi.MessageDetails

func (s byDateDesc) Len() int {
	return len(s)
}

func (s byDateDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byDateDesc) Less(i, j int) bool {
	a := s[i]
	b := s[j]

	return b.Created < a.Created
}
