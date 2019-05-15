package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"strings"
)

type metaDataV3 struct {
	Cluster string `json:"Cluster"`
	TaskARN string `json:"TaskARN"`
	Family  string `json:"Family"`
}

// GetPrefix ...
func GetPrefix(readCloser io.ReadCloser) string {
	defer readCloser.Close()

	m := &metaDataV3{}

	var prefix string

	if err := json.NewDecoder(readCloser).Decode(m); err != nil {
		log.Fatalln("json:", err)
	}

	_, clusterName := path.Split(m.Cluster)
	_, taskID := path.Split(m.TaskARN)
	safeClusterName := strings.Replace(clusterName, "-", "_", -1)
	safeTaskID := strings.Replace(taskID, "-", "_", -1)
	safeFamilyName := strings.Replace(m.Family, "-", "_", -1)

	prefix = fmt.Sprintf("%s_%s_%s", safeClusterName, safeFamilyName, safeTaskID)
	fmt.Println("prefix: ", prefix)
	return prefix
}
