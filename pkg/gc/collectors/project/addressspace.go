/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package project

import (
	corev1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/enmasse/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (p *projectCollector) collectAddressSpaces() error {

	log.Info("Collect AddressSpaces")

	opts := metav1.ListOptions{}

	list, err := p.client.EnmasseV1beta1().
		AddressSpaces(p.namespace).
		List(opts)

	if err != nil {
		return err
	}

	for _, as := range list.Items {
		if err := p.checkAddressSpace(&as); err != nil {
			return err
		}
	}

	return nil
}
func (p *projectCollector) checkAddressSpace(as *corev1alpha1.AddressSpace) error {
	log.Info("Checking adddress space", "AddressSpace", as)
	return nil
}
