package happy

import (
	"bufio"
	"encoding/binary"
	"hash/fnv"
	"io"
	"os"
	"time"

	//"github.com/petar/GoLLRB/llrb"
)

const (
	nodetableVersion             = uint16(1)
	nodetablePersistentThreshold = 1e6
	nodetableBloomSize           = 5e8
	nodetableBloomIterations     = 3
	nodetableBloomArraysize      = nodetableBloomSize / 8
)

// TODO:
// - Add commit log and log removal on nodetable write-to-disk

type nodetableIdx int

type nodetable struct {
	filename string
	created  uint64 // timestamp in ns
	bitmap   []byte
	idx      nodetableIdx
}

func loadNodetable(filename string) *nodetable {
	// TODO
	return nil
}

type bufferedWriteCounter struct {
	bw   *bufio.Writer
	size int
}

func newBufferedWriteCounter(w io.Writer) *bufferedWriteCounter {
	return &bufferedWriteCounter{
		bw: bufio.NewWriter(w),
	}
}

func (wc *bufferedWriteCounter) Size() int {
	return wc.size
}

func (wc *bufferedWriteCounter) Write(p []byte) (nn int, err error) {
	wc.size += len(p)
	return wc.bw.Write(p)
}

func (wc *bufferedWriteCounter) Flush() error {
	return wc.bw.Flush()
}

func createNodetable(memtable map[uint64]*node, filename string) (*nodetable, error) {
	nt := &nodetable{
		filename: filename,
		created:  uint64(time.Now().Nanosecond()),
		bitmap:   make([]byte, nodetableBloomArraysize),
	}

	fd, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := fd.Close()
		if err != nil {
			panic(err)
		}
	}()

	wd := newBufferedWriteCounter(fd)
	defer func() {
		err := wd.Flush()
		if err != nil {
			panic(err)
		}
	}()

	// Version nodetable_version
	err = binary.Write(wd, binary.BigEndian, nodetableVersion)
	if err != nil {
		return nil, err
	}

	// Timestamp
	err = binary.Write(wd, binary.BigEndian, nt.created)
	if err != nil {
		return nil, err
	}

	// Build the bitmap
	for k := range memtable {
		markBitmap(k, nt)
	}

	// Write bitmap
	_, err = wd.Write(nt.bitmap)
	if err != nil {
		return nil, err
	}

	todoList := make(map[uint64]struct{})
	empty := struct{}{}

	index := make(map[uint64]int)

	var start *node
	for k, v := range memtable {
		if start == nil {
			start = v
		}

		todoList[k] = empty
	}

	for {
		// Remove from todo list
		delete(todoList, start.id)

		// Write start node
		index[start.id] = wd.Size()
		err := start.write(wd)
		if err != nil {
			return nil, err
		}

		// Write all connected nodes recursively
		err = start.writeNeighbours(wd, memtable, index, todoList)
		if err != nil {
			return nil, err
		}

		// TODO: Optimize this bookkeeping and checking
		if len(index) == len(memtable) {
			break
		}

		// We still have a node left which is not recorded yet
		for k := range todoList {
			start = memtable[k]
			break
		}
	}

	// Write index

	return nt, nil
}

func markBitmap(x uint64, n *nodetable) {
	for i := 0; i < nodetableBloomIterations; i++ {
		h := fnv.New64a()
		err := binary.Write(h, binary.BigEndian, x)
		if err != nil {
			panic(err)
		}
		x = h.Sum64()

		bitpos := x % nodetableBloomSize
		arraypos := bitpos / 8
		n.bitmap[arraypos] = n.bitmap[arraypos] | (1 << (bitpos % 8))
	}
}
