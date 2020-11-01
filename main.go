/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	rtm "runtime"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	rbacdefinitionsv1beta1 "access-manager/api/v1beta1"
	"access-manager/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var (
	// Version sets the current Operator version
	Version = "0.0.1"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(rbacdefinitionsv1beta1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func printVersion() {
	setupLog.Info(fmt.Sprintf("Operator Version: %s", Version))
	setupLog.Info(fmt.Sprintf("Go Version: %s", rtm.Version()))
	setupLog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", rtm.GOOS, rtm.GOARCH))
}

func main() {
	var enableLeaderElection bool
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	printVersion()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "85a69c09.access-manager.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start access-manager")
		os.Exit(1)
	}

	if err = (&controllers.RbacDefinitionReconciler{
		Client: mgr.GetClient(),
		Logger: ctrl.Log.WithName("controllers").WithName("RbacDefinition"),
		Scheme: mgr.GetScheme(),
		Config: mgr.GetConfig(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RbacDefinition")
		os.Exit(1)
	}

	if err = (&controllers.NamespaceReconciler{
		Client: mgr.GetClient(),
		Logger: ctrl.Log.WithName("controllers").WithName("Namespace"),
		Scheme: mgr.GetScheme(),
		Config: mgr.GetConfig(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Namespace")
		os.Exit(1)
	}

	if err = (&controllers.ServiceAccountReconciler{
		Client: mgr.GetClient(),
		Logger: ctrl.Log.WithName("controllers").WithName("ServiceAccount"),
		Scheme: mgr.GetScheme(),
		Config: mgr.GetConfig(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ServiceAccount")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting access-manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running access-manager")
		os.Exit(1)
	}
}
