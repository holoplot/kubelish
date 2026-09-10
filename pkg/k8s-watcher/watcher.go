package watcher

import (
	"fmt"
	"log/slog"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/holoplot/kubelish/pkg/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	HasIPv4     bool
	HasIPv6     bool
	Port        int
}

type OnUpdateFunc func(*corev1.Service, *ServiceMDNS)
type OnDeleteFunc func(*corev1.Service, *ServiceMDNS)

// addIP records an address and the protocol it belongs to. Unparsable
// addresses are ignored.
func (s *ServiceMDNS) addIP(ip net.IP) {
	if ip == nil {
		return
	}

	if ip.To4() != nil {
		s.HasIPv4 = true
	} else {
		s.HasIPv6 = true
	}

	s.IPs = append(s.IPs, ip)
}

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
		service.addIP(net.ParseIP(ip))
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			// Ingress entries that only carry a host name have no IP.
			service.addIP(net.ParseIP(ingress.IP))
		}
	}

	// Without any usable address, fall back to the families the cluster
	// assigned to the service. Those follow the service CIDRs, so an IPv6
	// entry proves the cluster serves IPv6.
	if !service.HasIPv4 && !service.HasIPv6 {
		for _, family := range svc.Spec.IPFamilies {
			switch family {
			case corev1.IPv4Protocol:
				service.HasIPv4 = true
			case corev1.IPv6Protocol:
				service.HasIPv6 = true
			}
		}
	}

	// Still nothing known about the service: assume both protocols rather
	// than announcing nothing at all.
	if !service.HasIPv4 && !service.HasIPv6 {
		slog.Debug("Service has no addresses and no IP families, assuming both protocols",
			"namespace", svc.Namespace,
			"service", svc.Name)

		service.HasIPv4 = true
		service.HasIPv6 = true
	}

	if len(svc.Spec.Ports) == 1 {
		service.Port = int(svc.Spec.Ports[0].Port)
	} else {
		for _, port := range svc.Spec.Ports {
			if port.Name == service.Annotations.ServiceName {
				service.Port = int(port.Port)
				break
			}
		}

		if service.Port == 0 {
			slog.Warn("Service has multiple ports but none is named like the service",
				"namespace", svc.Namespace,
				"service", svc.Name,
				"serviceName", service.Annotations.ServiceName)
		}
	}

	if service.Port == 0 {
		return nil
	}

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

func (w *Watcher) addHandler(obj any) {
	if svc, ok := obj.(*corev1.Service); ok {
		w.updateService(svc)
	}
}

func (w *Watcher) updateHandler(oldObj, newObj any) {
	if svc, ok := newObj.(*corev1.Service); ok {
		w.updateService(svc)
	}
}

func (w *Watcher) deleteHandler(obj any) {
	if svc, ok := obj.(*corev1.Service); ok {
		w.deleteService(svc)
	}
}

func (w *Watcher) Close() {
	close(w.stopCh)
}

func New(kubeConfigPath, namespace string, serviceType corev1.ServiceType, onUpdate OnUpdateFunc, onDelete OnDeleteFunc) (*Watcher, error) {
	var config *rest.Config
	var err error

	config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
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
