package ensure

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

func TestDealResource(t *testing.T) {
	var tests = []struct {
		in                  *v1.PodSpec
		cpuRequestsExpected string
	}{
		{
			createPod("0.1", "100m", "100m", "100m"),
			"100m",
		},
		{
			createPod("1", "100m", "100m", "100m"),
			"1",
		},
		{
			createPod("100m", "100m", "100m", "100m"),
			"100m",
		},
		{
			createPod("2Gi", "100m", "100m", "100m"),
			"2Gi",
		},
	}
	for _, tt := range tests {
		DealResource(tt.in)
		if tt.in.Containers[0].Resources.Requests.Cpu().String() != tt.cpuRequestsExpected {
			t.Fatalf("actual = %s; expected = %s", tt.in.Containers[0].Resources.Requests.Cpu().String(), tt.cpuRequestsExpected)
		}
	}
}

func createPod(cpuRequests, memRequests, cpuLimits, memLimits string) *v1.PodSpec {
	return &v1.PodSpec{
		Containers: []v1.Container{
			{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						"cpu":    resource.MustParse(cpuRequests),
						"memory": resource.MustParse(memRequests),
					},
					Limits: v1.ResourceList{
						"cpu":    resource.MustParse(cpuLimits),
						"memory": resource.MustParse(memLimits),
					},
				},
			},
		},
	}
}
