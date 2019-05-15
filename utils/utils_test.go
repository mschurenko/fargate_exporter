package utils

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

var metaDataJSON = `
{
"Cluster": "arn:aws:ecs:us-west-2:123456789101:cluster/mycluster",
"TaskARN": "arn:aws:ecs:us-west-2:123456789101:task/1fd2daad-60b5-4272-9aeb-ffe35edaa4bd",
"Family": "thing",
"Revision": "1"
}
`

var expectedPrefix = "mycluster_thing_1fd2daad_60b5_4272_9aeb_ffe35edaa4bd"

func readCloserString() io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(metaDataJSON))
}

func TestGetPrefix(t *testing.T) {
	if GetPrefix(readCloserString()) != expectedPrefix {
		t.Error("prefix is not valid")
	}
}
