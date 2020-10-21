package parser

import (
	"github.com/tliron/kutil/ard"
)

//
// KubernetesArtifact
//

type KubernetesArtifact struct {
	Tag        string
	Inventory  string
	SourcePath string
}

func ParseKubernetesArtifact(value ard.Value) (*KubernetesArtifact, bool) {
	artifact := ard.NewNode(value)
	if tag, ok := artifact.Get("tag").String(false); ok {
		if inventory, ok := artifact.Get("inventory").String(false); ok {
			if sourcePath, ok := artifact.Get("sourcePath").String(false); ok {
				return &KubernetesArtifact{tag, inventory, sourcePath}, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}

//
// KubernetesArtifacts
//

type KubernetesArtifacts []*KubernetesArtifact

func ParseKubernetesArtifacts(value ard.Value) (KubernetesArtifacts, bool) {
	if artifacts, ok := ard.NewNode(value).Get("artifacts").List(false); ok {
		self := make(KubernetesArtifacts, len(artifacts))
		for index, artifact := range artifacts {
			if artifact_, ok := ParseKubernetesArtifact(artifact); ok {
				self[index] = artifact_
			} else {
				return nil, false
			}
		}
		return self, true
	} else {
		return nil, false
	}
}
