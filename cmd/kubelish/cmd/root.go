package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	debug     *bool
	namespace string
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "kubelish",
		Short: "kubelish is a service discovery tool for Kubernetes",
		Long:  `kubelish is a service discovery tool for Kubernetes. It publishes Kubernetes services to the local network using mDNS.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setupLogger()
		},
	}

	debug = rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug log output")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "The namespace to operate on")

	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(removeCmd())
	rootCmd.AddCommand(watchCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute root command")
	}
}
