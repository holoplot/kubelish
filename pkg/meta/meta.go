package meta

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
)

type Annotations struct {
	ServiceName string
	ServiceType string
	Txt         string
}

const (
	serviceNameKey = "kubelish/service-name"
	serviceTypeKey = "kubelish/service-type"
	txtKey         = "kubelish/txt"
)

func (a *Annotations) Equal(b *Annotations) bool {
	return reflect.DeepEqual(a, b)
}

func (a *Annotations) Empty() bool {
	return a == nil || (a.ServiceName == "" && a.ServiceType == "" && a.Txt == "")
}

func AddAnnotationsToService(svc *corev1.Service, a *Annotations) {
	svc.Annotations[serviceNameKey] = a.ServiceName
	svc.Annotations[serviceTypeKey] = a.ServiceType
	svc.Annotations[txtKey] = a.Txt
}

func AnnotationsFromService(svc *corev1.Service) *Annotations {
	a := &Annotations{}

	var ok bool

	if a.ServiceName, ok = svc.Annotations[serviceNameKey]; !ok {
		return nil
	}

	if a.ServiceType, ok = svc.Annotations[serviceTypeKey]; !ok {
		return nil
	}

	a.Txt = svc.Annotations[txtKey]

	return a
}

func RemoveAnnotationsFromService(svc *corev1.Service) {
	delete(svc.Annotations, serviceNameKey)
	delete(svc.Annotations, serviceTypeKey)
	delete(svc.Annotations, txtKey)
}
