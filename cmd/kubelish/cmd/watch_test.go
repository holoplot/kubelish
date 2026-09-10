package cmd

import (
	"testing"

	watcher "github.com/holoplot/kubelish/pkg/k8s-watcher"
	"github.com/holoplot/kubelish/pkg/publisher"
)

func TestResolveProtocols(t *testing.T) {
	both := publisher.ProtocolIPv4 | publisher.ProtocolIPv6

	for _, tc := range []struct {
		name    string
		mode    string
		m       *watcher.ServiceMDNS
		want    publisher.Protocol
		wantErr bool
	}{
		{
			name: "auto follows a v4 only service",
			mode: protocolModeAuto,
			m:    &watcher.ServiceMDNS{HasIPv4: true},
			want: publisher.ProtocolIPv4,
		},
		{
			name: "auto follows a v6 only service",
			mode: protocolModeAuto,
			m:    &watcher.ServiceMDNS{HasIPv6: true},
			want: publisher.ProtocolIPv6,
		},
		{
			name: "auto follows a dual stack service",
			mode: protocolModeAuto,
			m:    &watcher.ServiceMDNS{HasIPv4: true, HasIPv6: true},
			want: both,
		},
		{
			name: "auto announces both when nothing is known",
			mode: protocolModeAuto,
			m:    &watcher.ServiceMDNS{},
			want: both,
		},
		{
			name: "ipv4 overrides a dual stack service",
			mode: protocolModeIPv4,
			m:    &watcher.ServiceMDNS{HasIPv4: true, HasIPv6: true},
			want: publisher.ProtocolIPv4,
		},
		{
			name: "ipv6 overrides a v4 only service",
			mode: protocolModeIPv6,
			m:    &watcher.ServiceMDNS{HasIPv4: true},
			want: publisher.ProtocolIPv6,
		},
		{
			name: "both overrides a v4 only service",
			mode: protocolModeBoth,
			m:    &watcher.ServiceMDNS{HasIPv4: true},
			want: both,
		},
		{
			name:    "unknown mode fails",
			mode:    "v6only",
			m:       &watcher.ServiceMDNS{HasIPv6: true},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveProtocols(tc.mode, tc.m)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveProtocols(%q) = %v, want error", tc.mode, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("resolveProtocols(%q) failed: %v", tc.mode, err)
			}

			if got != tc.want {
				t.Errorf("resolveProtocols(%q) = %v, want %v", tc.mode, got, tc.want)
			}
		})
	}
}
