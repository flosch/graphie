package graphie

import (
	"errors"
)

var (
	ErrDriverNotFound = errors.New("Driver not found")
)

type Graph struct {
	name string
	s    Storage
}

func NewGraph(driverName string, driverAttrs string, dbname string) (*Graph, error) {
	g := &Graph{
		name: dbname,
	}

	storageSetupFn, has := drivers[driverName]
	if !has {
		return nil, ErrDriverNotFound
	}
	s, err := storageSetupFn(g)
	if err != nil {
		return nil, err
	}
	err = s.Start(driverAttrs, dbname)
	if err != nil {
		return nil, err
	}
	g.s = s

	return g, nil
}

func (g *Graph) Close() error {
	return g.s.Stop()
}

func (g *Graph) Labels(labels ...string) *LabelGroup {
	return &LabelGroup{
		g:      g,
		labels: labels,
	}
}

func (g *Graph) Link(nodeFrom, nodeTo NodeID, attrs Attrs) error {
	return g.s.Link(nodeFrom, nodeTo, attrs)
}
