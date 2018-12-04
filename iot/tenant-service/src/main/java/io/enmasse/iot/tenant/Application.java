/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package io.enmasse.iot.tenant;

import java.util.ArrayList;
import java.util.List;

import org.eclipse.hono.service.AbstractApplication;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ApplicationContext;
import org.springframework.context.annotation.ComponentScan;

import io.enmasse.iot.tenant.impl.TenantServiceImpl;
import io.vertx.core.CompositeFuture;
import io.vertx.core.Future;
import io.vertx.core.Vertx;

@SpringBootApplication
@ComponentScan("io.enmasse.iot.tenant.config")
@ComponentScan("io.enmasse.iot.tenant.impl")
public class Application extends AbstractApplication {

    public static void main(final String[] args) {
        SpringApplication.run(Application.class, args);
    }

    private ApplicationContext context;

    @Autowired
    public void context(ApplicationContext context) {
        this.context = context;
    }

    @Override
    protected Future<Void> deployRequiredVerticles(int maxInstances) {

        final Vertx vertx = getVertx();

        @SuppressWarnings("rawtypes")
        final List<Future> deploymentTracker = new ArrayList<>();

        for (int i = 0; i < maxInstances; i++) {

            final TenantServiceImpl serviceInstance = this.context.getBean(TenantServiceImpl.class);

            final Future<String> deployTracker = Future.future();
            vertx.deployVerticle(serviceInstance, deployTracker.completer());
            deploymentTracker.add(deployTracker);

            registerHealthchecks(serviceInstance);
        }

        final Future<Void> result = Future.future();
        CompositeFuture.all(deploymentTracker)
                .setHandler(r -> {
                    if (r.failed()) {
                        result.fail(r.cause());
                    } else {
                        result.complete(null);
                    }
                });

        return result;
    }

}
