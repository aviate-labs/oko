package config

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/internet-computer/oko/config/schema"
)

type Package struct {
	CompilerVersion        *string             `json:"compiler,omitempty"`
	Dependencies           []PackageInfoRemote `json:"dependencies"`
	LocalDependencies      []PackageInfoLocal  `json:"localDependencies,omitempty"`
	TransitiveDependencies []PackageInfoRemote `json:"transitiveDependencies,omitempty"`
}

func EmptyPackage() Package {
	return Package{
		Dependencies: make([]PackageInfoRemote, 0),
	}
}

func LoadPackage(path string) (*Package, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := schema.Validate(raw); err != nil {
		return nil, err
	}
	return NewPackage(raw)
}

func NewPackage(raw []byte) (*Package, error) {
	var pkg Package
	return &pkg, json.Unmarshal(raw, &pkg)
}

func (p *Package) Add(dependencies ...PackageInfoRemote) {
	set := p.Set()
	for _, dep := range dependencies {
		if _, ok := set[dep.Name]; !ok {
			set[dep.Name] = dep
			p.Dependencies = append(p.Dependencies, dep)
		}
	}
}

func (p *Package) AddDependency(dependencies ...PackageInfoRemote) {
	set := p.SetTransitive()
	for _, dep := range dependencies {
		if _, ok := set[dep.Name]; !ok {
			set[dep.Name] = dep
			p.TransitiveDependencies = append(p.TransitiveDependencies, dep)
		}
	}
}

func (p *Package) AddLocal(dependencies ...PackageInfoLocal) {
	set := p.SetLocal()
	for _, dep := range dependencies {
		if _, ok := set[dep.Name]; !ok {
			set[dep.Name] = dep
			p.LocalDependencies = append(p.LocalDependencies, dep)
		}
	}
}

func (p *Package) Contains(info PackageInfoRemote) (string, bool) {
	for k, dep := range p.Dependencies {
		if n, eq := p.Equals(info, k, dep); eq {
			return n, eq
		}
	}
	for k, dep := range p.TransitiveDependencies {
		if n, eq := p.Equals(info, k, dep); eq {
			return n, eq
		}
	}
	return "", false
}

func (p *Package) ContainsLocal(info PackageInfoLocal) (string, error) {
	for _, dep := range p.LocalDependencies {
		if info.Path == dep.Path {
			if info.Name != dep.Name {
				return "", fmt.Errorf("local package %q already exists %q", info.Path, info.Name)
			}
			return dep.Name, nil
		}
	}
	return "", nil
}

func (p Package) Download() error {
	for _, dep := range p.Dependencies {
		if err := dep.Download(); err != nil {
			return err
		}
	}
	for _, dep := range p.TransitiveDependencies {
		if err := dep.Download(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Package) Equals(info PackageInfoRemote, k int, dep PackageInfoRemote) (string, bool) {
	if info.RelativePath() == dep.RelativePath() {
		if info.GetName() != dep.GetName() {
			dep.AddName(info.GetName())
			p.Dependencies[k] = dep
		}
		return dep.Name, true
	}
	return "", false
}

func (p Package) Save(path string) error {
	sort.Slice(p.Dependencies, func(i, j int) bool {
		return strings.Compare(p.Dependencies[i].Name, p.Dependencies[j].Name) == -1
	})
	sort.Slice(p.LocalDependencies, func(i, j int) bool {
		return strings.Compare(p.Dependencies[i].Name, p.Dependencies[j].Name) == -1
	})
	sort.Slice(p.TransitiveDependencies, func(i, j int) bool {
		return strings.Compare(p.Dependencies[i].Name, p.Dependencies[j].Name) == -1
	})
	dataM, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, dataM, os.ModePerm)
}

func (p Package) Set() map[string]PackageInfoRemote {
	var set = make(map[string]PackageInfoRemote)
	for _, dep := range p.Dependencies {
		set[dep.Name] = dep
	}
	return set
}

func (p Package) SetLocal() map[string]PackageInfoLocal {
	var set = make(map[string]PackageInfoLocal)
	for _, dep := range p.LocalDependencies {
		set[dep.Name] = dep
	}
	return set
}

func (p Package) SetTransitive() map[string]PackageInfoRemote {
	var set = make(map[string]PackageInfoRemote)
	for _, dep := range p.TransitiveDependencies {
		set[dep.Name] = dep
	}
	return set
}

type PackageInfo interface {
	GetName() string
	RelativePath() string
}

type PackageInfoLocal struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (p PackageInfoLocal) GetName() string {
	return p.Name
}

func (p PackageInfoLocal) RelativePath() string {
	return p.Path
}

type PackageInfoRemote struct {
	Name             string   `json:"name"`
	AlternativeNames []string `json:"alts,omitempty"`
	Repository       string   `json:"repository"`
	Version          string   `json:"version,omitempty"`
	Dependencies     []string `json:"dependencies,omitempty"`
}

func (p *PackageInfoRemote) AddName(name string) {
	for _, n := range p.AlternativeNames {
		if n == name {
			return
		}
	}
	p.AlternativeNames = append(p.AlternativeNames, name)
}

func (p PackageInfoRemote) Download() error {
	raw, err := http.Get(fmt.Sprintf(
		"%s/archive/%s/.tar.gz",
		strings.TrimSuffix(p.Repository, ".git"), p.Version,
	))
	if err != nil {
		return err
	}
	if raw.StatusCode != 200 {
		return fmt.Errorf("%d", raw.StatusCode)
	}
	if err := os.MkdirAll(".oko", os.ModePerm); err != nil {
		return err
	}
	gzr, err := gzip.NewReader(raw.Body)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gzr)
	for h, err := tr.Next(); err == nil; h, err = tr.Next() {
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(fmt.Sprintf(".oko/%s", h.Name), os.ModePerm); err != nil {
				if os.IsExist(err) {
					return nil
				}

				return err
			}
		case tar.TypeReg:
			file, err := os.Create(fmt.Sprintf(".oko/%s", h.Name))
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(file, tr); err != nil {
				return err
			}
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

func (p PackageInfoRemote) GetName() string {
	return p.Name
}

func (p PackageInfoRemote) RelativePath() string {
	repo := strings.TrimSuffix(p.Repository, ".git")
	version := strings.TrimPrefix(p.Version, "v")
	return fmt.Sprintf(".oko/%s-%s", repo[strings.LastIndex(repo, "/")+1:], version)
}
