// Package happy is the main backend storage for graphie. happy stores
// graphs efficiently and supports high-write throughput with a write complexity
// of O(1).
package happy

import (
	"container/list"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"github.com/flosch/graphie"
)

const (
	maxWorkers = 4
)

var (
	ErrNotFound = errors.New("Node not found")
)

type nodetables []*nodetable

func (nt nodetables) Len() int {
	return len(nt)
}

func (nt nodetables) Less(i, j int) bool {
	return nt[i].created > nt[j].created
}

func (nt nodetables) Swap(i, j int) {
	nt[i], nt[j] = nt[j], nt[i]
}

type storage struct {
	g    *graphie.Graph
	path string

	wg sync.WaitGroup

	counterNodes  uint64
	counterLabels uint16

	lock sync.RWMutex

	labelIndex          map[string]uint16
	memtable            map[uint64]*node
	memtableQueueLock   sync.Mutex
	memtableQueue       *list.List
	memtableWorkersChan chan *list.Element
	tables              nodetables
}

func init() {
	graphie.RegisterDriver("happy", registerHappy)
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func registerHappy(g *graphie.Graph) (graphie.Storage, error) {
	return &storage{
		g:                   g,
		labelIndex:          make(map[string]uint16),
		memtable:            make(map[uint64]*node),
		memtableQueue:       list.New(),
		memtableWorkersChan: make(chan *list.Element),
	}, nil
}

func (s *storage) Start(attrs string, dbname string) error {
	// Load all table indexes (bitmaps)
	path, err := filepath.Abs(attrs)
	if err != nil {
		return err
	}
	s.path = path

	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		if !fi.IsDir() {
			return fmt.Errorf("Path '%s' is not a directory.", attrs)
		}
	}

	// Start all workers
	for i := 0; i < maxWorkers; i++ {
		s.wg.Add(1)
		go s.memtableWorker()
	}

	return nil
}

func (s *storage) Stop() error {
	s.lock.RLock()
	if len(s.memtable) > 0 {
		s.lock.RUnlock()
		s.memtableFlush()
	} else {
		s.lock.RUnlock()
	}

	close(s.memtableWorkersChan)
	s.wg.Wait()
	return nil
}

// memtable_flush manages the lock itself!
func (s *storage) memtableFlush() {
	s.lock.Lock()

	// Make old one persistent
	oldMemtable := s.memtable

	// Make the memtables accessible while they are being written on disk
	s.memtableQueueLock.Lock()
	el := s.memtableQueue.PushBack(oldMemtable)
	s.memtableQueueLock.Unlock()

	// Create an empty memtable for new nodes
	s.memtable = make(map[uint64]*node)

	s.lock.Unlock()

	// This blocks when all workers are busy and there's no chance to full the memory
	// Otherwise it will pass the memtable to a worker who will write the memtable down
	// in nodetable format
	s.memtableWorkersChan <- el
}

func (s *storage) Add(labels []string, attrs graphie.Attrs) (graphie.NodeID, error) {
	// Add to memtable
	s.lock.Lock()

	s.counterNodes++

	n := &node{
		s:        s,
		id:       s.counterNodes,
		attrs:    attrs,
		labels:   make([]uint16, 0, len(labels)),
		linksOut: make([]*link, 0, 10), // some guesses with 10 outlinks on avg
		linksIn:  make([]*link, 0, 10), // same
	}

	for _, lbl := range labels {
		n.labels = append(n.labels, s.labelindex(lbl))
	}

	s.memtable[s.counterNodes] = n

	s.lock.Unlock()

	// Check if we reached the treshold
	if len(s.memtable) >= nodetablePersistentThreshold {
		s.memtableFlush()
	}

	return graphie.NodeID(s.counterNodes), nil
}

// s.lock must be held outside of get()
func (s *storage) get(id graphie.NodeID) (*node, error) {
	// First search for the node
	n, err := s.getRaw(id)
	if err != nil {
		return nil, err
	}

	// Second, make a copy of this node
	newNode := &node{
		s:        s,
		id:       n.id,
		attrs:    make(graphie.Attrs),
		linksOut: make([]*link, 0, len(n.linksOut)),
		linksIn:  make([]*link, 0, len(n.linksIn)),
	}
	copy(newNode.labels, n.labels)

	for _, lnk := range n.linksOut {
		newNode.linksOut = append(newNode.linksOut, &link{
			other: lnk.other,
			attrs: lnk.attrs,
		})
	}

	for _, lnk := range n.linksIn {
		newNode.linksIn = append(newNode.linksIn, &link{
			other: lnk.other,
			attrs: lnk.attrs,
		})
	}

	for k, v := range n.attrs {
		newNode.attrs[k] = v
	}

	return newNode, nil
}

func (s *storage) getRaw(id graphie.NodeID) (*node, error) {
	// First, check current memtable
	n, has := s.memtable[uint64(id)]
	if has {
		// We were lucky
		return n, nil
	}

	// Second, check all remaining memtables in the persisting-queue
	s.memtableQueueLock.Lock()
	f := s.memtableQueue.Front()
	for f != nil {
		m := f.Value.(map[uint64]*node)
		n, has = m[uint64(id)]
		if has {
			s.memtableQueueLock.Unlock()
			return n, nil
		}
		f = f.Next()
	}
	s.memtableQueueLock.Unlock()

	// Last, check all persistent nodetables; (1) using bitmap (2) using binary search on disk
	// something like

	panic("not implemented yet")
}

func (s *storage) Unlink(from, to graphie.NodeID, attrs graphie.Attrs) error {
	panic("not implemented")
}

func (s *storage) Link(from, to graphie.NodeID, attrs graphie.Attrs) error {
	// Get both nodes

	// TODO: Do locking on a per node-id basis, not using a global lock
	s.lock.Lock()
	defer s.lock.Unlock()

	// Receive a copy of the node's data
	nodeFrom, err := s.get(from)
	if err != nil {
		return err
	}
	nodeTo, err := s.get(to)
	if err != nil {
		return err
	}

	nodeFrom.linksOut = append(nodeFrom.linksOut, &link{
		other: nodeTo.id,
		attrs: attrs,
	})

	nodeTo.linksIn = append(nodeTo.linksIn, &link{
		other: nodeFrom.id,
		attrs: attrs,
	})

	// Both nodes must be rewritten, add them to the memtable
	s.memtable[nodeFrom.id] = nodeFrom
	s.memtable[nodeTo.id] = nodeTo

	return nil
}

func (s *storage) memtableWorker() {
	defer s.wg.Done()

	for el := range s.memtableWorkersChan {
		mem := el.Value.(map[uint64]*node)

		// Make the memtable persistent
		h := md5.New()
		_, err := io.CopyN(h, rand.Reader, 32)
		if err != nil {
			panic(err)
		}
		nt, err := createNodetable(mem, filepath.Join(s.path, fmt.Sprintf("%x.nt", h.Sum(nil))))
		if err != nil {
			panic(err)
		}

		// Remove it now from the memtable-queue and add it to the nodetables atomically
		s.lock.Lock()

		s.memtableQueueLock.Lock()
		s.memtableQueue.Remove(el)
		s.memtableQueueLock.Unlock()

		s.tables = append(s.tables, nt)

		// It is important to keep all tables sorted by their creation date
		sort.Sort(s.tables)

		s.lock.Unlock()
	}
}

// Does not hold s.lock; must be held outside
func (s *storage) labelindex(l string) uint16 {
	i, has := s.labelIndex[l]
	if has {
		return i
	}

	if s.counterLabels >= 2^16-1 {
		panic("Too many labels; max supported by happy is 2^16-1")
	}

	s.counterLabels++
	s.labelIndex[l] = s.counterLabels
	return s.counterLabels

}

func (s *storage) Merge(labels []string, attrs graphie.Attrs) (graphie.NodeID, error) {
	return s.Add(labels, attrs)
}

func (s *storage) Remove(id graphie.NodeID) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// TODO (important):
	// To keep the database consistent, we have to remove all references to this
	// node (outgoing and incoming nodes in other nodes and we have to write them
	// again)

	// TODO: We're currently ignoring that the node might not exist
	s.memtable[uint64(id)] = nil
	return nil
}

func (s *storage) EnsureIndexNodes(labels []string, attrName string) error {
	return nil
}

func (s *storage) EnsureIndexLinks(labels []string, attrName string) error {
	return nil
}

func (s *storage) In(id graphie.NodeID) ([]*graphie.Link, error) {
	return nil, nil
}

func (s *storage) Out(id graphie.NodeID) ([]*graphie.Link, error) {
	return nil, nil
}

// Attribute handling
func (s *storage) Set(id graphie.NodeID, key string, value interface{}) error {
	return nil
}

func (s *storage) Get(id graphie.NodeID, key string) (interface{}, error) {
	return nil, nil
}

func (s *storage) Has(id graphie.NodeID, key string) (bool, error) {
	return false, nil
}

func (s *storage) Attrs(id graphie.NodeID) (graphie.Attrs, error) {
	return nil, nil
}
