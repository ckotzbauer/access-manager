package rbacdefinition_test

import (
	"context"
	"fmt"
	"sync/atomic"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteNamespace(ctx context.Context, ns *corev1.Namespace) {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, ns.Name, metav1.GetOptions{})
	if err == nil {
		err = clientset.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
	}
}

var _ = Describe("Reconciliation", func() {
	var namespace *corev1.Namespace
	var count uint64 = 0
	ctx := context.TODO()

	BeforeEach(func(done Done) {
		atomic.AddUint64(&count, 1)
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("ns-%v", count), Labels: map[string]string{"foo": "bar"}},
			Spec:       corev1.NamespaceSpec{},
		}

		close(done)
	})

	AfterEach(func(done Done) {
		deleteNamespace(ctx, namespace)
		close(done)
	})

	Describe("Test", func() {
		It("should work", func(done Done) {
			fmt.Println("It2")
			close(done)
		})
	})
})
