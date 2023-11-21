package main

import (
	"context"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"io"
)

type RegistroyDestination struct {
	destinationRef types.ImageReference
	destination    types.ImageDestination
	ctx            context.Context
	sysctx         *types.SystemContext

	// 目标镜像地址参数
	registry   string
	repository string
	tag        string
}

func NewRegistroyDestination(registry, repository, tag, username, password string, insecure bool) (*RegistroyDestination, error) {

	// 判断tag是否为空,如果非空则拼接tagWithColon参数
	tagWithColon := ""
	if tag != "" {
		tagWithColon = ":" + tag
	}

	// 如果tag为空，则tag默认为latest
	destRef, err := docker.ParseReference("//" + registry + "/" + repository + tagWithColon)
	if err != nil {
		return nil, err
	}

	var sysctx *types.SystemContext
	if insecure {
		// 使用http访问镜像服务
		sysctx = &types.SystemContext{
			DockerInsecureSkipTLSVerify: types.OptionalBoolTrue,
		}
	} else {
		sysctx = &types.SystemContext{}
	}


	ctx := context.WithValue(context.Background(), interface{}("RegistroyDestination"), repository)
	// 设置账号密码信息
	if username != "" && password != "" {
		sysctx.DockerAuthConfig = &types.DockerAuthConfig{
			Username: username,
			Password: password,
		}
	}

	// 创建目标地址镜像对象
	rawDestination, err := destRef.NewImageDestination(ctx, sysctx)
	if err != nil {
		return nil, err
	}

	return &RegistroyDestination{
		destinationRef: destRef,
		destination:    rawDestination,
		ctx:            ctx,
		sysctx:         sysctx,
		registry:       registry,
		repository:     repository,
		tag:            tag,
	}, nil
}

// 判断目标地址Blob是否存在
func (i *RegistroyDestination) CheckBlobExist(blobInfo types.BlobInfo) (bool, error) {
	exist, _, err := i.destination.TryReusingBlob(i.ctx, types.BlobInfo{
		Digest: blobInfo.Digest,
		Size:   blobInfo.Size,
	}, NoCache, false)

	return exist, err
}

// push 镜像manifest数据
func (i *RegistroyDestination) PushManifest(manifestByte []byte) error {
	return i.destination.PutManifest(i.ctx, manifestByte, nil)
}

// push 镜像blob数据
func (i *RegistroyDestination) PutABlob(blob io.ReadCloser, blobInfo types.BlobInfo) error {
	_, err := i.destination.PutBlob(i.ctx, blob, types.BlobInfo{
		Digest: blobInfo.Digest,
		Size:   blobInfo.Size,
	}, NoCache, true)


	// 关闭io.ReadCloser
	defer blob.Close()

	return err
}

// 获取目标镜像Registry地址
func (i *RegistroyDestination) GetRegistry() string {
	return i.registry
}

// 获取目标镜像Repository
func (i *RegistroyDestination) GetRepository() string {
	return i.repository
}

// 获取目标镜像Tag
func (i *RegistroyDestination) GetTag() string {
	return i.tag
}

func (i *RegistroyDestination) Close() error {
	return i.destination.Close()
}