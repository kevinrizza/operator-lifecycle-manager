package sat

import (
	"fmt"
	"strings"

	"github.com/irifrance/gini/z"
)

type solver interface {
	ActivateWith(act z.Lit)
	Add(m z.Lit)
	Assume(ms ...z.Lit)
}

type Constraint interface {
	String(subject Identifier) string
	apply(s solver, d *dict, subject Identifier) z.Lit
}

type AppliedConstraint struct {
	Installable Installable
	Constraint  Constraint
}

func (a AppliedConstraint) String() string {
	return a.Constraint.String(a.Installable.Identifier())
}

type zeroConstraint struct{}

func (zeroConstraint) String(subject Identifier) string {
	return ""
}

func (zeroConstraint) apply(s solver, d *dict, subject Identifier) z.Lit {
	return z.LitNull
}

type mandatory struct{}

func (c mandatory) String(subject Identifier) string {
	return fmt.Sprintf("%s is mandatory", subject)
}

func (c mandatory) apply(s solver, d *dict, subject Identifier) z.Lit {
	m := d.LitOf(subject)
	s.Assume(m)
	return m
}

type prohibited struct{}

func (c prohibited) String(subject Identifier) string {
	return fmt.Sprintf("%s is prohibited", subject)
}

func (c prohibited) apply(s solver, d *dict, subject Identifier) z.Lit {
	m := d.LitOf(subject).Not()
	s.Assume(m)
	return m
}

type dependency []Identifier

func (c dependency) String(subject Identifier) string {
	s := make([]string, len(c))
	for i, each := range c {
		s[i] = string(each)
	}
	return fmt.Sprintf("%s requires at least one of %s", subject, strings.Join(s, ", "))
}

func (c dependency) apply(s solver, d *dict, subject Identifier) z.Lit {
	s.Add(d.LitOf(subject).Not())
	for _, each := range c {
		s.Add(d.LitOf(Identifier(each)))
	}
	m := d.FreeLit()
	s.ActivateWith(m)
	s.Assume(m)
	return m
}

type conflict Identifier

func (c conflict) String(subject Identifier) string {
	return fmt.Sprintf("%s conflicts with %s", subject, Identifier(c))
}

func (c conflict) apply(s solver, d *dict, subject Identifier) z.Lit {
	s.Add(d.LitOf(subject).Not())
	s.Add(d.LitOf(Identifier(c)).Not())
	m := d.FreeLit()
	s.ActivateWith(m)
	s.Assume(m)
	return m

}
