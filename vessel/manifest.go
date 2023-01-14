package vessel

import (
	"github.com/internet-computer/oko/config"
	"github.com/philandstuff/dhall-golang/v6"
)

type Manifest struct {
	Compiler     *string  `dhall:"compiler"`
	Dependencies []string `dhall:"dependencies"`
}

func NewManifest(raw []byte) (Manifest, error) {
	var manifest Manifest
	return manifest, dhall.Unmarshal(raw, &manifest)
}

func (m Manifest) Oko(set PackageSet) config.Package {
	return config.Package{
		CompilerVersion: m.Compiler,
		Dependencies:    set.Oko(),
	}
}
