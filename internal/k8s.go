package internal

import (
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetKubernetesConfig from the file system,
// expects the runner to be a pod inside a K8s cluster with the appropriate service account.
func GetKubernetesConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to get kubernetes configuration")
	}
	return config
}

// GetKubernetesClient creates a client using GetKubernetesConfig
func GetKubernetesClient() *kubernetes.Clientset {
	client, err := kubernetes.NewForConfig(GetKubernetesConfig())
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to get kubernetes client")
	}

	return client
}
