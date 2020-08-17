package metrics

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("metrics")

var (
	version  prometheus.Gauge
	gaugeMap = make(map[string]map[string]prometheus.Gauge)

	rbacTotalCreated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "accessmanager_rbac_total_created",
		Help: "Count of created RoleBindings and ClusterRoleBindings",
	}, []string{"rbacdefinition", "kind", "namespace"})
)

// InitMetrics initializes prometheus metrics and the endpoint
func InitMetrics(v string, host string, port int32) {
	r := prometheus.NewRegistry()
	r.MustRegister(rbacTotalCreated)

	version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "accessmanager_version",
		Help: "Version information about the access-manager",
		ConstLabels: map[string]string{
			"version":    v,
			"go_version": runtime.Version(),
		},
	})
	version.Set(1)
	r.MustRegister(version)

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	go func() {
		err := http.ListenAndServe(fmt.Sprint(host, ":", port), nil)
		if err != nil {
			log.Error(err, "Could not serve metrics.")
		}
	}()

	log.Info(fmt.Sprint("Serve /metrics on port ", port))
}

func AddGauge(ns string, kind string, def string, inc float64) {
	defMap := gaugeMap[def]

	if defMap == nil {
		defMap = map[string]prometheus.Gauge{}
		gaugeMap[def] = defMap
	}

	sLabels := fmt.Sprint(map[string]string{"kind": kind, "namespace": ns})

	if gaugeMap[def][sLabels] == nil {
		gaugeMap[def][sLabels] = rbacTotalCreated.WithLabelValues(def, kind, ns)
	}

	gaugeMap[def][sLabels].Add(inc)
}

func ResetRbacDefinition(def string) {
	defMap := gaugeMap[def]

	if defMap != nil {
		for _, gauge := range defMap {
			gauge.Set(0)
		}
	}
}
