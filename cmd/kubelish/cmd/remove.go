package cmd

import (
	"log/slog"
	"os"

	"github.com/holoplot/kubelish/pkg/meta"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func doRemove(cmd *cobra.Command, args []string) {
	kubeConfig := getKubeConfig()

	if len(args) != 1 {
		slog.Error("You must specify a k8s service name")
		os.Exit(1)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)

	if err != nil {
		slog.Error("Failed to build config from flags", "error", err)
		os.Exit(1)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("Failed to create clientset", "error", err)
		os.Exit(1)
	}

	svc, err := clientSet.CoreV1().Services(namespace).Get(cmd.Context(), args[0], metav1.GetOptions{})
	if err != nil {
		slog.Error("Failed to get service", "error", err)
		os.Exit(1)
	}

	meta.RemoveAnnotationsFromService(svc)

	dumpServiceYAML(svc)
}

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "remove <k8s-service>",
		Long: `Remove annotations from a Kubernetes service and print the resulting yaml to stdout`,
		Run: func(cmd *cobra.Command, args []string) {
			doRemove(cmd, args)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			setupLogger()
		},
	}
}
