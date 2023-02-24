package config

import "fmt"

type DependencyError struct {
	PackageName    string
	DependencyName string
}

func NewDependencyError(packageName, dependencyName string) *DependencyError {
	return &DependencyError{
		PackageName:    packageName,
		DependencyName: dependencyName,
	}
}

func (e DependencyError) Error() string {
	return fmt.Sprintf(
		"package %q depends on %q",
		e.PackageName, e.DependencyName,
	)
}

type IOError struct {
	Err error
}

func NewIOError(err error) *IOError {
	return &IOError{
		Err: err,
	}
}

func (e IOError) Error() string {
	return fmt.Sprintf(
		"io error: %s",
		e.Err.Error(),
	)
}

type PackageAlreadyExistsError struct {
	Name string
}

func (e PackageAlreadyExistsError) Error() string {
	return fmt.Sprintf(
		"package with name %q already exists",
		e.Name,
	)
}

type PackageNotFoundError struct {
	Name string
}

func NewPackageAlreadyExistsError(name string) *PackageNotFoundError {
	return &PackageNotFoundError{
		Name: name,
	}
}

func NewPackageNotFoundError(name string) *PackageNotFoundError {
	return &PackageNotFoundError{
		Name: name,
	}
}

func (e PackageNotFoundError) Error() string {
	return fmt.Sprintf(
		"package with name %q not found",
		e.Name,
	)
}

type ValidationError struct {
	Err error
}

func NewValidationError(err error) *ValidationError {
	return &ValidationError{
		Err: err,
	}
}

func (e ValidationError) Error() string {
	return fmt.Sprintf(
		"validation error: %s",
		e.Err.Error(),
	)
}
