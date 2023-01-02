package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	"minecraftk8s/internal"
	"minecraftk8s/pkg/mck8s"
	"os"
)

const VERSION = "v0.0.1"

var K8sClient *kubernetes.Clientset

func main() {
	// Overwrite version flag to be capital 'V'
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}

	app := &cli.App{
		Name:    "mck8s",
		Usage:   "configure mck8s deployments via commandline",
		Version: VERSION,
		Authors: []*cli.Author{{
			Name:  "GDWR",
			Email: "gregory.dwr@gmail.com",
		}},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Usage: "Set logging to debug level", Aliases: []string{"v"}},
			&cli.PathFlag{Name: "kubeconfig", Usage: "Path to kubeconfig"},
		},
		Before: func(ctx *cli.Context) error {
			internal.ConfigureLogging(ctx.Bool("verbose"))

			var err error
			K8sClient, err = kubernetes.NewForConfig(internal.GetKubernetesConfig(ctx.Path("kubeconfig")))
			return err
		},
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "create a new mck8s deployment",
				ArgsUsage: "[name] [hostname]",
				Action: func(cCtx *cli.Context) error {
					return mck8s.CreateMck8sDeployment(K8sClient, cCtx.Args().Get(0), "default", cCtx.Args().Get(1))
				},
			},
			{
				Name:      "delete",
				Usage:     "delete an existing mck8s deployment",
				ArgsUsage: "[name] [hostname]",
				Action: func(cCtx *cli.Context) error {
					return mck8s.DeleteMck8sDeployment(K8sClient, cCtx.Args().Get(0), "default")
				},
			},
			{
				Name:      "disable",
				Usage:     "disable an existing mck8s deployment",
				ArgsUsage: "[name]",
				Action: func(cCtx *cli.Context) error {

					service, err := mck8s.GetMck8sService(K8sClient, cCtx.Args().Get(0), "default")
					if err != nil {
						return err
					}

					deployment, err := mck8s.GetMck8sDeploymentFromService(K8sClient, *service)
					if err != nil {
						return err
					}

					deployment.Spec.Replicas = pointer.Int32(0)

					_, err = K8sClient.AppsV1().
						Deployments("default").
						Update(context.TODO(), deployment, metav1.UpdateOptions{})

					return err
				},
			},
			{
				Name:      "enable",
				Usage:     "enable an existing mck8s deployment",
				ArgsUsage: "[name]",
				Action: func(cCtx *cli.Context) error {
					service, err := mck8s.GetMck8sService(K8sClient, cCtx.Args().Get(0), "default")
					if err != nil {
						return err
					}

					deployment, err := mck8s.GetMck8sDeploymentFromService(K8sClient, *service)
					if err != nil {
						return err
					}

					deployment.Spec.Replicas = pointer.Int32(1)

					_, err = K8sClient.AppsV1().
						Deployments("default").
						Update(context.TODO(), deployment, metav1.UpdateOptions{})

					return err
				},
			},
			{
				Name:  "list",
				Usage: "list all mck8s deployments",
				Action: func(cCtx *cli.Context) error {
					mck8Deployments, err := mck8s.GetAllMck8sServices(K8sClient, "") // All namespaces
					if err != nil {
						return err
					}

					for _, service := range mck8Deployments {

						var hostname string
						for k, v := range service.Annotations {
							if k == "ingress.mck8s/hostname" {
								hostname = v
								break
							}
						}

						deployment, err := mck8s.GetMck8sDeploymentFromService(K8sClient, service)
						if err != nil {
							return err
						}

						log.Info().
							Str("name", service.Name).
							Str("hostname", hostname).
							Bool("enabled", *deployment.Spec.Replicas != 0).
							Msg("deployment found")
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().
			Err(err).
			Msg("Error while running cli")
	}
}
