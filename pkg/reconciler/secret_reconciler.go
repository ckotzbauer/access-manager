package reconciler

import (
	"context"
	"fmt"
	"reflect"

	v1 "github.com/ckotzbauer/access-manager/apis/access-manager.io/v1"
	"github.com/ckotzbauer/access-manager/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var secretName = "SyncSecretDefinition"

// ReconcileSyncSecretDefinition applies all desired changes of the SyncSecretDefinition
func (r *Reconciler) ReconcileSyncSecretDefinition(instance *v1.SyncSecretDefinition) (reconcile.Result, error) {
	secrets := r.BuildAllSecrets(instance)
	ownedSecrets, err := r.GetOwnedSecrets(instance.Name)

	if err != nil {
		r.Logger.Error(err, "Failed to fetch all owned Secrets.")
	}

	r.RemoveOwnedSecretsNotInList(ownedSecrets, secrets)

	if err != nil {
		r.Logger.Error(err, "Failed to fetch owned Secrets.")
		return reconcile.Result{}, err
	}

	for _, s := range secrets {
		// Set SyncSecretDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &s, r.Scheme); err != nil {
			r.Logger.WithValues("Secret", s.Namespace+"/"+s.Name).Error(err, "Failed to set controllerReference.")
			continue
		}

		existingSecret, err := r.getSecretFromSlice(ownedSecrets, s)

		if err == nil && r.hasSecretChanged(existingSecret, s) {
			r.Logger.Info("Reconciling Secret", "Name", fmt.Sprintf("%s/%s", existingSecret.Namespace, existingSecret.Name))
			r.removeSecret(existingSecret)
		} else if err == nil {
			continue
		} else {
			r.Logger.Info("Reconciling Secret", "Name", fmt.Sprintf("%s/%s", s.Namespace, s.Name))
		}

		if _, err := r.CreateSecret(s); err != nil {
			r.Logger.WithValues("Secret", s.Namespace+"/"+s.Name).Error(err, "Failed to reconcile Secret.")
		}
	}

	return reconcile.Result{}, nil
}

// ReconcileSecret applies all desired changes of the Secret
func (r *Reconciler) ReconcileSecret(instance *corev1.Secret) (reconcile.Result, error) {
	list := &v1.SyncSecretDefinitionList{}
	err := r.ControllerClient.List(context.TODO(), list)

	if err != nil {
		r.Logger.Error(err, "Unexpected error occurred!")
		return reconcile.Result{}, err
	}

	for _, def := range list.Items {
		if def.Spec.Paused {
			continue
		}

		if !r.isSecretRelevant(def, instance) {
			continue
		}

		_, err = r.ReconcileSyncSecretDefinition(&def)

		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// BuildAllSecrets returns an array of Secrets for the given SyncSecretDefinition
func (r *Reconciler) BuildAllSecrets(cr *v1.SyncSecretDefinition) []corev1.Secret {
	var secrets []corev1.Secret = []corev1.Secret{}
	sourceSecret, err := r.getSourceSecret(cr.Spec.Source.Name, cr.Spec.Source.Namespace)

	if err != nil {
		if errors.IsNotFound(err) {
			return secrets
		}

		r.Logger.WithValues("Secret", cr.Spec.Source.Namespace+"/"+cr.Spec.Source.Name).Error(err, "Failed to fetch source secret.")
		return secrets
	}

	for _, nsSpec := range cr.Spec.Targets {
		relevantNamespaces := r.GetRelevantNamespaces(nsSpec.NamespaceSelector, nsSpec.Namespace)
		if relevantNamespaces == nil {
			return nil
		}

		r.Logger.WithValues("Namespaces", util.MapNamespaces(relevantNamespaces)).Info("Found matching Namespaces.")

		for _, ns := range relevantNamespaces {
			secret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sourceSecret.Name,
					Namespace: ns.Name,
				},
				Type:      sourceSecret.Type,
				Data:      sourceSecret.Data,
				Immutable: sourceSecret.Immutable,
			}

			secrets = append(secrets, secret)
		}
	}

	return secrets
}

// CreateSecret creates a new Secret
func (r *Reconciler) CreateSecret(s corev1.Secret) (*corev1.Secret, error) {
	existing, err := r.Client.CoreV1().Secrets(s.Namespace).Get(context.TODO(), s.Name, metav1.GetOptions{})
	if err == nil {
		if !HasNamedOwner(existing.OwnerReferences, secretName, "") {
			r.Logger.Info("Existing Secret is not owned by a SyncSecretDefinition. Ignoring", "Name", fmt.Sprintf("%s/%s", existing.Namespace, existing.Name))
			return existing, nil
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	r.Logger.Info("Creating new Secret", "Name", fmt.Sprintf("%s/%s", s.Namespace, s.Name))
	return r.Client.CoreV1().Secrets(s.Namespace).Create(context.TODO(), &s, metav1.CreateOptions{})
}

// RemoveOwnedSecretsNotInList deletes all secrets which are owned from the given object name and not in the slice.
func (r *Reconciler) RemoveOwnedSecretsNotInList(ownedSecrets []corev1.Secret, secrets []corev1.Secret) {
	for _, s := range ownedSecrets {
		if !r.containsSecret(s, secrets) {
			r.removeSecret(s)
		}
	}
}

func (r *Reconciler) removeSecret(s corev1.Secret) {
	r.Logger.Info("Deleting Secret", "Name", fmt.Sprintf("%s/%s", s.Namespace, s.Name))
	err := r.Client.CoreV1().Secrets(s.Namespace).Delete(context.TODO(), s.Name, metav1.DeleteOptions{})

	if err != nil {
		r.Logger.WithValues("Name", fmt.Sprintf("%s/%s", s.Namespace, s.Name)).Error(err, "Failed to delete Secret.")
	}
}

// GetOwnedSecrets returns a slice of all secrets which are owned by the given definition name.
func (r *Reconciler) GetOwnedSecrets(defName string) ([]corev1.Secret, error) {
	list, err := r.Client.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	var secrets []corev1.Secret = []corev1.Secret{}

	for _, s := range list.Items {
		if HasNamedOwner(s.OwnerReferences, secretName, defName) {
			secrets = append(secrets, s)
		}
	}

	return secrets, nil
}

func (r *Reconciler) getSourceSecret(name, ns string) (*corev1.Secret, error) {
	return r.Client.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
}

func (r *Reconciler) isSecretRelevant(spec v1.SyncSecretDefinition, secret *corev1.Secret) bool {
	return spec.Spec.Source.Name == secret.Name && spec.Spec.Source.Namespace == secret.Namespace
}

func (r *Reconciler) getSecretFromSlice(secrets []corev1.Secret, secret corev1.Secret) (corev1.Secret, error) {
	for _, s := range secrets {
		if s.Namespace == secret.Namespace && s.Name == secret.Name {
			return s, nil
		}
	}

	return corev1.Secret{}, fmt.Errorf("no secret found")
}

func (r *Reconciler) hasSecretChanged(existingSecret corev1.Secret, secret corev1.Secret) bool {
	return existingSecret.Namespace != secret.Namespace ||
		existingSecret.Name != secret.Name ||
		existingSecret.Type != secret.Type ||
		!reflect.DeepEqual(existingSecret.Data, secret.Data)
}

func (r *Reconciler) containsSecret(secret corev1.Secret, secrets []corev1.Secret) bool {
	for _, s := range secrets {
		if s.Namespace == secret.Namespace && s.Name == secret.Name {
			return true
		}
	}

	return false
}
