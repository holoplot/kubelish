package cmd

import (
	"log/slog"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

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
		slog.Error("Failed to marshal service", "error", err)
		os.Exit(1)
	}

	os.Stdout.Write(b)
}
