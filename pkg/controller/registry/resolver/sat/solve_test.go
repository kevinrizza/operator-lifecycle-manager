package sat

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestInstallable struct {
	identifier  Identifier
	constraints []Constraint
}

func (i TestInstallable) Identifier() Identifier {
	return i.identifier
}

func (i TestInstallable) Constraints() []Constraint {
	return i.constraints
}

func installable(id Identifier, constraints ...Constraint) Installable {
	return TestInstallable{
		identifier:  id,
		constraints: constraints,
	}
}

func TestNotSatisfiableError(t *testing.T) {
	type tc struct {
		Name   string
		Error  NotSatisfiable
		String string
	}

	for _, tt := range []tc{
		{
			Name:   "nil",
			String: "constraints not satisfiable",
		},
		{
			Name:   "empty",
			String: "constraints not satisfiable",
		},
		{
			Name: "single failure",
			Error: NotSatisfiable{
				AppliedConstraint{
					Installable: installable("a", mandatory{}),
					Constraint:  mandatory{},
				},
			},
			String: fmt.Sprintf("constraints not satisfiable: %s",
				mandatory{}.String("a")),
		},
		{
			Name: "multiple failures",
			Error: NotSatisfiable{
				AppliedConstraint{
					Installable: installable("a", mandatory{}),
					Constraint:  mandatory{},
				},
				AppliedConstraint{
					Installable: installable("b", prohibited{}),
					Constraint:  prohibited{},
				},
			},
			String: fmt.Sprintf("constraints not satisfiable: %s, %s",
				mandatory{}.String("a"), prohibited{}.String("b")),
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			assert.Equal(t, tt.String, tt.Error.Error())
		})
	}
}

func TestSolve(t *testing.T) {
	type tc struct {
		Name         string
		Installables []Installable
		Installed    []Installable
		Error        error
	}

	for _, tt := range []tc{
		{
			Name: "no installables",
		},
		{
			Name:         "single unconstrained installable",
			Installables: []Installable{installable("a")},
		},
		{
			Name:         "single preinstalled installable",
			Installables: []Installable{installable("a", mandatory{})},
			Installed:    []Installable{installable("a", mandatory{})},
		},
		{
			Name:         "both mandatory and prohibited",
			Installables: []Installable{installable("a", mandatory{}, prohibited{})},
			Error: NotSatisfiable{
				{
					Installable: installable("a", mandatory{}, prohibited{}),
					Constraint:  mandatory{},
				},
				{
					Installable: installable("a", mandatory{}, prohibited{}),
					Constraint:  prohibited{},
				},
			},
		},
		{
			Name: "dependency is installed",
			Installables: []Installable{
				installable("a"),
				installable("b", mandatory{}, dependency{"a"}),
			},
			Installed: []Installable{
				installable("a"),
				installable("b", mandatory{}, dependency{"a"}),
			},
		},
		{
			Name: "transitive dependency is installed",
			Installables: []Installable{
				installable("a"),
				installable("b", dependency{"a"}),
				installable("c", mandatory{}, dependency{"b"}),
			},
			Installed: []Installable{
				installable("a"),
				installable("b", dependency{"a"}),
				installable("c", mandatory{}, dependency{"b"}),
			},
		},
		{
			Name: "both dependencies are installed",
			Installables: []Installable{
				installable("a"),
				installable("b", dependency{"a"}),
				installable("c", mandatory{}, dependency{"a"}, dependency{"b"}),
			},
			Installed: []Installable{
				installable("a"),
				installable("b", dependency{"a"}),
				installable("c", mandatory{}, dependency{"a"}, dependency{"b"}),
			},
		},
		{
			Name: "conflicting dependency with earliest input position is avoided",
			Installables: []Installable{
				installable("a"),
				installable("b", conflict("a")),
				installable("c", mandatory{}, dependency{"a", "b"}),
			},
			Installed: []Installable{
				installable("b", conflict("a")),
				installable("c", mandatory{}, dependency{"a", "b"}),
			},
		},
		{
			Name: "two mandatory but conflicting packages",
			Installables: []Installable{
				installable("a", mandatory{}),
				installable("b", mandatory{}, conflict("a")),
			},
			Error: NotSatisfiable{
				{
					Installable: installable("a", mandatory{}),
					Constraint:  mandatory{},
				},
				{
					Installable: installable("b", mandatory{}, conflict("a")),
					Constraint:  mandatory{},
				},
				{
					Installable: installable("b", mandatory{}, conflict("a")),
					Constraint:  conflict("a"),
				},
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			assert := assert.New(t)

			installed, err := Solve(tt.Installables)

			if installed != nil {
				sort.SliceStable(installed, func(i, j int) bool {
					return installed[i].Identifier() < installed[j].Identifier()
				})
			}

			// Failed constraints are sorted in lexically
			// increasing order of the identifier of the
			// constraint's installable, with ties broken
			// in favor of the constraint that appears
			// earliest in the installable's list of
			// constraints.
			if ns, ok := err.(NotSatisfiable); ok {
				sort.SliceStable(ns, func(i, j int) bool {
					if ns[i].Installable.Identifier() != ns[j].Installable.Identifier() {
						return ns[i].Installable.Identifier() < ns[j].Installable.Identifier()
					}
					var x, y int
					for i, c := range ns[i].Installable.Constraints() {
						if reflect.DeepEqual(c, ns[i].Constraint) {
							x = i
							break
						}
					}
					for i, c := range ns[j].Installable.Constraints() {
						if reflect.DeepEqual(c, ns[j].Constraint) {
							y = i
							break
						}
					}
					return x < y
				})
			}

			assert.Equal(tt.Installed, installed)
			assert.Equal(tt.Error, err)
		})
	}
}
