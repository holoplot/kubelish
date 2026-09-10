package watcher

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func lbService(ingressIPs []string, externalIPs []string, families []corev1.IPFamily) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc",
			Namespace: "default",
			Annotations: map[string]string{
				"kubelish/service-name": "Test",
				"kubelish/service-type": "_http._tcp",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:        corev1.ServiceTypeLoadBalancer,
			ExternalIPs: externalIPs,
			IPFamilies:  families,
			Ports:       []corev1.ServicePort{{Port: 80}},
		},
	}

	for _, ip := range ingressIPs {
		svc.Status.LoadBalancer.Ingress = append(svc.Status.LoadBalancer.Ingress,
			corev1.LoadBalancerIngress{IP: ip})
	}

	return svc
}

func TestMeshDetailsIPFamilies(t *testing.T) {
	for _, tc := range []struct {
		name        string
		ingressIPs  []string
		externalIPs []string
		families    []corev1.IPFamily
		wantIPs     int
		wantV4      bool
		wantV6      bool
	}{
		{
			name:       "v4 only load balancer",
			ingressIPs: []string{"10.0.0.1"},
			families:   []corev1.IPFamily{corev1.IPv4Protocol},
			wantIPs:    1,
			wantV4:     true,
		},
		{
			name:       "v6 only load balancer",
			ingressIPs: []string{"2001:db8::1"},
			families:   []corev1.IPFamily{corev1.IPv6Protocol},
			wantIPs:    1,
			wantV6:     true,
		},
		{
			name:       "dual stack load balancer",
			ingressIPs: []string{"10.0.0.1", "2001:db8::1"},
			families:   []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
			wantIPs:    2,
			wantV4:     true,
			wantV6:     true,
		},
		{
			name:       "dual stack cluster with v4 only load balancer address",
			ingressIPs: []string{"10.0.0.1"},
			families:   []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
			wantIPs:    1,
			wantV4:     true,
		},
		{
			name:        "external ips are considered",
			externalIPs: []string{"2001:db8::2"},
			families:    []corev1.IPFamily{corev1.IPv4Protocol},
			wantIPs:     1,
			wantV6:      true,
		},
		{
			name:       "hostname only ingress falls back to ip families",
			ingressIPs: []string{""},
			families:   []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
			wantIPs:    0,
			wantV4:     true,
			wantV6:     true,
		},
		{
			name:    "no information at all announces both",
			wantIPs: 0,
			wantV4:  true,
			wantV6:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &Watcher{serviceType: corev1.ServiceTypeLoadBalancer}

			m := w.meshDetails(lbService(tc.ingressIPs, tc.externalIPs, tc.families))
			if m == nil {
				t.Fatal("meshDetails() returned nil")
			}

			if len(m.IPs) != tc.wantIPs {
				t.Errorf("len(IPs) = %d, want %d (%v)", len(m.IPs), tc.wantIPs, m.IPs)
			}

			if m.HasIPv4 != tc.wantV4 {
				t.Errorf("HasIPv4 = %v, want %v", m.HasIPv4, tc.wantV4)
			}

			if m.HasIPv6 != tc.wantV6 {
				t.Errorf("HasIPv6 = %v, want %v", m.HasIPv6, tc.wantV6)
			}
		})
	}
}
