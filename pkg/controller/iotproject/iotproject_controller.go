/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package iotproject

import (
	"context"
	"fmt"
	enmassev1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/enmasse/v1beta1"
	iotv1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/iot/v1alpha1"
	userv1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/user/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

	/*
	   cfg, err := config.GetConfig()
	   if err != nil {
		   klog.Fatalf("Error getting in-cluster config: %v", err.Error())
	   }

	   clientset, err := enmasse.NewForConfig(cfg)
	   if err != nil {
		   klog.Fatalf("Error building EnMasse client: t%v", err.Error())
	   }
	*/

	return &ReconcileIoTProject{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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

	// Watch for enmasse address space

	ownerHandler := ForkedEnqueueRequestForOwner{
		OwnerType:    &iotv1alpha1.IoTProject{},
		IsController: true,
	}

	err = c.Watch(&source.Kind{Type: &enmassev1alpha1.AddressSpace{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {

				log.V(2).Info("Change event", "kind", "AddressSpace", "object", a)

				// check if we have an owner

				result := ownerHandler.GetOwnerReconcileRequest(a.Meta)

				if result != nil && len(result) > 0 {
					log.Info("Owned resource")
					// looks like an owned resource ... take this is a result
					return result
				}

				// we need to actively look for a mapped resource

				// a is the AddressSpace that changed
				addressSpaceNamespace := a.Meta.GetNamespace()
				addressSpaceName := a.Meta.GetName()

				log.Info("Looking up IoT project for un-owned addressspace", "AddressSpaceNamespace", addressSpaceNamespace, "AddressSpaceName", addressSpaceName)

				// look for an iot project, that references this address space

				return convertToRequests(r.findIoTProjectsByMappedAddressSpaces(addressSpaceNamespace, addressSpaceName))
			}),
		})
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

	// enmasseclientset *enmasse.Clientset
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

			reqLogger.Info("Project was not found")

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

		reqLogger.Info("Handle as external")

		status, err := r.reconcileExternal(context.TODO(), &request, project)
		return r.applyUpdate(status, err, &request, project)

	} else if project.Spec.DownstreamStrategy.ProvidedDownstreamStrategy != nil {

		// handling as provided

		reqLogger.Info("Handle as provided")

		status, err := r.reconcileProvided(context.TODO(), &request, project)
		return r.applyUpdate(status, err, &request, project)

	} else if project.Spec.DownstreamStrategy.ManagedDownstreamStrategy != nil {

		// handling as managed

		reqLogger.Info("Handle as managed")

		status, err := r.reconcileManaged(context.TODO(), &request, project)
		return r.applyUpdate(status, err, &request, project)

	} else {

		// unknown strategy, we don't know how to handle this
		// so re-queuing doesn't make any sense

		reqLogger.Info("Missing or unknown downstream strategy")

		err = r.updateProjectStatusError(context.TODO(), &request, project)
		return reconcile.Result{}, err

	}

}

func (r *ReconcileIoTProject) reconcileExternal(ctx context.Context, request *reconcile.Request, project *iotv1alpha1.IoTProject) (*iotv1alpha1.ExternalDownstreamStrategy, error) {
	// we simply copy over the external information

	return project.Spec.DownstreamStrategy.ExternalDownstreamStrategy, nil
}

func getOrDefaults(strategy *iotv1alpha1.ProvidedDownstreamStrategy) (string, string, iotv1alpha1.EndpointMode, error) {
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
		return "", "", 0, fmt.Errorf("missing namespace")
	}
	if len(strategy.AddressSpaceName) == 0 {
		return "", "", 0, fmt.Errorf("missing address space name")
	}

	return endpointName, portName, endpointMode, nil
}

func (r *ReconcileIoTProject) reconcileProvided(ctx context.Context, request *reconcile.Request, project *iotv1alpha1.IoTProject) (*iotv1alpha1.ExternalDownstreamStrategy, error) {

	log.Info("Reconcile project with provided strategy")

	strategy := project.Spec.DownstreamStrategy.ProvidedDownstreamStrategy
	endpointName, portName, endpointMode, err := getOrDefaults(strategy)

	if err != nil {
		return nil, err
	}

	return r.processProvided(strategy, endpointMode, endpointName, portName)
}

func (r *ReconcileIoTProject) processProvided(strategy *iotv1alpha1.ProvidedDownstreamStrategy, endpointMode iotv1alpha1.EndpointMode, endpointName string, portName string) (*iotv1alpha1.ExternalDownstreamStrategy, error) {

	addressSpace := &enmassev1alpha1.AddressSpace{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: strategy.Namespace, Name: strategy.AddressSpaceName}, addressSpace)

	// addressSpace, err := r.enmasseclientset.EnmasseV1alpha1().AddressSpaces(strategy.Namespace).Get(strategy.AddressSpaceName, v1.GetOptions{})
	if err != nil {
		log.WithValues("namespace", strategy.Namespace, "name", strategy.AddressSpaceName).Info("Failed to get address space")
		return nil, err
	}

	return extractEndpointInformation(endpointName, endpointMode, portName, &strategy.Credentials, addressSpace, strategy.TLS)
}

func extractEndpointInformation(
	endpointName string,
	endpointMode iotv1alpha1.EndpointMode,
	portName string,
	credentials *iotv1alpha1.Credentials,
	addressSpace *enmassev1alpha1.AddressSpace,
	forceTls *bool,
) (*iotv1alpha1.ExternalDownstreamStrategy, error) {

	if !addressSpace.Status.IsReady {
		// not ready, yet … wait
		return nil, fmt.Errorf("address space is not ready yet")
	}

	endpoint := new(iotv1alpha1.ExternalDownstreamStrategy)

	endpoint.Credentials = *credentials

	foundEndpoint := false
	for _, es := range addressSpace.Status.EndpointStatus {
		if es.Name != endpointName {
			continue
		}

		foundEndpoint = true

		var ports []enmassev1alpha1.Port

		switch endpointMode {
		case iotv1alpha1.Service:
			endpoint.Host = es.ServiceHost
			ports = es.ServicePorts
		case iotv1alpha1.External:
			endpoint.Host = es.ExternalHost
			ports = es.ExternalPorts
		}

		log.V(2).Info("Ports to scan", "ports", ports)

		endpoint.Certificate = es.Certificate

		foundPort := false
		for _, port := range ports {
			if port.Name == portName {
				foundPort = true

				endpoint.Port = port.Port

				tls, err := isTls(addressSpace, &es, &port, forceTls)
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

func findEndpointSpec(addressSpace *enmassev1alpha1.AddressSpace, endpointStatus *enmassev1alpha1.EndpointStatus) *enmassev1alpha1.EndpointSpec {
	for _, end := range addressSpace.Spec.Ednpoints {
		if end.Name != endpointStatus.Name {
			continue
		}
		return &end
	}
	return nil
}

// get a an estimate if TLS should be enabled for a port, or not
func isTls(
	addressSpace *enmassev1alpha1.AddressSpace,
	endpointStatus *enmassev1alpha1.EndpointStatus,
	port *enmassev1alpha1.Port,
	forceTls *bool) (bool, error) {

	if forceTls != nil {
		return *forceTls, nil
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

func (r *ReconcileIoTProject) ensureOwnerIsSet(owner, object v1.Object) error {

	ts := object.GetCreationTimestamp()
	if ts.IsZero() {
		err := controllerutil.SetControllerReference(owner, object, r.scheme)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileIoTProject) reconcileManaged(ctx context.Context, request *reconcile.Request, project *iotv1alpha1.IoTProject) (*iotv1alpha1.ExternalDownstreamStrategy, error) {

	log.Info("Reconcile project with managed strategy")

	strategy := project.Spec.DownstreamStrategy.ManagedDownstreamStrategy

	addressSpace := &enmassev1alpha1.AddressSpace{
		ObjectMeta: v1.ObjectMeta{Namespace: project.Namespace, Name: strategy.AddressSpaceName},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, addressSpace, func(existing runtime.Object) error {
		existingAddressSpace := existing.(*enmassev1alpha1.AddressSpace)

		if err := r.ensureOwnerIsSet(project, existingAddressSpace); err != nil {
			return err
		}

		log.Info("Reconcile address space", "AddressSpace", existingAddressSpace)

		return reconcileAddressSpace(project, strategy, existingAddressSpace)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return nil, err
	}

	adapterUser := &userv1alpha1.MessagingUser{
		ObjectMeta: v1.ObjectMeta{Namespace: project.Namespace, Name: strategy.AddressSpaceName + ".adapter"},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.client, adapterUser, func(existing runtime.Object) error {
		existingUser := existing.(*userv1alpha1.MessagingUser)

		if err := r.ensureOwnerIsSet(project, existingUser); err != nil {
			return err
		}

		log.Info("Reconcile messaging user", "MessagingUser", existingUser)

		return reconcileAdapterMessagingUser(project, existingUser)
	})

	if err != nil {
		log.Error(err, "Failed calling CreateOrUpdate")
		return nil, err
	}

	/*
	   addressSpace := newAddressSpace(project, strategy)
	   if err := controllerutil.SetControllerReference(project, addressSpace, r.scheme); err != nil {
		   return nil, err
	   }

	   found := &enmassev1alpha1.AddressSpace{}
	   err := r.client.Get(ctx, types.NamespacedName{Name: addressSpace.Name, Namespace: addressSpace.Namespace}, found)
	   if err != nil && errors.IsNotFound(err) {
		   log.Info("Creating a new AddressSpace", "AddressSpace.Namespace", addressSpace.Namespace, "AddressSpace.Name", addressSpace.Name)
		   if err = r.client.Create(context.TODO(), addressSpace); err != nil {
			   return nil, err
		   }
	   }

	   credentials := newMessagingUser(project, addressSpace)
	*/

	credentials := iotv1alpha1.Credentials{
		Username: adapterUser.Spec.Username,
		Password: "bar",
	}

	forceTls := true

	return extractEndpointInformation("messaging", iotv1alpha1.Service, "amqps", &credentials, addressSpace, &forceTls)
}

func reconcileAddressSpace(project *iotv1alpha1.IoTProject, strategy *iotv1alpha1.ManagedDownstreamStrategy, existing *enmassev1alpha1.AddressSpace) error {

	if existing.CreationTimestamp.IsZero() {
		existing.ObjectMeta.Labels = project.Labels
	}

	existing.Spec = enmassev1alpha1.AddressSpaceSpec{
		Type: "standard",
		Plan: "standard-unlimited",
	}

	return nil
}

func reconcileAdapterMessagingUser(project *iotv1alpha1.IoTProject, existing *userv1alpha1.MessagingUser) error {

	username := "adapter"
	password := "bar"
	tenant := project.Namespace + "." + project.Name

	existing.Spec = userv1alpha1.MessagingUserSpec{

		Username: username,

		Authentication: userv1alpha1.AuthenticationSpec{
			Type:     "password",
			Password: password,
		},

		Authorization: []userv1alpha1.AuthorizationSpec{
			{
				Addresses: []string{
					"telemetry/" + tenant + "/#",
					"event/" + tenant + "/#",
					"command/" + tenant + "/#",
				},
				Operations: []string{
					"send",
					"recv",
				},
			},
		},
	}

	return nil
}
