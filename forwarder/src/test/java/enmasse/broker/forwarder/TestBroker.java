/*
 * Copyright 2016 Red Hat Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package enmasse.broker.forwarder;

import org.apache.activemq.artemis.api.core.TransportConfiguration;
import org.apache.activemq.artemis.core.config.Configuration;
import org.apache.activemq.artemis.core.config.impl.ConfigurationImpl;
import org.apache.activemq.artemis.core.remoting.impl.netty.NettyAcceptorFactory;
import org.apache.activemq.artemis.core.server.embedded.EmbeddedActiveMQ;

import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;


public class TestBroker {
    private final String host;
    private final int port;
    private final String address;
    private final EmbeddedActiveMQ server = new EmbeddedActiveMQ();

    public TestBroker(String host, int port, String address) {
        this.host = host;
        this.port = port;
        this.address = address;
    }

    public void start() throws Exception {
        Configuration config = new ConfigurationImpl();

        Map<String, Object> params = new LinkedHashMap<>();
        params.put("protocols", "AMQP");
        params.put("host", host);
        params.put("port", port);
        TransportConfiguration transport = new TransportConfiguration(NettyAcceptorFactory.class.getName(), params, "amqp");

        config.setAcceptorConfigurations(Collections.singleton(transport));
        config.setSecurityEnabled(false);
        config.setName("broker-" + port);

        server.setConfiguration(config);

        server.start();
    }

    public int numConnected() {
        return server.getActiveMQServer().getConnectionCount();
    }

    public void stop() throws Exception {
        server.stop();
    }
}
