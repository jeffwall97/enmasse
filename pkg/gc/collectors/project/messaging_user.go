/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package project

import (
    "github.com/enmasseproject/enmasse/pkg/apis/iot/v1alpha1"
    userv1alpha1 "github.com/enmasseproject/enmasse/pkg/apis/user/v1beta1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    "strings"
)

func (p *projectCollector) collectMessagingUsers() error {

    log.Info("Collect MessagingUsers")

    opts := metav1.ListOptions{}

    list, err := p.client.UserV1alpha1().
        MessagingUsers(p.namespace).
        List(opts)

    if err != nil {
        return err
    }

    for _, user := range list.Items {
        if err := p.checkMessagingUser(&user); err != nil {
            return err
        }
    }

    return nil
}

func (p *projectCollector) checkMessagingUser(user *userv1alpha1.MessagingUser) error {
    log.Info("Checking messaging user", "MessagingUser", user)

    if user.ObjectMeta.OwnerReferences != nil {
        for _, owner := range user.ObjectMeta.OwnerReferences {

            if owner.Kind != "IoTProject" {
                continue
            }

            project, err := p.findOwnerProject(user.Namespace, &owner)
            if err != nil {
                log.Error(err, "Failed to locate owning project", "MessagingUser", user)
                continue
            }

            if ! p.shouldDeleteUserForProject(user, project) {
                continue
            }

            // no valid owner anymore

            if err := p.deleteMessagingUser(user, &owner.UID); err != nil {
                log.Error(err, "Failed to delete MessagingUser", "MessagingUser", user)
            }
        }
    }

    return nil
}

func (p *projectCollector) shouldDeleteUserForProject(user *userv1alpha1.MessagingUser, project *v1alpha1.IoTProject) bool {

    if project == nil {
        // owning project is gone
        return true
    }

    if project.Spec.DownstreamStrategy.ManagedDownstreamStrategy == nil {
        // owning project is no longer of typed "managed"
        return true
    }

    toks := strings.Split(user.Name, ".")
    if len(toks) != 2 {
        // invalid user name format ... better delete this
        return true
    }

    if toks[0] != project.Spec.DownstreamStrategy.ManagedDownstreamStrategy.AddressSpaceName {
        // address space name doesn't match user name ... as we own it, we delete it
        return true
    }

    // keep it … for now ;-)
    return false

}

// find the owner project of a resource
func (p *projectCollector) findOwnerProject(namespace string, owner *metav1.OwnerReference) (*v1alpha1.IoTProject, error) {
    if owner.Kind != "IoTProject" {
        return nil, nil
    }

    project, err := p.client.IotV1alpha1().
        IoTProjects(namespace).
        Get(owner.Name, metav1.GetOptions{})

    if err != nil {
        if errors.IsNotFound(err) {
            return nil, nil
        }
        return nil, err
    }

    if project.ObjectMeta.UID != owner.UID {
        return nil, nil
    }

    return project, nil
}

func (p *projectCollector) deleteMessagingUser(user *userv1alpha1.MessagingUser, uid *types.UID) error {
    log.Info("Deleting Messaging User", "MessagingUser", user, "UID", uid)
    return p.client.UserV1alpha1().
        MessagingUsers(user.Namespace).
        Delete(user.Name, &metav1.DeleteOptions{
            Preconditions: &metav1.Preconditions{UID: uid},
        })
}
