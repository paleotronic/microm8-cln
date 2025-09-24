package forumtool

import (
	"testing"
	"time"

	"paleotronic.com/server/forumapi"
)

func TestTree(t *testing.T) {

	r := NewRootNode()
	r.AddMsg(
		&forumapi.MessageDetails{
			MessageId: 1,
			Subject:   "Welcome",
			Created:   time.Now().UnixNano(),
		},
	)
	r.AddMsg(
		&forumapi.MessageDetails{
			MessageId: 2,
			Subject:   "About ducks",
			Created:   time.Now().UnixNano(),
		},
	)
	r.AddMsg(
		&forumapi.MessageDetails{
			MessageId: 3,
			Subject:   "Rules and stuff",
			Created:   time.Now().UnixNano(),
		},
	)

	m1 := r.FindMessageById(1)
	if m1 == nil {
		t.Errorf("Could not find message by Id")
	}

	_, err := r.PlaceMsg(
		&forumapi.MessageDetails{
			MessageId: 4,
			ParentId:  1,
			Subject:   "Re: Welcome",
			Created:   time.Now().UnixNano(),
		},
	)
	if err != nil {
		t.Fatalf("Failed to place msg 4: %v", err)
	}

	_, err = r.PlaceMsg(
		&forumapi.MessageDetails{
			MessageId: 5,
			ParentId:  1,
			Subject:   "Re: Welcome",
			Created:   time.Now().UnixNano() - 100000000000,
		},
	)
	if err != nil {
		t.Fatalf("Failed to place msg 5: %v", err)
	}

	r.SetDesc(false)

	nodes := r.GetAll()
	if len(nodes) != 5 {
		t.Fatalf("Expected nodes to be 5, but got %d", len(nodes))
	}

	for i, n := range nodes {
		t.Logf("Node #%d: Message ID = %d, Parent ID = %d", i, n.Message.MessageId, n.Message.ParentId)
	}

	t.Fail()

}
