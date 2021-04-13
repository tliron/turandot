package parser

import (
	"github.com/tliron/kutil/util"
	"gopkg.in/yaml.v3"
)

//
// KubernetesArtifact
//

type KubernetesArtifact struct {
	Name       string `yaml:"name"`
	Registry   string `yaml:"registry"`
	SourcePath string `yaml:"sourcePath"`
}

//
// KubernetesArtifacts
//

type KubernetesArtifacts []*KubernetesArtifact

func DecodeKubernetesArtifacts(code string) (KubernetesArtifacts, bool) {
	var artifacts struct {
		Artifacts KubernetesArtifacts `yaml:"artifacts"`
	}
	if err := yaml.Unmarshal(util.StringToBytes(code), &artifacts); err == nil {
		return artifacts.Artifacts, true
	} else {
		return nil, false
	}
}
