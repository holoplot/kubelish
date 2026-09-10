package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	watcher "github.com/holoplot/kubelish/pkg/k8s-watcher"
	"github.com/holoplot/kubelish/pkg/publisher"
	avahi "github.com/holoplot/kubelish/pkg/publisher/avahi"
	"github.com/okzk/sdnotify"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

const (
	// protocolModeAuto announces each service on the protocols the cluster
	// actually serves it on. The other modes force a fixed set.
	protocolModeAuto = "auto"
	protocolModeIPv4 = "ipv4"
	protocolModeIPv6 = "ipv6"
	protocolModeBoth = "both"
)

var (
	interfaces    []string
	publisherImpl string
	protocolMode  string
)

// resolveProtocols determines the protocols a single service is announced on.
func resolveProtocols(mode string, m *watcher.ServiceMDNS) (publisher.Protocol, error) {
	both := publisher.ProtocolIPv4 | publisher.ProtocolIPv6

	switch mode {
	case protocolModeIPv4:
		return publisher.ProtocolIPv4, nil
	case protocolModeIPv6:
		return publisher.ProtocolIPv6, nil
	case protocolModeBoth:
		return both, nil
	case protocolModeAuto:
		var protocols publisher.Protocol

		if m.HasIPv4 {
			protocols |= publisher.ProtocolIPv4
		}

		if m.HasIPv6 {
			protocols |= publisher.ProtocolIPv6
		}

		if protocols == 0 {
			return both, nil
		}

		return protocols, nil
	}

	return 0, fmt.Errorf("unknown protocol mode %q", mode)
}

func doWatch(cmd *cobra.Command, args []string) {
	kubeConfig := getKubeConfig()
	var pub publisher.Publisher

	switch publisherImpl {
	case "avahi":
		var err error
		pub, err = avahi.New()
		if err != nil {
			slog.Error("Failed to create Avahi publisher", "error", err)
			os.Exit(1)
		}
	default:
		slog.Error("Unknown publisher", "publisher", publisherImpl)
		os.Exit(1)
	}

	publishedServices := make(map[string]publisher.PublishedService)

	updateService := func(svc *corev1.Service, m *watcher.ServiceMDNS) {
		if ps, ok := publishedServices[string(svc.UID)]; ok {
			ps.Close()
			// Give the network some time to digest the loss of a service.
			// Otherwise the new service will replace the old one, but
			// clients might miss it.
			time.Sleep(2 * time.Second)
		}

		if m == nil {
			slog.Info("Service unpublished",
				"name", svc.Name,
				"namespace", svc.Namespace,
				"id", string(svc.UID))

			return
		}

		an := m.Annotations
		if an == nil {
			slog.Debug("No annotations found for service",
				"name", svc.Name,
				"namespace", svc.Namespace,
				"id", string(svc.UID))

			return
		}

		protocols, err := resolveProtocols(protocolMode, m)
		if err != nil {
			slog.Error("Failed to determine protocols for service", "error", err)
			return
		}

		ps, err := pub.Publish(an.ServiceName, an.ServiceType, an.Txt, protocols, m.Port)
		if err != nil {
			slog.Error("Failed to publish service", "error", err)
			return
		}

		publishedServices[string(svc.UID)] = ps

		slog.Info("Service updated",
			"name", svc.Name,
			"namespace", svc.Namespace,
			"mdns-name", an.ServiceName,
			"mdns-type", an.ServiceType,
			"txt", an.Txt,
			"protocols", protocols.String(),
			"id", string(svc.UID))
	}

	deleteService := func(svc *corev1.Service, m *watcher.ServiceMDNS) {
		if ps, ok := publishedServices[string(svc.UID)]; ok {
			ps.Close()
		}

		slog.Info("Service deleted",
			"name", svc.Name,
			"namespace", svc.Namespace,
			"mdns-name", m.Annotations.ServiceName,
			"mdns-type", m.Annotations.ServiceType,
			"id", string(svc.UID))
	}

	w, err := watcher.New(kubeConfig, namespace, corev1.ServiceTypeLoadBalancer, updateService, deleteService)
	if err != nil {
		slog.Error("Failed to create watcher", "error", err, "namespace", namespace)
		os.Exit(1)
	}

	slog.Info("Watching for annotated Kubernetes services")

	sdnotify.Ready()

	<-cmd.Context().Done()
	w.Close()
}

func watchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch for changes in Kubernetes services and publish them to the local network",
		Long:  "Watch for changes in Kubernetes services and publish them to the local network",
		Run:   doWatch,
		PreRun: func(cmd *cobra.Command, args []string) {
			setupLogger()

			if _, err := resolveProtocols(protocolMode, &watcher.ServiceMDNS{}); err != nil {
				slog.Error("Invalid protocol mode", "error", err,
					"valid", []string{protocolModeAuto, protocolModeIPv4, protocolModeIPv6, protocolModeBoth})
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringArrayP("interface", "i", interfaces,
		"Network interface to publish services on (can be specified multiple times) (default: all interfaces)")
	cmd.Flags().StringVarP(&publisherImpl, "publisher", "p", "avahi", "mDNS Publisher to use")
	cmd.Flags().StringVar(&protocolMode, "protocols", protocolModeAuto,
		"Protocols to announce services on: auto (follow the protocols the cluster serves each service on), ipv4, ipv6 or both")

	return cmd
}
