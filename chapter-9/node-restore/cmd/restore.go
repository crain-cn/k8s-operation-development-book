package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/rest"
	"os"
	"sigs.k8s.io/yaml"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// 初始化cobra
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context())
	},
}

var (
	backDir        string
	kubeConfig     string
	deleteInterval int64
)

func init() {
	rootCmd.AddCommand(restoreCmd)
	// 备份目录
	restoreCmd.Flags().StringVar(&backDir, "back-dir", "./backdir", "back up node yaml dir")
	// kubeconfig配置
	restoreCmd.Flags().StringVar(&kubeConfig, "kubeconfig", "", "back up node yaml dir")
	// 删除节点等待时间
	restoreCmd.Flags().Int64Var(&deleteInterval, "delete-interval", 5, "delete node interval")
}

func run(ctx context.Context) {
	// 创建备份目录
	err := os.MkdirAll(backDir, 0755)
	if err != nil {
		logrus.Fatalf("mkdir %s err: %v", backDir, err)
	}
	// 初始化clientset
	var clientset *kubernetes.Clientset
	if len(kubeConfig) == 0 {
		config, err := rest.InClusterConfig()
		if err != nil {
			logrus.Fatalf("new incluster client set err: %v", err)
		}
		clientset, err = kubernetes.NewForConfig(config)
	} else {
		config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			logrus.Fatalf("build config failed, err: %v", err)
		}
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			logrus.Fatalf("new out of cluster client set err: %v", err)
		}
	}

	// 通过节点标签过滤机器列表
	nodeRequirement, err := labels.NewRequirement("kubernetes.io/role", selection.Equals, []string{"node"})
	if err != nil {
		logrus.Fatalf("create node requirement fail, err: %v ", err)
	}
	nodeRoleRequirement, err := labels.NewRequirement("node-role.kubernetes.io/node", selection.Exists, []string{})
	if err != nil {
		logrus.Fatalf("create node role requirement fail, err: %v ", err)
	}

	selector := labels.NewSelector().Add(*nodeRequirement).Add(*nodeRoleRequirement)

	for {
		select {
		case <-ctx.Done():
			logrus.Debugf("program will exit")
			return
		default:
			// 获得节点信息
			nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{
				LabelSelector: selector.String(),
			})
			if err != nil {
				logrus.Fatalf("get node list err: %v", err)
			}

			totalNode := len(nodeList.Items)
			// 执行备份变更操作
			restoreNode(ctx, clientset, totalNode, nodeList)
			logrus.Debugf("%d all node restore success", totalNode)
			return
		}
	}

}

// 执行备份变更操作
func restoreNode(ctx context.Context, clientset *kubernetes.Clientset, totalNode int, nodeList *corev1.NodeList) {
	for i := 0; i < totalNode; i++ {
		node := nodeList.Items[i]
		nodeBs, err := json.Marshal(node)
		if err != nil {
			logrus.Fatalf("%d/%d node: %s, json marshal node err: %v", i+1, totalNode, node.Name, err)
		}

		bs, err := yaml.JSONToYAML(nodeBs)
		if err != nil {
			logrus.Fatalf("%d/%d node: %s, json to yaml err: %v", i+1, totalNode, node.Name, err)
		}

		fileName := fmt.Sprintf("%s/%s.yaml", backDir, node.Name)
		// 备份node数据到备份目录
		err = os.WriteFile(fileName, bs, 0755)
		if err != nil {
			logrus.Fatalf("%d/%d back up node: %s yaml err: %v", i+1, totalNode, node.Name, err)
		}

		// 清空node spec pod cidr
		node.Spec.PodCIDRs = nil
		node.Spec.PodCIDR = ""
		// 清空node status
		node.Status = corev1.NodeStatus{}
		node.ResourceVersion = ""

		// 删除节点
		err = clientset.CoreV1().Nodes().Delete(ctx, node.Name, metav1.DeleteOptions{})
		if err != nil {
			logrus.Fatalf("%d/%d delete node: %s failed, err: %v", i+1, totalNode, node.Name, err)
		}

		logrus.Debugf("%d/%d delete node: %s, success", i+1, totalNode, node.Name)

		// 创建节点
		newNode, err := clientset.CoreV1().Nodes().Create(ctx, &node, metav1.CreateOptions{})
		if err != nil {
			logrus.Fatalf("%d/%d delete node: %s failed, err: %v", i+1, totalNode, node.Name, err)
		}

		logrus.Debugf("%d/%d create node: %s, success", i+1, totalNode, newNode.Name)

		// 等待节点Ready
		waitNodeReady(ctx, i+1, totalNode, clientset, node.Name)
		if deleteInterval > 0 {
			logrus.Debugf("%d/%d will sleep %d second, then to restore next node", i+1, totalNode, deleteInterval)
			time.Sleep(time.Second * time.Duration(deleteInterval))
		}
	}
}

// 等待节点Ready
func waitNodeReady(ctx context.Context, idx, total int, clientset *kubernetes.Clientset, nodeName string) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			node, err := clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			if err != nil {
				logrus.Fatalf("%d/%d get node: %s err: %v", idx, total, nodeName, err)
			}

			for i := 0; i < len(node.Status.Conditions); i++ {
				if node.Status.Conditions[i].Type == corev1.NodeReady {
					logrus.Debugf("%d/%d node: %s ready", idx, total, nodeName)
					return
				}
			}

			logrus.Debugf("%d/%d node: %s not ready", idx, total, nodeName)
		}
	}
}
