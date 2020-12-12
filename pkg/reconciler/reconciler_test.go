package reconciler_test

import (
	accessmanagerv1beta1 "access-manager/apis/access-manager.io/v1beta1"
	"access-manager/pkg/reconciler"
	"access-manager/pkg/util"
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestReconciliation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reconciliation Suite")
}

var testenv *envtest.Environment
var clientset *kubernetes.Clientset

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter)))

	testenv = &envtest.Environment{}

	var err error
	cfg, err := testenv.Start()
	Expect(err).NotTo(HaveOccurred())

	clientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(clientset).NotTo(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	Expect(testenv.Stop()).To(Succeed())
})

func createNamespaces(ctx context.Context, nss ...*corev1.Namespace) {
	for _, ns := range nss {
		_, err := clientset.CoreV1().Namespaces().Get(ctx, ns.Name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			ns, err = clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

func createServiceAccounts(ctx context.Context, accounts ...*corev1.ServiceAccount) {
	for _, account := range accounts {
		_, err := clientset.CoreV1().ServiceAccounts(account.Namespace).Get(ctx, account.Name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			account, err = clientset.CoreV1().ServiceAccounts(account.Namespace).Create(ctx, account, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

func createClusterRoleBindings(ctx context.Context, crbs ...*rbacv1.ClusterRoleBinding) {
	for _, crb := range crbs {
		_, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, crb.Name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			crb, err = clientset.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

func createRoleBindings(ctx context.Context, rbs ...*rbacv1.RoleBinding) {
	for _, rb := range rbs {
		_, err := clientset.RbacV1().RoleBindings("default").Get(ctx, rb.Name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			rb, err = clientset.RbacV1().RoleBindings("default").Create(ctx, rb, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

var _ = Describe("Reconciler", func() {
	var namespace1 *corev1.Namespace
	var namespace2 *corev1.Namespace
	var namespace3 *corev1.Namespace
	var namespace4 *corev1.Namespace
	var roleBinding1 *rbacv1.RoleBinding
	var count uint64 = 0
	var scheme *runtime.Scheme
	var logger logr.Logger
	var rec *reconciler.Reconciler
	ctx := context.TODO()
	flag := true

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
		namespace4 = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   fmt.Sprintf("ns-four-%v", count),
				Labels: map[string]string{"team": fmt.Sprintf("four-%v", count)},
			},
			Spec: corev1.NamespaceSpec{},
		}
		clusterRoleBinding1 := rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:            fmt.Sprintf("existing-crb1-%v", count),
				OwnerReferences: []metav1.OwnerReference{{Kind: "RbacDefinition", APIVersion: "access-manager.io/v1beta1", Controller: &flag, Name: "xx", UID: "123456"}},
			},
			RoleRef: rbacv1.RoleRef{
				Name: "test-role",
				Kind: "ClusterRole",
			},
			Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
		}
		clusterRoleBinding2 := rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("existing-crb2-%v", count),
			},
			RoleRef: rbacv1.RoleRef{
				Name: "test-role",
				Kind: "ClusterRole",
			},
			Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
		}
		roleBinding1 = &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:            fmt.Sprintf("existing-rb1-%v", count),
				OwnerReferences: []metav1.OwnerReference{{Kind: "RbacDefinition", APIVersion: "access-manager.io/v1beta1", Controller: &flag, Name: "xx", UID: "123456"}},
			},
			RoleRef: rbacv1.RoleRef{
				Name: "test-role",
				Kind: "ClusterRole",
			},
			Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
		}
		roleBinding2 := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("existing-rb2-%v", count),
			},
			RoleRef: rbacv1.RoleRef{
				Name: "test-role",
				Kind: "ClusterRole",
			},
			Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
		}
		serviceAccount1 := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("one-%v", count),
				Namespace: fmt.Sprintf("ns-four-%v", count),
			},
		}
		serviceAccount2 := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("two-%v", count),
				Namespace: fmt.Sprintf("ns-four-%v", count),
			},
		}

		scheme = kscheme.Scheme
		logger = log.Log.WithName("testLogger")
		rec = &reconciler.Reconciler{Client: *clientset, Scheme: scheme, Logger: logger}
		createNamespaces(ctx, namespace1, namespace2, namespace3, namespace4)
		createClusterRoleBindings(ctx, &clusterRoleBinding1, &clusterRoleBinding2)
		createRoleBindings(ctx, roleBinding1, roleBinding2)
		createServiceAccounts(ctx, serviceAccount1, serviceAccount2)
		close(done)
	})

	AfterEach(func(done Done) {
		close(done)
	})

	Describe("GetRelevantNamespaces", func() {
		It("should not match any namespace", func(done Done) {
			spec := &accessmanagerv1beta1.NamespacedSpec{NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"no": "match"}}}

			found := rec.GetRelevantNamespaces(*spec)
			Expect(found).NotTo(BeNil())
			Expect(found).To(BeEmpty())
			close(done)
		})

		It("should match namespace1", func(done Done) {
			spec := &accessmanagerv1beta1.NamespacedSpec{
				NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"team": fmt.Sprintf("one-%v", count)}},
			}

			found := rec.GetRelevantNamespaces(*spec)
			Expect(found).NotTo(BeNil())
			Expect(util.MapNamespaces(found)).To(BeEquivalentTo([]string{namespace1.Name}))
			close(done)
		})

		It("should match namespace2 and namespace3", func(done Done) {
			spec := &accessmanagerv1beta1.NamespacedSpec{
				NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"ci": fmt.Sprintf("true-%v", count)}},
			}

			found := rec.GetRelevantNamespaces(*spec)
			Expect(found).NotTo(BeNil())
			Expect(util.MapNamespaces(found)).To(BeEquivalentTo([]string{namespace3.Name, namespace2.Name}))
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

			clusterRoles := rec.BuildAllClusterRoleBindings(cr)
			Expect(clusterRoles).To(BeEmpty())
			close(done)
		})

		It("should return nothing if no subjects are provided", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Cluster: []accessmanagerv1beta1.ClusterSpec{
						{
							ClusterRoleName: "test-role",
							Subjects:        []rbacv1.Subject{},
						},
					},
				},
			}

			clusterRoles := rec.BuildAllClusterRoleBindings(cr)
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
							Name:            "my-awesome-clusterrolebinding",
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
					ObjectMeta: metav1.ObjectMeta{Name: "my-awesome-clusterrolebinding"},
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

			clusterRoles := rec.BuildAllClusterRoleBindings(cr)
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

			roles := rec.BuildAllRoleBindings(cr)
			Expect(roles).NotTo(BeNil())
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

			roles := rec.BuildAllRoleBindings(cr)
			Expect(roles).NotTo(BeNil())
			Expect(roles).To(BeEmpty())
			close(done)
		})

		It("should return empty array - no subjects", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Namespaced: []accessmanagerv1beta1.NamespacedSpec{
						{
							NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"team": fmt.Sprintf("one-%v", count)}},
							Bindings: []accessmanagerv1beta1.BindingsSpec{
								{
									Kind:     "ClusterRole",
									RoleName: "admin-role",
									Subjects: []rbacv1.Subject{},
								},
							},
						},
					},
				},
			}

			roles := rec.BuildAllRoleBindings(cr)
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
									Name:     "my-awesome-rolebinding",
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
					ObjectMeta: metav1.ObjectMeta{Name: "my-awesome-rolebinding", Namespace: namespace1.Name},
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

			clusterRoles := rec.BuildAllRoleBindings(cr)
			Expect(clusterRoles).NotTo(BeNil())
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

			clusterRoles := rec.BuildAllRoleBindings(cr)
			Expect(clusterRoles).NotTo(BeNil())
			Expect(clusterRoles).To(BeEquivalentTo(expectedBindings))
			close(done)
		})

		It("should return correct RoleBindings - allServiceAccounts 1", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Namespaced: []accessmanagerv1beta1.NamespacedSpec{
						{
							NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"team": fmt.Sprintf("four-%v", count)}},
							Bindings: []accessmanagerv1beta1.BindingsSpec{
								{
									Kind:               "Role",
									Name:               "my-awesome-rolebinding",
									RoleName:           "test-role",
									AllServiceAccounts: true,
								},
							},
						},
					},
				},
			}

			expectedBindings := []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "my-awesome-rolebinding", Namespace: namespace4.Name},
					RoleRef: rbacv1.RoleRef{
						Name: "test-role",
						Kind: "Role",
					},
					Subjects: []rbacv1.Subject{
						{APIGroup: "", Kind: "ServiceAccount", Name: fmt.Sprintf("one-%v", count)},
						{APIGroup: "", Kind: "ServiceAccount", Name: fmt.Sprintf("two-%v", count)},
					},
				},
			}

			clusterRoles := rec.BuildAllRoleBindings(cr)
			Expect(clusterRoles).NotTo(BeNil())
			Expect(clusterRoles).To(BeEquivalentTo(expectedBindings))
			close(done)
		})

		It("should return correct RoleBindings - allServiceAccounts 2", func(done Done) {
			cr := &accessmanagerv1beta1.RbacDefinition{
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Namespaced: []accessmanagerv1beta1.NamespacedSpec{
						{
							NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"team": fmt.Sprintf("four-%v", count)}},
							Bindings: []accessmanagerv1beta1.BindingsSpec{
								{
									Kind:               "Role",
									Name:               "my-awesome-rolebinding",
									RoleName:           "test-role",
									AllServiceAccounts: true,
									Subjects: []rbacv1.Subject{
										{APIGroup: "", Kind: "ServiceAccount", Name: fmt.Sprintf("one-%v", count)},
										{APIGroup: "", Kind: "ServiceAccount", Name: "myacc"},
									},
								},
							},
						},
					},
				},
			}

			expectedBindings := []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "my-awesome-rolebinding", Namespace: namespace4.Name},
					RoleRef: rbacv1.RoleRef{
						Name: "test-role",
						Kind: "Role",
					},
					Subjects: []rbacv1.Subject{
						{APIGroup: "", Kind: "ServiceAccount", Name: fmt.Sprintf("one-%v", count)},
						{APIGroup: "", Kind: "ServiceAccount", Name: "myacc"},
						{APIGroup: "", Kind: "ServiceAccount", Name: fmt.Sprintf("two-%v", count)},
					},
				},
			}

			clusterRoles := rec.BuildAllRoleBindings(cr)
			Expect(clusterRoles).NotTo(BeNil())
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

			_, err := rec.CreateOrRecreateClusterRoleBinding(crb)
			Expect(err).NotTo(HaveOccurred())

			_, err = clientset.RbacV1().ClusterRoleBindings().Get(ctx, crb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			close(done)
		})

		It("should recreate a existing ClusterRoleBinding", func(done Done) {
			crb := rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("existing-crb1-%v", count)},
				RoleRef: rbacv1.RoleRef{
					Name: "new-role",
					Kind: "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "ci", Namespace: "default"}},
			}

			_, err := rec.CreateOrRecreateClusterRoleBinding(crb)
			Expect(err).NotTo(HaveOccurred())

			updated, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, crb.Name, metav1.GetOptions{})
			Expect(updated.RoleRef.Name == "new-role").To(BeTrue())
			Expect(updated.Subjects[0].Name == "ci").To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
			close(done)
		})

		It("should not touch a unchanged ClusterRoleBinding", func(done Done) {
			original, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, fmt.Sprintf("existing-crb2-%v", count), metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			crb := rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("existing-crb2-%v", count),
					Namespace: "default",
				},
				RoleRef: rbacv1.RoleRef{
					Name: "test-role",
					Kind: "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
			}

			_, err = rec.CreateOrRecreateClusterRoleBinding(crb)
			Expect(err).NotTo(HaveOccurred())

			unchanged, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, crb.Name, metav1.GetOptions{})
			Expect(unchanged.UID).To(BeEquivalentTo(original.UID))
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

			_, err := rec.CreateOrRecreateRoleBinding(rb)
			Expect(err).NotTo(HaveOccurred())

			_, err = clientset.RbacV1().RoleBindings("default").Get(ctx, rb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			close(done)
		})

		It("should recreate a existing RoleBinding", func(done Done) {
			original, err := clientset.RbacV1().RoleBindings("default").Get(ctx, fmt.Sprintf("existing-rb1-%v", count), metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			rb := rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("existing-rb1-%v", count), Namespace: "default"},
				RoleRef: rbacv1.RoleRef{
					Name: "new-role",
					Kind: "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "ci", Namespace: "default"}},
			}

			_, err = rec.CreateOrRecreateRoleBinding(rb)
			Expect(err).NotTo(HaveOccurred())

			updated, err := clientset.RbacV1().RoleBindings("default").Get(ctx, rb.Name, metav1.GetOptions{})
			Expect(updated.UID).ToNot(BeEquivalentTo(original.UID))
			Expect(updated.RoleRef.Name == "new-role").To(BeTrue())
			Expect(updated.Subjects[0].Name == "ci").To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
			close(done)
		})

		It("should not touch a unchanged RoleBinding", func(done Done) {
			original, err := clientset.RbacV1().RoleBindings("default").Get(ctx, fmt.Sprintf("existing-rb2-%v", count), metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			rb := rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("existing-rb2-%v", count),
					Namespace: "default",
				},
				RoleRef: rbacv1.RoleRef{
					Name: "test-role",
					Kind: "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
			}

			_, err = rec.CreateOrRecreateRoleBinding(rb)
			Expect(err).NotTo(HaveOccurred())

			unchanged, err := clientset.RbacV1().RoleBindings("default").Get(ctx, rb.Name, metav1.GetOptions{})
			Expect(unchanged.UID).To(BeEquivalentTo(original.UID))
			Expect(err).NotTo(HaveOccurred())
			close(done)
		})
	})

	Describe("DeleteOwnedRoleBindings", func() {
		It("should create a new RoleBinding", func(done Done) {
			flag := true

			ownedRb := rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("owned-rb-%v", count),
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "access-manager.io/v1beta1",
							Controller: &flag,
							Kind:       "RbacDefinition",
							Name:       "test-def",
							UID:        "123456",
						},
					},
				},
				RoleRef: rbacv1.RoleRef{
					Name: "test-role",
					Kind: "ClusterRole",
				},
				Subjects: []rbacv1.Subject{{APIGroup: "", Kind: "ServiceAccount", Name: "default", Namespace: "default"}},
			}

			def := &accessmanagerv1beta1.RbacDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "test-def"},
				Spec: accessmanagerv1beta1.RbacDefinitionSpec{
					Namespaced: []accessmanagerv1beta1.NamespacedSpec{
						{
							Namespace: accessmanagerv1beta1.NamespaceSpec{Name: "default"},
						},
					},
				},
			}

			createRoleBindings(ctx, &ownedRb)

			err := rec.DeleteOwnedRoleBindings("default", *def)
			Expect(err).NotTo(HaveOccurred())

			_, err = clientset.RbacV1().RoleBindings("default").Get(ctx, ownedRb.Name, metav1.GetOptions{})
			Expect(errors.IsNotFound(err)).To(BeTrue())

			ex, err := clientset.RbacV1().RoleBindings("default").Get(ctx, roleBinding1.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(ex).NotTo(BeNil())

			close(done)
		})
	})
})
