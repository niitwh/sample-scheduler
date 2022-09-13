package podstate

import (
	"context"
	"fmt"
	"math"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type PodState struct {
	handle framework.Handle
}

var _ = framework.ScorePlugin(&PodState{})

// Name is the name of the plugin used in the Registry and configurations.
const Name = "PodState"

func (ps *PodState) Name() string {
	return Name
}

// Score invoked at the score extension point.
func (ps *PodState) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := ps.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}

	// ps.score favors nodes with less same kind pods
	// It calculates the sum of the node's same kind pods
	return ps.score(nodeInfo, pod)
}

// ScoreExtensions of the Score plugin.
func (ps *PodState) ScoreExtensions() framework.ScoreExtensions {
	return ps
}

func (ps *PodState) score(nodeInfo *framework.NodeInfo, pod *v1.Pod) (int64, *framework.Status) {
	var sameKindPodNum int64
	for _, p := range nodeInfo.Pods {
		if p.Pod.Labels["app"] == pod.Labels["app"] {
			sameKindPodNum++;
		}
	}
	return 100 - sameKindPodNum, nil
}

func (ps *PodState) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	// Find highest and lowest scores.
	var highest int64 = -math.MaxInt64
	var lowest int64 = math.MaxInt64
	for _, nodeScore := range scores {
		if nodeScore.Score > highest {
			highest = nodeScore.Score
		}
		if nodeScore.Score < lowest {
			lowest = nodeScore.Score
		}
	}

	// Transform the highest to lowest score range to fit the framework's min to max node score range.
	oldRange := highest - lowest
	newRange := framework.MaxNodeScore - framework.MinNodeScore
	for i, nodeScore := range scores {
		if oldRange == 0 {
			scores[i].Score = framework.MinNodeScore
		} else {
			scores[i].Score = ((nodeScore.Score - lowest) * newRange / oldRange) + framework.MinNodeScore
		}
	}

	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	return &PodState{handle: h}, nil
}
