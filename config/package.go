package config

import (
	"encoding/json"

	"github.com/internet-computer/oko/internal"
)

type PackageConfig struct {
	CompilerVersion        *string             `json:"compiler,omitempty"`
	Dependencies           []PackageInfoRemote `json:"dependencies"`
	LocalDependencies      []PackageInfoLocal  `json:"localDependencies,omitempty"`
	TransitiveDependencies []PackageInfoRemote `json:"transitiveDependencies,omitempty"`
}

func NewPackageConfig(raw []byte) (*PackageConfig, error) {
	var pkg PackageConfig
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return nil, internal.Error(err)
	}
	return &pkg, nil
}
