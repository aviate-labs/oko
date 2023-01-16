package vessel

import (
	"os"
	"strings"

	"github.com/internet-computer/oko/config"
	"github.com/philandstuff/dhall-golang/v6"
)

type Package struct {
	Name         string   `dhall:"name"`
	Repo         string   `dhall:"repo"`
	Version      string   `dhall:"version"`
	Dependencies []string `dhall:"dependencies"`
}

type PackageSet struct {
	Packages map[string]Package
}

func LoadPackageSet(path string) (*PackageSet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewPackageSet(raw)
}

func NewPackageSet(raw []byte) (*PackageSet, error) {
	var (
		set = PackageSet{
			Packages: make(map[string]Package),
		}
		list []Package
	)
	if err := dhall.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	for _, pkg := range list {
		if v, ok := set.Packages[pkg.Name]; ok {
			return &set, DuplicatePackageName(pkg, v)
		}
		set.Packages[pkg.Name] = pkg
	}
	return &set, nil
}

func (set PackageSet) Filter(packages []string) (PackageSet, error) {
	var result = PackageSet{
		Packages: make(map[string]Package),
	}
	for len(packages) != 0 {
		name := packages[0]
		packages = packages[1:]

		if _, ok := result.Packages[name]; !ok {
			if pkg, ok := set.Packages[name]; !ok {
				return result, MissingPackageDependency(name)
			} else {
				packages = append(packages, pkg.Dependencies...)
				result.Packages[name] = pkg
			}
		}
	}
	return result, nil
}

func (set PackageSet) Oko() []config.PackageInfoRemote {
	var packages []config.PackageInfoRemote
	for _, pkg := range set.Packages {
		packages = append(packages, config.PackageInfoRemote{
			Name:         pkg.Name,
			Repository:   strings.TrimSuffix(pkg.Repo, ".git"),
			Version:      pkg.Version,
			Dependencies: pkg.Dependencies,
		})
	}
	return packages
}
