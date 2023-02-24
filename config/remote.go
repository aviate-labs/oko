package config

import (
	"fmt"
	"strings"

	"github.com/internet-computer/oko/internal"
	"github.com/internet-computer/oko/internal/tar"
	"golang.org/x/exp/slices"
)

type PackageInfoRemote struct {
	Name             string   `json:"name"`
	AlternativeNames []string `json:"alts,omitempty"`
	Repository       string   `json:"repository"`
	Version          string   `json:"version"`
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
	if err := tar.Download(
		fmt.Sprintf(
			"%s/archive/%s/.tar.gz",
			strings.TrimSuffix(p.Repository, ".git"), p.Version,
		),
		".oko",
	); err != nil {
		return internal.Error(err)
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

// equals returns true if both the repository and version match.
func (p PackageInfoRemote) equals(o PackageInfoRemote) bool {
	return p.Repository == o.Repository && p.Version == o.Version
}

// hashName returns true if either it has the same name, or an alternative name matches.
func (p PackageInfoRemote) hasName(name string) bool {
	return p.Name == name || slices.Contains(p.AlternativeNames, name)
}

// hasSameName returns true if either both have the same name, or an alternative name matches.
func (p PackageInfoRemote) hasSameName(o PackageInfoRemote) bool {
	if p.Name == o.Name {
		return true
	}
	for _, p := range p.AlternativeNames {
		if slices.Contains(o.AlternativeNames, p) {
			return true
		}
	}
	return false
}
