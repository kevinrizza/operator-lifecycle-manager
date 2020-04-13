package resolver

import (
	"fmt"

	"github.com/blang/semver"
)

type CatalogDependencyCache interface {
	GetCatalogs() map[string][]Operator
	GetCSVNameFromCatalog(csvName, catalogName, catalogNamespace string) (Operator, error)
	GetCSVNameFromAllCatalogs(csvName string) ([]Operator, error)
	GetPackageFromAllCatalogs(pkg string) ([]Operator, error)
	GetPackageVersionFromAllCatalogs(pkg string, version semver.Version) ([]Operator, error)
	Resync()
}

type VersionDependency struct {
	Package string
	Version semver.Version
	// TODO: BundleImage string
	// TODO: VersionRange string
}

// TODO: Do we need this at all?
func (d *VersionDependency) CanBeSatisfiedBy(bundle Operator) bool {
	var bundleVersion semver.Version
	if bundle.Version() != nil {
		bundleVersion = *bundle.Version()
	}

	return (d.Version.Equals(bundleVersion)) && (d.Package == bundle.Package())
}

type Cache struct {
	operatorCatalogs map[string][]Operator
}

func (c *Cache) GetCatalogs() map[string][]Operator {
	return c.operatorCatalogs
}

/*

func (c *Cache) GetCatalog(catalog string) ([]Operator, error) {
	if pkgOperators, ok := c.operatorCatalogs[catalog]; ok {
		return pkgOperators, nil
	}

	return fmt.Errorf("Unable to find package in catalog")
}
*/

func (c *Cache) GetCSVNameFromCatalog(csvName, catalogName, catalogNamespace string) (Operator, error) {
	catalogKey := fmt.Sprintf("%s/%s", catalogNamespace, catalogName)
	catalog := c.operatorCatalogs[catalogKey]
	for _, op := range catalog {
		if op.Identifier() == csvName {
			return op, nil
		}
	}

	return Operator{}, fmt.Errorf("Unable to find csv %s in catalog %s", csvName, catalogKey)
}

func (c *Cache) GetCSVNameFromAllCatalogs(csvName string) ([]Operator, error) {
	operators := make([]Operator, 0)

	for _, catalog := range c.operatorCatalogs {
		for _, op := range catalog {
			if op.Identifier() == csvName {
				operators = append(operators, op)
			}
		}
	}

	if len(operators) == 0 {
		return nil, fmt.Errorf("Unable to find csv %s in any catalog", csvName)
	}

	return operators, nil
}

func (c *Cache) GetPackageVersionFromAllCatalogs(pkg string, version semver.Version) ([]Operator, error) {
	operators := make([]Operator, 0)

	for _, catalog := range c.operatorCatalogs {
		for _, op := range catalog {
			opVersion := op.Version()
			if op.Package() == pkg && opVersion.Equals(version) {
				operators = append(operators, op)
			}
		}
	}

	if len(operators) == 0 {
		return nil, fmt.Errorf("Unable to find package in any catalog")
	}

	return operators, nil
}

func (c *Cache) GetPackageFromAllCatalogs(pkg string) ([]Operator, error) {
	operators := make([]Operator, 0)

	for _, catalog := range c.operatorCatalogs {
		for _, op := range catalog {
			if op.Package() == pkg {
				operators = append(operators, op)
			}
		}
	}

	if len(operators) == 0 {
		return nil, fmt.Errorf("Unable to find package in any catalog")
	}

	return operators, nil
}

// TODO: implement real resync
// func Resync(source CatalogKey) Cache

func (c *Cache) Resync() {
	c.operatorCatalogs = make(map[string][]Operator)
}
