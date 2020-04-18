package common

import (
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcore "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

func CreateEventRecorder(kubeClientset kubernetes.Interface, component string) record.EventRecorder {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&typedcore.EventSinkImpl{Interface: kubeClientset.CoreV1().Events("")})
	return broadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: component})
}
