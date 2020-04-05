package gkeresizer

import (
	"fmt"
	"google.golang.org/api/container/v1"
)

type Resizer struct {
	service container.Service
}

func NewResizer(svc container.Service) (*Resizer, error) {
	return &Resizer{
		service: svc,
	}, nil
}

func (r *Resizer) Resize(project string, location string, cluster string, nodePool string, nodeCount int64, autoscaling bool) (*string, error) {
	rb := &container.SetNodePoolSizeRequest{
		NodeCount: nodeCount,
	}

	resize, err := r.service.Projects.Locations.Clusters.NodePools.SetSize("projects/"+project+"/locations/"+location+"/clusters/"+cluster+"/nodePools/"+nodePool, rb).Do()

	if err != nil {
		return nil, err
	}

	sprintf := fmt.Sprintf("Resize status %q (%s)", resize.Name, resize.Status)

	return &sprintf, nil
}
