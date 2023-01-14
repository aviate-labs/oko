package vessel

import "fmt"

type DuplicatePackageNameError struct {
	pkg, duplicate Package
}

func DuplicatePackageName(pkg Package, duplicate Package) *DuplicatePackageNameError {
	return &DuplicatePackageNameError{
		pkg:       pkg,
		duplicate: duplicate,
	}
}

func (e DuplicatePackageNameError) Error() string {
	return fmt.Sprintf(
		"duplicate package name %q: got %q and %q",
		e.pkg.Name,
		fmt.Sprintf("%s-%s", e.pkg.Repo, e.pkg.Version),
		fmt.Sprintf("%s-%s", e.pkg.Repo, e.pkg.Version),
	)
}

type MissingPackageDependencyError struct {
	dependency string
}

func MissingPackageDependency(dependency string) *MissingPackageDependencyError {
	return &MissingPackageDependencyError{
		dependency: dependency,
	}
}

func (e MissingPackageDependencyError) Error() string {
	return fmt.Sprintf(
		"missing package dependency: %q not found",
		e.dependency,
	)
}
