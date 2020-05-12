package common

import (
	"github.com/op/go-logging"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcore "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

func CreateEventRecorder(kubernetes kubernetes.Interface, component string, log *logging.Logger) record.EventRecorder {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(log.Infof)
	broadcaster.StartRecordingToSink(&typedcore.EventSinkImpl{Interface: kubernetes.CoreV1().Events("")})
	return broadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: component})
}
