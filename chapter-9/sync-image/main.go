package main

import (
	"context"
	"fmt"
	"github.com/containers/image/v5/manifest"
)

func main() {

	username := "" 	// 账号
	password := "" // 密码

	sourceRegistry := "" 						// 源镜像地址
	sourceRepository := ""		 				// 源Repository信息
	sourceTag := ""								// 源tag信息

	destinationRegistry := ""  					// 目标镜像地址
	destinationRepository := ""				    // 目标Repository信息
	destinationTag := ""				 		// 目标tag信息


	// 初始化源信息
	source, err := NewRegistrySource(sourceRegistry, sourceRepository, sourceTag, username, password, false)
	if err != nil {
		fmt.Println("初始化源信息失败： ", err.Error())
		return
	}

	// 获取源地址Manifest信息
	sourceByte, manifestType, err := source.source.GetManifest(context.TODO(), nil)
	if err != nil {
		fmt.Println("获取manifest信息失败： ", err.Error())
		return
	}

	// 初始化目标信息
	destination, err := NewRegistroyDestination(destinationRegistry, destinationRepository, destinationTag, username, password, false)
	if err != nil {
		fmt.Println("初始化目标信息失败： ", err.Error())
		return
	}

	// 获取层信息数据
	blobInfos, err := source.GetBlobInfos(sourceByte, manifestType)
	if err != nil {
		fmt.Println("获取Blobinfos信息: ", err.Error())
		return
	}

	for _, b := range blobInfos {
		// 检测Blob层数据再目标地址中是否存在
		blobExist, err := destination.CheckBlobExist(b)
		if err != nil {
			fmt.Println("检测Blob失败: ", err.Error())
			return
		}

		// 判断Blob层数据再目标地址中是否存在
		if !blobExist {
			// pull一个Blob信息
			blob, size, err := source.GetBlob(b)
			if err != nil {
				fmt.Println("获取Blob信息失败", err.Error())
				return
			}

			// Push Blob层信息到目标地址中
			b.Size = size
			if err := destination.PutABlob(blob, b); err != nil {
				fmt.Println("Push Blob数据失败", err.Error())
				return
			}

		} else {
			//存在则不处理
			fmt.Println(fmt.Sprintf(" Blob %s(%v) 已经拉取 %s, 不需要再拉取", b.Digest, b.Size, destination.GetRegistry()+"/"+destination.GetRepository()))
		}
	}

	// 处理Manifest信息，如果manifestType是application/vnd.docker.distribution.manifest.list.v2+json 类型则循环Push
	if manifestType == manifest.DockerV2ListMediaType {
		// 获得Manifest列表
		manifestSchemaListInfo, err := manifest.Schema2ListFromManifest(sourceByte)
		if err != nil {
			fmt.Println("获取Manifest List数据失败: ", err.Error())
			return
		}
		var subManifestByte []byte

		// 循环处理 manifest 到目标地址
		for _, manifestDescriptorElem := range manifestSchemaListInfo.Manifests {
			// 获得源Manifest
			subManifestByte, _, err = source.source.GetManifest(source.ctx, &manifestDescriptorElem.Digest)
			if err != nil {
				fmt.Println("获取源 Manifest 信息错误: ", err.Error())
				return
			}

			// push manifest 到目标地址
			if err := destination.PushManifest(subManifestByte); err != nil {
				fmt.Println("Push Manifest 错误: ", err.Error())
				return
			}

		}
		// push manifest 列表到目标地址
		if err := destination.PushManifest(sourceByte); err != nil {
			fmt.Println("Push Manifest 错误: ", err.Error())
			return
		}
	} else {
		// push manifest 到目标地址
		if err := destination.PushManifest(sourceByte); err != nil {
			fmt.Println("Push Manifest 错误: ", err.Error())
			return
		}

	}

	fmt.Println("Push 镜像成功，地址为 ",fmt.Sprintf("%s/%s:%s",destination.GetRegistry(), destination.GetRepository(), destination.GetTag()))
	source.Close()
	destination.Close()
}
