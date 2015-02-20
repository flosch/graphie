package happy

import (
	"github.com/flosch/graphie"
	"sync"
)

type storage struct {
	c             int
	g             *graphie.Graph
	nodes         []*node
	indexes_nodes map[string]map[string]struct{} // label -> attr-key
	indexes_links map[string]map[string]struct{} // label -> attr-key
	m             sync.RWMutex
}

func (s *storage) Start(attrs string, dbname string) error {
	s.indexes_nodes = make(map[string]map[string]struct{})
	s.indexes_links = make(map[string]map[string]struct{})
	return nil
}

func (s *storage) Stop() error {
	return nil
}

func (s *storage) Add(labels []string, attrs graphie.Attrs) (graphie.Node, error) {
	n := &node{
		attrs: attrs,
	}
	return n, nil
}

func (s *storage) Merge(labels []string, attrs graphie.Attrs) (graphie.Node, error) {
	return nil, nil
}

func (s *storage) Link(labels []string, from, to graphie.Node, attrs graphie.Attrs) (*graphie.Link, error) {
	return nil, nil
}

func (s *storage) Remove(labels []string, name string, attrs map[string]interface{}) error {
	return nil
}

func (s *storage) EnsureIndexNodes(labels []string, attr_name string) error {
	for _, lbl := range labels {
		m, has := s.indexes_nodes[lbl]
		if !has {
			m = make(map[string]struct{})
			s.indexes_nodes[lbl] = m
		}
		m[attr_name] = struct{}{}
	}
	return nil
}

func (s *storage) EnsureIndexLinks(labels []string, attr_name string) error {
	return nil
}
