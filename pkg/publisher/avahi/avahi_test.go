//go:build linux

package avahipublisher

import (
	"testing"

	"github.com/holoplot/go-avahi"
	"github.com/holoplot/kubelish/pkg/publisher"
)

func TestAvahiProtocol(t *testing.T) {
	for _, tc := range []struct {
		name      string
		protocols publisher.Protocol
		want      int32
		wantErr   bool
	}{
		{
			name:      "ipv4 only",
			protocols: publisher.ProtocolIPv4,
			want:      avahi.ProtoInet,
		},
		{
			name:      "ipv6 only",
			protocols: publisher.ProtocolIPv6,
			want:      avahi.ProtoInet6,
		},
		{
			name:      "both protocols",
			protocols: publisher.ProtocolIPv4 | publisher.ProtocolIPv6,
			want:      avahi.ProtoUnspec,
		},
		{
			name:      "no protocol",
			protocols: 0,
			wantErr:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := avahiProtocol(tc.protocols)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("avahiProtocol(%v) = %d, want error", tc.protocols, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("avahiProtocol(%v) failed: %v", tc.protocols, err)
			}

			if got != tc.want {
				t.Errorf("avahiProtocol(%v) = %d, want %d", tc.protocols, got, tc.want)
			}
		})
	}
}
