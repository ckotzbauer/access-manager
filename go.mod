module access-manager

go 1.14

require (
    github.com/go-logr/logr v0.1.0
    github.com/onsi/ginkgo v1.14.0
    github.com/onsi/gomega v1.10.1
    github.com/operator-framework/operator-sdk v0.18.2
    github.com/spf13/pflag v1.0.5
    golang.org/x/sys v0.0.0-20200610111108-226ff32320da // indirect
    k8s.io/api v0.18.4
    k8s.io/apimachinery v0.18.5
    k8s.io/client-go v12.0.0+incompatible
    sigs.k8s.io/controller-runtime v0.6.0
)

replace (
    github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
    k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
