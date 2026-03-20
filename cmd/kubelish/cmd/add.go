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

var (
	serviceName string
	serviceType string
	txt         string
)

func doAdd(cmd *cobra.Command, args []string) {
	kubeConfig := getKubeConfig()

	if len(args) != 1 {
		slog.Error("You must specify a k8s service name")
		os.Exit(1)
	}

	if serviceName == "" {
		slog.Error("You must specify an mDNS service name")
		os.Exit(1)
	}

	if serviceType == "" {
		slog.Error("You must specify an mDNS service type")
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

	an := meta.Annotations{
		ServiceName: serviceName,
		ServiceType: serviceType,
		Txt:         txt,
	}

	meta.AddAnnotationsToService(svc, &an)

	dumpServiceYAML(svc)
}

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <k8s-service>",
		Short: "Add annotations to a Kubernetes service",
		Long:  "Add annotations to a Kubernetes service and print the resulting yaml to stdout",
		Run:   doAdd,
		PreRun: func(cmd *cobra.Command, args []string) {
			setupLogger()
		},
	}

	cmd.Flags().StringVarP(&serviceName, "service-name", "s", "", "The mDNS service name")
	cmd.Flags().StringVarP(&serviceType, "service-type", "t", "", "The mDNS service type")
	cmd.Flags().StringVarP(&txt, "txt", "x", "", "The mDNS TXT record")

	return cmd
}
