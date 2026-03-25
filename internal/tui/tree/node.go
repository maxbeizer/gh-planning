package tree

import "github.com/maxbeizer/gh-planning/internal/github"

// Node represents a single item in the tree hierarchy.
type Node struct {
	Item     *github.ProjectItem
	Children []*Node
	Expanded bool
	Depth    int
}

// hasChildren returns true if this node has child nodes.
func (n *Node) hasChildren() bool {
	return len(n.Children) > 0
}
