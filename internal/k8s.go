package internal

import (
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubernetesConfig from the file system,
// if kubeconfigPath is provided as an empty string, the host is expected to be inside a pod.
func GetKubernetesConfig(kubeconfigPath string) *rest.Config {
	var config *rest.Config
	var err error

	if kubeconfigPath == "" { // No kubeconfigPath provided, assume in pod
		log.Debug().
			Msg("no kubeconfig kubeconfigPath provided, assuming running in pod")
		config, err = rest.InClusterConfig()
	} else {
		log.Debug().
			Str("path", kubeconfigPath).
			Msg("kubeconfig kubeconfigPath provided")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to get kubernetes configuration")
	}

	return config
}
