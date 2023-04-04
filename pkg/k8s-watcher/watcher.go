package watcher

import (
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/holoplot/kubelish/pkg/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type Watcher struct {
	onUpdate    OnUpdateFunc
	onDelete    OnDeleteFunc
	serviceType corev1.ServiceType
	stopCh      chan struct{}

	services     map[string]*ServiceMDNS
	servicesLock sync.Mutex
}

type ServiceMDNS struct {
	Annotations *meta.Annotations
	IPs         []net.IP
	Port        int
}

type OnUpdateFunc func(*corev1.Service, *ServiceMDNS)
type OnDeleteFunc func(*corev1.Service, *ServiceMDNS)

func (w *Watcher) meshDetails(svc *corev1.Service) *ServiceMDNS {
	if svc.Spec.Type != corev1.ServiceType(w.serviceType) {
		return nil
	}

	service := &ServiceMDNS{}

	if service.Annotations = meta.AnnotationsFromService(svc); service.Annotations == nil {
		return nil
	}

	service.IPs = make([]net.IP, 0)

	for _, ip := range svc.Spec.ExternalIPs {
		service.IPs = append(service.IPs, net.ParseIP(ip))
	}

	if len(svc.Spec.Ports) != 1 {
		return nil
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			service.IPs = append(service.IPs, net.ParseIP(ingress.IP))
		}
	}

	service.Port = int(svc.Spec.Ports[0].Port)

	return service
}

func (w *Watcher) updateService(svc *corev1.Service) {
	w.servicesLock.Lock()
	defer w.servicesLock.Unlock()

	if m := w.meshDetails(svc); m != nil {
		if e, ok := w.services[string(svc.UID)]; !ok || !reflect.DeepEqual(e, m) {
			w.onUpdate(svc, m)
			w.services[string(svc.UID)] = m
		}
	} else {
		if _, ok := w.services[string(svc.UID)]; ok {
			w.onUpdate(svc, nil)
			delete(w.services, string(svc.UID))
		}
	}
}

func (w *Watcher) deleteService(svc *corev1.Service) {
	if m := w.meshDetails(svc); m != nil {
		w.servicesLock.Lock()
		defer w.servicesLock.Unlock()

		delete(w.services, string(svc.UID))
		w.onDelete(svc, m)
	}
}

func (w *Watcher) addHandler(obj interface{}) {
	if svc, ok := obj.(*corev1.Service); ok {
		w.updateService(svc)
	}
}

func (w *Watcher) updateHandler(oldObj, newObj interface{}) {
	if svc, ok := newObj.(*corev1.Service); ok {
		w.updateService(svc)
	}
}

func (w *Watcher) deleteHandler(obj interface{}) {
	if svc, ok := obj.(*corev1.Service); ok {
		w.deleteService(svc)
	}
}

func (w *Watcher) Close() {
	close(w.stopCh)
}

func New(kubeConfigPath, namespace string, serviceType corev1.ServiceType, onUpdate OnUpdateFunc, onDelete OnDeleteFunc) (*Watcher, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})
	factory := informers.NewSharedInformerFactoryWithOptions(clientSet, time.Minute, informers.WithNamespace(namespace))

	svcInformer := factory.Core().V1().Services().Informer()

	go factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, svcInformer.HasSynced) {
		return nil, fmt.Errorf("Timeout waiting for caches to sync")
	}

	w := &Watcher{
		onUpdate:    onUpdate,
		onDelete:    onDelete,
		serviceType: serviceType,
		stopCh:      stopCh,
		services:    make(map[string]*ServiceMDNS),
	}

	if _, err := svcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    w.addHandler,
		UpdateFunc: w.updateHandler,
		DeleteFunc: w.deleteHandler,
	}); err != nil {
		return nil, err
	}

	return w, nil
}
