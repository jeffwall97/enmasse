/*
 * Copyright 2019, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */
package io.enmasse.systemtest.utils;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.enmasse.admin.model.v1.AuthenticationService;
import io.enmasse.admin.model.v1.AuthenticationServiceBuilder;
import io.enmasse.admin.model.v1.AuthenticationServiceType;
import io.enmasse.admin.model.v1.DoneableAuthenticationService;
import io.enmasse.systemtest.CustomLogger;
import io.vertx.core.json.JsonObject;
import org.slf4j.Logger;

import java.util.HashMap;
import java.util.Map;

public class AuthServiceUtils {
    private static Logger log = CustomLogger.getLogger();

    public static AuthenticationService createNoneAuthServiceObject(String name) {
        return createAuthService(name, AuthenticationServiceType.none).done();
    }

    public static AuthenticationService createStandardAuthServiceObject(String name, boolean persistent) {
        Map<String, String> storage = new HashMap<>();
        storage.put("type", persistent ? "persistent-claim" : "ephemeral");
        if (persistent) {
            storage.put("delete-claim", "true");
            storage.put("size", "2Gi");
        }
        return createAuthService(name, AuthenticationServiceType.standard)
                .editSpec()
                .withNewStandard()
                .addToAdditionalProperties("storage", storage)
                .endStandard()
                .endSpec()
                .done();
    }

    private static DoneableAuthenticationService createAuthService(String name, AuthenticationServiceType type) {
        return new DoneableAuthenticationService(new AuthenticationServiceBuilder()
                .withNewMetadata()
                .withName(name)
                .endMetadata()
                .withNewSpec()
                .withType(type)
                .endSpec()
                .build());
    }

    public static JsonObject authenticationServiceToJson(AuthenticationService service) throws JsonProcessingException {
        return new JsonObject(new ObjectMapper().writeValueAsString(service));
    }

    public static AuthenticationService jsonToAuthenticationService(JsonObject authService) throws Exception {
        return new ObjectMapper().readValue(authService.toString(), AuthenticationService.class);
    }
}
