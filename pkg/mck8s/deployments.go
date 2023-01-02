package mck8s

import (
	"context"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
)

func GetAllMck8sServices(client *kubernetes.Clientset, namespace string) ([]v1.Service, error) {
	services, err := client.CoreV1().
		Services(namespace).
		List(context.TODO(), meta.ListOptions{})

	if err != nil {
		return nil, err
	}

	return services.Items, nil
}

func GetMck8sDeployment(client *kubernetes.Clientset, name string, namespace string) (*apps.Deployment, error) {
	return client.AppsV1().
		Deployments(namespace).
		Get(context.TODO(), name, meta.GetOptions{})
}

func EnableMck8sDeployment(client *kubernetes.Clientset, name string, namespace string) error {
	deploymentsClient := client.AppsV1().Deployments(namespace)

	deployment, err := deploymentsClient.Get(context.TODO(), name, meta.GetOptions{})
	if err != nil {
		return err
	}

	deployment.Spec.Replicas = pointer.Int32Ptr(1)
	_, updateErr := deploymentsClient.Update(context.TODO(), deployment, meta.UpdateOptions{})
	if updateErr != nil {
		return err
	}

	return nil
}

func DisableMck8sDeployment(client *kubernetes.Clientset, name string, namespace string) error {
	deploymentsClient := client.AppsV1().Deployments(namespace)
	deployment, err := deploymentsClient.Get(context.TODO(), name, meta.GetOptions{})

	if err != nil {
		return err
	}

	deployment.Spec.Replicas = pointer.Int32Ptr(0)
	_, updateErr := deploymentsClient.Update(context.TODO(), deployment, meta.UpdateOptions{})
	if updateErr != nil {
		return err
	}

	return nil
}
