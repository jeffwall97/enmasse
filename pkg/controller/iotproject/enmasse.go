/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package iotproject

import (
    "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/util/workqueue"
    "sigs.k8s.io/controller-runtime/pkg/event"
    "sigs.k8s.io/controller-runtime/pkg/handler"
    "sigs.k8s.io/controller-runtime/pkg/predicate"
    "time"

    enmasse "github.com/enmasseproject/enmasse/pkg/client/clientset/versioned"
)

type listerSource struct {
    GVK      schema.GroupVersionKind
    Interval time.Duration
    Client   *enmasse.Clientset
}

func NewListerSource(interval time.Duration, client *enmasse.Clientset) listerSource {

    return listerSource{
        Interval: interval,
        Client:   client,
    }
}

func (ls *listerSource) Start(handler handler.EventHandler, workqueue workqueue.RateLimitingInterface, predicates ...predicate.Predicate) error {
    Stop := make(chan struct{})

    go func() {
        ticker := time.NewTicker(ls.Interval)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                // List all object for the GVK
                // ul := &unstructured.UnstructuredList{}
                // ul.SetGroupVersionKind(ls.GVK)
                opts := v1.ListOptions{}
                list, err := ls.Client.EnmasseV1alpha1().AddressSpaces("").List(opts)
                if err != nil {
                    log.WithValues("gvk", ls.GVK).Info("unable to list resources for GV during reconcilation")
                    continue
                }
                for _, u := range list.Items {
                    e := event.GenericEvent{
                        Meta:   &u,
                        Object: &u,
                    }
                    handler.Generic(e, workqueue)
                }
            case <-Stop:
                return
            }
        }
    }()

    return nil
}
