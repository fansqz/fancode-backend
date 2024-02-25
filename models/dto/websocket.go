package dto

import "FanCode/constants"

// WsRequest ws 请求类型
type WsRequest struct {
	Type constants.WsType
	Data interface{}
}

// WsResponse ws响应类型
type WsResponse struct {
	Type constants.WsType
	Data interface{}
}
