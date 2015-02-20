package mongo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flosch/graphie"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type mongodbStorage struct {
	g          *graphie.Graph
	hostname   string
	session    *mgo.Session
	coll_nodes *mgo.Collection
	coll_edges *mgo.Collection
}

func (s *mongodbStorage) Init(hostname, dbname string) error {
	sess, err := mgo.Dial(s.hostname)
	if err != nil {
		return err
	}
	s.session = sess
	s.coll_nodes = sess.DB(dbname).C("nodes")
	s.coll_edges = sess.DB(dbname).C("edges")

	s.coll_edges.EnsureIndexKey("_from")
	s.coll_edges.EnsureIndexKey("_to")

	return nil
}

func (s *mongodbStorage) Shutdown() error {
	s.session.Close()
	return nil
}

func (s *mongodbStorage) Add(labels []string, attrs graphie.Attrs) (graphie.Node, error) {
	id := bson.NewObjectId()

	d := bson.M{
		"_id": id,
	}

	for _, lbl := range labels {
		d[fmt.Sprintf("_lbl_%s", lbl)] = true
	}

	if attrs != nil {
		for k, v := range attrs {
			if strings.HasPrefix(k, "_") {
				return nil, errors.New("Attribute with '_'-prefix is not allowed.")
			}
			d[k] = v
		}
	}

	err := s.coll_nodes.Insert(d)
	if err != nil {
		return nil, err
	}

	return &graphie.Node{
		G:      s.g,
		Labels: labels,
		Id:     string(id),
		Attrs:  attrs,
	}, nil
}

func (s *mongodbStorage) Merge(labels []string, attrs graphie.Attrs) (graphie.Node, error) {
	d := make(bson.M)

	for _, lbl := range labels {
		d[fmt.Sprintf("_lbl_%s", lbl)] = true
	}

	if attrs != nil {
		for k, v := range attrs {
			if strings.HasPrefix(k, "_") {
				return nil, errors.New("Attribute with '_'-prefix is not allowed.")
			}
			d[k] = v
		}
	}

	inf, err := s.coll_nodes.Upsert(d, d)
	if err != nil {
		return nil, err
	}
	if inf.UpsertedId != nil {
		// We have an ID
		return &graphie.Node{
			G:      s.g,
			Labels: labels,
			Id:     string(inf.UpsertedId.(bson.ObjectId)),
			Attrs:  attrs,
		}, nil
	} else {
		// We have to fetch the ID
		var t bson.M
		err = s.coll_nodes.Find(d).One(&t)
		if err != nil {
			return nil, err
		}

		return &graphie.Node{
			G:      s.g,
			Labels: labels,
			Id:     string(t["_id"].(bson.ObjectId)),
			Attrs:  attrs,
		}, nil

	}
}

func (s *mongodbStorage) Link(labels []string, from, to *graphie.Node, attrs graphie.Attrs) error {
	id := bson.NewObjectId()

	d := bson.M{
		"_id":   id,
		"_from": bson.ObjectId(from.Id),
		"_to":   bson.ObjectId(to.Id),
	}

	for _, lbl := range labels {
		d[fmt.Sprintf("_lbl_%s", lbl)] = true
	}

	if attrs != nil {
		for k, v := range attrs {
			if strings.HasPrefix(k, "_") {
				return nil, errors.New("Attribute with '_'-prefix is not allowed.")
			}
			d[k] = v
		}
	}

	err := s.coll_edges.Insert(d)
	if err != nil {
		return nil, err
	}

	return nil

}

func (s *mongodbStorage) EnsureIndexNodes(labels []string, attr_name string) error {
	if strings.HasPrefix(attr_name, "_") {
		return errors.New("Attribute with '_'-prefix is not allowed.")
	}

	keys := make([]string, 0, 1)

	keys = append(keys, attr_name)

	for _, lbl := range labels {
		keys = append(keys, fmt.Sprintf("_lbl_%s", lbl))
	}

	idx := mgo.Index{
		Key:    keys,
		Sparse: true,
	}

	return s.coll_nodes.EnsureIndex(idx)
}

func (s *mongodbStorage) EnsureIndexLinks(labels []string, attr_name string) error {
	if strings.HasPrefix(attr_name, "_") {
		return errors.New("Attribute with '_'-prefix is not allowed.")
	}

	keys := make([]string, 0, 1)
	keys = append(keys, attr_name)

	for _, lbl := range labels {
		keys = append(keys, fmt.Sprintf("_lbl_%s", lbl))
	}

	idx := mgo.Index{
		Key:    keys,
		Sparse: true,
	}

	return s.coll_edges.EnsureIndex(idx)
}

func registerMongodb(g *graphie.Graph) (graphie.Storage, error) {
	return &mongodbStorage{
		g: g,
	}, nil
}

func init() {
	// RegisterDriver
	graphie.RegisterDriver("mongodb", registerMongodb)
}
