package sat

import (
	"fmt"
	"strings"

	"github.com/irifrance/gini"
)

type Identifier string

type Installable interface {
	Identifier() Identifier
	Constraints() []Constraint
}

type NotSatisfiable []AppliedConstraint

func (e NotSatisfiable) Error() string {
	const msg = "constraints not satisfiable"
	if len(e) == 0 {
		return msg
	}
	s := make([]string, len(e))
	for i, a := range e {
		s[i] = a.String()
	}
	return fmt.Sprintf("%s: %s", msg, strings.Join(s, ", "))
}

const (
	satisfiable   = 1
	unsatisfiable = -1
)

// Solve takes a slice containing all Installables and returns a slice
// containing only those Installables that were selected for
// installation. If no solution is possible, an error is returned.
func Solve(installables []Installable) (result []Installable, err error) {
	d := compileDict(installables)
	defer func() {
		// This likely indicates a bug, so discard whatever
		// return values were produced.
		if derr := d.Error(); derr != nil {
			result = nil
			err = derr
		}
	}()

	g := gini.New()

	for _, i := range installables {
		for _, c := range i.Constraints() {
			m := c.apply(g, d, i.Identifier())
			d.constraints[m] = AppliedConstraint{i, c}
		}
	}

	switch g.Solve() {
	case satisfiable:
		var result []Installable
		for _, i := range installables {
			if g.Value(d.LitOf(i.Identifier())) {
				result = append(result, i)
			}
		}
		return result, nil
	case unsatisfiable:
		whys := g.Why(nil)
		as := make([]AppliedConstraint, len(whys))
		for i, why := range whys {
			as[i] = d.ConstraintOf(why)
		}
		return nil, NotSatisfiable(as)
	}

	return nil, fmt.Errorf("failed to solve in the allotted time")
}
