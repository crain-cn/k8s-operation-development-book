package main

import (
	"context"
	"fmt"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhaoyu/.kube/config")
	if err != nil {
		panic(err.Error())
	}

	// 创建Pipeline ClientSet
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespace :=  "default" // Namespace空间

	pipelineRun := &v1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pipeline-run-go1-19", // pipelinerun 名称
			Namespace: namespace,
		},
		Spec: v1.PipelineRunSpec{
			PipelineSpec: &v1.PipelineSpec{
				Tasks: []v1.PipelineTask{
					v1.PipelineTask{
						Name: "task-go1-19",  // task名称
						Params: v1.Params{
							v1.Param{ 		// 传递镜像地址参数
								Name:  "image",
								Value: v1.ParamValue{Type: v1.ParamTypeString, StringVal: "image.zhihu.com/zhihu/test:v1"},
							},
						},
						TaskSpec: &v1.EmbeddedTask{
							TaskSpec: v1.TaskSpec{
								Params: v1.ParamSpecs{ // 传递的参数
									v1.ParamSpec{ // image地址
										Name:        "image",
										Type:        v1.ParamTypeString,
										Description: "image url",
									},
								},
								Steps: []v1.Step{
									v1.Step{
										Name:       "go-build",      // 执行步骤名称
										Image:      "golang:1.19.0", // 执行时容器镜像
										WorkingDir: "/workspace",    // 工作目录
										Command: []string{ // 执行编译文件
											"/bin/bash",
										},
										Args: []string{
											"-c",
											"-exu",
											`sh $(workspaces.config.path)/go.sh`,
										},
										Env: []corev1.EnvVar{ // 环境变量、这里增加了GOPROXY 和 GO111MODULE，用于go vendor代理
											corev1.EnvVar{
												Name:  "GOPROXY",
												Value: "https://mirrors.aliyun.com/goproxy",
											},
											corev1.EnvVar{
												Name:  "GO111MODULE",
												Value: "on",
											},
										},
										VolumeMounts: []corev1.VolumeMount{
											corev1.VolumeMount{ // 挂载工作目录
												Name:      "temp",
												MountPath: "/workspace",
											},
										},
									},
									v1.Step{
										Name:       "docker-build",   // 执行步骤名称
										Image:      "docker:18.09.9", // 执行时容器镜像
										WorkingDir: "/workspace",     // 工作目录
										Command: []string{ // 执行内容，这里主要做 docker build, 编译的地址是通过$image传递
											"/bin/sh",
										},
										Args: []string{
											"-c",
											"-exu",
											`image=$(params.image)
											 docker build -t $image -f $(workspaces.config.path)/Dockerfile  ./
											 docker push $image
											 docker rmi $image`,
										},
										VolumeMounts: []corev1.VolumeMount{
											corev1.VolumeMount{ // 挂载docker.sock
												Name:      "dockersocket",
												MountPath: "/var/run/docker.sock",
											},
											corev1.VolumeMount{ // 挂载工作目录
												Name:      "temp",
												MountPath: "/workspace",
											},
										},
									},
								},
								Workspaces: []v1.WorkspaceDeclaration{
									v1.WorkspaceDeclaration{ //configmap内容
										Name:        "config",
										Description: "configmap workspace",
									},
								},
								Volumes: []corev1.Volume{
									corev1.Volume{ // 挂载本地docker.sock 用于docker build或者是docker push
										Name: "dockersocket",
										VolumeSource: corev1.VolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/var/run/docker.sock",
											},
										},
									},
									corev1.Volume{ // 挂载本地temp目录，用于编译时临时存储二进制，针对一些编译文件可以加载缓存等等
										Name: "temp",
										VolumeSource: corev1.VolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/tmp",
											},
										},
									},
								},
							},
						},
					},
				},
				Workspaces: []v1.PipelineWorkspaceDeclaration{
					v1.PipelineWorkspaceDeclaration{	// configmap模版
						Name: "config",
					},
				},
			},
			Workspaces: []v1.WorkspaceBinding{
				v1.WorkspaceBinding{	// 绑定configmap模版
					Name: "config",
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "pipeline-go1-19",
						},
					},
				},
			},
		},
	}
	// 创建PipelineRun
	_, err = clientset.TektonV1().PipelineRuns(namespace).Create(context.TODO(), pipelineRun, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("create PipelineRun err: ", err.Error())
		return
	}
	fmt.Println("创建成功")

}