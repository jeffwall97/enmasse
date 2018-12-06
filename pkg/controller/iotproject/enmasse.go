/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package iotproject

import (
    "context"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/util/workqueue"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/event"
    "sigs.k8s.io/controller-runtime/pkg/handler"
    "sigs.k8s.io/controller-runtime/pkg/predicate"
    "time"
)

type listerSource struct {
    GVK      schema.GroupVersionKind
    Interval time.Duration
    Client   client.Client
}

func NewListerSource(interval time.Duration, gvk schema.GroupVersionKind, client client.Client) listerSource {
    return listerSource{
        GVK:      gvk,
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
                ul := &unstructured.UnstructuredList{}
                ul.SetGroupVersionKind(ls.GVK)
                err := ls.Client.List(context.Background(), nil, ul)
                if err != nil {
                    log.WithValues("gvk", ls.GVK).Info("unable to list resources for GV during reconcilation")
                    continue
                }
                for _, u := range ul.Items {
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
