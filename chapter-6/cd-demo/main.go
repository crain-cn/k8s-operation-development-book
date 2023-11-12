package main
import (
	"context"
	"fmt"
	"github.com/openkruise/rollouts/api/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/utils/pointer"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func createDeployment(dynamicClient dynamic.Interface) error {
	rolloutDeploy := appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rollout-demo", // Deployment名称
			Labels: map[string]string{
				"app" : "rollout-demo",
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas:  pointer.Int32(10), // 副本数
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app" : "rollout-demo",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app" : "rollout-demo",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name: "rollout-demo", // 容器名称
							Image: "nginx:1.7.9", // 镜像地址
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{  // 映射端口
									ContainerPort: 80,
									Name: "http-port",
									Protocol: "TCP",
								},
							},
						},
					},
				},
			},
		},
	}

	// 转换Unstructured
	deploy ,err := runtime.DefaultUnstructuredConverter.ToUnstructured(&rolloutDeploy)
	if err != nil {
		return err
	}
	// 创建Deployment
	_, err = dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}).Namespace("default").Create(context.TODO(), &unstructured.Unstructured{deploy} ,metav1.CreateOptions{})

	return err
}

func createService(dynamicClient dynamic.Interface) error {
	iPFamilyPolicy := corev1.IPFamilyPolicySingleStack
	internalTrafficPolicy := corev1.ServiceInternalTrafficPolicyCluster
	rolloutService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rollout-demo", // Servcie名称
			Labels: map[string]string{ // 标签
				"app" : "rollout-demo",
			},
		},
		Spec: corev1.ServiceSpec{
			InternalTrafficPolicy: &internalTrafficPolicy,
			IPFamilies: []corev1.IPFamily{
				corev1.IPv4Protocol,
			},
			IPFamilyPolicy: &iPFamilyPolicy,
			Ports: []corev1.ServicePort{ // 映射端口
				corev1.ServicePort{
					Name: "http",
					Port: 80,
					Protocol: "TCP",
					TargetPort: intstr.IntOrString{
						Type: intstr.String,
						IntVal: 80,
					},
				},
			},
			SessionAffinity: corev1.ServiceAffinityNone,
			Type: corev1.ServiceTypeClusterIP, 	// Service 类型
		},
	}

	// 转换Unstructured
	service ,err := runtime.DefaultUnstructuredConverter.ToUnstructured(&rolloutService)
	if err != nil {
		return err
	}
	// 创建Service
	_, err = dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}).Namespace("default").Create(context.TODO(), &unstructured.Unstructured{service} ,metav1.CreateOptions{})

	return err
}

func createIngress(dynamicClient dynamic.Interface) error {
	prefix := v1.PathTypePrefix
	rolloutIngress := v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingress-demo", // ingress名称
		},
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				v1.IngressRule{
					Host: "rollout-demo.zhihu.com", // host配置
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{ // 配置路径
								v1.HTTPIngressPath{
									Path: "/", // 路径
									PathType: &prefix,
									Backend: v1.IngressBackend{ //配置后端服务
										Service: &v1.IngressServiceBackend{
											Name: "rollout-demo",  // 对应Servcie名称
											Port: v1.ServiceBackendPort{ // 对应Service 端口
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	// 转换Unstructured
	ingress ,err := runtime.DefaultUnstructuredConverter.ToUnstructured(&rolloutIngress)
	if err != nil {
		return err
	}
	// 创建ingress
	_, err = dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}).Namespace("default").Create(context.TODO(), &unstructured.Unstructured{ingress} ,metav1.CreateOptions{})
	return err
}

func createRollout(dynamicClient dynamic.Interface) error {
	rollout := v1alpha1.Rollout{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rollout-demo",  // rollot名称
			Annotations: map[string]string{
				"rollouts.kruise.io/rolling-style": "partition", //设置滚动类型
			},
		},
		Spec: v1alpha1.RolloutSpec{
			ObjectRef: v1alpha1.ObjectRef{
				WorkloadRef: &v1alpha1.WorkloadRef{ // 关联workload
					APIVersion: "apps/v1",
					Kind: "Deployment",
					Name: "rollout-demo",
				},
			},
			Strategy: v1alpha1.RolloutStrategy{
				Canary: &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						v1alpha1.CanaryStep{ // 第一批
							Weight: pointer.Int32(5), // 将 5% 的流量路由到新版本
							Replicas: &intstr.IntOrString{ // 发布副本的第一步。如果未设置，则默认使用“weight”，如上所示为 5%。
								Type: intstr.Int,
								IntVal: 1,
							},
						},
						v1alpha1.CanaryStep{ // 第二批
							Replicas: &intstr.IntOrString{ // 滚动50% Pod
								Type: intstr.String,
								StrVal: "50%",
							},
							Pause: v1alpha1.RolloutPause{ // 等待 1 小时后自动进入下一批
								Duration: pointer.Int32(3600),
							},
						},
						v1alpha1.CanaryStep{ // 第三批
							Replicas: &intstr.IntOrString{ // 滚动100% Pod
								Type: intstr.String,
								StrVal: "100%",
							},
						},
					},
					TrafficRoutings: []*v1alpha1.TrafficRouting{
						&v1alpha1.TrafficRouting{ // 与工作负载相关的服务名称
							Service: "rollout-demo",  // service 名称
							Ingress: &v1alpha1.IngressTrafficRouting{ // ingress 配置
								Name: "ingress-demo",
								ClassType: "nginx",
							},
						},
					},
				},
			},
		},
	}

	// 转化Unstructured
	rolloutobj ,err := runtime.DefaultUnstructuredConverter.ToUnstructured(&rollout)
	if err != nil {
		return err
	}

	// 创建rollou
	_, err = dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "rollouts.kruise.io",
		Version:  "v1alpha1",
		Resource: "rollouts",
	}).Namespace("default").Create(context.TODO(), &unstructured.Unstructured{rolloutobj} ,metav1.CreateOptions{})
	return err
}

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhaoyu/.kube/config")
	if err != nil {
		panic(err.Error())
	}

	// 创建 DynamicClient
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// 创建Deployment
	err = createDeployment(dynamicClient)
	if err != nil {
		fmt.Println("创建Deployment失败，", err.Error())
		return
	}
	// 创建service
	err = createService(dynamicClient)
	if err != nil {
		fmt.Println("创建Service失败，", err.Error())
		return
	}
	// 创建ingress
	err = createIngress(dynamicClient)
	if err != nil {
		fmt.Println("创建Ingress失败，", err.Error())
		return
	}
	// 创建rollout
	err = createRollout(dynamicClient)
	if err != nil {
		fmt.Println("创建Rollout失败，", err.Error())
		return
	}
	fmt.Println("创建成功")
}

