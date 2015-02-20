package graphie

type Attrs map[string]interface{}

type Storage interface {
	// Initializes the underlying storage, creates appropriate indexes and
	// establishes a connection, if not done yet
	Start(attrs string, dbname string) error

	// Stops the storage and shuts it down properly
	Stop() error

	EnsureIndexNodes(labels []string, attrName string) error
	EnsureIndexLinks(labels []string, attrName string) error

	// Elementary CRUD operations for nodes
	Add(labels []string, attrs Attrs) (NodeID, error)
	Merge(labels []string, attrs Attrs) (NodeID, error)
	Link(from, to NodeID, attrs Attrs) error
	Unlink(from, to NodeID, attrs Attrs) error
	Remove(id NodeID) error

	// Edge handling
	In(id NodeID) ([]*Link, error)
	Out(id NodeID) ([]*Link, error)

	// Attribute handling
	Set(id NodeID, key string, value interface{}) error
	Get(id NodeID, key string) (interface{}, error)
	Has(id NodeID, key string) (bool, error)
	Attrs(id NodeID) (Attrs, error)
}

type Link struct {
	Other NodeID
	Attrs Attrs
}

type NodeID uint64
type Node interface {
	//ID() NodeID

	Link(to Node, attrs Attrs) error

	// Attributes
	/*Set(key string, value interface{}) error
	Has(key string) (bool, error)
	Get(key string) (interface{}, error)

	// In/out nodes
	In() INodeSet
	Out() INodeSet
	Both() INodeSet

	// The following methods are for convenient use; they panic on error.
	SafeSet(key string, value interface{})
	SafeHas(key string) bool
	SafeGet(key string) interface{}*/
}
