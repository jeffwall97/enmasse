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
    enmasse "github.com/enmasseproject/enmasse/pkg/client/clientset/versioned"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/klog"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/client/config"
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

func newReconciler(mgr manager.Manager) *ReconcileIoTProject {

    cfg, err := config.GetConfig()
    if err != nil {
        klog.Fatalf("Error getting in-cluster config: %v", err.Error())
    }

    clientset, err := enmasse.NewForConfig(cfg)
    if err != nil {
        klog.Fatalf("Error building EnMasse client: t%v", err.Error())
    }

    return &ReconcileIoTProject{client: mgr.GetClient(), scheme: mgr.GetScheme(), enmasseclientset: clientset}
}

func add(mgr manager.Manager, r *ReconcileIoTProject) error {

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

    ls := NewListerSource(30*time.Second, r.enmasseclientset)
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

    enmasseclientset *enmasse.Clientset
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

    endpointName := strategy.EndpointName
    if len(endpointName) == 0 {
        endpointName = DefaultEndpointName
    }
    portName := strategy.PortName
    if len(portName) == 0 {
        portName = DefaultPortName
    }
    var endpointMode iotv1alpha1.EndpointMode
    if strategy.EndpointMode != nil {
        endpointMode = *strategy.EndpointMode
    } else {
        endpointMode = DefaultEndpointMode
    }

    if len(strategy.Namespace) == 0 {
        return nil, fmt.Errorf("missing namespace")
    }
    if len(strategy.AddressSpaceName) == 0 {
        return nil, fmt.Errorf("missing address space name")
    }

    return r.processProvided(strategy, endpointMode, endpointName, portName)
}

func (r *ReconcileIoTProject) processProvided(strategy *iotv1alpha1.ProvidedDownstreamStrategy, endpointMode iotv1alpha1.EndpointMode, endpointName string, portName string) (*iotv1alpha1.ExternalDownstreamStrategy, error) {

    // FIXME: use cached version, when enmasse#1280 is fixed
    // addressSpace := &enmassealpha1.AddressSpace{}
    // err := r.client.Get(ctx, types.NamespacedName{Namespace: strategy.Namespace, Name: strategy.AddressSpaceName}, addressSpace)

    addressSpace, err := r.enmasseclientset.EnmasseV1alpha1().AddressSpaces(strategy.Namespace).Get(strategy.AddressSpaceName, v1.GetOptions{})
    if err != nil {
        log.WithValues("namespace", strategy.Namespace, "name", strategy.AddressSpaceName).Info("Failed to get address space")
        return nil, err
    }

    if !addressSpace.Status.IsReady {
        // not ready, yet … wait
        return nil, fmt.Errorf("address space is not ready yet")
    }

    endpoint := new(iotv1alpha1.ExternalDownstreamStrategy)

    endpoint.Credentials = strategy.Credentials

    foundEndpoint := false
    for _, es := range addressSpace.Status.EndpointStatus {
        if es.Name != endpointName {
            continue
        }

        foundEndpoint = true

        var ports []enmassealpha1.Port

        switch endpointMode {
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
            if port.Name == portName {
                foundPort = true

                endpoint.Port = port.Port

                tls, err := isTls(addressSpace, &es, &port, strategy)
                if err != nil {
                    return nil, err
                }
                endpoint.TLS = tls

            }
        }

        if !foundPort {
            return nil, fmt.Errorf("unable to find port: %s for endpoint: %s", portName, endpointName)
        }

    }

    if !foundEndpoint {
        return nil, fmt.Errorf("unable to find endpoint: %s", endpointName)
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

    if strategy.TLS != nil {
        return *strategy.TLS, nil
    }

    endpoint := findEndpointSpec(addressSpace, endpointStatus)

    if endpoint == nil {
        return false, fmt.Errorf("failed to locate endpoint named: %v", endpointStatus.Name)
    }

    if endpointStatus.Certificate != nil {
        // if there is a certificate, enable tls
        return true, nil
    }

    if endpoint.Expose != nil {
        // anything set as tls termination counts as tls enabled = true
        return len(endpoint.Expose.RouteTlsTermination) > 0, nil
    }

    return false, nil

}
