package v1

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-access-manager-io-v1-SyncSecretDefinition,mutating=true,failurePolicy=ignore,groups=access-manager.io,resources=SyncSecretDefinitions,verbs=create;update,versions=v1,name=syncSecretdefinition.access-manager.io,sideEffects=None

var syncSecretDefinitionlog = ctrl.Log.WithName("syncSecretdefinition-resource")
var _ webhook.Defaulter = &SyncSecretDefinition{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SyncSecretDefinition) Default() {
	syncSecretDefinitionlog.Info("default", "name", r.Name)
}

func (r *SyncSecretDefinition) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
