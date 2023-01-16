package config

import (
	"fmt"
	"strings"

	"github.com/internet-computer/oko/internal/tar"
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
	return tar.Download(
		fmt.Sprintf(
			"%s/archive/%s/.tar.gz",
			strings.TrimSuffix(p.Repository, ".git"), p.Version,
		),
		".oko",
	)
}

func (p PackageInfoRemote) GetName() string {
	return p.Name
}

func (p PackageInfoRemote) RelativePath() string {
	repo := strings.TrimSuffix(p.Repository, ".git")
	version := strings.TrimPrefix(p.Version, "v")
	return fmt.Sprintf(".oko/%s-%s", repo[strings.LastIndex(repo, "/")+1:], version)
}

func (p PackageInfoRemote) equals(o PackageInfoRemote) bool {
	return p.Repository == o.Repository && p.Version == o.Version
}

func (p PackageInfoRemote) sameName(o PackageInfoRemote) bool {
	if p.Name == o.Name {
		return true
	}
	for _, p := range p.AlternativeNames {
		for _, o := range o.AlternativeNames {
			if p == o {
				return true
			}
		}
	}
	return false
}
