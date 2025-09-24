package forumtool

import (
	"sort"

	"paleotronic.com/log"
	"paleotronic.com/server/forumapi"
)

type MsgNode struct {
	Message  *forumapi.MessageDetails
	Children []*MsgNode
	Parent   *MsgNode
}

func NewMsgNode(parent *MsgNode, msg *forumapi.MessageDetails) *MsgNode {
	n := &MsgNode{
		Message:  msg,
		Parent:   parent,
		Children: make([]*MsgNode, 0),
	}
	if parent != nil {
		// associate with parent node
		parent.AddChild(n)
	}
	return n
}

func (n *MsgNode) AddChild(c *MsgNode) {
	idx := -1
	for i, nn := range n.Children {
		if nn.Message.MessageId == c.Message.MessageId {
			idx = i
			break
		}
	}
	if idx != -1 {
		n.Children[idx] = c
	} else {
		n.Children = append(n.Children, c)
	}
	// enforce parent relationship
	c.Parent = n
}

func (n *MsgNode) AddMsg(m *forumapi.MessageDetails) *MsgNode {
	nn := NewMsgNode(n, m)
	return nn
}

func (n *MsgNode) NumChildren() int {
	return len(n.Children)
}

func (n *MsgNode) HasChildren() bool {
	return n.NumChildren() > 0
}

func (n *MsgNode) GetChild(idx int) *MsgNode {
	if idx >= 0 && idx < n.NumChildren() {
		return n.Children[idx]
	}
	return nil
}

// How many hops to the root?
func (n *MsgNode) Generation() int {
	hopCount := 0
	nn := n
	for nn.Parent != nil {
		nn = nn.Parent
		hopCount++
	}
	return hopCount
}

// ---------------------

type RootNode struct {
	*MsgNode
	Desc bool
}

func NewRootNode() *RootNode {
	return &RootNode{
		MsgNode: NewMsgNode(nil, nil),
	}
}

func (n *MsgNode) walkChildrenForId(mid int32) *MsgNode {

	var found *MsgNode
	for _, nn := range n.Children {
		if nn.Message.MessageId == mid {
			return nn
		}
		found = nn.walkChildrenForId(mid)
		if found != nil {
			return found
		}
	}
	return nil

}

func (r *RootNode) FindMessageById(mid int32) *MsgNode {

	return r.walkChildrenForId(mid)

}

func (r *RootNode) PlaceMsg(msg *forumapi.MessageDetails) (*MsgNode, error) {

	log.Printf("Placing message id %d (parent id = %d)", msg.MessageId, msg.ParentId)

	if msg.ParentId == 0 {
		mm := r.AddMsg(msg)
		return mm, nil
	}
	pp := r.FindMessageById(msg.ParentId)
	if pp == nil {
		return r.AddMsg(msg), nil // rehome to parent
	}
	return pp.AddMsg(msg), nil
}

func (n *MsgNode) OrderChildrenAsc() {
	sort.Sort(byTimeAsc(n.Children))

	for _, nn := range n.Children {
		nn.OrderChildrenAsc()
	}
}

func (n *MsgNode) OrderChildrenDesc() {
	sort.Sort(byTimeDesc(n.Children))

	for _, nn := range n.Children {
		nn.OrderChildrenDesc()
	}
}

func (n *MsgNode) walkIntoSlice(s *[]*MsgNode) {

	for _, cc := range n.Children {
		*s = append(*s, cc)
		cc.walkIntoSlice(s)
	}

}

func (r *RootNode) Order() {
	if r.Desc {
		r.OrderChildrenDesc()
	} else {
		r.OrderChildrenAsc()
	}
}

func (r *RootNode) SetDesc(d bool) {
	r.Desc = d
	r.Order()
}

func (r *RootNode) GetAll() []*MsgNode {
	out := make([]*MsgNode, 0)
	r.walkIntoSlice(&out)
	return out
}

// sorting
type byTimeAsc []*MsgNode

func (s byTimeAsc) Len() int {
	return len(s)
}

func (s byTimeAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byTimeAsc) Less(i, j int) bool {
	return s[i].Message.Created < s[j].Message.Created
}

type byTimeDesc []*MsgNode

func (s byTimeDesc) Len() int {
	return len(s)
}

func (s byTimeDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byTimeDesc) Less(i, j int) bool {
	return s[i].Message.Created > s[j].Message.Created
}
