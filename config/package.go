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

func New() Package {
	return Package{
		Dependencies: make([]PackageInfo, 0),
	}
}

type Package struct {
	CompilerVersion *string       `json:"compiler,omitempty"`
	Dependencies    []PackageInfo `json:"dependencies"`
}

func LoadPackage(path string) (*Package, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := schema.Validate(raw); err != nil {
		return nil, err
	}
	var pkg Package
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (p *Package) Add(dependencies ...PackageInfo) {
	set := p.Set()
	for _, dep := range dependencies {
		if _, ok := set[dep.Name]; !ok {
			set[dep.Name] = dep
			p.Dependencies = append(p.Dependencies, dep)
		}
	}
	sort.Slice(p.Dependencies, func(i, j int) bool {
		return strings.Compare(p.Dependencies[i].Name, p.Dependencies[j].Name) == -1
	})
}

func (p *Package) Contains(info PackageInfo) (string, bool) {
	for k, dep := range p.Dependencies {
		if info.RelativePathDownload() == dep.RelativePathDownload() {
			if info.Name != dep.Name {
				dep.AddName(info.Name)
				p.Dependencies[k] = dep
			}
			return dep.Name, true
		}
	}
	return "", false
}

func (p Package) Download() error {
	for _, dep := range p.Dependencies {
		if err := dep.Download(); err != nil {
			return err
		}
	}
	return nil
}

func (p Package) Set() map[string]PackageInfo {
	var set = make(map[string]PackageInfo)
	for _, dep := range p.Dependencies {
		set[dep.Name] = dep
	}
	return set
}

type PackageInfo struct {
	Name             string   `json:"name"`
	AlternativeNames []string `json:"alts,omitempty"`
	Repository       string   `json:"repository"`
	Version          string   `json:"version"`
	Dependencies     []string `json:"dependencies,omitempty"`
}

func (p *PackageInfo) AddName(name string) {
	for _, n := range p.AlternativeNames {
		if n == name {
			return
		}
	}
	p.AlternativeNames = append(p.AlternativeNames, name)
}

func (p PackageInfo) Download() error {
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

func (p PackageInfo) RelativePathDownload() string {
	repo := strings.TrimSuffix(p.Repository, ".git")
	version := strings.TrimPrefix(p.Version, "v")
	return fmt.Sprintf(".oko/%s-%s", repo[strings.LastIndex(repo, "/")+1:], version)
}
