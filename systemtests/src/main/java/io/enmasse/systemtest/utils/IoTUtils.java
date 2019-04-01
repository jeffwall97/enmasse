/*
 * Copyright 2019, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */
package io.enmasse.systemtest.utils;

import java.util.concurrent.TimeUnit;

import org.slf4j.Logger;

import io.enmasse.address.model.AddressSpace;
import io.enmasse.iot.model.v1.IoTConfig;
import io.enmasse.iot.model.v1.IoTProject;
import io.enmasse.systemtest.CustomLogger;
import io.enmasse.systemtest.Kubernetes;
import io.enmasse.systemtest.TimeoutBudget;
import io.enmasse.systemtest.apiclients.AddressApiClient;
import io.enmasse.systemtest.apiclients.IoTConfigApiClient;
import io.enmasse.systemtest.apiclients.IoTProjectApiClient;

public class IoTUtils {

    private static Logger log = CustomLogger.getLogger();

    public static void waitForIoTConfigReady(IoTConfigApiClient apiClient, IoTConfig config) throws Exception {
        boolean isReady = false;
        TimeoutBudget budget = new TimeoutBudget(5, TimeUnit.MINUTES);
        while (budget.timeLeft() >= 0 && !isReady) {
            config = apiClient.getIoTConfig(config.getMetadata().getName());
            isReady = config.getStatus()!=null && config.getStatus().isInitialized();
            if (!isReady) {
                Thread.sleep(10000);
            }
            log.info("Waiting until IoTConfig: '{}' will be in ready state", config.getMetadata().getName());
        }
        if (!isReady) {
            String jsonStatus = config != null ? config.getStatus().getState() : "";
            throw new IllegalStateException("IoTConfig " + config.getMetadata().getName() + " is not in Ready state within timeout: " + jsonStatus);
        }
    }

    public static void waitForIoTProjectReady(IoTProjectApiClient apiClient, IoTProject project) throws Exception {
        boolean isReady = false;
        TimeoutBudget budget = new TimeoutBudget(10, TimeUnit.MINUTES);
        while (budget.timeLeft() >= 0 && !isReady) {
            project = apiClient.getIoTProject(project.getMetadata().getName());
            isReady = project.getStatus()!=null && project.getStatus().isReady();
            if (!isReady) {
                Thread.sleep(10000);
            }
            log.info("Waiting until IoTProject: '{}' will be in ready state", project.getMetadata().getName());
        }
        if (!isReady) {
            String jsonStatus = project != null && project.getStatus() != null ? project.getStatus().toString() : "Project doesn't have status";
            throw new IllegalStateException("IoTProject " + project.getMetadata().getName() + " is not in Ready state within timeout: " + jsonStatus);
        }
    }

    public static void waitForIoTProjectDeleted(Kubernetes kubernetes, AddressApiClient addressApiClient, IoTProject project) throws Exception {
        if(project.getSpec().getDownstreamStrategy().getManagedStrategy() != null) {
            String addressSpaceName = project.getSpec().getDownstreamStrategy().getManagedStrategy().getAddressSpace().getName();
            AddressSpace addressSpace = AddressSpaceUtils.getAddressSpacesObjects(addressApiClient)
                .stream()
                .filter(space -> space.getMetadata().getName().equals(addressSpaceName))
                .findFirst()
                .orElse(null);
            if(addressSpace != null) {
                AddressSpaceUtils.waitForAddressSpaceDeleted(kubernetes, addressSpace);
            }
        }
    }

}
