package controllers

import (
	e "FanCode/error"
	"FanCode/models/dto"
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetPageQueryByQuery(ctx *gin.Context) (*dto.PageQuery, *e.Error) {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	var page int
	var pageSize int
	var convertErr error
	page, convertErr = strconv.Atoi(pageStr)
	if convertErr != nil {
		return nil, e.ErrBadRequest
	}
	pageSize, convertErr = strconv.Atoi(pageSizeStr)
	if convertErr != nil {
		return nil, e.ErrBadRequest
	}
	sortProperty := ctx.Query("sortProperty")
	sortRule := ctx.Query("sortRule")
	answer := &dto.PageQuery{
		Page:         page,
		PageSize:     pageSize,
		SortProperty: sortProperty,
		SortRule:     sortRule,
	}
	return answer, nil
}
