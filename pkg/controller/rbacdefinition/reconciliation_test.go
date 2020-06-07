package rbacdefinition_test

import (
	accessmanagerv1beta1 "access-manager/pkg/apis/accessmanager/v1beta1"
	"access-manager/pkg/controller/rbacdefinition"
	"context"
	"fmt"
	"sync/atomic"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func createNamespaces(ctx context.Context, nss ...*corev1.Namespace) {
	for _, ns := range nss {
		_, err := clientset.CoreV1().Namespaces().Get(ctx, ns.Name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			ns, err = clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

func createClusterRoleBinding(ctx context.Context, crb *rbacv1.ClusterRoleBinding) {
	_, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, crb.Name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		crb, err = clientset.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
	}
}

func createRoleBinding(ctx context.Context, rb *rbacv1.RoleBinding) {
	_, err := clientset.RbacV1().RoleBindings("default").Get(ctx, rb.Name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		rb, err = clientset.RbacV1().RoleBindings("default").Create(ctx, rb, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
	}
}

var _ = Describe("Reconciliation", func() {
	var namespace1 *corev1.Namespace
	var namespace2 *corev1.Namespace
	var namespace3 *corev1.Namespace
	//var clusterRoleBinding *rbacv1.ClusterRoleBinding
	//var roleBinding *rbacv1.RoleBinding
	var count uint64 = 0
	var scheme *runtime.Scheme
	var logger logr.Logger
	var def *rbacdefinition.ReconcileRbacDefinition
	ctx := context.TODO()

	BeforeEach(func(done Done) {
		atomic.AddUint64(&count, 1)
		namespace1 = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   fmt.Sprintf("ns-one-%v", count),
				Labels: map[string]string{"team": fmt.Sprintf("one-%v", count)},
			},
			Spec: corev1.NamespaceSpec{},
		}
		namespace2 = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   fmt.Sprintf("ns-two-%v", count),
				Labels: map[string]string{"team": fmt.Sprintf("two-%v", count), "ci": fmt.Sprintf("true-%v", count)},
			},
			Spec: corev1.NamespaceSpec{},
		}
		namespace3 = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   fmt.Sprintf("ns-three-%v", count),
				Labels: map[string]string{"team": fmt.Sprintf("three-%v", count), "ci": fmt.Sprintf("true-%v", count)},
			},
			Spec: corev1.NamespaceSpec{},
		}
		clusterRoleBinding := rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("existing-crb-%v", count)},
			RoleRef: rbacv1.RoleRef{
				Name: "test-role",
				Kind: "ClusterRole",
			},
			Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
		}
		roleBinding := rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("existing-rb-%v", count)},
			RoleRef: rbacv1.RoleRef{
				Name: "test-role",
				Kind: "ClusterRole",
			},
			Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
		}

		scheme = kscheme.Scheme
		logger = log.Log.WithName("testLogger")
		def = &rbacdefinition.ReconcileRbacDefinition{Client: *clientset, Scheme: scheme, Logger: logger}
		createNamespaces(ctx, namespace1, namespace2, namespace3)
		createClusterRoleBinding(ctx, &clusterRoleBinding)
		createRoleBinding(ctx, &roleBinding)
		close(done)
	})

	AfterEach(func(done Done) {
		close(done)
	})

	Describe("GetRelevantNamespaces", func() {
		It("should not match any namespace", func(done Done) {
			spec := &accessmanagerv1beta1.NamespacedSpec{NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"no": "match"}}}

			found, err := rbacdefinition.GetRelevantNamespaces(*spec, def)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeEmpty())
			close(done)
		})

		It("should match namespace1", func(done Done) {
			spec := &accessmanagerv1beta1.NamespacedSpec{
				NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"team": fmt.Sprintf("one-%v", count)}},
			}

			found, err := rbacdefinition.GetRelevantNamespaces(*spec, def)
			Expect(err).NotTo(HaveOccurred())
			Expect(rbacdefinition.MapNamespaces(found, rbacdefinition.MapNamespaceName)).To(BeEquivalentTo([]string{namespace1.Name}))
			close(done)
		})

		It("should match namespace2 and namespace3", func(done Done) {
			spec := &accessmanagerv1beta1.NamespacedSpec{
				NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"ci": fmt.Sprintf("true-%v", count)}},
			}

			found, err := rbacdefinition.GetRelevantNamespaces(*spec, def)
			Expect(err).NotTo(HaveOccurred())
			Expect(rbacdefinition.MapNamespaces(found, rbacdefinition.MapNamespaceName)).To(BeEquivalentTo([]string{namespace3.Name, namespace2.Name}))
			close(done)
		})
	})

	Describe("BuildAllClusterRoleBindings", func() {
		It("should return empty array", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Cluster: []accessmanagerv1beta1.ClusterSpec{},
				},
			}

			clusterRoles := rbacdefinition.BuildAllClusterRoleBindings(cr)
			Expect(clusterRoles).To(BeEmpty())
			close(done)
		})

		It("should return correct ClusterRoleBindings", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Cluster: []accessmanagerv1beta1.ClusterSpec{
						{
							ClusterRoleName: "test-role",
							Subjects:        []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "default"}},
						},
						{
							ClusterRoleName: "admin-role",
							Subjects:        []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "Group", Name: "admins"}},
						},
						{
							ClusterRoleName: "john-role",
							Subjects:        []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: "john"}},
						},
					},
				},
			}

			expectedBindings := []rbacv1.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "test-role"},
					RoleRef: rbacv1.RoleRef{
						Name: "test-role",
						Kind: "ClusterRole",
					},
					Subjects: []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "default"}},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "admin-role"},
					RoleRef: rbacv1.RoleRef{
						Name: "admin-role",
						Kind: "ClusterRole",
					},
					Subjects: []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "Group", Name: "admins"}},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "john-role"},
					RoleRef: rbacv1.RoleRef{
						Name: "john-role",
						Kind: "ClusterRole",
					},
					Subjects: []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: "john"}},
				},
			}

			clusterRoles := rbacdefinition.BuildAllClusterRoleBindings(cr)
			Expect(clusterRoles).To(BeEquivalentTo(expectedBindings))
			close(done)
		})
	})

	Describe("BuildAllRoleBindings", func() {
		It("should return empty array - no specs", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Namespaced: []accessmanagerv1beta1.NamespacedSpec{},
				},
			}

			roles, err := rbacdefinition.BuildAllRoleBindings(cr, def)
			Expect(err).NotTo(HaveOccurred())
			Expect(roles).To(BeEmpty())
			close(done)
		})

		It("should return empty array - no namespaces", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Namespaced: []accessmanagerv1beta1.NamespacedSpec{
						{
							NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"not": "existent"}},
						},
					},
				},
			}

			roles, err := rbacdefinition.BuildAllRoleBindings(cr, def)
			Expect(err).NotTo(HaveOccurred())
			Expect(roles).To(BeEmpty())
			close(done)
		})

		It("should return correct RoleBindings - one namespace", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Namespaced: []accessmanagerv1beta1.NamespacedSpec{
						{
							NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"team": fmt.Sprintf("one-%v", count)}},
							Bindings: []accessmanagerv1beta1.BindingsSpec{
								{
									Kind:     "ClusterRole",
									RoleName: "admin-role",
									Subjects: []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: "john"}},
								},
								{
									Kind:     "Role",
									RoleName: "test-role",
									Subjects: []rbacv1.Subject{
										{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "default"},
										{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "manager"},
									},
								},
							},
						},
					},
				},
			}

			expectedBindings := []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "admin-role", Namespace: namespace1.Name},
					RoleRef: rbacv1.RoleRef{
						Name: "admin-role",
						Kind: "ClusterRole",
					},
					Subjects: []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: "john"}},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "test-role", Namespace: namespace1.Name},
					RoleRef: rbacv1.RoleRef{
						Name: "test-role",
						Kind: "Role",
					},
					Subjects: []rbacv1.Subject{
						{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "default"},
						{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "manager"},
					},
				},
			}

			clusterRoles, err := rbacdefinition.BuildAllRoleBindings(cr, def)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterRoles).To(BeEquivalentTo(expectedBindings))
			close(done)
		})

		It("should return correct RoleBindings - multiple namespace", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Namespaced: []accessmanagerv1beta1.NamespacedSpec{
						{
							NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"ci": fmt.Sprintf("true-%v", count)}},
							Bindings: []accessmanagerv1beta1.BindingsSpec{
								{
									Kind:     "ClusterRole",
									RoleName: "reader-role",
									Subjects: []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: "john"}},
								},
								{
									Kind:     "Role",
									RoleName: "ci-role",
									Subjects: []rbacv1.Subject{
										{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "default"},
										{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "ci"},
									},
								},
							},
						},
					},
				},
			}

			expectedBindings := []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "reader-role", Namespace: namespace3.Name},
					RoleRef: rbacv1.RoleRef{
						Name: "reader-role",
						Kind: "ClusterRole",
					},
					Subjects: []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: "john"}},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "ci-role", Namespace: namespace3.Name},
					RoleRef: rbacv1.RoleRef{
						Name: "ci-role",
						Kind: "Role",
					},
					Subjects: []rbacv1.Subject{
						{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "default"},
						{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "ci"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "reader-role", Namespace: namespace2.Name},
					RoleRef: rbacv1.RoleRef{
						Name: "reader-role",
						Kind: "ClusterRole",
					},
					Subjects: []rbacv1.Subject{{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: "john"}},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "ci-role", Namespace: namespace2.Name},
					RoleRef: rbacv1.RoleRef{
						Name: "ci-role",
						Kind: "Role",
					},
					Subjects: []rbacv1.Subject{
						{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "default"},
						{APIGroup: "rbac.authorization.k8s.io", Kind: "ServiceAccount", Name: "ci"},
					},
				},
			}

			clusterRoles, err := rbacdefinition.BuildAllRoleBindings(cr, def)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterRoles).To(BeEquivalentTo(expectedBindings))
			close(done)
		})
	})

	Describe("CreateOrRecreateClusterRoleBinding", func() {
		It("should create a new ClusterRoleBinding", func(done Done) {
			crb := rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-crb-%v", count)},
				RoleRef: rbacv1.RoleRef{
					Name: "test-role",
					Kind: "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
			}

			_, err := rbacdefinition.CreateOrRecreateClusterRoleBinding(crb, def)
			Expect(err).NotTo(HaveOccurred())

			_, err = clientset.RbacV1().ClusterRoleBindings().Get(ctx, crb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			close(done)
		})

		It("should recreate a existing ClusterRoleBinding", func(done Done) {
			crb := rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("existing-crb-%v", count)},
				RoleRef: rbacv1.RoleRef{
					Name: "new-role",
					Kind: "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "ci", Namespace: "default"}},
			}

			_, err := rbacdefinition.CreateOrRecreateClusterRoleBinding(crb, def)
			Expect(err).NotTo(HaveOccurred())

			updated, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, crb.Name, metav1.GetOptions{})
			Expect(updated.RoleRef.Name == "new-role").To(BeTrue())
			Expect(updated.Subjects[0].Name == "ci").To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
			close(done)
		})
	})

	Describe("CreateOrRecreateRoleBinding", func() {
		It("should create a new RoleBinding", func(done Done) {
			rb := rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-rb-%v", count), Namespace: "default"},
				RoleRef: rbacv1.RoleRef{
					Name: "test-role",
					Kind: "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
			}

			_, err := rbacdefinition.CreateOrRecreateRoleBinding(rb, def)
			Expect(err).NotTo(HaveOccurred())

			_, err = clientset.RbacV1().RoleBindings("default").Get(ctx, rb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			close(done)
		})

		It("should recreate a existing RoleBinding", func(done Done) {
			rb := rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("existing-rb-%v", count), Namespace: "default"},
				RoleRef: rbacv1.RoleRef{
					Name: "new-role",
					Kind: "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "ci", Namespace: "default"}},
			}

			_, err := rbacdefinition.CreateOrRecreateRoleBinding(rb, def)
			Expect(err).NotTo(HaveOccurred())

			updated, err := clientset.RbacV1().RoleBindings("default").Get(ctx, rb.Name, metav1.GetOptions{})
			Expect(updated.RoleRef.Name == "new-role").To(BeTrue())
			Expect(updated.Subjects[0].Name == "ci").To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
			close(done)
		})
	})
})
