/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package iotproject

import (
	"fmt"
	"strings"

	iotv1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/iot/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func splitUserName(name string) (string, string, error) {
	tokens := strings.Split(name, ".")

	if len(tokens) != 2 {
		return "", "", fmt.Errorf("username in wrong format, must be <addressspace>.<name>: %v", name)
	}

	return tokens[0], tokens[1], nil
}

// Convert projects to reconcile requests
func convertToRequests(projects []iotv1alpha1.IoTProject, err error) []reconcile.Request {
	if err != nil {
		return []reconcile.Request{}
	}

	var result []reconcile.Request

	for _, project := range projects {
		result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{
			Namespace: project.Namespace,
			Name:      project.Name,
		}})
	}

	return result
}

func NewOwnerRef(owner v1.Object, gvk schema.GroupVersionKind) *v1.OwnerReference {
	blockOwnerDeletion := false
	isController := false
	return &v1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}

func isSameRef(ref1, ref2 v1.OwnerReference) bool {

	gv1, err := schema.ParseGroupVersion(ref1.APIVersion)
	if err != nil {
		return false
	}

	gv2, err := schema.ParseGroupVersion(ref2.APIVersion)
	if err != nil {
		return false
	}

	return ref1.Kind == ref2.Kind &&
		gv1.Group == gv2.Group &&
		ref1.Name == ref2.Name &&
		ref1.UID == ref2.UID
}
