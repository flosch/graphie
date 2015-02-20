package graphie

type INode interface {
	ID() NodeID

	// Attributes
	Set(key string, value interface{}) error
	Has(key string) (bool, error)
	Get(key string) (interface{}, error)

	// In/out nodes
	In() INodeSet
	Out() INodeSet
	Both() INodeSet

	// The following methods are for convenient use; they panic on error.
	SafeSet(key string, value interface{})
	SafeHas(key string) bool
	SafeGet(key string) interface{}
}

type INodeSet interface {
	Len() int
}

/*
Examples:
Andere Universitäten
tx.Get("Wikidata").HasAttrValue("name", "TU Berlin")
	.Out("Wikidata", "instanceOf").In("Wikidata", "

Schauspieler, die mit Tom Hanks im selben Film gespielt haben
played_in_same_movie := tx.Build().
tx.Get("Wikidata").


*/

type FilterFn func(n INode) bool

type IQueryBuilder interface {
	// Node properties
	HasLabel(label string) IQueryBuilder                      // filters all vertices for
	HasAttrKey(key string) IQueryBuilder                      // TODO: Uses Filter(), for convenient use
	HasAttrValue(key string, value interface{}) IQueryBuilder // TODO: Uses Filter(), for convenient use
	// HasValueGt[e](), HasValueLt[e](), HasValueBetween(), ... <-- for int* and time., for convenient use, using Filter()
	Filter(filterFn FilterFn) IQueryBuilder // filters all nodes at that stage

	Attr(key, value string) IQueryBuilder

	// Vertices
	In(edgeAttrs ...Attrs) IQueryBuilder
	Out(edgeAttrs ...Attrs) IQueryBuilder
	Both(edgeAttrs ...Attrs) IQueryBuilder

	// Set operations
	Intersect(b IQueryBuilder) IQueryBuilder
	Union(b IQueryBuilder) IQueryBuilder

	// Morphisms
	Follow(m IQueryBuilder) IQueryBuilder

	// Executors
	Count() int
	All() INodeSet
	Limit(n int) INodeSet
	Iterate() chan<- INode
}

type IAtomicInstructions interface {
	// Adds exactly one new node to the database (with a new unique ID)
	Add(label string, attrs Attrs) (INode, error)

	// Removes one node
	Remove(id NodeID) error

	// Query
	Get(label string) (IQueryBuilder, error)
	Build() IQueryBuilder

	// Links
	Link(from, to INode, attrs Attrs) error
}

type ICollection interface {
	Add(attrs Attrs) (INode, error)
	Get(attrs Attrs) (INode, error)
	Has(attrs Attrs) (bool, error)
	Remove(node INode) error
	Link(from, to INode, attrs Attrs) error
	Unlink(from, to INode, attrs Attrs) error
}

type IStorage interface {
	ICollection

	/*
		Create(name string) (ICollection, error)
		Get(name string) (ICollection, error)
		Has(name string) (bool, error)
		Remove(name string) (bool, error)
	*/

	Close() error
}

/*
Starting points:
.V([name], [fuzzyness]) returns all vertexes with name (name can be nil for ALL vertexes in the graph)
.M()					starts a morphism (not finalizable)

Path traversals:
.In(type, [Attrs])		returns all vertexes pointing TOWARDS the subject
.Out(type, [Attrs])		returns all vertexes in the opposite direction
.Both()					both directions (in+out)
.Is(name...)			returns (filters) all paths with the given vertexes with name...

Joining:
.Intersect(q)
.Union(q)

Applying morphisms:
.Follow(m)

Finalizers:
.Count()				count the resulting vertexes
.Get(n)					return n vertexes from the resulting set in no particular order
.All()					return all resulting vertexes
.ForEach(...) ?
*/
