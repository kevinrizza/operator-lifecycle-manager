package resolver

import (
	//"fmt"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver/sat"
)

type SatResolver struct {
	cache Cache
}

func (s *SatResolver) SolveOperators(csvs []*v1alpha1.ClusterServiceVersion, subs []*v1alpha1.Subscription, add map[OperatorSourceInfo]struct{}) (OperatorSet, error) {
	var errs []error

	installables := make([]sat.Installable, 0)

	// for each subscription and operator to add, create a mandatory virtual "package" installable that has a
	// dependency on the bundles that could be used to install it
	// then make an installable for each version of that bundle along with its set of dependency constraints
	// then make an installable for each version of the dependency

	for _, sub := range subs {
		pkg := sub.Spec.Package
		packageInstallables, err := s.getPackageInstallables(pkg)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for _, pkgInstallable := range packageInstallables {
			installables = append(installables, pkgInstallable)
		}
	}

	for opToAdd := range add {
		pkg := opToAdd.Package

		packageInstallables, err := s.getPackageInstallables(pkg)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for _, pkgInstallable := range packageInstallables {
			installables = append(installables, pkgInstallable)
		}
		// todo: condition on starting csv
	}

	if len(errs) > 0 {
		return nil, utilerrors.NewAggregate(errs)
	}

	// TODO: Consider csvs not attached to subscriptions
	/*
		for _, csv := range csvs {

			// TODO: Mandatory
			installables = append(installables, sat.NewBundleInstallable(csv.Name))
		}
	*/

	solvedInstallables, err := sat.Solve(installables)
	if err != nil {
		return nil, err
	}

	// get the set of bundle installables from the result solved installables
	operatorInstallables := make([]sat.BundleInstallable, 0)
	for _, installable := range solvedInstallables {
		if bundleInstallable, ok := installable.(sat.BundleInstallable); ok {
			operatorInstallables = append(operatorInstallables, bundleInstallable)
		}
	}

	operators := make(map[string]OperatorSurface, 0)
	for _, installableOperator := range operatorInstallables {
		catalogName, catalogNamespace, csvName, err := installableOperator.BundleSourceInfo()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		op, err := s.cache.GetCSVNameFromCatalog(csvName, catalogName, catalogNamespace)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		operators[csvName] = &op
	}

	if len(errs) > 0 {
		return nil, utilerrors.NewAggregate(errs)
	}

	return operators, nil
}

func (s *SatResolver) getPackageInstallables(pkg string) ([]sat.Installable, error) {
	installables := make([]sat.Installable, 0)

	virtualInstallable := sat.NewVirtualPackageInstallable(pkg)

	bundles, err := s.cache.GetPackageFromAllCatalogs(pkg)
	if err != nil {
		return installables, err
	}

	virtDependencies := make([]sat.Identifier, 0)
	// add installable for each bundle version of the package
	for _, bundle := range bundles {
		// add a bundle installable
		bundleIdentifiers, bundleInstallables, err := s.getBundleInstallables(bundle.Identifier())
		if err != nil {
			// todo: return aggregate error
			return nil, err
		}

		for _, bundleInstallable := range bundleInstallables {
			installables = append(installables, bundleInstallable)
		}

		for _, bundleIdentifier := range bundleIdentifiers {
			virtDependencies = append(virtDependencies, bundleIdentifier)
		}
	}

	virtualInstallable.AddDependency(virtDependencies)
	installables = append(installables, virtualInstallable)

	return installables, nil
}

func (s *SatResolver) getBundleInstallables(csvName string) ([]sat.Identifier, []sat.Installable, error) {
	var errs []error
	installables := make([]sat.Installable, 0)
	identifiers := make([]sat.Identifier, 0)

	bundles, err := s.cache.GetCSVNameFromAllCatalogs(csvName)
	if err != nil {
		return nil, nil, err
	}

	for _, bundle := range bundles {
		bundleCatalog := bundle.SourceInfo().Catalog // TODO: Nil check here
		bundleInstallable := sat.NewBundleInstallable(csvName, bundleCatalog.Name, bundleCatalog.Namespace)
		for _, depVersion := range bundle.VersionDependencies() {
			depCandidates, err := s.cache.GetPackageVersionFromAllCatalogs(depVersion.Package, depVersion.Version)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			bundleDependencies := make([]sat.Identifier, 0)
			for _, dep := range depCandidates {
				depIdentifiers, depInstallables, err := s.getBundleInstallables(dep.Identifier())
				if err != nil {
					errs = append(errs, err)
					continue
				}
				for _, depInstallable := range depInstallables {
					installables = append(installables, depInstallable)
				}
				for _, depIdentifier := range depIdentifiers {
					bundleDependencies = append(bundleDependencies, depIdentifier)
				}
			}
			bundleInstallable.AddDependency(bundleDependencies)
		}
		installables = append(installables, bundleInstallable)
		identifiers = append(identifiers, bundleInstallable.Identifier())
	}

	if len(errs) > 0 {
		return nil, nil, utilerrors.NewAggregate(errs)
	}

	return identifiers, installables, nil
}
