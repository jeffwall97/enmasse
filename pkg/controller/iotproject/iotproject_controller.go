/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package iotproject

import (
    "context"
    "encoding/json"
    "fmt"

    iotv1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/iot/v1alpha1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller"
    "sigs.k8s.io/controller-runtime/pkg/handler"
    "sigs.k8s.io/controller-runtime/pkg/manager"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
    "sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_iotproject")

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

    return r.client.Status().Update(ctx, newProject)
}

func (r *ReconcileIoTProject) updateProjectStatusReady(ctx context.Context, request *reconcile.Request, project *iotv1alpha1.IoTProject, endpointStatus *iotv1alpha1.ExternalDownstreamStrategy) error {

    newProject := project.DeepCopy()

    newProject.Status.IsReady = true
    newProject.Status.DownstreamEndpoint = endpointStatus.DeepCopy()

    data, _ := json.Marshal(newProject)
    fmt.Println(string(data))

    return r.client.Update(ctx, newProject)
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

        // we simply copy over the externally provided information

        err = r.updateProjectStatusReady(context.TODO(), &request, project, project.Spec.DownstreamStrategy.ExternalDownstreamStrategy)
        return reconcile.Result{}, err

    } else if project.Spec.DownstreamStrategy.ProvidedDownstreamStrategy != nil {

        // handling as provided

        // FIXME: implement

        return reconcile.Result{}, nil

    } else {

        // unknown strategy, we don't know how to handle this
        // so re-queuing doesn't make any sense

        err = r.updateProjectStatusError(context.TODO(), &request, project)

        return reconcile.Result{}, err

    }

    /*

       // Define a new Pod object
       pod := newPodForCR(instance)

       // Set IoTProject instance as the owner and controller
       if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
           return reconcile.Result{}, err
       }

       // Check if this Pod already exists
       found := &corev1.Pod{}
       err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
       if err != nil && errors.IsNotFound(err) {
           reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
           err = r.client.Create(context.TODO(), pod)
           if err != nil {
               return reconcile.Result{}, err
           }

           // Pod created successfully - don't requeue
           return reconcile.Result{}, nil
       } else if err != nil {
           return reconcile.Result{}, err
       }

       // Pod already exists - don't requeue
       reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
       return reconcile.Result{}, nil
    */
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *iotv1alpha1.IoTProject) *corev1.Pod {
    labels := map[string]string{
        "app": cr.Name,
    }
    return &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      cr.Name + "-pod",
            Namespace: cr.Namespace,
            Labels:    labels,
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name:    "busybox",
                    Image:   "busybox",
                    Command: []string{"sleep", "3600"},
                },
            },
        },
    }
}
