package dto

type PageInfo struct {
	Size  uint        `json:"size"`
	Total uint        `json:"total"`
	List  interface{} `json:"list"`
}
