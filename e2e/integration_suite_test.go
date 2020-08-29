package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var clientset *kubernetes.Clientset
var client dynamic.Interface

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	kubeconfig := os.Getenv("KUBECONFIG")
	cwd, _ := os.Getwd()
	cfg, err := clientcmd.BuildConfigFromFlags("", filepath.Join(cwd, kubeconfig))
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	client, err = dynamic.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(client).NotTo(BeNil())

	clientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(clientset).NotTo(BeNil())

	close(done)
}, 60)
