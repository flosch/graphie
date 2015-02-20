package happy

import (
	"encoding/binary"
	"io"

	"github.com/flosch/graphie"

	"github.com/vmihailenco/msgpack"
)

type link struct {
	other uint64
	attrs graphie.Attrs
}

type node struct {
	s      *storage
	id     uint64
	labels []uint16
	attrs  graphie.Attrs

	linksOut []*link
	linksIn  []*link
}

func (n *node) write(w io.Writer) error {
	// Write number of labels
	err := binary.Write(w, binary.BigEndian, uint16(len(n.labels)))
	if err != nil {
		return err
	}

	// Write all labels
	for _, lid := range n.labels {
		err = binary.Write(w, binary.BigEndian, lid)
		if err != nil {
			return err
		}
	}

	enc := msgpack.NewEncoder(w)

	// Write out-links
	/*err = binary.Write(w, binary.BigEndian, uint64(len(n.links_out)))
	if err != nil {
		return err
	}
	for _, id := range n.links_out {
		err = binary.Write(w, binary.BigEndian, id)
		if err != nil {
			return err
		}
	}*/
	err = enc.Encode(n.linksOut)
	if err != nil {
		return err
	}

	// Write in-links
	/*err = binary.Write(w, binary.BigEndian, uint64(len(n.links_in)))
	if err != nil {
		return err
	}
	for _, id := range n.links_in {
		err = binary.Write(w, binary.BigEndian, id)
		if err != nil {
			return err
		}
	}*/
	err = enc.Encode(n.linksIn)
	if err != nil {
		return err
	}

	// Write attributes
	err = enc.Encode(n.attrs)
	if err != nil {
		return err
	}

	return nil
}

func (n *node) writeNeighbours(w *bufferedWriteCounter, memtable map[uint64]*node, index map[uint64]int, todoList map[uint64]struct{}) error {
	toVisit := make([]*node, 0, 10)

	// Write all connected nodes close together (outgoing)
	for _, lnk := range n.linksOut {
		// Visited? Ignore
		_, has := index[lnk.other]
		if has {
			continue
		}

		// In current memtable?
		neighbour, has := memtable[lnk.other]
		if has {
			// TODO: Optimize this (for example by looking at in/out-degrees of the connected nodes)
			index[neighbour.id] = w.Size()
			delete(todoList, neighbour.id)
			neighbour.write(w)

			toVisit = append(toVisit, neighbour)
		}
	}

	for _, lnk := range n.linksIn {
		// Visited? Ignore
		_, has := index[lnk.other]
		if has {
			continue
		}

		// In current memtable?
		neighbour, has := memtable[lnk.other]
		if has {
			// TODO: Optimize this (for example by looking at in/out-degrees of the connected nodes)
			index[neighbour.id] = w.Size()
			delete(todoList, neighbour.id)
			neighbour.write(w)

			toVisit = append(toVisit, neighbour)
		}
	}

	for _, neighbour := range toVisit {
		err := neighbour.writeNeighbours(w, memtable, index, todoList)
		if err != nil {
			return err
		}
	}

	return nil
}
