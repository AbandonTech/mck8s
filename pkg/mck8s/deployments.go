package mck8s

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const MinecraftImage = "itzg/minecraft-server:2022.10.0"

// Selector to gather all Services labelled with "mck8s/managed"
var selectorSet = map[string]string{"mck8s/managed": ""}
var Selector = labels.SelectorFromValidatedSet(selectorSet)

// GetAllMck8sServices for the connected cluster
func GetAllMck8sServices(client *kubernetes.Clientset, namespace string) ([]corev1.Service, error) {
	services, err := client.CoreV1().
		Services(namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: Selector.String()})

	if err != nil {
		return nil, err
	}

	return services.Items, nil
}

func GetMck8sService(client *kubernetes.Clientset, name string, namespace string) (*corev1.Service, error) {
	return client.CoreV1().
		Services(namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
}

func GetMck8sDeploymentFromService(client *kubernetes.Clientset, service corev1.Service) (*appsv1.Deployment, error) {
	return client.
		AppsV1().
		Deployments(service.Namespace).
		Get(context.TODO(), service.Spec.Selector["app"], metav1.GetOptions{})
}

func CreateMck8sDeployment(client *kubernetes.Clientset, name string, namespace string, hostname string) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"mck8s/managed": "",
			},
			Annotations: map[string]string{
				"ingress.mck8s/hostname": hostname,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "minecraft",
					Protocol: "TCP",
					Port:     25565,
				},
			},
		},
	}

	_, err := client.CoreV1().
		Services(namespace).
		Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	volume := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"mck8s/managed": "",
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: "manual",
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("10Gi"),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: fmt.Sprintf("/mnt/minecraft/%s", name),
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
		},
	}
	_, err = client.CoreV1().
		PersistentVolumes().
		Create(context.TODO(), volume, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	manual := "manual"
	volumeClaim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"mck8s/managed": "",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &manual,

			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
			VolumeName: name,
		},
	}
	_, err = client.CoreV1().
		PersistentVolumeClaims(namespace).
		Create(context.TODO(), volumeClaim, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":           name,
				"mck8s/managed": "",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: name,
									ReadOnly:  false,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "minecraft",
							Image: MinecraftImage,
							Env: []corev1.EnvVar{
								{Name: "EULA", Value: "TRUE"},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "minecraft",
									ContainerPort: 25565,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = client.AppsV1().
		Deployments(namespace).
		Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func DeleteMck8sDeployment(client *kubernetes.Clientset, name string, namespace string) error {
	err := client.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = client.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
