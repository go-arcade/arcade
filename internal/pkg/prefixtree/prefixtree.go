// Package prefixtree defines a simple prefix tree used in the pipeline demo.
package prefixtree

// Node is a prefix tree node.
type Node struct {
	Count    int
	IsWord   bool
	Children map[byte]*Node
}

// New creates and returns a new, empty Node.
func New() *Node {
	return &Node{
		Children: map[byte]*Node{},
	}
}

// Clear clears out the receiver, recursively emptying (but not deleting)
// child nodes.
func (n *Node) Clear() {
	if n.Count > 0 {
		for _, child := range n.Children {
			child.Clear()
		}
	}
	n.Count = 0
	n.IsWord = false
}

// MergeFrom updates the receiver recursively from the argument.
func (n *Node) MergeFrom(other *Node) {
	n.Count += other.Count
	n.IsWord = n.IsWord || other.IsWord
	for char, otherChild := range other.Children {
		if otherChild.Count == 0 {
			continue
		}
		child, ok := n.Children[char]
		if !ok {
			child = New()
			n.Children[char] = child
		}
		child.MergeFrom(otherChild)
	}
}

// Insert inserts the specified string into this node.  If the receiver is
// the root of a tree, 'suffix' is a whole word; otherwise, it's a suffix
// after discarding as many prefix characters as the receiver node is deep.
func (n *Node) Insert(suffix string) {
	n.Count++
	if len(suffix) == 0 {
		n.IsWord = true
		return
	}
	char := suffix[0]
	suffix = suffix[1:]
	child, ok := n.Children[char]
	if !ok {
		child = New()
		n.Children[char] = child
	}
	child.Insert(suffix)
}
