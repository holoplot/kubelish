package cmd

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

func setupLogger() {
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func getKubeConfig() string {
	kubeConfig, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	return kubeConfig
}

func dumpServiceYAML(svc *corev1.Service) {
	svc.Kind = "Service"
	svc.APIVersion = "v1"
	svc.ManagedFields = make([]metav1.ManagedFieldsEntry, 0)

	b, err := yaml.Marshal(svc)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal service")
	}

	os.Stdout.Write(b)
}
