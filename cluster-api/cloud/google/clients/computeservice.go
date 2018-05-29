package clients

import (
	compute "google.golang.org/api/compute/v1"
	"net/http"
	"net/url"
)

type ComputeService struct {
	service *compute.Service
}

func NewComputeService(client *http.Client) (*ComputeService, error) {
	return newComputeService(client)
}

func NewComputeServiceForURL(client *http.Client, baseURL string) (*ComputeService, error) {
	ComputeServiceImpl, err := newComputeService(client)
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(ComputeServiceImpl.service.BasePath)
	if err != nil {
		return nil, err
	}
	ComputeServiceImpl.service.BasePath = baseURL + url.Path
	return ComputeServiceImpl, err
}

func newComputeService(client *http.Client) (*ComputeService, error) {
	service, err := compute.New(client)
	if err != nil {
		return nil, err
	}
	return &ComputeService{
		service: service,
	}, nil
}

func (c *ComputeService) ImagesGet(project string, image string) (*compute.Image, error) {
	return c.service.Images.Get(project, image).Do()
}

func (c *ComputeService) ImagesGetFromFamily(project string, family string) (*compute.Image, error) {
	return c.service.Images.GetFromFamily(project, family).Do()
}

func (c *ComputeService) InstancesDelete(project string, zone string, targetInstance string) (*compute.Operation, error) {
	return c.service.Instances.Delete(project, zone, targetInstance).Do()
}

func (c *ComputeService) InstancesGet(project string, zone string, instance string) (*compute.Instance, error) {
	return c.service.Instances.Get(project, zone, instance).Do()
}

func (c *ComputeService) InstancesInsert(project string, zone string, instance *compute.Instance) (*compute.Operation, error) {
	return c.service.Instances.Insert(project, zone, instance).Do()
}

func (c *ComputeService) ZoneOperationsGet(project string, zone string, operation string) (*compute.Operation, error) {
	return c.service.ZoneOperations.Get(project, zone, operation).Do()
}
