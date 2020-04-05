package sat

import (
	"fmt"
	"strings"

	"github.com/irifrance/gini/z"
)

// dict performs translation between the input and output types of
// Solve (Constraints, Installables, etc.) and the variables that
// appear in the SAT formula.
type dict struct {
	installables []Installable
	constraints  map[z.Lit]AppliedConstraint
	indices      map[Identifier]int
	errs         dictError
	next         int
}

type dictError []error

func (dictError) Error() string {
	return "internal solver failure"
}

// LitOf returns the positive literal corresponding to the Installable
// with the given Identifier.
func (d *dict) LitOf(id Identifier) z.Lit {
	index, ok := d.indices[id]
	if !ok {
		d.errs = append(d.errs, fmt.Errorf("installable %q referenced but not provided", id))
		return z.LitNull
	}
	return z.Var(index + 1).Pos()
}

type zeroInstallable struct{}

func (zeroInstallable) Identifier() Identifier {
	return ""
}

func (zeroInstallable) Constraints() []Constraint {
	return nil
}

// InstallableOf returns the Installable corresponding to the provided
// literal, or a zeroInstallable if no such Installable exists.
func (d *dict) InstallableOf(m z.Lit) Installable {
	v := int(m.Var())
	index := v - 1
	if index < 0 || index >= len(d.indices) {
		d.errs = append(d.errs, fmt.Errorf("no installable corresponding to %s", m))
		return zeroInstallable{}
	}
	return d.installables[index]
}

// ConstraintOf returns the constraint application corresponding to
// the provided literal, or a zeroConstraint if no such constraint
// exists.
func (d *dict) ConstraintOf(m z.Lit) AppliedConstraint {
	if a, ok := d.constraints[m]; ok {
		return a
	}
	d.errs = append(d.errs, fmt.Errorf("no constraint corresponding to %s", m))
	return AppliedConstraint{
		Installable: zeroInstallable{},
		Constraint:  zeroConstraint{},
	}
}

// Error returns a single error value that is an aggregation of all
// errors encountered during a dict's lifetime, or nil if there have
// been no errors. A non-nil return value likely indicates a problem
// with the solver or constraint implementations.
func (d *dict) Error() error {
	if len(d.errs) == 0 {
		return nil
	}
	s := make([]string, len(d.errs))
	for i, err := range d.errs {
		s[i] = err.Error()
	}
	return fmt.Errorf("%d errors encountered: %s", len(s), strings.Join(s, ", "))
}

// FreeLit returns an unused literal. Constraints that need to
// introduce a new literal should get it from this method to avoid
// accidental reuse.
func (d *dict) FreeLit() z.Lit {
	m := z.Var(d.next).Pos()
	d.next++
	return m
}

func compileDict(installables []Installable) *dict {
	d := dict{
		installables: installables,
		constraints:  make(map[z.Lit]AppliedConstraint),
		indices:      make(map[Identifier]int, len(installables)),
		next:         len(installables) + 1,
	}

	for index, installable := range installables {
		d.indices[installable.Identifier()] = index
	}

	return &d
}
