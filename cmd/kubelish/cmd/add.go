package cmd

import (
	"github.com/holoplot/kubelish/pkg/meta"
	"github.com/rs/zerolog/log"
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
		log.Fatal().Msg("You must specify a k8s service name")
	}

	if serviceName == "" {
		log.Fatal().Msg("You must specify an mDNS service name")
	}

	if serviceType == "" {
		log.Fatal().Msg("You must specify an mDNS service type")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build config from flags")
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create clientset")
	}

	svc, err := clientSet.CoreV1().Services(namespace).Get(cmd.Context(), args[0], metav1.GetOptions{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get service")
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
