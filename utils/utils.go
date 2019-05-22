package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/docker/docker/api/types"
	"golang.org/x/sys/unix"
)

func getBody(path string) io.ReadCloser {
	if path != "task" && path != "task/stats" {
		log.Fatalln(path, "must be task or task/stats")
	}

	uri := os.Getenv("ECS_CONTAINER_METADATA_URI")
	if uri == "" {
		log.Fatalln("metadata: Are you even in AWS?")
	}

	url := uri + "/" + path
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	return resp.Body
}

// Container represents task metadata containers
type Container struct {
	DockerID   string            `json:"DockerId"`
	Name       string            `json:"Name"`
	DockerName string            `json:"DockerName"`
	Image      string            `json:"Image"`
	Labels     map[string]string `json:"Labels"`
	Type       string            `json:"Type"`
}

type taskMetaDataV3 struct {
	Cluster    string      `json:"Cluster"`
	TaskARN    string      `json:"TaskARN"`
	Revision   string      `json:"Revision"`
	Family     string      `json:"Family"`
	Containers []Container `json:"Containers"`
}

func getTaskMetaDataV3() *taskMetaDataV3 {
	v3 := &taskMetaDataV3{}

	body := getBody("task")

	// http://blog.manugarri.com/how-to-reuse-http-response-body-in-golang/
	bodyBytes, _ := ioutil.ReadAll(body)
	defer body.Close()
	// fmt.Println("body:", string(bodyBytes))

	body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.NewDecoder(body).Decode(v3); err != nil {
		log.Fatalln("json:", err)
	}

	return v3
}

func getDockerStats() *map[string]types.StatsJSON {
	v := &map[string]types.StatsJSON{}

	body := getBody("task/stats")
	defer body.Close()

	if err := json.NewDecoder(body).Decode(v); err != nil {
		log.Fatalln("json:", err)
	}

	return v
}

type taskAttributes struct {
	ClusterName string
	TaskID      string
	Revision    string
	Family      string
	Containers  map[string]string
}

func getTaskAttributes() taskAttributes {
	v3 := getTaskMetaDataV3()
	_, taskID := path.Split(v3.TaskARN)
	_, clusterName := path.Split(v3.Cluster)

	containers := make(map[string]string)

	for _, c := range v3.Containers {
		if c.Type == "CNI_PAUSE" {
			continue
		}
		containers[c.Name] = c.DockerID
	}

	return taskAttributes{
		ClusterName: clusterName,
		TaskID:      taskID,
		Revision:    v3.Revision,
		Family:      v3.Family,
		Containers:  containers,
	}
}

// Task ...
type Task struct {
	ClusterName string
	TaskID      string
	Family      string
}

// ContainerStats ...
// https://github.com/mschurenko/signalfx-agent/blob/master/internal/monitors/docker/conversion.go#L95
type ContainerStats struct {
	Task
	ContainerName string
	ContainerID   string
	TotalCPU      int64
	SystemCPU     int64
	MemoryUsage   int64
	MemoryLimit   int64
}

// GetContainerStats ...
func GetContainerStats() []ContainerStats {
	taskAttr := getTaskAttributes()
	stats := getDockerStats()
	nStats := *stats

	var cs []ContainerStats

	for name, id := range taskAttr.Containers {
		t := Task{
			ClusterName: taskAttr.ClusterName,
			TaskID:      taskAttr.TaskID,
			Family:      taskAttr.Family,
		}
		c := ContainerStats{
			Task:          t,
			ContainerName: name,
			ContainerID:   id,
			TotalCPU:      int64(nStats[id].CPUStats.CPUUsage.TotalUsage),
			SystemCPU:     int64(nStats[id].CPUStats.SystemUsage),
			MemoryUsage:   int64(nStats[id].MemoryStats.Usage),
			// https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/6/html/resource_management_guide/sec-memory
			MemoryLimit: int64(nStats[id].MemoryStats.Stats["hierarchical_memory_limit"]),
		}

		cs = append(cs, c)
	}

	return cs
}

// https://gitlab.cncf.ci/prometheus/node_exporter/blob/master/collector/filesystem_linux.go

// DiskStats ...
type DiskStats struct {
	Task
	Size  int
	Free  int
	Avail int
	Used  int
}

// GetDiskStats ...
func GetDiskStats(path string) DiskStats {
	taskAttr := getTaskAttributes()
	t := Task{
		ClusterName: taskAttr.ClusterName,
		TaskID:      taskAttr.TaskID,
		Family:      taskAttr.Family,
	}
	fs := unix.Statfs_t{}
	err := unix.Statfs(path, &fs)
	if err != nil {
		log.Fatal(err)
	}

	ds := DiskStats{Task: t}
	ds.Size = int(float64(fs.Blocks) * float64(fs.Bsize))
	ds.Free = int(float64(fs.Bfree) * float64(fs.Bsize))
	ds.Avail = int(float64(fs.Bavail) * float64(fs.Bsize))

	return ds
}
