/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package iotproject

import (
	"fmt"
	iotv1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/iot/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
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
