package controller

import (
	"errors"

	"github.com/tliron/puccini/ard"
	cloutpkg "github.com/tliron/puccini/clout"
	puccinicommon "github.com/tliron/puccini/common"
	resources "github.com/tliron/turandot/resources/turandot.puccini.cloud/v1alpha1"
)

func (self *Controller) processArtifacts(artifacts interface{}, service *resources.Service) (map[string]string, error) {
	if artifacts_, ok := NewKubernetesArtifacts(artifacts); ok {
		artifactMappings := make(map[string]string)
		if len(artifacts_) > 0 {
			if ips, err := self.getPodIps(service.Namespace, "turandot-inventory"); err == nil {
				for _, artifact := range artifacts_ {
					if name, err := self.Push(artifact.Name, artifact.SourcePath, ips); err == nil {
						artifactMappings[artifact.SourcePath] = name
					} else {
						return nil, err
					}
				}
			}
		}
		return artifactMappings, nil
	} else {
		return nil, errors.New("could not parse artifacts")
	}
}

func (self *Controller) updateCloutArtifacts(clout *cloutpkg.Clout, artifactMappings map[string]string) {
	history := ard.StringMap{
		"description": "artifacts",
		"timestamp":   puccinicommon.Timestamp(false),
	}
	ard.NewNode(clout.Metadata).Get("puccini-tosca").Get("history").Append(history)

	for _, vertex := range clout.Vertexes {
		tosca := ard.NewNode(vertex.Metadata).Get("puccini-tosca")
		if kind, ok := tosca.Get("kind").String(true); ok {
			if kind != "NodeTemplate" {
				continue
			}
		} else {
			continue
		}
		if version, ok := tosca.Get("version").String(true); ok {
			if version != "1.0" {
				continue
			}
		} else {
			continue
		}

		if artifacts, ok := ard.NewNode(vertex.Properties).Get("artifacts").StringMap(true); ok {
			for _, artifact := range artifacts {
				artifact_ := ard.NewNode(artifact)
				if sourcePath, ok := artifact_.Get("sourcePath").String(true); ok {
					if name, ok := artifactMappings[sourcePath]; ok {
						if artifact_.Put("$artifact", name) {
							self.log.Infof("setting $artifact for %s to %s", sourcePath, name)
						}
					}
				}
			}
		}
	}
}

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
