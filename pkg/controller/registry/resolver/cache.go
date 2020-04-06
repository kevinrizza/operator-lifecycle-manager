package resolver

import (
	"github.com/blang/semver"
)

type CatalogDependencyCache interface {
	FindPackageState()
}

/*
type Bundle struct {
	CatalogSource CatalogKey
	//	CsvName       string
	Version semver.Version
	Package string
	//	Channels     []string
	//	BundleImage  string
	ProvidedApis []*v1.GroupVersionKind
	RequiredApis []*v1.GroupVersionKind
	//	Dependecies  []Dependency
	//  Provides  []Provide //TODO: decide if we want this
}
*/

type VersionDependency struct {
	Package string
	Version semver.Version
	// TODO: BundleImage string
	// TODO: VersionRange string
}

func (d *VersionDependency) CanBeSatisfiedBy(bundle Operator) bool {
	var bundleVersion semver.Version
	if bundle.Version() != nil {
		bundleVersion = *bundle.Version()
	}

	return (d.Version.Equals(bundleVersion)) && (d.Package == bundle.Package())
}

type Cache struct {
	packages []Operator
}

func (c *Cache) FindPackageState() []Operator {
	return c.packages
}

// TODO: implement resync
// func Resync(source CatalogKey)
