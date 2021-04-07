package main

import (
	"bufio"
	"fmt"
	"io"
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

	label struct {
		name   string
		values []string
	}
)

func writeMetrics(out io.Writer, deployments []deployment, value int) int {

	w := bufio.NewWriter(out)
	total := 0

	for _, d := range deployments {
		for _, p := range d.pods {
			// inbound response_total
			boundLabels := map[string]string{
				"direction":      "inbound",
				"authority":      p.name,
				"target_addr":    p.addr,
				"classification": "failure",
				"tls":            "true",
				"namespace":      p.namespace,
				"pod":            p.name,
				"workload_name":  d.name,
				"workload_kind":  "Deployment",
			}
			unboundLabels := []label{
				{
					name:   "client_id",
					values: clientIDs(d.fanIn),
				},
				{
					name:   "status_code",
					values: []string{"200", "500", "404"},
				},
			}
			total += renderMetric(w, "response_total", boundLabels, unboundLabels, value)

			// outbound response_total
			for i := 0; i < d.fanOut; i++ {
				target := fmt.Sprintf("dst-%d", i)
				boundLabels = map[string]string{
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
					"classification":        "failure",
					"tls":                   "true",
					"grpc_status":           "0",
					"namespace":             p.namespace,
					"pod":                   p.name,
					"workload_name":         d.name,
					"workload_kind":         "Deployment",
					"dst_workload_kind":     "Deployment",
					"dst_workload_name":     target,
				}
				unboundLabels = []label{
					{
						name:   "status_code",
						values: []string{"200", "500", "404"},
					},
				}
				total += renderMetric(w, "response_total", boundLabels, unboundLabels, value)
			}

			// inbound response_latency_ms_bucket
			boundLabels = map[string]string{
				"direction":     "inbound",
				"authority":     p.name,
				"target_addr":   p.addr,
				"tls":           "true",
				"namespace":     p.namespace,
				"pod":           p.name,
				"workload_name": d.name,
				"workload_kind": "Deployment",
			}
			unboundLabels = []label{
				{
					name:   "client_id",
					values: clientIDs(d.fanIn),
				},
				{
					name:   "status_code",
					values: []string{"200", "500", "404"},
				},
				{
					name:   "le",
					values: []string{"1", "2", "3", "4", "5", "10", "20", "30", "40", "50", "100", "200", "300", "400", "500", "1000", "2000", "3000", "4000", "5000", "10000", "20000", "30000", "40000", "50000", "+Inf"},
				},
			}
			total += renderMetric(w, "response_latency_ms_bucket", boundLabels, unboundLabels, value)

			// outbound response_latency_ms_bucket
			for i := 0; i < d.fanOut; i++ {
				target := fmt.Sprintf("dst-%d", i)
				boundLabels = map[string]string{
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
				}
				unboundLabels = []label{
					{
						name:   "status_code",
						values: []string{"200", "500", "404"},
					},
					{
						name:   "le",
						values: []string{"1", "2", "3", "4", "5", "10", "20", "30", "40", "50", "100", "200", "300", "400", "500", "1000", "2000", "3000", "4000", "5000", "10000", "20000", "30000", "40000", "50000", "+Inf"},
					},
				}
				total += renderMetric(w, "response_latency_ms_bucket", boundLabels, unboundLabels, value)
			}

			total += renderTcpMetric(w, p, d, "tcp_open_connections", value)
			total += renderTcpMetric(w, p, d, "tcp_read_bytes_total", value)
			total += renderTcpMetric(w, p, d, "tcp_write_bytes_total", value)

		}
	}
	w.Flush()
	return total
}

func renderTcpMetric(w io.Writer, p pod, d deployment, metric string, value int) int {
	// inbound
	boundLabels := map[string]string{
		"direction":     "inbound",
		"target_addr":   p.addr,
		"tls":           "true",
		"namespace":     p.namespace,
		"pod":           p.name,
		"workload_name": d.name,
		"workload_kind": "Deployment",
	}
	unboundLabels := []label{
		{
			name:   "client_id",
			values: clientIDs(d.fanIn),
		},
		{
			name:   "peer",
			values: []string{"src", "dst"},
		},
	}
	total := renderMetric(w, metric, boundLabels, unboundLabels, value)

	// outbound
	for i := 0; i < d.fanOut; i++ {
		target := fmt.Sprintf("dst-%d", i)
		boundLabels = map[string]string{
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
		}
		unboundLabels = []label{
			{
				name:   "peer",
				values: []string{"src", "dst"},
			},
		}
		total += renderMetric(w, metric, boundLabels, unboundLabels, value)
	}
	return total
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

func renderMetric(w io.Writer, metric string, boundLabels map[string]string, labels []label, value int) int {
	if len(labels) == 0 {
		fmt.Fprint(w, metric)
		fmt.Fprint(w, " {")
		n := len(boundLabels)
		i := 0
		for k, v := range boundLabels {
			fmt.Fprint(w, k)
			fmt.Fprint(w, "=\"")
			fmt.Fprint(w, v)
			fmt.Fprint(w, "\"")
			i++
			if i != n {
				fmt.Fprint(w, ",")
			}
		}
		fmt.Fprint(w, "} ")
		fmt.Fprintf(w, "%d\n", value)
		return 1
	}

	total := 0
	label := labels[0]
	for _, v := range label.values {
		boundLabels[label.name] = v
		total += renderMetric(w, metric, boundLabels, labels[1:], value)
	}
	delete(boundLabels, label.name)
	return total
}

func clientIDs(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("client-%d.namespace.serviceaccount.identity.linkerd.cluster.local", i)
	}
	return ids
}
