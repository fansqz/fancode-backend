package setting

// @Description:oss的配置
type StoreConfig struct {
	Type            string `ini:"type"`
	EndPoint        string `ini:"endPoint"`
	AccessKeyId     string `ini:"accessKeyId"`
	AccessKeySecret string `ini:"accessKeySecret"`
	BucketName      string `ini:"bucketName"`
}
