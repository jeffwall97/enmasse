/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */
package io.enmasse.systemtest.iot;

import static io.enmasse.systemtest.TestTag.sharedIot;
import static org.hamcrest.collection.IsIterableContainingInAnyOrder.containsInAnyOrder;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.List;
import java.util.Optional;
import java.util.concurrent.TimeUnit;

import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import io.enmasse.address.model.Address;
import io.enmasse.address.model.AddressSpace;
import io.enmasse.iot.model.v1.IoTConfigBuilder;
import io.enmasse.iot.model.v1.IoTProject;
import io.enmasse.iot.model.v1.IoTProjectBuilder;
import io.enmasse.systemtest.ability.ITestBaseStandard;
import io.enmasse.systemtest.bases.IoTTestBase;
import io.enmasse.systemtest.utils.UserUtils;
import io.enmasse.user.model.v1.Operation;
import io.enmasse.user.model.v1.User;
import io.enmasse.user.model.v1.UserAuthorization;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.OwnerReference;

@Tag(sharedIot)
class IoTProjectManagedTest extends IoTTestBase implements ITestBaseStandard {

    @Test
    void testCreate() throws Exception {
        createIoTConfig(new IoTConfigBuilder()
                .withNewMetadata()
                .withName("default")
                .endMetadata()
                .withNewSpec()
                .withEnableDefaultRoutes(false)
                .endSpec()
                .build());

        String addressSpaceName = "managed-address-space";

        IoTProject project = new IoTProjectBuilder()
                .withNewMetadata()
                .withName("iot-project-managed")
                .endMetadata()
                .withNewSpec()
                .withNewDownstreamStrategy()
                .withNewManagedStrategy()
                .withNewAddressSpace()
                .withName(addressSpaceName)
                .withPlan("standard-unlimited")
                .withType("standard")
                .endAddressSpace()
                .withNewAddresses()
                .withNewTelemetry()
                .withPlan("standard-small-anycast")
                .withType("anycast")
                .endTelemetry()
                .withNewEvent()
                .withPlan("standard-small-queue")
                .withType("queue")
                .endEvent()
                .withNewCommand()
                .withPlan("standard-small-anycast")
                .withType("anycast")
                .endCommand()
                .endAddresses()
                .endManagedStrategy()
                .endDownstreamStrategy()
                .endSpec()
                .build();

        createIoTProject(project);// waiting until ready

        IoTProject created = iotProjectApiClient.getIoTProject(project.getMetadata().getName());

        assertNotNull(created);
        assertEquals(iotProjectNamespace, created.getMetadata().getNamespace());
        assertEquals(project.getMetadata().getName(), created.getMetadata().getName());
        assertEquals(
                project.getSpec().getDownstreamStrategy().getManagedStrategy().getAddressSpace().getName(),
                created.getSpec().getDownstreamStrategy().getManagedStrategy().getAddressSpace().getName());

        assertManaged(created);

    }

    private void assertManaged(IoTProject project) throws Exception {
        //address space s
        AddressSpace addressSpace = getAddressSpace(project.getSpec().getDownstreamStrategy().getManagedStrategy().getAddressSpace().getName());
        assertEquals(project.getSpec().getDownstreamStrategy().getManagedStrategy().getAddressSpace().getName(), addressSpace.getMetadata().getName());
        assertEquals("standard", addressSpace.getSpec().getType());
        assertEquals("standard-unlimited", addressSpace.getSpec().getPlan());

        //addresses
        //{event/control/telemetry}/"project-namespace"."project-name"
        String addressSuffix = "/"+project.getMetadata().getNamespace()+"."+project.getMetadata().getName();
        List<Address> addresses = getAddressesObjects(addressSpace, Optional.empty()).get(30, TimeUnit.SECONDS);
        assertEquals(3, addresses.size());
        assertEquals(3, addresses.stream()
            .map(Address::getMetadata)
            .map(ObjectMeta::getOwnerReferences)
            .flatMap(List::stream)
            .filter(reference -> isOwner(project, reference))
            .count());
        int correctAddressesCounter = 0;
        for(Address address : addresses) {
            if ( address.getSpec().getAddress().equals(IOT_ADDRESS_EVENT + addressSuffix) ) {
                assertEquals("queue", address.getSpec().getType());
                assertEquals("standard-small-queue", address.getSpec().getPlan());
                correctAddressesCounter++;
            } else if ( address.getSpec().getAddress().equals(IOT_ADDRESS_CONTROL + addressSuffix)
                    || address.getSpec().getAddress().equals(IOT_ADDRESS_TELEMETRY + addressSuffix) ) {
                assertEquals("anycast", address.getSpec().getType());
                assertEquals("standard-small-anycast", address.getSpec().getPlan());
                correctAddressesCounter++;
            }
        }
        assertEquals(3, correctAddressesCounter, "There are incorrect IoT addresses "+addresses);

        //username "adapter"
        //name "project-address-space"+".adapter"
        User user = UserUtils.getUserObject(getUserApiClient(), addressSpace.getMetadata().getName(), "adapter");
        assertNotNull(user);
        assertEquals(1, user.getMetadata().getOwnerReferences().size());
        assertTrue(isOwner(project, user.getMetadata().getOwnerReferences().get(0)));

        UserAuthorization actualAuthorization = user.getSpec().getAuthorization().stream().findFirst().get();

        assertThat(actualAuthorization.getOperations(), containsInAnyOrder(Operation.recv, Operation.send));

        assertThat(actualAuthorization.getAddresses(), containsInAnyOrder(IOT_ADDRESS_EVENT + addressSuffix,
                                                                            IOT_ADDRESS_CONTROL + addressSuffix,
                                                                            IOT_ADDRESS_TELEMETRY + addressSuffix,
                                                                            IOT_ADDRESS_EVENT + addressSuffix + "/*",
                                                                            IOT_ADDRESS_CONTROL + addressSuffix + "/*",
                                                                            IOT_ADDRESS_TELEMETRY + addressSuffix + "/*"));
    }

    private boolean isOwner(IoTProject project, OwnerReference ownerReference) {
        return ownerReference.getKind().equals(IoTProject.KIND) && project.getMetadata().getName().equals(ownerReference.getName());
    }

}
