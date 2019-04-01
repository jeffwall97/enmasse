/*
 * Copyright 2019, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */
package io.enmasse.iot.model.v1;

import static com.fasterxml.jackson.annotation.JsonInclude.Include.NON_NULL;

import com.fasterxml.jackson.annotation.JsonInclude;

import io.fabric8.kubernetes.api.model.Doneable;
import io.sundr.builder.annotations.Buildable;
import io.sundr.builder.annotations.Inline;

@Buildable(
        editableEnabled = false,
        generateBuilderPackage = false,
        builderPackage = "io.fabric8.kubernetes.api.builder",
        inline = @Inline(
                type = Doneable.class,
                prefix = "Doneable",
                value = "done"
                )
        )
@JsonInclude(NON_NULL)
public class AddressesConfig {

    private AddressConfig telemetry;
    private AddressConfig event;
    private AddressConfig command;

    public AddressConfig getTelemetry() {
        return telemetry;
    }

    public void setTelemetry(AddressConfig telemetry) {
        this.telemetry = telemetry;
    }

    public AddressConfig getEvent() {
        return event;
    }

    public void setEvent(AddressConfig event) {
        this.event = event;
    }

    public AddressConfig getCommand() {
        return command;
    }

    public void setCommand(AddressConfig command) {
        this.command = command;
    }

}
