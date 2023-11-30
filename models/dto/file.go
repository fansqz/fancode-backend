package dto

// 文件相关类型
type FileDto struct {
	Name    string `json:name`
	Content string `json:content`
	Path    string `json:path`
}
