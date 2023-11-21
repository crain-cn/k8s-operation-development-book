package main

import (
	"fmt"
	"github.com/containers/image/v5/manifest"
	"strings"
)

// 判断是否增加tag
func CheckIfIncludeTag(repository string) bool {
	return strings.Contains(repository, ":")
}

// 解析Manifest数据
func ManifestHandler(m []byte, t string, i *RegistrySource) ([]manifest.Manifest, error) {

	var manifestInfoSlice []manifest.Manifest

	if t == manifest.DockerV2Schema2MediaType {
		manifestInfo, err := manifest.Schema2FromManifest(m)
		if err != nil {
			return nil, err
		}
		manifestInfoSlice = append(manifestInfoSlice, manifestInfo)
		return manifestInfoSlice, nil
	} else if t == manifest.DockerV2Schema1MediaType || t == manifest.DockerV2Schema1SignedMediaType {
		manifestInfo, err := manifest.Schema1FromManifest(m)
		if err != nil {
			return nil, err
		}
		manifestInfoSlice = append(manifestInfoSlice, manifestInfo)
		return manifestInfoSlice, nil
	} else if t == manifest.DockerV2ListMediaType {
		manifestSchemaListInfo, err := manifest.Schema2ListFromManifest(m)
		if err != nil {
			return nil, err
		}

		for _, manifestDescriptorElem := range manifestSchemaListInfo.Manifests {
			manifestByte, manifestType, err := i.source.GetManifest(i.ctx, &manifestDescriptorElem.Digest)
			if err != nil {
				return nil, err
			}

			platformSpecManifest, err := ManifestHandler(manifestByte, manifestType, i)
			if err != nil {
				return nil, err
			}

			manifestInfoSlice = append(manifestInfoSlice, platformSpecManifest...)
		}
		return manifestInfoSlice, nil
	}

	return nil, fmt.Errorf("unsupported manifest type: %v", t)
}
