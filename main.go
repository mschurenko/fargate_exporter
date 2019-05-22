package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mschurenko/fargate_exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// set to something lower than scrape interval
	collectSleep = 5
)

var (
	// task metrics
	taskLabels = []string{
		"cluster_name",
		"task_family",
		"task_id",
	}
	dfSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fargate_overlay_disk_size",
			Help: "disk size",
		},
		taskLabels,
	)
	dfFree = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fargate_overlay_disk_free",
			Help: "disk free",
		},
		taskLabels,
	)
	dfAvail = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fargate_overlay_disk_avail",
			Help: "disk available",
		},
		taskLabels,
	)

	// container metrics
	containerLabels = append(taskLabels, []string{"container_name"}...)

	totalCPU = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fargate_cpu_usage_total",
			Help: "Total CPU time consumed",
		},
		containerLabels,
	)
	systemCPU = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fargate_cpu_usage_system",
			Help: "System Usage",
		},
		containerLabels,
	)

	memoryLimit = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fargate_memory_limit",
			Help: "number of times memory usage hits limits",
		},
		containerLabels,
	)
	memoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fargate_memory_usage",
			Help: "current res_counter usage for memory",
		},
		containerLabels,
	)
)

func collectDiskFree() {
	for {
		ds := utils.GetDiskStats("/")

		labels := []string{
			ds.ClusterName,
			ds.Family,
			ds.TaskID,
		}

		dfSize.WithLabelValues(labels...).Set(float64(ds.Size))
		dfFree.WithLabelValues(labels...).Set(float64(ds.Free))
		dfAvail.WithLabelValues(labels...).Set(float64(ds.Avail))

		time.Sleep(time.Second * collectSleep)
	}
}

func collectContainerStats() {
	for {
		cs := utils.GetContainerStats()

		for _, c := range cs {
			labels := []string{
				c.ClusterName,
				c.Family,
				c.TaskID,
				c.ContainerName,
			}
			totalCPU.WithLabelValues(labels...).Set(float64(c.TotalCPU))
			systemCPU.WithLabelValues(labels...).Set(float64(c.SystemCPU))
			memoryUsage.WithLabelValues(labels...).Set(float64(c.MemoryUsage))
			memoryLimit.WithLabelValues(labels...).Set(float64(c.MemoryLimit))
		}

		time.Sleep(time.Second * collectSleep)
	}
}

func main() {
	go collectDiskFree()
	go collectContainerStats()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}
