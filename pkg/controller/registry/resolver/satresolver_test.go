package resolver

import (
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"

	"github.com/operator-framework/operator-registry/pkg/api"
	"github.com/operator-framework/operator-registry/pkg/registry"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
)

func TestSolveOperators(t *testing.T) {
	APISet := APISet{registry.APIKey{"g", "v", "k", "ks"}: struct{}{}}
	Provides := APISet

	namespace := "ns1"
	catalog := CatalogKey{"catsrc", namespace}

	csv := existingOperator(namespace, "packageA.v1", "packageA", "alpha", "", Provides, nil, nil, nil)
	csvs := []*v1alpha1.ClusterServiceVersion{csv}
	sub := existingSub(namespace, "packageA.v1", "packageA", "alpha", catalog)
	subs := []*v1alpha1.Subscription{sub}

	opToAdd := OperatorSourceInfo{
		Package: "packageB",
		Channel: "alpha",
		Catalog: catalog,
	}
	opsToAdd := map[OperatorSourceInfo]struct{}{
		opToAdd: struct{}{},
	}

	satResolver := SatResolver{
		cache: Cache{
			operatorCatalogs: map[string][]Operator{
				"olm/community": []Operator{
					genOperator("packageA.v1", "0.0.1", "packageA", "community", "olm", nil),
					genOperator("packageB.v1", "1.0.0", "packageB", "community", "olm", nil),
				},
			},
		},
	}

	operators, err := satResolver.SolveOperators(csvs, subs, opsToAdd)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(operators))
}

func TestSolveOperators_WithDependencies(t *testing.T) {
	APISet := APISet{registry.APIKey{"g", "v", "k", "ks"}: struct{}{}}
	Provides := APISet

	namespace := "ns1"
	catalog := CatalogKey{"catsrc", namespace}

	csv := existingOperator(namespace, "packageA.v1", "packageA", "alpha", "", Provides, nil, nil, nil)
	csvs := []*v1alpha1.ClusterServiceVersion{csv}
	sub := existingSub(namespace, "packageA.v1", "packageA", "alpha", catalog)
	subs := []*v1alpha1.Subscription{sub}

	opToAdd := OperatorSourceInfo{
		Package: "packageB",
		Channel: "alpha",
		Catalog: catalog,
	}
	opsToAdd := map[OperatorSourceInfo]struct{}{
		opToAdd: struct{}{},
	}
	depVersion, _ := semver.Make("0.1.0")
	opToAddVersionDeps := []VersionDependency{
		VersionDependency{
			Package: "packageC",
			Version: depVersion,
		},
	}

	satResolver := SatResolver{
		cache: Cache{
			operatorCatalogs: map[string][]Operator{
				"olm/community": []Operator{
					genOperator("packageA.v1", "0.0.1", "packageA", "community", "olm", nil),
					genOperator("packageB.v1", "1.0.0", "packageB", "community", "olm", opToAddVersionDeps),
					genOperator("packageC.v1", "0.1.0", "packageC", "community", "olm", nil),
				},
			},
		},
	}

	operators, err := satResolver.SolveOperators(csvs, subs, opsToAdd)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(operators))
}

func genOperator(name, version, pkg, catalogName, catalogNamespace string, versionDependencies []VersionDependency) Operator {
	semversion, _ := semver.Make(version)
	return Operator{
		name:    name,
		version: &semversion,
		bundle: &api.Bundle{
			PackageName: pkg,
		},
		versionDependencies: versionDependencies,
		sourceInfo: &OperatorSourceInfo{
			Catalog: CatalogKey{
				Name:      catalogName,
				Namespace: catalogNamespace,
			},
		},
	}
}
