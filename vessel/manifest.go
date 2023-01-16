package vessel

import (
	"os"

	"github.com/internet-computer/oko/config"
	"github.com/philandstuff/dhall-golang/v6"
)

type Manifest struct {
	Compiler     *string  `dhall:"compiler"`
	Dependencies []string `dhall:"dependencies"`
}

func LoadManifest(path string) (*Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewManifest(raw)
}

func NewManifest(raw []byte) (*Manifest, error) {
	var manifest Manifest
	return &manifest, dhall.Unmarshal(raw, &manifest)
}

func (m Manifest) Oko(set PackageSet) config.Package {
	return config.Package{
		CompilerVersion: m.Compiler,
		Dependencies:    set.Oko(),
	}
}

func (m Manifest) Save(path string, set PackageSet) error {
	pkg := m.Oko(set)
	return config.NewPackageState(&pkg).Save(path)
}
