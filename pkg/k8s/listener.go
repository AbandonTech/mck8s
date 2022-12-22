package k8s

import (
	"context"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

// HostnameToService contains information gathered via the Listener
type HostnameToService struct {
	Delete   bool
	Hostname string
	Service  *v1.Service
}

// Listener uses the kubernetes API to listen to events,
// forwarding them via the Listener.onServiceDiscovered() function to be handled.
type Listener struct {
	client              kubernetes.Interface
	onServiceDiscovered func([]HostnameToService)
}

// NewListener creates a new Listener, onServiceDiscovered is a callback expected to handle each
// event listened to.
func NewListener(client kubernetes.Interface, onServiceDiscovered func([]HostnameToService)) *Listener {
	return &Listener{
		client:              client,
		onServiceDiscovered: onServiceDiscovered,
	}
}

// Run a Listener in a Blocking fashion.
// Any event listened to will be constructed into a HostnameToService with the expectation
// to be handled by the Listener.onServiceDiscovered() method.
func (listener *Listener) Run(ctx context.Context) error {
	log.Debug().Msg("Starting kubernetes event listener")
	factory := informers.NewSharedInformerFactory(listener.client, time.Minute)
	serviceLister := factory.Core().V1().Services().Lister()

	constructAndForwardPayload := func(delete bool) {
		hostnameToServices := make([]HostnameToService, 0)

		services, err := serviceLister.List(labels.Everything())
		if err != nil {
			log.Error().
				Err(err).
				Msg("failed to list services")

			return
		}

		for _, service := range services {
			for annotationKey, hostname := range service.Annotations {
				if annotationKey == "ingress.mck8s/hostname" {
					hostnameToServices = append(
						hostnameToServices,
						HostnameToService{
							Delete:   delete,
							Hostname: hostname,
							Service:  service,
						})

					log.Debug().
						Str("Service", service.Name).
						Str("Namespace", service.Namespace).
						Str("Hostname", hostname).
						Msg("Service found for ingress")
					break
				}
			}
		}

		listener.onServiceDiscovered(hostnameToServices)
	}

	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			constructAndForwardPayload(false)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			constructAndForwardPayload(false)
		},
		DeleteFunc: func(obj interface{}) {
			constructAndForwardPayload(true)
		},
	}

	informer := factory.Core().V1().Services().Informer()
	informer.AddEventHandler(handler)
	informer.Run(ctx.Done())
	return nil
}
