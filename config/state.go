package config

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/internet-computer/oko/config/schema"
	"github.com/internet-computer/oko/internal"
	"golang.org/x/exp/slices"
)

// PackageState is the in-memory state of the packages.
type PackageState struct {
	CompilerVersion        *string
	Dependencies           map[string]*PackageInfoRemote
	LocalDependencies      map[string]*PackageInfoLocal
	TransitiveDependencies map[string]*PackageInfoRemote
}

// EmptyState returns an empty package state.
func EmptyState() PackageState {
	return PackageState{
		Dependencies:           make(map[string]*PackageInfoRemote),
		LocalDependencies:      make(map[string]*PackageInfoLocal),
		TransitiveDependencies: make(map[string]*PackageInfoRemote),
	}
}

// LoadPackageState loads a package config file.
func LoadPackageState(path string) (*PackageState, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, NewIOError(err)
	}
	if err := schema.Validate(raw); err != nil {
		return nil, NewValidationError(err)
	}
	pkg, err := NewPackageConfig(raw)
	if err != nil {
		return nil, err
	}
	return NewPackageState(pkg), nil
}

// NewPackageState creates a new package state based on the given package config.
func NewPackageState(pkg *PackageConfig) *PackageState {
	state := EmptyState()
	if pkg == nil {
		return &state
	}
	state.CompilerVersion = pkg.CompilerVersion
	for _, dep := range pkg.Dependencies {
		d := dep // copy
		state.Dependencies[dep.Name] = &d
	}
	for _, dep := range pkg.LocalDependencies {
		d := dep // copy
		state.LocalDependencies[dep.Name] = &d
	}
	for _, dep := range pkg.TransitiveDependencies {
		d := dep // copy
		state.TransitiveDependencies[dep.Name] = &d
	}
	return &state
}

// AddLocalPackage adds the given package to the local package state.
func (s *PackageState) AddLocalPackage(pkg PackageInfoLocal) error {
	p, err := s.GetLocal(pkg)
	if err != nil {
		return err
	}
	if p != nil {
		return NewPackageAlreadyExistsError(p.Name)
	}

	// Add new local package.
	s.LocalDependencies[pkg.Name] = &pkg
	return nil
}

// AddPackage adds the given package and its dependencies to the remote package state.
func (s *PackageState) AddPackage(pkg PackageInfoRemote, dependencies ...PackageInfoRemote) error {
	p, same, err := s.Get(pkg)
	if err != nil {
		return err
	}
	if p != nil {
		if same {
			// A package with the same repository and version already exists.
			return NewPackageAlreadyExistsError(p.Name)
		}
		// Add an alternative name, since it does not exist yet.
		p.AlternativeNames = append(p.AlternativeNames, pkg.Name)
	} else {
		// Move from transitive if necessary.
		t, same, err := s.GetTransitive(pkg)
		if err != nil {
			return err
		}
		if t != nil {
			// A package with the same repository and version exists in the transitive dependencies.
			delete(s.TransitiveDependencies, t.Name)
			if !same {
				// Package not found with the exact same name, add it to alternative names.
				t.AlternativeNames = append(t.AlternativeNames, pkg.Name)
			}
			// Move the transitive dependency.
			s.Dependencies[t.Name] = t
		} else {
			// Package not found, add it.
			s.Dependencies[pkg.Name] = &pkg
		}
	}

	// Also add all dependencies.
	return s.addPackageDependencies(dependencies...)
}

// Download downloads all dependencies (including transitive dependencies).
func (s PackageState) Download() error {
	for _, dep := range s.Dependencies {
		if err := dep.Download(); err != nil {
			return err
		}
	}
	for _, dep := range s.TransitiveDependencies {
		if err := dep.Download(); err != nil {
			return err
		}
	}
	return nil
}

// Get returns the package matching the given package info.
// `true` get returns if the package also has the same name.
func (s PackageState) Get(p PackageInfoRemote) (*PackageInfoRemote, bool, error) {
	for _, dep := range s.Dependencies {
		// Check if repository and version match.
		if dep.equals(p) {
			return dep, dep.hasSameName(p), nil
		}

		if dep.hasSameName(p) {
			// No match found, but another package with the same name already exists.
			return nil, false, NewPackageAlreadyExistsError(p.Name)
		}
	}
	// Nothing found.
	return nil, false, nil
}

// Get returns the package matching the given package info.
// Returns an error if a package with the same name already exists.
func (s PackageState) GetLocal(p PackageInfoLocal) (*PackageInfoLocal, error) {
	for _, dep := range s.LocalDependencies {
		// Check if name and path match.
		if dep.equals(p) {
			return dep, nil
		}

		if dep.Name == p.Name {
			// No match found, but another package with the same name already exists.
			return nil, NewPackageAlreadyExistsError(p.Name)
		}
	}
	// Nothing found.
	return nil, nil
}

// GetPackageDependencies returns a list of (copied) package dependencies.
func (s PackageState) GetPackageDependencies(info *PackageInfoRemote) ([]PackageInfoRemote, error) {
	dependencyMap, err := s.getPackageDependencies(info)
	if err != nil {
		return nil, err
	}
	var dependencies []PackageInfoRemote
	for _, dep := range dependencyMap {
		dependencies = append(dependencies, *dep)
	}
	return dependencies, nil
}

// Get returns the package matching the given package info.
// `true` get returns if the package also has the same name.
func (s PackageState) GetTransitive(p PackageInfoRemote) (*PackageInfoRemote, bool, error) {
	for _, dep := range s.TransitiveDependencies {
		// Check if repository and version match.
		if dep.equals(p) {
			return dep, dep.hasSameName(p), nil
		}

		if dep.hasSameName(p) {
			// No match found, but another package with the same name already exists.
			return nil, false, NewPackageAlreadyExistsError(p.Name)
		}
	}
	// Nothing found.
	return nil, false, nil
}

// LoadState loads in another package state.
func (s PackageState) LoadState(state *PackageState) error {
	for _, dep := range state.Dependencies {
		dependencies, err := s.GetPackageDependencies(dep)
		if err != nil {
			return err
		}
		if err := s.AddPackage(*dep, dependencies...); err != nil {
			return err
		}
	}
	// TODO: local packages!
	return nil
}

// MarshalJSON converts the state to raw (formatted) JSON.
func (s PackageState) MarshalJSON() ([]byte, error) {
	raw, err := json.MarshalIndent(PackageConfig{
		CompilerVersion:        s.CompilerVersion,
		Dependencies:           s.dependencyList(),
		LocalDependencies:      s.localDependencyList(),
		TransitiveDependencies: s.transitiveDependencyList(),
	}, "", "\t")
	if err != nil {
		return nil, internal.Error(err)
	}
	return raw, nil
}

// RemoveLocalPackage removes the local package with the given name.
func (s *PackageState) RemoveLocalPackage(name string) error {
	for _, dep := range s.LocalDependencies {
		if dep.Name == name {
			delete(s.LocalDependencies, name)
			return nil
		}
	}
	// No local dependency found with a matching name.
	return NewPackageNotFoundError(name)
}

// RemovePackage removes the package with the given name
func (s *PackageState) RemovePackage(name string) error {
	var pkg *PackageInfoRemote

	// First check if other packages depend on the package with the given name.
	for _, dep := range s.Dependencies {
		if dep.hasName(name) {
			// Ignore the package itself.
			pkg = dep
			continue
		}

		// Check if another package depends on this package.
		if slices.Contains(dep.Dependencies, name) {
			if pkg := s.getDependencyByName(name); pkg != nil {
				// Move dependency to transitive dependencies.
				delete(s.Dependencies, pkg.Name)
				s.TransitiveDependencies[pkg.Name] = pkg
				return nil
			}
			return NewDependencyError(dep.Name, name)
		}
	}

	if pkg == nil {
		// Did not encounter package in dependency list.
		return NewPackageNotFoundError(name)
	}

	for _, dep := range s.TransitiveDependencies {
		// Check if a (transitive) package depends on this package.
		if slices.Contains(dep.Dependencies, name) {
			return NewDependencyError(dep.Name, name)
		}
	}

	// Try to remove the dependencies too.
	for _, name := range pkg.Dependencies {
		// If not possible (error), ignore.
		_ = s.removeTransitivePackage(name)
	}

	// No alternative names, can be safely removed.
	if len(pkg.AlternativeNames) == 0 {
		delete(s.Dependencies, name)
		return nil
	}

	// Remove the current name, but do not remove the package.
	if pkg.Name == name {
		pkg.Name = pkg.AlternativeNames[0]
		pkg.AlternativeNames = pkg.AlternativeNames[1:]
		return nil
	}

	// Remove the name from the alternative name list.
	for i, n := range pkg.AlternativeNames {
		if n == name {
			pkg.AlternativeNames = append(pkg.AlternativeNames[:i], pkg.AlternativeNames[i+1:]...)
			return nil
		}
	}

	// No dependency found with a matching name.
	return NewPackageNotFoundError(name)
}

// Save writes the state to the given path.
func (s PackageState) Save(path string) error {
	json, err := s.MarshalJSON()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, json, os.ModePerm); err != nil {
		return NewIOError(err)
	}
	return nil
}

// addPackageDependencies adds the given packages to the transitive package list.
func (s *PackageState) addPackageDependencies(dependencies ...PackageInfoRemote) error {
	for _, dep := range dependencies {
		p, _, err := s.Get(dep)
		if err != nil {
			return err
		}

		if p == nil {
			d, _, err := s.GetTransitive(dep)
			if err != nil {
				return err
			}

			if d == nil {
				d := dep //copy
				s.TransitiveDependencies[dep.Name] = &d
			}
		}
	}

	// No errors occurred.
	return nil
}

// DependencyList returns a sorted list of dependencies. Does not include transitive and local dependencies.
func (s PackageState) dependencyList() []PackageInfoRemote {
	dependencies := make([]PackageInfoRemote, 0)
	for _, d := range s.Dependencies {
		dependencies = append(dependencies, *d)
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return strings.Compare(dependencies[i].Name, dependencies[j].Name) == -1
	})
	return dependencies
}

// getDependencyByName return the package that matches the given name.
func (s PackageState) getDependencyByName(name string) *PackageInfoRemote {
	for _, dep := range s.Dependencies {
		if dep.hasName(name) {
			return dep
		}
	}
	return nil
}

// Returns a map of package dependencies.
func (s PackageState) getPackageDependencies(info *PackageInfoRemote) (map[string]*PackageInfoRemote, error) {
	dependencies := make(map[string]*PackageInfoRemote)
	for _, name := range info.Dependencies {
		var hit bool
		for _, d := range s.Dependencies {
			if d.hasName(name) {
				if _, ok := dependencies[name]; !ok {
					dependencies[name] = d
				}

				// Dependencies of the dependencies...
				m, err := s.getPackageDependencies(d)
				if err != nil {
					return nil, err
				}
				for name, dep := range m {
					if _, ok := dependencies[name]; !ok {
						dependencies[name] = dep
					}
				}

				hit = true
				break
			}
		}
		if hit {
			// Already found in dependency list.
			continue
		}

		for _, d := range s.TransitiveDependencies {
			if d.hasName(name) {
				if _, ok := dependencies[name]; !ok {
					dependencies[name] = d
				}

				// Dependencies of the dependencies...
				m, err := s.getPackageDependencies(d)
				if err != nil {
					return nil, err
				}
				for name, dep := range m {
					if _, ok := dependencies[name]; !ok {
						dependencies[name] = dep
					}
				}

				hit = true
				break
			}
		}
		if !hit {
			return nil, NewPackageNotFoundError(name)
		}
	}
	return dependencies, nil
}

// LocalDependencyList returns a sorted list of local dependencies.
func (s PackageState) localDependencyList() []PackageInfoLocal {
	var dependencies []PackageInfoLocal
	for _, d := range s.LocalDependencies {
		dependencies = append(dependencies, *d)
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return strings.Compare(dependencies[i].Name, dependencies[j].Name) == -1
	})
	return dependencies
}

// removeTransitivePackage removes transitive dependencies if they are not in use.
func (s *PackageState) removeTransitivePackage(name string) error {
	var pkg *PackageInfoRemote

	// First check if other packages depend on the package with the given name.
	for _, dep := range s.Dependencies {
		// Package is still needed.
		if slices.Contains(dep.Dependencies, name) {
			return NewDependencyError(dep.Name, name)
		}
	}
	for _, dep := range s.TransitiveDependencies {
		if dep.hasName(name) {
			// Ignore the package itself.
			pkg = dep
			continue
		}

		// Package is needed by another transitive dependency.
		if slices.Contains(dep.Dependencies, name) {
			return NewDependencyError(dep.Name, name)
		}
	}

	if pkg == nil {
		// Did not encounter package in dependency list.
		return NewPackageNotFoundError(name)
	}

	// Try to remove the dependencies too.
	for _, name := range pkg.Dependencies {
		_ = s.removeTransitivePackage(name)
	}

	// No alternative names, can be safely removed.
	if len(pkg.AlternativeNames) == 0 {
		delete(s.TransitiveDependencies, name)
		return nil
	}

	// Remove the current name, but do not remove the package.
	if pkg.Name == name {
		pkg.Name = pkg.AlternativeNames[0]
		pkg.AlternativeNames = pkg.AlternativeNames[1:]
		return nil
	}

	// Remove the name from the alternative name list.
	for i, n := range pkg.AlternativeNames {
		if n == name {
			pkg.AlternativeNames = append(pkg.AlternativeNames[:i], pkg.AlternativeNames[i+1:]...)
			return nil
		}
	}

	// No dependency found with a matching name.
	return NewPackageNotFoundError(name)
}

// transitiveDependencyList returns a sorted list of transitive dependencies.
func (s PackageState) transitiveDependencyList() []PackageInfoRemote {
	var dependencies []PackageInfoRemote
	for _, d := range s.TransitiveDependencies {
		dependencies = append(dependencies, *d)
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return strings.Compare(dependencies[i].Name, dependencies[j].Name) == -1
	})
	return dependencies
}
