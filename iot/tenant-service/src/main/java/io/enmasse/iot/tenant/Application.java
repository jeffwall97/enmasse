/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package io.enmasse.iot.tenant;

import org.eclipse.hono.service.AbstractApplication;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.ComponentScan;

@SpringBootApplication
@ComponentScan("io.enmasse.iot.tenant.config")
@ComponentScan("io.enmasse.iot.tenant.impl")
public class Application extends AbstractApplication {

    public static void main(final String[] args) {
        SpringApplication.run(Application.class, args);
    }

}
