package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sys/unix"

	"github.com/mschurenko/fargate_exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	volPath = "/"
	dfSleep = 2
	kB      = 1024
	mB      = 1024 * kB
)

func getMetaData() io.ReadCloser {
	var err error
	var resp *http.Response

	uri := os.Getenv("ECS_CONTAINER_METADATA_URI")
	if uri == "" {
		fmt.Println("Are you in AWS?")
		log.Fatalln("metadata: Are you even in AWS?")
	}

	resp, err = http.Get(uri + "/task")
	if err != nil {
		log.Fatalln(err)
	}

	return resp.Body
}

var (
	prefix = utils.GetPrefix(getMetaData())
	dfSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_fargate_overlay_disk_size", prefix),
			Help: "disk size",
		},
	)
	dfFree = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_fargate_overlay_disk_free", prefix),
			Help: "disk free",
		},
	)
	dfAvail = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_fargate_overlay_disk_avail", prefix),
			Help: "disk available",
		},
	)
)

// https://gitlab.cncf.ci/prometheus/node_exporter/blob/master/collector/filesystem_linux.go
type diskStats struct {
	size, free, avail, used int
}

func diskFree(path string) diskStats {
	fs := unix.Statfs_t{}
	err := unix.Statfs(path, &fs)
	if err != nil {
		log.Println(err)
	}

	ds := diskStats{}

	ds.size = int(float64(fs.Blocks) * float64(fs.Bsize) / mB)
	ds.free = int(float64(fs.Bfree) * float64(fs.Bsize) / mB)
	ds.avail = int(float64(fs.Bavail) * float64(fs.Bsize) / mB)
	// ds.used = int((float64(fs.Blocks) - float64(fs.Bavail)) / gB)
	fmt.Printf("size: %v\n", ds.size)
	fmt.Printf("free: %v\n", ds.free)
	fmt.Printf("avail: %v\n", ds.avail)
	// fmt.Printf("used: %v\n", ds.used)
	return ds
}

func df() {
	go func() {
		for {
			ds := diskFree(volPath)
			dfSize.Set(float64(ds.size))
			dfFree.Set(float64(ds.free))
			dfAvail.Set(float64(ds.avail))
			time.Sleep(time.Second * dfSleep)
		}
	}()
}

func dfBin(w http.ResponseWriter, r *http.Request) {
	cmdPath, err := exec.LookPath("df")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(cmdPath, "-B", "M")

	bs, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", string(bs))
}

func main() {
	df()
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/df", dfBin)
	http.ListenAndServe(":2112", nil)
}
