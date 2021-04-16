package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type (
	pod struct {
		namespace string
		name      string
		addr      string
		identity  string
	}

	deployment struct {
		name   string
		pods   []pod
		fanIn  int
		fanOut int
	}
)

var responseTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{Name: "response_total", Help: "Response total"},
	[]string{
		"direction", "authority", "target_addr", "classification", "tls", "namespace", "pod", "workload_name", "workload_kind", "client_id", "status_code",
		"server_id", "dst_control_plane_ns", "dst_deployment", "dst_namespace", "dst_pod", "dst_pod_template_hash", "dst_service", "dst_serviceaccount", "grpc_status", "dst_workload_kind", "dst_workload_name",
	},
)

var responseLatency = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "response_latency_ms",
		Help: "Histogram of response latencies",
		Buckets: []float64{
			1, 2, 3, 4, 5,
			10, 20, 30, 40, 50,
			100, 200, 300, 400, 500,
			1000, 2000, 3000, 4000, 5000,
			10000, 20000, 30000, 40000, 50000,
		},
	},
	[]string{
		"direction", "authority", "target_addr", "tls", "namespace", "pod", "workload_name", "workload_kind", "client_id", "status_code",
		"server_id", "dst_control_plane_ns", "dst_deployment", "dst_namespace", "dst_pod", "dst_pod_template_hash", "dst_service", "dst_serviceaccount", "dst_workload_kind", "dst_workload_name",
	},
)

var tcpOpenConnections = promauto.NewGaugeVec(
	prometheus.GaugeOpts{Name: "tcp_open_connections", Help: "TCP open connections"},
	[]string{
		"direction", "authority", "target_addr", "tls", "namespace", "pod", "workload_name", "workload_kind", "client_id", "peer",
		"server_id", "dst_control_plane_ns", "dst_deployment", "dst_namespace", "dst_pod", "dst_pod_template_hash", "dst_service", "dst_serviceaccount", "dst_workload_kind", "dst_workload_name",
	},
)

var tcpReads = promauto.NewCounterVec(
	prometheus.CounterOpts{Name: "tcp_read_bytes_total", Help: "TCP read bytes total"},
	[]string{
		"direction", "authority", "target_addr", "tls", "namespace", "pod", "workload_name", "workload_kind", "client_id", "peer",
		"server_id", "dst_control_plane_ns", "dst_deployment", "dst_namespace", "dst_pod", "dst_pod_template_hash", "dst_service", "dst_serviceaccount", "dst_workload_kind", "dst_workload_name",
	},
)

var tcpWrites = promauto.NewCounterVec(
	prometheus.CounterOpts{Name: "tcp_writes_bytes_total", Help: "TCP writes bytes total"},
	[]string{
		"direction", "authority", "target_addr", "tls", "namespace", "pod", "workload_name", "workload_kind", "client_id", "peer",
		"server_id", "dst_control_plane_ns", "dst_deployment", "dst_namespace", "dst_pod", "dst_pod_template_hash", "dst_service", "dst_serviceaccount", "dst_workload_kind", "dst_workload_name",
	},
)

func start(deployments []deployment) {
	ticker := time.Tick(1 * time.Second)
	for {
		incMetrics(deployments)
		<-ticker
	}
}

func incMetrics(deployments []deployment) {
	for _, d := range deployments {
		for _, p := range d.pods {
			// inbound
			inboundResponse := responseTotal.MustCurryWith(withClassification(inboundLabels(p, d)))
			inboundLatency := responseLatency.MustCurryWith(inboundLabels(p, d))

			for _, clientID := range clientIDs(d.fanIn) {
				for _, statusCode := range []string{"200", "500", "404"} {
					labels := prometheus.Labels{
						"client_id":   clientID,
						"status_code": statusCode,
					}
					inboundResponse.With(labels).Inc()
					inboundLatency.With(labels).Observe(rand.Float64() * 50000)
				}
			}

			// outbound
			for i := 0; i < d.fanOut; i++ {
				target := fmt.Sprintf("dst-%d", i)
				outboundResponse := responseTotal.MustCurryWith(withClassification(outboundLabels(p, d, target)))
				outboundLatency := responseLatency.MustCurryWith(outboundLabels(p, d, target))

				for _, statusCode := range []string{"200", "500", "404"} {
					labels := prometheus.Labels{
						"status_code": statusCode,
					}
					outboundResponse.With(labels).Inc()
					outboundLatency.With(labels).Observe(rand.Float64() * 50000)
				}
			}

			setTCPMetricsGauge(p, d, tcpOpenConnections)
			incTCPMetricsCounter(p, d, tcpReads)
			incTCPMetricsCounter(p, d, tcpWrites)
		}
	}
}

func setTCPMetricsGauge(p pod, d deployment, metric *prometheus.GaugeVec) {
	inbound := metric.MustCurryWith(inboundLabels(p, d))

	for _, clientID := range clientIDs(d.fanIn) {
		for _, peer := range []string{"src", "dst"} {
			inbound.With(
				prometheus.Labels{
					"client_id": clientID,
					"peer":      peer,
				},
			).Set(rand.Float64() * 1000)
		}
	}

	for i := 0; i < d.fanOut; i++ {
		target := fmt.Sprintf("dst-%d", i)
		outbound := metric.MustCurryWith(outboundLabels(p, d, target))

		for _, peer := range []string{"src", "dst"} {
			outbound.With(
				prometheus.Labels{
					"peer": peer,
				},
			).Set(rand.Float64() * 1000)
		}
	}
}

func incTCPMetricsCounter(p pod, d deployment, metric *prometheus.CounterVec) {
	inbound := metric.MustCurryWith(inboundLabels(p, d))

	for _, clientID := range clientIDs(d.fanIn) {
		for _, peer := range []string{"src", "dst"} {
			inbound.With(
				prometheus.Labels{
					"client_id": clientID,
					"peer":      peer,
				},
			).Inc()
		}
	}

	for i := 0; i < d.fanOut; i++ {
		target := fmt.Sprintf("dst-%d", i)
		outbound := metric.MustCurryWith(outboundLabels(p, d, target))

		for _, peer := range []string{"src", "dst"} {
			outbound.With(
				prometheus.Labels{
					"peer": peer,
				},
			).Inc()
		}
	}
}

func makePods(n int, namespace string) []pod {
	pods := make([]pod, n)
	for i := 0; i < n; i++ {
		pods[i] = pod{
			namespace: namespace,
			name:      fmt.Sprintf("pod-%d", i),
			addr:      fmt.Sprintf("pod-%d:8080", i),
			identity:  fmt.Sprintf("pod-%d.%s.serviceacount.identity.linkerd.cluster.local", i, namespace),
		}
	}
	return pods
}

func makeDeployments(config DeploymentConfig) []deployment {
	deployments := make([]deployment, config.Quantity)
	for i := 0; i < config.Quantity; i++ {
		deployments[i] = deployment{
			name:   fmt.Sprintf("deployment-%d", i),
			pods:   makePods(config.Pods, config.Namespace),
			fanIn:  config.FanIn,
			fanOut: config.FanOut,
		}
	}
	return deployments
}

func clientIDs(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("client-%d.namespace.serviceaccount.identity.linkerd.cluster.local", i)
	}
	return ids
}

func inboundLabels(p pod, d deployment) prometheus.Labels {
	return prometheus.Labels{
		"direction":     "inbound",
		"authority":     p.name,
		"target_addr":   p.addr,
		"tls":           "true",
		"namespace":     p.namespace,
		"pod":           p.name,
		"workload_name": d.name,
		"workload_kind": "Deployment",

		// outbound
		"server_id":             "",
		"dst_control_plane_ns":  "",
		"dst_deployment":        "",
		"dst_namespace":         "",
		"dst_pod":               "",
		"dst_pod_template_hash": "",
		"dst_service":           "",
		"dst_serviceaccount":    "",
		"dst_workload_kind":     "",
		"dst_workload_name":     "",
	}
}

func outboundLabels(p pod, d deployment, target string) prometheus.Labels {
	return prometheus.Labels{
		"direction":             "outbound",
		"authority":             target,
		"target_addr":           target,
		"server_id":             target,
		"dst_control_plane_ns":  "linkerd",
		"dst_deployment":        target,
		"dst_namespace":         "default",
		"dst_pod":               target,
		"dst_pod_template_hash": target,
		"dst_service":           target,
		"dst_serviceaccount":    target,
		"tls":                   "true",
		"namespace":             p.namespace,
		"pod":                   p.name,
		"workload_name":         d.name,
		"workload_kind":         "Deployment",
		"dst_workload_kind":     "Deployment",
		"dst_workload_name":     target,

		// inbound
		"client_id": "",
	}
}

// withClassification mutates the input
func withClassification(labels prometheus.Labels) prometheus.Labels {
	labels["classification"] = "failure"
	labels["grpc_status"] = "0"
	return labels
}
