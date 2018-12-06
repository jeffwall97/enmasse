/*
 * Copyright 2018, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package main

import (
    "flag"
    "fmt"
    "github.com/enmasseproject/enmasse/pkg/client/clientset/versioned/scheme"
    "os"
    "runtime"

    "github.com/enmasseproject/enmasse/pkg/controller"
    sdkVersion "github.com/operator-framework/operator-sdk/version"
    _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
    "sigs.k8s.io/controller-runtime/pkg/client/config"
    "sigs.k8s.io/controller-runtime/pkg/manager"
    logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
    "sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var log = logf.Log.WithName("cmd")

func printVersion() {
    log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
    log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
    log.Info(fmt.Sprintf("operator-sdk Version: %v", sdkVersion.Version))
}

func main() {
    flag.Parse()

    logf.SetLogger(logf.ZapLogger(true /* FIXME: switch to production, or make configurable */))

    printVersion()

    /*
       namespace, err := k8sutil.GetWatchNamespace()
       if err != nil {
           log.Error(err, "failed to get watch namespace")
           os.Exit(1)
       }
    */

    cfg, err := config.GetConfig()
    if err != nil {
        log.Error(err, "Failed to get configuration")
        os.Exit(1)
    }

    mgr, err := manager.New(cfg, manager.Options{Namespace: "" /*NOTE: watching all namespaces*/ })
    if err != nil {
        log.Error(err, "")
        os.Exit(1)
    }

    log.Info("Registering components...")

    // register APIs

    scheme.AddToScheme(mgr.GetScheme())

    // register controller

    if err := controller.AddToManager(mgr); err != nil {
        log.Error(err, "Failed to register controller")
        os.Exit(1)
    }

    log.Info("Starting the operator")

    if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
        log.Error(err, "manager exited non-zero")
        os.Exit(1)
    }
}
