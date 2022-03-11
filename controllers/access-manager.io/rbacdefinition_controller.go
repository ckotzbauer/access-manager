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

package controllers

import (
	"context"

	v1beta1 "github.com/ckotzbauer/access-manager/apis/access-manager.io/v1beta1"
	"github.com/ckotzbauer/access-manager/pkg/reconciler"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbacdefinitionsv1beta1 "github.com/ckotzbauer/access-manager/apis/access-manager.io/v1beta1"
)

// RbacDefinitionReconciler reconciles a RbacDefinition object
type RbacDefinitionReconciler struct {
	Client client.Client
	Config *rest.Config
	Scheme *runtime.Scheme
	Logger logr.Logger
}

func (r *RbacDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Logger.WithValues("rbacdefinition", req.NamespacedName)

	// Fetch the RbacDefinition instance
	instance := &v1beta1.RbacDefinition{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.Logger.Error(err, "Unexpected error occurred!")
		return reconcile.Result{}, err
	}

	if instance.Spec.Paused {
		return reconcile.Result{}, nil
	}

	r.Logger.Info("Reconciling RbacDefinition", "Name", req.Name)
	rec := reconciler.Reconciler{Client: *kubernetes.NewForConfigOrDie(r.Config), Logger: r.Logger, Scheme: r.Scheme}
	return rec.ReconcileRbacDefinition(instance)
}

func (r *RbacDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacdefinitionsv1beta1.RbacDefinition{}).
		Complete(r)
}
