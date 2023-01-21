package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/internet-computer/oko/config/schema"
	"golang.org/x/exp/slices"
)

type PackageState struct {
	CompilerVersion        *string
	Dependencies           map[string]*PackageInfoRemote
	LocalDependencies      map[string]*PackageInfoLocal
	TransitiveDependencies map[string]*PackageInfoRemote
}

func EmptyState() PackageState {
	return PackageState{
		Dependencies:           make(map[string]*PackageInfoRemote),
		LocalDependencies:      make(map[string]*PackageInfoLocal),
		TransitiveDependencies: make(map[string]*PackageInfoRemote),
	}
}

func LoadPackageState(path string) (*PackageState, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := schema.Validate(raw); err != nil {
		return nil, err
	}
	pkg, err := NewPackage(raw)
	if err != nil {
		return nil, err
	}
	return NewPackageState(pkg), nil
}

func NewPackageState(pkg *Package) *PackageState {
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

func (s *PackageState) AddLocalPackage(pkg PackageInfoLocal) error {
	p, err := s.ContainsLocal(pkg)
	if err != nil {
		return err
	}
	if p != nil {
		return fmt.Errorf("package with name %q already exists", p.Name)
	}
	s.LocalDependencies[pkg.Name] = &pkg
	return nil
}

func (s *PackageState) AddPackage(pkg PackageInfoRemote, dependencies ...PackageInfoRemote) error {
	p, same, err := s.Contains(pkg)
	if err != nil {
		return err
	}
	if p != nil {
		if same {
			return fmt.Errorf("package with name %q already exists", p.Name)
		}
		p.AlternativeNames = append(p.AlternativeNames, pkg.Name)
	} else {
		// Move from transitive if necessary.
		t, same, err := s.ContainsTransitive(pkg)
		if err != nil {
			return err
		}
		if t != nil {
			delete(s.TransitiveDependencies, t.Name)
			if !same && !slices.Contains(t.AlternativeNames, pkg.Name) {
				t.AlternativeNames = append(t.AlternativeNames, pkg.Name)
			}
			s.Dependencies[t.Name] = t
		} else {
			s.Dependencies[pkg.Name] = &pkg
		}
	}
	return s.addPackageDependencies(dependencies...)
}

func (s PackageState) Contains(p PackageInfoRemote) (*PackageInfoRemote, bool, error) {
	for _, dep := range s.Dependencies {
		if dep.equals(p) {
			return dep, dep.sameName(p), nil
		} else if dep.sameName(p) {
			return nil, false, fmt.Errorf("package with name %q already exists", p.Name)
		}
	}
	return nil, false, nil
}

func (s PackageState) ContainsLocal(p PackageInfoLocal) (*PackageInfoLocal, error) {
	for _, dep := range s.LocalDependencies {
		if dep.equals(p) {
			return dep, nil
		} else if dep.Name == p.Name {
			return nil, fmt.Errorf("package with name %q already exists", p.Name)
		}
	}
	return nil, nil
}

func (s PackageState) ContainsTransitive(p PackageInfoRemote) (*PackageInfoRemote, bool, error) {
	for _, dep := range s.TransitiveDependencies {
		if dep.equals(p) {
			return dep, dep.sameName(p), nil
		} else if dep.sameName(p) {
			return nil, false, fmt.Errorf("package with name %q already exists", p.Name)
		}
	}
	return nil, false, nil
}

func (s PackageState) DependencyList() []PackageInfoRemote {
	dependencies := make([]PackageInfoRemote, 0)
	for _, d := range s.Dependencies {
		dependencies = append(dependencies, *d)
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return strings.Compare(dependencies[i].Name, dependencies[j].Name) == -1
	})
	return dependencies
}

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

func (s PackageState) LocalDependencyList() []PackageInfoLocal {
	var dependencies []PackageInfoLocal
	for _, d := range s.LocalDependencies {
		dependencies = append(dependencies, *d)
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return strings.Compare(dependencies[i].Name, dependencies[j].Name) == -1
	})
	return dependencies
}

func (s *PackageState) RemoveLocalPackage(name string) error {
	for _, dep := range s.LocalDependencies {
		if dep.Name == name {
			delete(s.LocalDependencies, name)
			return nil
		}
	}
	return fmt.Errorf("package with name %q not found", name)
}

func (s *PackageState) RemovePackage(name string) error {
	for _, dep := range s.Dependencies {
		if dep.Name == name || slices.Contains(dep.AlternativeNames, name) {
			continue
		}
		if slices.Contains(dep.Dependencies, name) {
			pkg := s.getDependencyByName(name)
			if pkg == nil {
				return fmt.Errorf("package %q depends on %q", dep.Name, name)
			}
			delete(s.Dependencies, pkg.Name)
			s.TransitiveDependencies[pkg.Name] = pkg
			return nil
		}
	}
	for _, dep := range s.TransitiveDependencies {
		if slices.Contains(dep.Dependencies, name) {
			return fmt.Errorf("package %q depends on %q", dep.Name, name)
		}
	}

	for _, dep := range s.Dependencies {
		if dep.Name == name {
			if len(dep.AlternativeNames) == 0 {
				delete(s.Dependencies, name)

				// Try to remove the dependencies too.
				for _, name := range dep.Dependencies {
					_ = s.removeTransitivePackage(name)
				}
			} else {
				dep.Name = dep.AlternativeNames[0]
				dep.AlternativeNames = dep.AlternativeNames[1:]
			}
			return nil
		}
		for i, n := range dep.AlternativeNames {
			if n == name {
				dep.AlternativeNames = append(dep.AlternativeNames[:i], dep.AlternativeNames[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("package with name %q not found", name)
}

func (s PackageState) Save(path string) error {
	json, err := s.ToJSON()
	if err != nil {
		return err
	}
	return os.WriteFile(path, json, os.ModePerm)
}

func (s PackageState) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s.ToPackage(), "", "\t")
}

func (s PackageState) ToPackage() Package {
	return Package{
		CompilerVersion:        s.CompilerVersion,
		Dependencies:           s.DependencyList(),
		LocalDependencies:      s.LocalDependencyList(),
		TransitiveDependencies: s.TransitiveDependencyList(),
	}
}

func (s PackageState) TransitiveDependencyList() []PackageInfoRemote {
	var dependencies []PackageInfoRemote
	for _, d := range s.TransitiveDependencies {
		dependencies = append(dependencies, *d)
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return strings.Compare(dependencies[i].Name, dependencies[j].Name) == -1
	})
	return dependencies
}

func (s *PackageState) addPackageDependencies(dependencies ...PackageInfoRemote) error {
	for _, dep := range dependencies {
		p, _, err := s.Contains(dep)
		if err != nil {
			return err
		}
		if p == nil {
			d, same, err := s.ContainsTransitive(dep)
			if err != nil {
				return err
			}
			if d != nil {
				if !same {
					d.AlternativeNames = append(d.AlternativeNames, dep.Name)
				}
			} else {
				d := dep //copy
				s.TransitiveDependencies[dep.Name] = &d
			}
		}
	}
	return nil
}

func (s PackageState) getDependencyByName(name string) *PackageInfoRemote {
	for _, dep := range s.Dependencies {
		if dep.Name == name {
			return dep
		}
		for _, n := range dep.AlternativeNames {
			if n == name {
				return dep
			}
		}
	}
	return nil
}

func (s *PackageState) removeTransitivePackage(name string) error {
	for _, dep := range s.Dependencies {
		if slices.Contains(dep.Dependencies, name) {
			return fmt.Errorf("package %q depends on %q", dep.Name, name)
		}
	}
	for _, dep := range s.TransitiveDependencies {
		if dep.Name == name || slices.Contains(dep.AlternativeNames, name) {
			continue
		}
		if slices.Contains(dep.Dependencies, name) {
			return fmt.Errorf("package %q depends on %q", dep.Name, name)
		}
	}

	for _, dep := range s.TransitiveDependencies {
		if dep.Name == name {
			if len(dep.AlternativeNames) == 0 {
				delete(s.TransitiveDependencies, name)

				// Try to remove the dependencies too.
				for _, name := range dep.Dependencies {
					_ = s.removeTransitivePackage(name)
				}
			} else {
				dep.Name = dep.AlternativeNames[0]
				dep.AlternativeNames = dep.AlternativeNames[1:]
			}
			return nil
		}
		for i, n := range dep.AlternativeNames {
			if n == name {
				dep.AlternativeNames = append(dep.AlternativeNames[:i], dep.AlternativeNames[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("package with name %q not found", name)
}
