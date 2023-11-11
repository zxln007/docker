package main

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"io"
	"os"
	"strings"
)

func (c *ClientAPI) ExportImage(imageId string) (io.ReadCloser, error) {

	r, err := c.DockerCli.ImageSave(c.Ctx, []string{imageId})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *ClientAPI) ImportImage(savePath string, repo string, tag string) error {

	fileReader, err := os.Open(savePath)
	defer fileReader.Close()

	source := types.ImageImportSource{
		Source:     fileReader,
		SourceName: "",
	}
	options := types.ImageImportOptions{
		Tag:      tag,
		Message:  "import",
		Changes:  nil,
		Platform: "",
	}

	response, err := c.DockerCli.ImageImport(c.Ctx, source, repo, options)
	defer response.Close()
	if err != nil {
		return err
	}
	// 解析响应并读取导入日志
	_, err = io.Copy(os.Stdout, response)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClientAPI) PullImage(imageUrl string) error {
	//cm.DockerCli.ImageList(cm.Ctx,types.ImageListOptions{Filters: })
	events, err := c.DockerCli.ImagePull(c.Ctx, imageUrl, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer events.Close()
	io.Copy(os.Stdout, events)
	return nil

}

func (c *ClientAPI) EnsureImageExist(image string) (bool, error) {
	var args = filters.NewArgs()
	if !strings.Contains(image, "@") {
		args.Add("reference", image)
	}
	images, err := c.DockerCli.ImageList(c.Ctx, types.ImageListOptions{Filters: args})
	if err != nil {
		return false, err
	}

	for _, _image := range images {
		//fmt.Println(_image.RepoDigests[0])
		if strings.Contains(image, "@") {
			if strings.Contains(_image.RepoDigests[0], image) {
				return true, nil
			}
		} else {
			if strings.Contains(_image.RepoTags[0], image) {
				return true, nil
			}
		}
	}
	return false, errors.New(fmt.Sprintf("can not find image %s", image))
}

func (c *ClientAPI) RemoveImage(imageId string, removeOpt types.ImageRemoveOptions) error {
	_, err := c.DockerCli.ImageRemove(c.Ctx, imageId, removeOpt)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClientAPI) ImageInfo(imageId string) (*types.ImageInspect, error) {
	inspect, _, err := c.DockerCli.ImageInspectWithRaw(c.Ctx, imageId)
	if err != nil {
		return nil, err
	}
	return &inspect, nil
}

func (c *ClientAPI) IsImageUsed(imageID string) (bool, error) {
	containerInfos, err := c.DockerCli.ContainerList(c.Ctx, types.ContainerListOptions{
		All: true})
	if err != nil {
		return false, err
	}
	for _, info := range containerInfos {
		if info.ImageID == imageID {
			return true, nil
		}
	}
	return false, nil
}
