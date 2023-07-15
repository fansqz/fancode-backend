package store

import (
	"FanCode/setting"
	"context"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
	"net/http"
	"net/url"
	"os"
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
	objects, _, err := c.client.Bucket.Get(context.TODO(), &cos.BucketGetOptions{
		Prefix:    storePath,
		Delimiter: "/",
	})
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	for _, obj := range objects.Contents {
		// 获取文件的名称
		filename := obj.Key[strings.LastIndex(obj.Key, "/")+1:]
		// 创建本地文件
		file, err := os.Create(filename)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		defer file.Close()

		// 下载文件
		_, err = c.client.Object.Get(context.TODO(), obj.Key, nil)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Downloaded %s\n", filename)
		}
	}
	return nil
}
