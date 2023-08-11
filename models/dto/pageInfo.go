package dto

type PageInfo struct {
	Size  int64       `json:"size"`
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}
