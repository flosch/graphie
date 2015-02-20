package main

import (
	//"encoding/csv"
	//"fmt"
	"log"

	"github.com/flosch/graphie"
	_ "github.com/flosch/graphie/storages/memory"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	my_graph, err := graphie.NewGraph("memory", "testdb", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := my_graph.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Create two labels
	date := my_graph.Labels("date")
	category := my_graph.Labels("category")
	person := my_graph.Labels("person")

	// Create indexes for each label; we're querying on these attributes
	must(category.EnsureIndexNodes("name"))
	// must(persons.EnsureIndexNodes("fullname"))
	//must(categories.EnsureIndexNodes

	// Create all nodes
	science := category.MustAdd(graphie.Attrs{"name": "Science"})
	formal_science := category.MustAdd(graphie.Attrs{"name": "Formal science"})
	maths := category.MustAdd(graphie.Attrs{"name": "Mathematics"})
	number_theory := category.MustAdd(graphie.Attrs{"name": "Number theory"})
	analysis := category.MustAdd(graphie.Attrs{"name": "Analysis"})
	algebra := category.MustAdd(graphie.Attrs{"name": "Algebra"})
	law := category.MustAdd(graphie.Attrs{"name": "Law"})
	theology := category.MustAdd(graphie.Attrs{"name": "Theology"})

	y_1665 := date.MustAdd(graphie.Attrs{"year": 1665})
	y_1918 := date.MustAdd(graphie.Attrs{"year": 1918})
	d_16january := date.MustAdd(graphie.Attrs{"day": 16, "month": 1})

	cantor := person.MustAdd(graphie.Attrs{"fullname": "Georg Cantor"})
	fermat := person.MustAdd(graphie.Attrs{"fullname": "Pierre de Fermat"})
	hilbert := person.MustAdd(graphie.Attrs{"fullname": "David Hilbert"})
	pappus := person.MustAdd(graphie.Attrs{"fullname": "Johannes Pappus"})

	// Create connections
	must(law.Link(science, graphie.Attrs{"name": "instance_of"}))
	must(theology.Link(science, graphie.Attrs{"name": "instance_of"}))
	must(formal_science.Link(science, graphie.Attrs{"name": "instance_of"}))
	must(maths.Link(formal_science, graphie.Attrs{"name": "instance_of"}))
	must(maths.Link(number_theory, graphie.Attrs{"name": "contains"}))
	must(maths.Link(analysis, graphie.Attrs{"name": "contains"}))
	must(maths.Link(algebra, graphie.Attrs{"name": "contains"}))
	must(cantor.Link(maths, graphie.Attrs{"name": "field_of_profession"}))
	must(fermat.Link(maths, graphie.Attrs{"name": "field_of_profession"}))
	must(fermat.Link(law, graphie.Attrs{"name": "field_of_profession"}))
	must(hilbert.Link(maths, graphie.Attrs{"name": "field_of_profession"}))
	must(fermat.Link(y_1665, graphie.Attrs{"name": "date_of_death"}))
	must(cantor.Link(y_1918, graphie.Attrs{"name": "date_of_death"}))
	must(cantor.Link(d_16january, graphie.Attrs{"name": "date_of_death"}))
	must(pappus.Link(d_16january, graphie.Attrs{"name": "date_of_birth"}))
	must(pappus.Link(theology, graphie.Attrs{"name": "field_of_profession"}))

	// Query

	// Given the number theory, get mathematicians
	/* mathematicians := category.Nodes(graphie.Attrs{"name": "number_theory"}).
		In(graphie.Attrs{"name": "contains"}).
		In(graphie.Attrs{"name": "field_of_profession"}).All()

	for _, mathematician := range mathematicians {
		fmt.Println(mathematician)
	} */

	// What field of professions did fermat had?
	// person.QueryNode(fermat).Out(nil).HasLabel("category")

	// Who is researching on mathematician and at least one other

	// What is the date of death of fermat?

	// What happened on january 12th?

	// Which sciences are there? (law, theology, mathematics), ...

	// Which subfields does mathematics have?

	// Get all persons with their year of birth and their field of professions
	/* prepared_path := person.Edges(graphie.Attrs{"name": "field_of_profession"}).
		Union(person.Query().Out("date_of_birth"))
	fmt.Println(person.QueryNodes(nil).Follow(prepared_path)) */
}
