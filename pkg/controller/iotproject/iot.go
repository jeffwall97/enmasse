/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package iotproject

import (
	"context"
	iotv1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/iot/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ReconcileIoTProject) findIoTProjectsByPredicate(predicate func(project *iotv1alpha1.IoTProject) bool) ([]iotv1alpha1.IoTProject, error) {

	var result []iotv1alpha1.IoTProject

	opts := &client.ListOptions{}
	list := &iotv1alpha1.IoTProjectList{}

	err := r.client.List(context.TODO(), opts, list)
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		if predicate(&item) {
			result = append(result, item)
		}
	}

	return result, nil
}

func (r *ReconcileIoTProject) findIoTProjectsByMappedAddressSpaces(addressSpaceNamespace string, addressSpaceName string) ([]iotv1alpha1.IoTProject, error) {

	// FIXME: brute force scanning through all IoT projects.
	//        This could be improved if field selectors for CRDs would be implemented, which they
	//        current are not.

	// Look for all IoTProjects where:
	//    spec.downstreamStrategy.providedStrategy.namespace = addressSpaceNamespace
	//      and
	//    spec.downstreamStrategy.providedStrategy.addressSpaceName = addressSpaceName

	return r.findIoTProjectsByPredicate(func(project *iotv1alpha1.IoTProject) bool {
		if project.Spec.DownstreamStrategy.ProvidedDownstreamStrategy == nil {
			return false
		}
		if project.Spec.DownstreamStrategy.ProvidedDownstreamStrategy.Namespace != addressSpaceNamespace {
			return false
		}
		if project.Spec.DownstreamStrategy.ProvidedDownstreamStrategy.AddressSpaceName != addressSpaceName {
			return false
		}
		return true
	})
}
