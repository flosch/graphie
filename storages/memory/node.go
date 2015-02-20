package happy

import (
	"github.com/flosch/graphie"
	//"sync"
)

type link struct {
	to    *node
	attrs graphie.Attrs
}

type node struct {
	attrs graphie.Attrs
	links []*link
}

func (n *node) Link(to graphie.Node, attrs graphie.Attrs) error {
	l := &link{
		to:    to.(*node),
		attrs: attrs,
	}
	n.links = append(n.links, l)
	return nil
}