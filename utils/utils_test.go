package utils

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	rc := m.Run()
	teardown()
	os.Exit(rc)
}

func task(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./testing/task.json")
}

func taskStats(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./testing/task_stats.json")
}

func setup() {
	os.Setenv("ECS_CONTAINER_METADATA_URI", "http://localhost:8080")
	go func() {
		http.HandleFunc("/task", task)
		http.HandleFunc("/task/stats", taskStats)
		http.ListenAndServe(":8080", nil)
	}()
}

func teardown() {
	fmt.Println("tearing down...")
}

func TestGetContainerStats(t *testing.T) {
	task := Task{
		ClusterName: "microservices-testing",
		TaskID:      "20c8a872-0311-4089-b1d4-bef05989f648",
		Family:      "foo",
	}
	expected := []ContainerStats{
		ContainerStats{
			Task:          task,
			ContainerName: "exporter",
			ContainerID:   "d2ddaf2c7c34d606f6f5f5854ea5767187ca69dda25b855a466d98da06feb450",
			TotalCPU:      3593656101,
			SystemCPU:     18397330000000,
			MemoryUsage:   8347648,
			MemoryLimit:   2147483648,
		},
		ContainerStats{
			Task:          task,
			ContainerName: "gotty",
			ContainerID:   "8991fb337b52d33a299a1fa3290e741a7f2b736ce5201d640ca7b21ea3602e1d",
			TotalCPU:      162431257,
			SystemCPU:     18397330000000,
			MemoryUsage:   5603328,
			MemoryLimit:   2147483648,
		},
		ContainerStats{
			Task:          task,
			ContainerName: "foo",
			ContainerID:   "b2cdf7287f0cd14ce8baa7b8900e34af2a02c5b0792cbd9ed8a59fe23fc5755e",
			TotalCPU:      11236383863,
			SystemCPU:     18397330000000, // fixed value per task
			MemoryUsage:   27295744,
			MemoryLimit:   2147483648,
		},
	}

	actual := GetContainerStats()

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Not equal.\n\nActual:\n%v\n\nExpected:\n%v", actual, expected)
	}

}
