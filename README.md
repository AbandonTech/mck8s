# Kubernetes Minecraft Ingress

Enable ingress routing for minecraft servers.

## Running locally

Set up a Kubernetes cluster and ensure you are able to address services by their cluster IP. I recommend doing this
using [minikube](https://minikube.sigs.k8s.io/docs/start/) for a local cluster and using their `minikube tunnel` to
enable addressing services via their cluster IP.

Once a cluster has been created, and you can address the local services. Run 
```shell
go run cmd/ingress-controller.go --kubeconfig ~/.kube/config # Assuming your kubeconfig is in the default location.
```

To create some Minecraft Server services, use the [provided example manifest](deploy/local/mck8s-server.yml). You can apply
these to your cluster using.
```shell
kubectl apply -f ./deploy/mck8s-server.yml
```


## Resources

- [Article | How to build a custom Kubernetes Ingress Controller](https://www.doxsey.net/blog/how-to-build-a-custom-kubernetes-ingress-controller-in-go/)
