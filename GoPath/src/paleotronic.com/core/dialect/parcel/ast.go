package parcel

const (
	ASTStatement int = iota
	ASTAssignment
	ASTOperator
	ASTCompare
	ASTKeyword
	ASTFunction
	ASTIdentifier
	ASTLiteral
	ASTVariable
	ASTString
	ASTFloat
	ASTInteger
	ASTList
)

// ASTNode represents a node in the abstract syntax tree.
type ASTNode struct {
	Content  string
	TypeID   int
	Parent   *ASTNode
	Children []*ASTNode
}

// NewASTNode creates a new ASTNode.
func NewASTNode(content string, typeId int) *ASTNode {
	return &ASTNode{
		Content:  content,
		TypeID:   typeId,
		Parent:   nil,
		Children: []*ASTNode{},
	}
}

func (n *ASTNode) AddChild(nn *ASTNode) *ASTNode {
	n.Children = append(n.Children, nn)
	nn.Parent = n
	return n
}

func (n *ASTNode) AddSibling(nn *ASTNode) *ASTNode {
	p := n.Parent
	if p == nil {
		panic("cannot add sibling to root node")
	}
	p.AddChild(nn)
	return n
}

func (n *ASTNode) String() string {
	return n.Content
}
