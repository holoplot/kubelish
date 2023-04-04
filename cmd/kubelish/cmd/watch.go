package cmd

import (
	"time"

	watcher "github.com/holoplot/kubelish/pkg/k8s-watcher"
	"github.com/holoplot/kubelish/pkg/publisher"
	avahi "github.com/holoplot/kubelish/pkg/publisher/avahi"
	"github.com/okzk/sdnotify"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var (
	interfaces    []string
	publisherImpl string
)

func doWatch(cmd *cobra.Command, args []string) {
	kubeConfig := getKubeConfig()
	var pub publisher.Publisher

	switch publisherImpl {
	case "avahi":
		var err error
		pub, err = avahi.New()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Avahi publisher")
		}
	default:
		log.Fatal().Str("publisher", publisherImpl).Msg("Unknown publisher")
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
			log.Info().
				Str("name", svc.Name).
				Str("namespace", svc.Namespace).
				Str("id", string(svc.UID)).
				Msg("Service unpublished")

			return
		}

		an := m.Annotations
		if an == nil {
			log.Debug().
				Str("name", svc.Name).
				Str("namespace", svc.Namespace).
				Str("id", string(svc.UID)).
				Msg("No annotations found for service")

			return
		}

		ps, err := pub.Publish(an.ServiceName, an.ServiceType, an.Txt, m.Port)
		if err != nil {
			log.Error().Err(err).Msg("Failed to publish service")
			return
		}

		publishedServices[string(svc.UID)] = ps

		log.Info().
			Str("name", svc.Name).
			Str("namespace", svc.Namespace).
			Str("mdns-name", an.ServiceName).
			Str("mdns-type", an.ServiceType).
			Str("txt", an.Txt).
			Str("id", string(svc.UID)).
			Msg("Service updated")
	}

	deleteService := func(svc *corev1.Service, m *watcher.ServiceMDNS) {
		if ps, ok := publishedServices[string(svc.UID)]; ok {
			ps.Close()
		}

		log.Info().
			Str("name", svc.Name).
			Str("namespace", svc.Namespace).
			Str("mdns-name", m.Annotations.ServiceName).
			Str("mdns-type", m.Annotations.ServiceType).
			Str("id", string(svc.UID)).
			Msg("Service deleted")
	}

	w, err := watcher.New(kubeConfig, namespace, corev1.ServiceTypeLoadBalancer, updateService, deleteService)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("namespace", namespace).
			Msg("Failed to create watcher")
	}

	log.Info().Msg("Watching for annotated Kubernetes services")

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
		},
	}

	cmd.Flags().StringArrayP("interface", "i", interfaces,
		"Network interface to publish services on (can be specified multiple times) (default: all interfaces)")
	cmd.Flags().StringVarP(&publisherImpl, "publisher", "p", "avahi", "mDNS Publisher to use")

	return cmd
}
