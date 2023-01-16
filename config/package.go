package config

import (
	"encoding/json"
)

type Package struct {
	CompilerVersion        *string             `json:"compiler,omitempty"`
	Dependencies           []PackageInfoRemote `json:"dependencies"`
	LocalDependencies      []PackageInfoLocal  `json:"localDependencies,omitempty"`
	TransitiveDependencies []PackageInfoRemote `json:"transitiveDependencies,omitempty"`
}

func NewPackage(raw []byte) (*Package, error) {
	var pkg Package
	return &pkg, json.Unmarshal(raw, &pkg)
}
