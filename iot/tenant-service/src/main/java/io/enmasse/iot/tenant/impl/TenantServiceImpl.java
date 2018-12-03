/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package io.enmasse.iot.tenant.impl;

import javax.security.auth.x500.X500Principal;

import org.eclipse.hono.service.HealthCheckProvider;
import org.eclipse.hono.service.tenant.BaseTenantService;
import org.eclipse.hono.util.TenantResult;
import org.springframework.context.annotation.Scope;
import org.springframework.stereotype.Service;

import io.vertx.core.AsyncResult;
import io.vertx.core.Handler;
import io.vertx.core.json.JsonObject;
import io.vertx.ext.healthchecks.HealthCheckHandler;

@Service
@Scope("prototype")
public class TenantServiceImpl extends BaseTenantService<TenantServiceConfigProperties> implements HealthCheckProvider {

    @Override
    public void setConfig(final TenantServiceConfigProperties configuration) {
    }

    @Override
    public void get(final String tenantId, final Handler<AsyncResult<TenantResult<JsonObject>>> resultHandler) {
        super.get(tenantId, resultHandler);
    }

    @Override
    public void get(final X500Principal subjectDn,
            final Handler<AsyncResult<TenantResult<JsonObject>>> resultHandler) {
        super.get(subjectDn, resultHandler);
    }

    @Override
    public void registerReadinessChecks(final HealthCheckHandler readinessHandler) {
    }

    @Override
    public void registerLivenessChecks(final HealthCheckHandler livenessHandler) {
    }

}
