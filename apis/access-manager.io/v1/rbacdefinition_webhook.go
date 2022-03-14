package v1

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-access-manager-io-v1-RbacDefinition,mutating=true,failurePolicy=ignore,groups=access-manager.io,resources=RbacDefinitions,verbs=create;update,versions=v1,name=rbacdefinition.access-manager.io,sideEffects=None

var rbacDefinitionlog = ctrl.Log.WithName("rbacdefinition-resource")
var _ webhook.Defaulter = &RbacDefinition{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *RbacDefinition) Default() {
	rbacDefinitionlog.Info("default", "name", r.Name)
}

func (r *RbacDefinition) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
