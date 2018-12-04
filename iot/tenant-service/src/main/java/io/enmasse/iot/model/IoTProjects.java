/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package io.enmasse.iot.model;

import java.lang.reflect.Proxy;

import io.enmasse.iot.model.v1.DoneableIoTProject;
import io.enmasse.iot.model.v1.IoTProject;
import io.enmasse.iot.model.v1.IoTProjectList;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.dsl.MixedOperation;
import io.fabric8.kubernetes.client.dsl.Resource;

public final class IoTProjects {
    private IoTProjects() {
    }

    public interface Client extends
            MixedOperation<IoTProject, IoTProjectList, DoneableIoTProject, Resource<IoTProject, DoneableIoTProject>> {
    }

    public static Client forClient(final KubernetesClient client) {
        final MixedOperation<IoTProject, IoTProjectList, DoneableIoTProject, Resource<IoTProject, DoneableIoTProject>> result = client
                .customResources(IoTProject.CRD, IoTProject.class, IoTProjectList.class, DoneableIoTProject.class);

        return (Client) Proxy.newProxyInstance(IoTProjects.class.getClassLoader(), new Class<?>[] { Client.class },
                (proxy, method, args) -> {
                    return method.invoke(result, args);
                });
    }

}
