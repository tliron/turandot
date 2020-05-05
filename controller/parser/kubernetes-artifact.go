package parser

import (
	"github.com/tliron/puccini/ard"
)

//
// KubernetesArtifact
//

type KubernetesArtifact struct {
	Name       string
	SourcePath string
}

func NewKubernetesArtifact(data interface{}) (*KubernetesArtifact, bool) {
	artifact := ard.NewNode(data)
	if name, ok := artifact.Get("name").String(false); ok {
		if sourcePath, ok := artifact.Get("sourcePath").String(false); ok {
			return &KubernetesArtifact{name, sourcePath}, true
		}
	}
	return nil, false
}

//
// KubernetesArtifacts
//

type KubernetesArtifacts []*KubernetesArtifact

func NewKubernetesArtifacts(data interface{}) (KubernetesArtifacts, bool) {
	if artifacts, ok := ard.NewNode(data).Get("artifacts").List(false); ok {
		self := make(KubernetesArtifacts, len(artifacts))
		for index, artifact := range artifacts {
			if artifact_, ok := NewKubernetesArtifact(artifact); ok {
				self[index] = artifact_
			} else {
				return nil, false
			}
		}
		return self, true
	}
	return nil, false
}
