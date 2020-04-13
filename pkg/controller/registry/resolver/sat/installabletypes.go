package sat

import (
	"fmt"
	"strings"
)

type BundleInstallable struct {
	identifier  Identifier
	constraints []Constraint
}

func (i BundleInstallable) Identifier() Identifier {
	return i.identifier
}

func (i BundleInstallable) Constraints() []Constraint {
	return i.constraints
}

func (i *BundleInstallable) AddDependency(dependencies []Identifier) {
	i.constraints = append(i.constraints, dependency(dependencies))
}

func (i *BundleInstallable) BundleSourceInfo() (string, string, string, error) {
	info := strings.Split(string(i.identifier), "/")
	if len(info) != 3 { // TODO: enforce this? does it even make sense? Should just redefine identifier type?
		return "", "", "", fmt.Errorf("Unable to parse identifier %s for source info", i.identifier)
	}
	return info[0], info[1], info[2], nil
}

// TODO: Identify this based on combination of CSV name and catalog
func NewBundleInstallable(bundle, catalogNamespace, catalogName string, constraints ...Constraint) BundleInstallable {
	return BundleInstallable{
		identifier:  Identifier(fmt.Sprintf("%s/%s/%s", catalogNamespace, catalogName, bundle)),
		constraints: constraints,
	}
}

type VirtPackageInstallable struct {
	identifier  Identifier
	constraints []Constraint
}

func (v VirtPackageInstallable) Identifier() Identifier {
	return v.identifier
}

func (v VirtPackageInstallable) Constraints() []Constraint {
	return v.constraints
}

func (v *VirtPackageInstallable) AddDependency(dependencies []Identifier) {
	v.constraints = append(v.constraints, dependency(dependencies))
}

func NewVirtualPackageInstallable(bundle string) VirtPackageInstallable {
	return VirtPackageInstallable{
		identifier:  Identifier(bundle),
		constraints: []Constraint{mandatory{}},
	}
}
