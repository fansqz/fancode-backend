package store

import (
	"FanCode/setting"
	"context"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
	"net/http"
	"net/url"
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
	c := cos.NewClient(b, &http.Client{
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
		client: c,
	}
}

func (cos *COS) SaveFile(storePath string, file *strings.Reader) error {
	_, err := cos.client.Object.Put(context.Background(), storePath, file, nil)
	return err
}

// storePaht:对象存储的路劲，loadPath:本地路径
func (cos *COS) LoadFile(storePath, localPath string) error {
	_, err := cos.client.Object.GetToFile(context.Background(), storePath, localPath, nil)
	return err
}
