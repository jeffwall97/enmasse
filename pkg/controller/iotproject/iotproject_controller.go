/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package iotproject

import (
    "context"
    "fmt"
    enmassealpha1 "github.com/enmasseproject/enmasse/pkg/apis/enmasse/v1alpha1"
    iotv1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/iot/v1alpha1"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller"
    "sigs.k8s.io/controller-runtime/pkg/handler"
    "sigs.k8s.io/controller-runtime/pkg/manager"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
    "sigs.k8s.io/controller-runtime/pkg/source"
    "time"
)

var log = logf.Log.WithName("controller_iotproject")

const DefaultEndpointName = "messaging"
const DefaultPortName = "amqps"
const DefaultEndpointMode = iotv1alpha1.Service

// Gets called by parent "init", adding as to the manager
func Add(mgr manager.Manager) error {
    return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
    return &ReconcileIoTProject{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
    // Create a new controller
    c, err := controller.New("iotproject-controller", mgr, controller.Options{Reconciler: r})
    if err != nil {
        return err
    }

    // Watch for changes to primary resource IoTProject
    err = c.Watch(&source.Kind{Type: &iotv1alpha1.IoTProject{}}, &handler.EnqueueRequestForObject{})
    if err != nil {
        return err
    }

    gvk := enmassealpha1.SchemeGroupVersion.WithKind("AddressSpace")
    ls := NewListerSource(30*time.Second, gvk, mgr.GetClient())
    err = c.Watch(&ls, &handler.EnqueueRequestForObject{})
    if err != nil {
        return err
    }

    return nil
}

var _ reconcile.Reconciler = &ReconcileIoTProject{}

type ReconcileIoTProject struct {
    // This client, initialized using mgr.Client() above, is a split client
    // that reads objects from the cache and writes to the apiserver
    client client.Client
    scheme *runtime.Scheme
}

func (r *ReconcileIoTProject) updateProjectStatusError(ctx context.Context, request *reconcile.Request, project *iotv1alpha1.IoTProject) error {

    newProject := project.DeepCopy()
    newProject.Status.IsReady = false
    newProject.Status.DownstreamEndpoint = nil

    return r.client.Update(ctx, newProject)
}

func (r *ReconcileIoTProject) updateProjectStatusReady(ctx context.Context, request *reconcile.Request, project *iotv1alpha1.IoTProject, endpointStatus *iotv1alpha1.ExternalDownstreamStrategy) error {

    newProject := project.DeepCopy()

    newProject.Status.IsReady = true
    newProject.Status.DownstreamEndpoint = endpointStatus.DeepCopy()

    return r.client.Update(ctx, newProject)
}

func (r *ReconcileIoTProject) applyUpdate(status *iotv1alpha1.ExternalDownstreamStrategy, err error, request *reconcile.Request, project *iotv1alpha1.IoTProject) (reconcile.Result, error) {

    if err != nil {
        log.Error(err, "failed to reconcile")
        err = r.updateProjectStatusError(context.TODO(), request, project)
        return reconcile.Result{}, err
    }

    err = r.updateProjectStatusReady(context.TODO(), request, project, status)
    return reconcile.Result{}, err
}

// Reconcile by reading the IoT project spec and making required changes
//
// returning an error will get the request re-queued
func (r *ReconcileIoTProject) Reconcile(request reconcile.Request) (reconcile.Result, error) {
    reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
    reqLogger.Info("Reconciling IoTProject")

    // Get project
    project := &iotv1alpha1.IoTProject{}
    err := r.client.Get(context.TODO(), request.NamespacedName, project)

    if err != nil {

        if errors.IsNotFound(err) {
            // Request object not found, could have been deleted after reconcile request.
            // Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
            // Return and don't requeue
            return reconcile.Result{}, nil
        }

        // Error reading the object - requeue the request.
        return reconcile.Result{}, err
    }

    if project.Spec.DownstreamStrategy.ExternalDownstreamStrategy != nil {

        // handling as external

        status, err := r.reconcileExternal(context.TODO(), &request, project)
        return r.applyUpdate(status, err, &request, project)

    } else if project.Spec.DownstreamStrategy.ProvidedDownstreamStrategy != nil {

        // handling as provided

        status, err := r.reconcileProvided(context.TODO(), &request, project)
        return r.applyUpdate(status, err, &request, project)

    } else {

        // unknown strategy, we don't know how to handle this
        // so re-queuing doesn't make any sense

        err = r.updateProjectStatusError(context.TODO(), &request, project)

        return reconcile.Result{}, err

    }

}

func (r *ReconcileIoTProject) reconcileExternal(ctx context.Context, request *reconcile.Request, project *iotv1alpha1.IoTProject) (*iotv1alpha1.ExternalDownstreamStrategy, error) {
    // we simply copy over the externally provided information

    return project.Spec.DownstreamStrategy.ExternalDownstreamStrategy, nil
}

func (r *ReconcileIoTProject) reconcileProvided(ctx context.Context, request *reconcile.Request, project *iotv1alpha1.IoTProject) (*iotv1alpha1.ExternalDownstreamStrategy, error) {

    log.Info("Reconcile project with provided strategy")

    strategy := project.Spec.DownstreamStrategy.ProvidedDownstreamStrategy

    if len(strategy.EndpointName) == 0 {
        strategy.EndpointName = DefaultEndpointName
    }
    if len(strategy.PortName) == 0 {
        strategy.PortName = DefaultPortName
    }
    if strategy.EndpointMode == nil {
        c := DefaultEndpointMode
        strategy.EndpointMode = &c
    }

    if len(strategy.Namespace) == 0 {
        return nil, fmt.Errorf("missing namespace")
    }
    if len(strategy.AddressSpaceName) == 0 {
        return nil, fmt.Errorf("missing address space name")
    }

    addressSpace := &enmassealpha1.AddressSpace{}
    err := r.client.Get(ctx, types.NamespacedName{Namespace: strategy.Namespace, Name: strategy.AddressSpaceName}, addressSpace)
    if err != nil {
        log.WithValues("namespace", strategy.Namespace, "name", strategy.AddressSpaceName).Info("Failed to get address space")
        return nil, err
    }

    if !addressSpace.Status.IsReady {
        err = r.updateProjectStatusError(ctx, request, project)
        // not ready, yet … wait
        return nil, err
    }

    endpoint := new(iotv1alpha1.ExternalDownstreamStrategy)

    endpoint.Credentials = strategy.Credentials

    foundEndpoint := false
    for _, es := range addressSpace.Status.EndpointStatus {
        if es.Name != strategy.EndpointName {
            continue
        }

        foundEndpoint = true

        var ports []enmassealpha1.Port

        switch *strategy.EndpointMode {
        case iotv1alpha1.Service:
            endpoint.Host = es.ServiceHost
            ports = es.ServicePorts
        case iotv1alpha1.External:
            endpoint.Host = es.ExternalHost
            ports = es.ExternalPorts
        }

        log.WithValues("ports", ports).Info("Ports to scan")

        endpoint.Certificate = es.Certificate

        foundPort := false
        for _, port := range ports {
            if port.Name == strategy.PortName {
                foundPort = true

                endpoint.Port = port.Port

                tls, err := isTls(addressSpace, &es, &port, strategy)
                if err != nil {
                    endpoint.TLS = tls
                } else {
                    return nil, err
                }

            }
        }

        if !foundPort {
            return nil, fmt.Errorf("unable to find port: %s for endpoint: %s", strategy.PortName, strategy.EndpointName)
        }

    }

    if !foundEndpoint {
        return nil, fmt.Errorf("unable to find endpoint: %s", strategy.EndpointName)
    }

    return endpoint, nil
}

func findEndpointSpec(addressSpace *enmassealpha1.AddressSpace, endpointStatus *enmassealpha1.EndpointStatus) *enmassealpha1.EndpointSpec {
    for _, end := range addressSpace.Spec.Ednpoints {
        if end.Name != endpointStatus.Name {
            continue
        }
        return &end
    }
    return nil
}

func isTls(
    addressSpace *enmassealpha1.AddressSpace,
    endpointStatus *enmassealpha1.EndpointStatus,
    port *enmassealpha1.Port,
    strategy *iotv1alpha1.ProvidedDownstreamStrategy) (bool, error) {

    if strategy.DisableTLS {
        // TLS is forced off
        return false, nil
    }

    endpoint := findEndpointSpec(addressSpace, endpointStatus)

    if endpoint == nil {
        return false, fmt.Errorf("failed to locate endpoint named: %v", endpointStatus.Name)
    }

    if endpointStatus.Certificate != nil {
        return true, nil
    }

    if endpoint.Expose != nil {
        switch endpoint.Expose.RouteTlsTermination {
        case "reencrypt":
        case "edge":
            return true, nil
        }
    }

    return false, nil

}
