package main

import (
	"context"
	"fmt"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/pkg/blobinfocache/none"
	"github.com/containers/image/v5/types"
	"io"
)

type RegistrySource struct {
	sourceRef types.ImageReference
	source    types.ImageSource
	ctx       context.Context
	sysctx    *types.SystemContext

	// 源镜像地址参数
	registry   string
	repository string
	tag        string
}

var (
	// 关闭cache
	NoCache = none.NoCache
)

func NewRegistrySource(registry, repository, tag, username, password string, insecure bool) (*RegistrySource, error) {

	// 判断tag是否为空,如果非空则拼接tagWithColon参数
	tagWithColon := ""
	if tag != "" {
		tagWithColon = ":" + tag
	}

	// 如果tag为空，则tag默认为latest
	srcRef, err := docker.ParseReference("//" + registry + "/" + repository + tagWithColon)
	if err != nil {
		return nil, err
	}

	var sysctx *types.SystemContext
	if insecure {
		// 使用http访问镜像那个地址
		sysctx = &types.SystemContext{
			DockerInsecureSkipTLSVerify: types.OptionalBoolTrue,
		}
	} else {
		sysctx = &types.SystemContext{}
	}

	ctx := context.WithValue(context.Background(), interface{}("RegistrySource"), repository)
	// 设置账号密码信息
	if username != "" && password != "" {
		sysctx.DockerAuthConfig = &types.DockerAuthConfig{
			Username: username,
			Password: password,
		}
	}

	var rawSource types.ImageSource
	if tag != "" {
		// 如果tag为空，则tag默认为latest
		rawSource, err = srcRef.NewImageSource(ctx, sysctx)
		if err != nil {
			return nil, err
		}
	}

	return &RegistrySource{
		sourceRef:  srcRef,
		source:     rawSource,
		ctx:        ctx,
		sysctx:     sysctx,
		registry:   registry,
		repository: repository,
		tag:        tag,
	}, nil
}

// 通过manifest层数据获取blob列表信息
func (i *RegistrySource) GetBlobInfos(manifestByte []byte, manifestType string) ([]types.BlobInfo, error) {
	if i.source == nil {
		return nil, fmt.Errorf("source 对象不能为空")
	}

	// 获取解析后Manifest数据
	manifestInfoSlice, err := ManifestHandler(manifestByte, manifestType, i)
	if err != nil {
		return nil, err
	}

	// 获取 Blob列表
	srcBlobs := []types.BlobInfo{}
	for _, manifestInfo := range manifestInfoSlice {
		// 获取层信息
		blobInfos := manifestInfo.LayerInfos()
		for _, l := range blobInfos {
			srcBlobs = append(srcBlobs, l.BlobInfo)
		}

		configBlob := manifestInfo.ConfigInfo()
		if configBlob.Digest != "" {
			srcBlobs = append(srcBlobs, configBlob)
		}
	}

	return srcBlobs, nil
}

// 获取Blob信息
func (i *RegistrySource) GetBlob(blobInfo types.BlobInfo) (io.ReadCloser, int64, error) {
	return i.source.GetBlob(i.ctx, types.BlobInfo{Digest: blobInfo.Digest, Size: -1}, NoCache)
}

func (i *RegistrySource) Close() error {
	return i.source.Close()
}

// 获取源镜像Registry地址
func (i *RegistrySource) GetRegistry() string {
	return i.registry
}

// 获取源镜像Repository
func (i *RegistrySource) GetRepository() string {
	return i.repository
}

// 获取源镜像Tag
func (i *RegistrySource) GetTag() string {
	return i.tag
}
