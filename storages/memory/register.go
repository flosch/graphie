package happy

import (
	"github.com/flosch/graphie"
)

func registerMemory(g *graphie.Graph) (graphie.Storage, error) {
	return &storage{g: g}, nil
}

func init() {
	graphie.RegisterDriver("memory", registerMemory)
}
