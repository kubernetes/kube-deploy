package cmd

import (
	"os"

	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/cluster-api/api"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [YAML_FILE]",
	Short: "upgrade kubernetes cluster",
	Long:  "Upgrade a kubernetes cluster with one command",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logger.Critical("Please provide yaml file for upgrade specification.")
			os.Exit(1)
		} else if len(args) > 1 {
			logger.Critical("Too many arguments.")
			os.Exit(1)
		}
		yamlFile := args[0]
		cluster, err := readAndValidateYaml(yamlFile)
		if err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}
		logger.Info("Parsing done [%s]", cluster)

		if err = upgradeCluster(cluster); err != nil {
			logger.Critical(err.Error())
			os.Exit(1)
		}
	},
}

func upgradeCluster(cluster *api.Cluster) error {
	//
	// The logic here would be to fetch a list of cluster-api.k8s.io/v1alpha1/Machine object, and update each of them
	// with new config.
	//
	logger.Info("Upgrade is in progress... [%s]", cluster)

	return nil
}

func init() {
	RootCmd.AddCommand(upgradeCmd)
}
