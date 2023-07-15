package store

import (
	"FanCode/setting"
	"context"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type COS struct {
	client *cos.Client
}

func NewCOS() *COS {
	storeConfig := setting.Conf.COSConfig
	u, _ := url.Parse(fmt.Sprintf("http://%s-%s.cos.%s.myqcloud.com",
		storeConfig.BucketName, storeConfig.AppID, storeConfig.Region))
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  storeConfig.SecretID,
			SecretKey: storeConfig.SecretKey,
			Transport: &debug.DebugRequestTransport{
				RequestHeader:  true,
				RequestBody:    true,
				ResponseHeader: true,
				ResponseBody:   true,
			},
		},
	})
	return &COS{
		client: client,
	}
}

func (c *COS) SaveFile(storePath string, file *strings.Reader) error {
	_, err := c.client.Object.Put(context.Background(), storePath, file, nil)
	return err
}

// storePaht:对象存储的路劲，loadPath:本地路径
func (c *COS) LoadFile(storePath, localPath string) error {
	_, err := c.client.Object.GetToFile(context.Background(), storePath, localPath, nil)
	return err
}

func (c *COS) LoadFolder(storePath, localPath string) error {
	// 列出文件夹下的对象
	// 获取指定前缀下的文件列表
	options := &cos.BucketGetOptions{
		Prefix: storePath,
	}
	res, _, err := c.client.Bucket.Get(context.Background(), options)
	if err != nil {
		fmt.Println("Failed to get bucket:", err)
		return err
	}

	// 遍历文件列表并下载每个文件
	for _, object := range res.Contents {
		// 构建文件下载的本地路径
		filePathInCos := object.Key

		// 下载文件
		resp, err1 := c.client.Object.Get(context.Background(), object.Key, nil)
		if err1 != nil {
			fmt.Println("Failed to download file:", err1)
			return err1
		}
		defer resp.Body.Close()

		// 读取文件内容
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Failed to read file:", err)
			return err
		}

		// 创建文件夹
		filePathInLocal := strings.Replace(filePathInCos, storePath, localPath, 1)
		err = os.MkdirAll(filepath.Dir(filePathInLocal), 0755)
		if err != nil {
			fmt.Println("Failed to save file:", err)
			return err
		}

		// 将文件内容写入本地文件
		err = ioutil.WriteFile(filePathInLocal, data, 0644)
		if err != nil {
			fmt.Println("Failed to save file:", err)
			return err
		}

		fmt.Println("File downloaded:", localPath)
	}
	return nil
}

func (c *COS) DeleteFolder(storePath string) error {
	// 列出文件夹下的对象
	// 获取指定前缀下的文件列表
	options := &cos.BucketGetOptions{
		Prefix: storePath,
	}
	res, _, err := c.client.Bucket.Get(context.Background(), options)
	if err != nil {
		fmt.Println("Failed to get bucket:", err)
		return err
	}
	// 遍历文件列表并下载每个文件
	for _, object := range res.Contents {
		// 构建文件下载的本地路径
		filePathInCos := object.Key

		// 下载文件
		_, deleteError := c.client.Object.Delete(context.Background(), filePathInCos, nil)
		if deleteError != nil {
			return deleteError
		}
	}
	return nil
}
