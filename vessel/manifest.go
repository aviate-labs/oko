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
		return nil, NewVesselError(err)
	}
	return NewManifest(raw)
}

func NewManifest(raw []byte) (*Manifest, error) {
	var manifest Manifest
	if err := dhall.Unmarshal(raw, &manifest); err != nil {
		return nil, NewVesselError(err)
	}
	return &manifest, nil
}

func (m Manifest) Oko(set PackageSet) config.PackageConfig {
	return config.PackageConfig{
		CompilerVersion: m.Compiler,
		Dependencies:    set.Oko(),
	}
}

func (m Manifest) Save(path string, set PackageSet) error {
	pkg := m.Oko(set)
	if err := config.NewPackageState(&pkg).Save(path); err != nil {
		return NewVesselError(err)
	}
	return nil
}
