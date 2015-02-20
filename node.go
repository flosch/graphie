package graphie

/*
type Node struct {
	G      *Graph
	Labels []string

	Id    string
	Attrs Attrs
}

func (n *Node) Link(other *Node, attrs Attrs) (*Link, error) {
	return n.G.s.Link(n.Labels, n, other, attrs)
}
*/

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
