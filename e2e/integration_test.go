package integration_test

import (
	accessmanagerv1beta1 "access-manager/apis/access-manager.io/v1beta1"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var (
	rbacDefGVR = schema.GroupVersionResource{
		Group:    "access-manager.io",
		Version:  "v1beta1",
		Resource: "rbacdefinitions",
	}
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func getRoleBinding(c kubernetes.Clientset, ctx context.Context, name string, namespace string) (*rbacv1.RoleBinding, error) {
	return c.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
}

func getClusterRoleBinding(c kubernetes.Clientset, ctx context.Context, name string) (*rbacv1.ClusterRoleBinding, error) {
	return c.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
}

func createServiceAccount(c kubernetes.Clientset, ctx context.Context, serviceAccount corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	return c.CoreV1().ServiceAccounts(serviceAccount.Namespace).Create(ctx, &serviceAccount, metav1.CreateOptions{})
}

func deleteServiceAccount(c kubernetes.Clientset, ctx context.Context, namespace, name string) error {
	return c.CoreV1().ServiceAccounts(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func patchNamespace(c kubernetes.Interface, ctx context.Context, cur, mod corev1.Namespace) error {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, corev1.Namespace{})
	if err != nil {
		return err
	}

	if len(patch) == 0 || string(patch) == "{}" {
		return nil
	}

	_, err = c.CoreV1().Namespaces().Patch(ctx, cur.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}

func addNamespaceLabel(c kubernetes.Clientset, ctx context.Context, namespace string, labelKey string, labelValue string) error {
	current, _ := c.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	mod := current.DeepCopy()

	if mod.ObjectMeta.Labels == nil {
		mod.ObjectMeta.Labels = map[string]string{}
	}

	mod.ObjectMeta.Labels[labelKey] = labelValue
	return patchNamespace(&c, ctx, *current, *mod)
}

func deleteNamespaceLabel(c kubernetes.Clientset, ctx context.Context, namespace string, labelKey string) error {
	current, _ := c.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	mod := current.DeepCopy()

	delete(mod.ObjectMeta.Labels, labelKey)
	return patchNamespace(&c, ctx, *current, *mod)
}

func checkRoleBindingToBeEquivalent(rb rbacv1.RoleBinding, expected rbacv1.RoleBinding) {
	Expect(rb.Name).To(BeEquivalentTo(expected.Name))
	Expect(rb.Namespace).To(BeEquivalentTo(expected.Namespace))
	Expect(rb.RoleRef).To(BeEquivalentTo(expected.RoleRef))
	Expect(rb.Subjects).To(BeEquivalentTo(expected.Subjects))
}

func checkClusterRoleBindingToBeEquivalent(crb rbacv1.ClusterRoleBinding, expected rbacv1.ClusterRoleBinding) {
	Expect(crb.Name).To(BeEquivalentTo(expected.Name))
	Expect(crb.RoleRef).To(BeEquivalentTo(expected.RoleRef))
	Expect(crb.Subjects).To(BeEquivalentTo(expected.Subjects))
}

func createRbacDefinition(c dynamic.Interface, ctx context.Context, def accessmanagerv1beta1.RbacDefinition) error {
	res := c.Resource(rbacDefGVR)
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&def)
	if err != nil {
		return err
	}

	unstructuredObj["kind"] = "RbacDefinition"
	unstructuredObj["apiVersion"] = rbacDefGVR.Group + "/" + rbacDefGVR.Version
	log.Printf("Creating RbacDefinition %s", def.Name)
	_, err = res.Create(ctx, &unstructured.Unstructured{Object: unstructuredObj}, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Failed to create RbacDefinition %#v", def)
	}

	return err
}

func deleteRbacDefinition(c dynamic.Interface, ctx context.Context, def accessmanagerv1beta1.RbacDefinition) error {
	res := c.Resource(rbacDefGVR)

	log.Printf("Deleting RbacDefinition %s", def.Name)
	err := res.Delete(ctx, def.Name, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Failed to delete RbacDefinition %#v", def)
	}

	return err
}

var _ = Describe("IntegrationTest", func() {
	var def1 accessmanagerv1beta1.RbacDefinition
	var def2 accessmanagerv1beta1.RbacDefinition
	var def3 accessmanagerv1beta1.RbacDefinition
	ctx := context.TODO()

	def1 = accessmanagerv1beta1.RbacDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rbac-def1",
		},
		Spec: accessmanagerv1beta1.RbacDefinitionSpec{
			Namespaced: []accessmanagerv1beta1.NamespacedSpec{
				{
					NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"ci": "true"}},
					Bindings: []accessmanagerv1beta1.BindingsSpec{
						{
							Kind:     "Role",
							RoleName: "test-role",
							Subjects: []rbacv1.Subject{
								{
									Kind:      "ServiceAccount",
									Name:      "default",
									Namespace: "namespace1",
								},
							},
						},
					},
				},
			},
		},
	}

	def2 = accessmanagerv1beta1.RbacDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rbac-def2",
		},
		Spec: accessmanagerv1beta1.RbacDefinitionSpec{
			Cluster: []accessmanagerv1beta1.ClusterSpec{
				{
					ClusterRoleName: "test-role",
					Subjects: []rbacv1.Subject{
						{
							Kind:      "ServiceAccount",
							Name:      "default",
							Namespace: "namespace2",
						},
					},
				},
			},
		},
	}

	def3 = accessmanagerv1beta1.RbacDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rbac-def3",
		},
		Spec: accessmanagerv1beta1.RbacDefinitionSpec{
			Namespaced: []accessmanagerv1beta1.NamespacedSpec{
				{
					Namespace: accessmanagerv1beta1.NamespaceSpec{
						Name: "namespace4",
					},
					Bindings: []accessmanagerv1beta1.BindingsSpec{
						{
							Name:               "test-rolebinding",
							RoleName:           "test-role",
							Kind:               "Role",
							AllServiceAccounts: true,
							Subjects:           []rbacv1.Subject{},
						},
					},
				},
			},
		},
	}

	Describe("RbacDefinition", func() {
		It("should apply new RoleBinding", func(done Done) {
			err := createRbacDefinition(client, ctx, def1)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(3 * time.Second)

			expectedRb := rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "test-role", Namespace: "namespace1"},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Name:     "test-role",
					Kind:     "Role",
				},
				Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Name: "default", Namespace: "namespace1"}},
			}

			rb, err := getRoleBinding(*clientset, ctx, "test-role", "namespace1")
			Expect(err).NotTo(HaveOccurred())
			checkRoleBindingToBeEquivalent(*rb, expectedRb)
			close(done)
		}, 5)

		It("should apply new ClusterRoleBinding", func(done Done) {
			err := createRbacDefinition(client, ctx, def2)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(3 * time.Second)

			expectedCrb := rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "test-role"},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Name:     "test-role",
					Kind:     "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Name: "default", Namespace: "namespace2"}},
			}

			crb, err := getClusterRoleBinding(*clientset, ctx, "test-role")
			Expect(err).NotTo(HaveOccurred())
			checkClusterRoleBindingToBeEquivalent(*crb, expectedCrb)
			close(done)
		}, 5)

		It("should delete ClusterRoleBinding on definition removal", func(done Done) {
			err := deleteRbacDefinition(client, ctx, def2)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(3 * time.Second)

			_, err = getClusterRoleBinding(*clientset, ctx, "test-role")
			Expect(errors.IsNotFound(err)).To(BeTrue())
			close(done)
		}, 5)

		It("should create a RoleBinding if namespace is labeled", func(done Done) {
			err := addNamespaceLabel(*clientset, ctx, "namespace3", "ci", "true")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(3 * time.Second)

			expectedRb := rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "test-role", Namespace: "namespace3"},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Name:     "test-role",
					Kind:     "Role",
				},
				Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Name: "default", Namespace: "namespace1"}},
			}

			rb, err := getRoleBinding(*clientset, ctx, "test-role", "namespace3")
			Expect(err).NotTo(HaveOccurred())
			checkRoleBindingToBeEquivalent(*rb, expectedRb)
			close(done)
		}, 5)

		It("should delete a RoleBinding if namespace is unlabeled", func(done Done) {
			err := deleteNamespaceLabel(*clientset, ctx, "namespace3", "ci")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(3 * time.Second)

			_, err = getRoleBinding(*clientset, ctx, "test-role", "namespace3")
			Expect(errors.IsNotFound(err)).To(BeTrue())
			close(done)
		}, 5)

		It("should modify RoleBinding on ServiceAccount creation", func(done Done) {
			err := createRbacDefinition(client, ctx, def3)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(3 * time.Second)

			expectedRb := rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "test-rolebinding", Namespace: "namespace4"},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Name:     "test-role",
					Kind:     "Role",
				},
				Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Name: "default", Namespace: ""}},
			}

			rb, err := getRoleBinding(*clientset, ctx, "test-rolebinding", "namespace4")
			Expect(err).NotTo(HaveOccurred())
			checkRoleBindingToBeEquivalent(*rb, expectedRb)

			createServiceAccount(*clientset, ctx, corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "new-sa", Namespace: "namespace4"}})
			time.Sleep(3 * time.Second)

			expectedRb = rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "test-rolebinding", Namespace: "namespace4"},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Name:     "test-role",
					Kind:     "Role",
				},
				Subjects: []rbacv1.Subject{
					{Kind: "ServiceAccount", Name: "default", Namespace: ""},
					{Kind: "ServiceAccount", Name: "new-sa", Namespace: ""},
				},
			}

			rb, err = getRoleBinding(*clientset, ctx, "test-rolebinding", "namespace4")
			Expect(err).NotTo(HaveOccurred())
			checkRoleBindingToBeEquivalent(*rb, expectedRb)
			close(done)
		}, 10)
	})
})
