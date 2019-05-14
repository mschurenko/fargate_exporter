package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sys/unix"

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

var (
	dfSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "fargate_overlay_disk_size",
			Help: "disk size",
		},
	)
	dfFree = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "fargate_overlay_disk_free",
			Help: "disk free",
		},
	)
	dfAvail = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "fargate_overlay_disk_avail",
			Help: "disk available",
		},
	)
)

// https://gitlab.cncf.ci/prometheus/node_exporter/blob/master/collector/filesystem_linux.go
type diskStats struct {
	size  int
	free  int
	avail int
	used  int
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

	execArgs := os.Args[1:]

	log.Println("entrypoint arguments:", strings.Join(execArgs, " "))

	if len(execArgs) == 0 {
		log.Fatal("error: must have args to run")
	}

	cmdPath, err := exec.LookPath(execArgs[0])
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		fmt.Println("going to start cmd")
		cmd := exec.Command(cmdPath, execArgs[1:]...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal("stdout pipe error:", err)
		}

		scanner := bufio.NewScanner(stdout)
		go func() {
			for scanner.Scan() {
				fmt.Printf("%s\n", scanner.Text())
			}
		}()

		if err := cmd.Start(); err != nil {
			log.Fatal("command error:", err)
		}

		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}

	}()

	fmt.Println("got here")

	// start prom endpoint
	df()
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/df", dfBin)
	http.ListenAndServe(":2112", nil)
}
