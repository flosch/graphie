package graphie

type LabelGroup struct {
	g      *Graph
	labels []string
}

func (lg *LabelGroup) Add(attrs Attrs) (NodeID, error) {
	return lg.g.s.Add(lg.labels, attrs)
}

func (lg *LabelGroup) MustAdd(attrs Attrs) NodeID {
	n, err := lg.Add(attrs)
	if err != nil {
		panic(err)
	}
	return n
}

func (lg *LabelGroup) Merge(attrs Attrs) (NodeID, error) {
	return lg.g.s.Merge(lg.labels, attrs)
}

func (lg *LabelGroup) EnsureIndexNodes(attr_name string) error {
	return lg.g.s.EnsureIndexNodes(lg.labels, attr_name)
}

func (lg *LabelGroup) EnsureIndexLinks(attr_name string) error {
	return lg.g.s.EnsureIndexLinks(lg.labels, attr_name)
}

func (lg *LabelGroup) Query() *Query {
	return &Query{
		g: lg.g,
	}
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
