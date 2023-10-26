package cosched

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	Name = "cosched" //插件名称
)

var (
	_ framework.ScorePlugin = &Plugin{} // 映射ScorePlugin到Plugin结构
)

type Plugin struct {
	handle framework.Handle
}

func New(args runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	klog.InfoS("cosched plugin init")

	return &Plugin{
		handle: handle,
	}, nil
}

func (p *Plugin) Name() string { return Name }

// 打分扩展函数
func (p *Plugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// 打分函数
func (p *Plugin) Score(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) (int64, *framework.Status) {

	nodeInfo, err := p.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}
	node := nodeInfo.Node()
	if node == nil {
		return 0, framework.NewStatus(framework.Error, "node not found")
	}

	// 判断有cosched=on标签的节点给100分
	if node.Labels["cosched"] == "on" {
		return 100, nil
	}

	return 0, nil
}
