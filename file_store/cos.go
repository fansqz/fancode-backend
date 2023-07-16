package file_store

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

type cosStore struct {
	client *cos.Client
}

func NewCOS() Store {
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
	return &cosStore{
		client: client,
	}
}

func (c *cosStore) SaveFile(storePath string, file *strings.Reader) error {
	_, err := c.client.Object.Put(context.Background(), storePath, file, nil)
	return err
}

// storePaht:对象存储的路劲，loadPath:本地路径
func (c *cosStore) DownloadFile(storePath, localPath string) error {
	_, err := c.client.Object.GetToFile(context.Background(), storePath, localPath, nil)
	return err
}

func (c *cosStore) DownloadFolder(storePath, localPath string) error {
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

func (c *cosStore) DeleteFolder(storePath string) error {
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

func (c *cosStore) UploadFolder(storePath string, localPath string) {

	// 对每个文件进行上传
	err := filepath.Walk(localPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Failed to access file %q: %v\n", filePath, err)
			return err
		}

		if !info.IsDir() {
			dstPath, _ := filepath.Rel(localPath, filePath) // 计算文件相对路径，用于指定远程路径
			key := filepath.ToSlash(dstPath)                // 将路径中的 \ 替换为 /

			// 使用 PutObject 接口上传文件
			_, err = c.client.Object.PutFromFile(context.Background(), storePath+"/"+key, filePath, nil)
			if err != nil {
				fmt.Printf("Failed to upload file %q: %v\n", filePath, err)
				return err
			}

			fmt.Printf("Successfully uploaded file: %s\n", key)
		}

		return nil
	})

	if err != nil {
		fmt.Println("Failed to upload folder:", err)
	} else {
		fmt.Println("Folder upload completed.")
	}
}
