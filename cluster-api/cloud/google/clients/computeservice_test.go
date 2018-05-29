package clients_test

import (
	compute "google.golang.org/api/compute/v1"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
	"k8s.io/kube-deploy/cluster-api/cloud/google/clients"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImagesGet(t *testing.T) {
	mux, server, client := createMuxServerAndClient()
	defer server.Close()
	responseImage := compute.Image{
		Name:             "imageName",
		ArchiveSizeBytes: 544,
	}
	mux.Handle("/compute/v1/projects/projectName/global/images/imageName", handler(nil, &responseImage))
	image, err := client.ImagesGet("projectName", "imageName")
	assert.Nil(t, err)
	assert.NotNil(t, image)
	assert.Equal(t, "imageName", image.Name)
	assert.Equal(t, int64(544), image.ArchiveSizeBytes)
}

func TestImagesGetFromFamily(t *testing.T) {
	mux, server, client := createMuxServerAndClient()
	defer server.Close()
	responseImage := compute.Image{
		Name:             "imageName",
		ArchiveSizeBytes: 544,
	}
	mux.Handle("/compute/v1/projects/projectName/global/images/family/familyName", handler(nil, &responseImage))
	image, err := client.ImagesGetFromFamily("projectName", "familyName")
	assert.Nil(t, err)
	assert.NotNil(t, image)
	assert.Equal(t, "imageName", image.Name)
	assert.Equal(t, int64(544), image.ArchiveSizeBytes)
}

func TestInstancesDelete(t *testing.T) {
	mux, server, client := createMuxServerAndClient()
	defer server.Close()
	responseOperation := compute.Operation{
		Id: 4501,
	}
	mux.Handle("/compute/v1/projects/projectName/zones/zoneName/instances/instanceName", handler(nil, &responseOperation))
	op, err := client.InstancesDelete("projectName", "zoneName", "instanceName")
	assert.Nil(t, err)
	assert.NotNil(t, op)
	assert.Equal(t, uint64(4501), responseOperation.Id)
}

func TestInstancesGet(t *testing.T) {
	mux, server, client := createMuxServerAndClient()
	defer server.Close()
	responseInstance := compute.Instance{
		Name: "instanceName",
		Zone: "zoneName",
	}
	mux.Handle("/compute/v1/projects/projectName/zones/zoneName/instances/instanceName", handler(nil, &responseInstance))
	instance, err := client.InstancesGet("projectName", "zoneName", "instanceName")
	assert.Nil(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "instanceName", instance.Name)
	assert.Equal(t, "zoneName", instance.Zone)
}

func TestInstancesInsert(t *testing.T) {
	mux, server, client := createMuxServerAndClient()
	defer server.Close()
	responseOperation := compute.Operation{
		Id: 3001,
	}
	mux.Handle("/compute/v1/projects/projectName/zones/zoneName/instances", handler(nil, &responseOperation))
	op, err := client.InstancesInsert("projectName", "zoneName", nil)
	assert.Nil(t, err)
	assert.NotNil(t, op)
	assert.Equal(t, uint64(3001), responseOperation.Id)
}

func createMuxServerAndClient() (*http.ServeMux, *httptest.Server, *clients.ComputeService) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	client, _ := clients.NewComputeServiceForURL(server.Client(), server.URL)
	return mux, server, client
}

func handler(err *googleapi.Error, obj interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		handleTestRequest(w, err, obj)
	}
}

func handleTestRequest(w http.ResponseWriter, handleErr *googleapi.Error, obj interface{}) {
	if handleErr != nil {
		http.Error(w, errMsg(handleErr), handleErr.Code)
		return
	}
	res, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, "json marshal error", http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func errMsg(e *googleapi.Error) string {
	res, err := json.Marshal(&errorReply{e})
	if err != nil {
		return "json marshal error"
	}
	return string(res)
}

type errorReply struct {
	Error *googleapi.Error `json:"error"`
}
