package cmd

import (
	"github.com/holoplot/kubelish/pkg/meta"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func doRemove(cmd *cobra.Command, args []string) {
	kubeConfig := getKubeConfig()

	if len(args) != 1 {
		log.Fatal().Msg("You must specify a k8s service name")
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
