package reconciler

import (
	accessmanagerv1beta1 "access-manager/apis/access-manager.io/v1beta1"
	"access-manager/pkg/util"
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var secretName = "SyncSecretDefinition"

// ReconcileSyncSecretDefinition applies all desired changes of the SyncSecretDefinition
func (r *Reconciler) ReconcileSyncSecretDefinition(instance *accessmanagerv1beta1.SyncSecretDefinition) (reconcile.Result, error) {
	secrets := r.BuildAllSecrets(instance)
	r.RemoveOwnedSecrets(instance.Name)

	for _, s := range secrets {
		// Set SyncSecretDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &s, r.Scheme); err != nil {
			r.Logger.WithValues("Secret", s.Namespace+"/"+s.Name).Error(err, "Failed to set controllerReference.")
			continue
		}

		if _, err := r.CreateSecret(s); err != nil {
			r.Logger.WithValues("Secret", s.Namespace+"/"+s.Name).Error(err, "Failed to reconcile Secret.")
		}
	}

	return reconcile.Result{}, nil
}

// ReconcileSecret applies all desired changes of the Secret
func (r *Reconciler) ReconcileSecret(instance *corev1.Secret) (reconcile.Result, error) {
	list := &accessmanagerv1beta1.SyncSecretDefinitionList{}
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
func (r *Reconciler) BuildAllSecrets(cr *accessmanagerv1beta1.SyncSecretDefinition) []corev1.Secret {
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
			r.Logger.Info("Existing Secret is not owned by a SyncSecretDefinition. Ignoring", "Secret.Name", existing.Name)
			return existing, nil
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	r.Logger.Info("Creating new Secret", "Secret.Name", s.Name)
	return r.Client.CoreV1().Secrets(s.Namespace).Create(context.TODO(), &s, metav1.CreateOptions{})
}

// RemoveOwnedSecrets deletes all secrets which are owned from the given object name.
func (r *Reconciler) RemoveOwnedSecrets(defName string) {
	ownedSecrets, err := r.getOwnedSecrets(defName)

	if err != nil {
		r.Logger.Error(err, "Failed to fetch all owned Secrets.")
	}

	for _, s := range ownedSecrets {
		r.Logger.Info("Deleting Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
		err = r.Client.CoreV1().Secrets(s.Namespace).Delete(context.TODO(), s.Name, metav1.DeleteOptions{})

		if err != nil {
			r.Logger.WithValues("Secret", s.Name, "Namespace", s.Namespace).Error(err, "Failed to delete Secret.")
		}
	}
}

func (r *Reconciler) getOwnedSecrets(defName string) ([]corev1.Secret, error) {
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

func (r *Reconciler) isSecretRelevant(spec accessmanagerv1beta1.SyncSecretDefinition, secret *corev1.Secret) bool {
	return spec.Spec.Source.Name == secret.Name && spec.Spec.Source.Namespace == secret.Namespace
}
