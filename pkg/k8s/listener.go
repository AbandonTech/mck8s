package k8s

import (
	"context"
	"github.com/bep/debounce"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

// Payload contains information gathered via the Listener
type Payload struct {
	Hostname string
	Service  *v1.Service
}

// Listener uses the kubernetes API to listen to events,
// forwarding them via the Listener.onPayload() function to be handled.
type Listener struct {
	client    kubernetes.Interface
	onPayload func(*Payload)
}

// NewListener creates a new Listener, onPayload is a callback expected to handle each
// event listened to.
func NewListener(client kubernetes.Interface, onPayload func(*Payload)) *Listener {
	return &Listener{
		client:    client,
		onPayload: onPayload,
	}
}

// Run a Listener in a Blocking fashion.
// Any event listened to will be constructed into a Payload with the expectation
// to be handled by the Listener.onPayload() method.
func (listener *Listener) Run(ctx context.Context) error {
	log.Debug().Msg("Starting kubernetes event listener")
	factory := informers.NewSharedInformerFactory(listener.client, time.Minute)
	serviceLister := factory.Core().V1().Services().Lister()

	constructAndForwardPayload := func() {
		payload := &Payload{}

		services, err := serviceLister.List(labels.Everything())
		if err != nil {
			log.Error().
				Err(err).
				Msg("failed to list services")

			return
		}

	forService:
		for _, service := range services {
			for annotationKey, hostname := range service.Annotations {
				if annotationKey == "ingress.mck8s/hostname" {
					payload.Hostname = hostname
					payload.Service = service

					log.Info().
						Str("Service", service.Name).
						Str("Namespace", service.Namespace).
						Str("Hostname", hostname).
						Msg("Service found for ingress")

					break forService
				}
			}
		}

		listener.onPayload(payload)
	}

	debounced := debounce.New(time.Second)
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Debug().
				Msg("Received an add event")
			debounced(constructAndForwardPayload)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Debug().
				Msg("Received an updated event")
			debounced(constructAndForwardPayload)
		},
		DeleteFunc: func(obj interface{}) {
			log.Debug().
				Msg("Received an delete event")
			debounced(constructAndForwardPayload)
		},
	}

	informer := factory.Core().V1().Services().Informer()
	informer.AddEventHandler(handler)
	informer.Run(ctx.Done())
	return nil
}
