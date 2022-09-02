package plugins

import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"time"
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// name of plugin
const Name = "sample-plugin"

type Sample struct {
	handle framework.Handle
}

func (s *Sample) Name() string {
	return Name
}

func getNodeNameHash(s string) int {
	value := int(crc32.ChecksumIEEE([]byte(s)))
	if value >= 0 {
		return value
	} 

	if -value >= 0 {
		return -value
	}

	return 100
}



func (s *Sample) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	klog.V(3).Infof("prefilter pod: %v", pod.Name)
	return framework.NewStatus(framework.Success, "")	
}


func (s *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.V(3).Infof("filter pod: %v, node: %v", pod.Name, nodeInfo.Node().Name)
	return framework.NewStatus(framework.Success, "")
}

// score based on node name
func (s *Sample) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	hashValue := getNodeNameHash(nodeName) % 100
	rand.Seed(time.Now().UnixNano())
	score := (int64)(hashValue + rand.Intn(100)) % 100
	klog.V(3).Infof("node: %v, score: %d", nodeName, score)
	return score, framework.NewStatus(framework.Success, "")
}


func (s *Sample) PreBind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) *framework.Status {
	klog.V(3).Infof("prebind node info: %v", nodeName)
	return framework.NewStatus(framework.Success, "")
}



func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	return &Sample{
		handle: h,
	}, nil
}