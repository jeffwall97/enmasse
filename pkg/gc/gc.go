/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package gc

import (
	"github.com/enmasseproject/enmasse/pkg/gc/collectors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"time"
)

var log = logf.Log.WithName("gc")

type garbageCollector struct {
	ticker     *time.Ticker
	stopChan   chan bool
	collectors []collectors.Collector
}

func NewGarbageCollector() *garbageCollector {

	ticker := time.NewTicker(1 * time.Minute)
	stopChan := make(chan bool)

	result := &garbageCollector{
		ticker:   ticker,
		stopChan: stopChan,
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				result.Collect()
			case <-stopChan:
				return
			}
		}
	}()

	return result
}

func (g *garbageCollector) Stop() {
	g.stopChan <- true
	g.ticker.Stop()
}

func (g *garbageCollector) Run(stopCh <-chan struct{}) {
	<-stopCh
}

func (g *garbageCollector) AddCollector(collector collectors.Collector) {
	g.collectors = append(g.collectors, collector)
}

func (g *garbageCollector) Collect() {

	for _, collector := range g.collectors {
		if err := collector.CollectOnce(); err != nil {
			log.Error(err, "Failed to process collector", "collector", collector)
		}
	}
}
