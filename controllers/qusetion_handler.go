package controllers

import (
	"FanCode/dao"
	"FanCode/models"
	r "FanCode/result"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
)

func (u *userController) InsertQuestion(ctx *gin.Context) {
	result := r.NewResult(ctx)
	questionNumber := ctx.PostForm("number")
	questionName := ctx.PostForm("name")
	description := ctx.PostForm("description")
	title := ctx.PostForm("title")
	path := ctx.PostForm("path")
	if dao.CheckQuestionNumber(questionNumber) {
		result.SimpleErrorMessage("题目编号已存在")
		return
	}
	question := &models.Question{}
	question.Number = questionNumber
	question.Name = questionName
	question.Description = description
	question.Title = title
	question.Path = path
	//插入
	dao.InsertQuestion(question)

	result.SuccessMessage("题库添加成功")

}

func (u *userController) UpdateQuestion(ctx *gin.Context) {
	result := r.NewResult(ctx)
	questionIDString := ctx.PostForm("id")
	quesetionID, err := strconv.Atoi(questionIDString)
	if err != nil {
		result.SimpleErrorMessage("题目id出错")
	}
	questionNumber := ctx.PostForm("number")
	questionName := ctx.PostForm("name")
	description := ctx.PostForm("description")
	title := ctx.PostForm("title")
	path := ctx.PostForm("path")

	question := &models.Question{}
	question.ID = uint(quesetionID)
	question.Number = questionNumber
	question.Name = questionName
	question.Description = description
	question.Title = title
	question.Path = path

	dao.UpdateQuestion(question)
	result.SuccessData("修改成功")
}

func (u *userController) DeleteQuestion(ctx *gin.Context) {
	result := r.NewResult(ctx)
	userNumber := ctx.PostForm("number")
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")
	if userNumber == "" {
		result.SimpleErrorMessage("用户名不可为空")
		return
	}
	if oldPassword == "" {
		result.SimpleErrorMessage("请输入原始密码")
		return
	}
	//检验用户名
	user, err := dao.GetUserByUserNumber(userNumber)
	if err != nil {
		log.Println(err)
		result.SimpleErrorMessage("系统错误")
		return
	}
	if user == nil {
		result.SimpleErrorMessage("用户不存在")
		return
	}
	//检验旧密码
	if !utils.ComparePwd(oldPassword, user.Password) {
		result.SimpleErrorMessage("原始密码输入错误")
		return
	}
	password, getPwdErr := utils.GetPwd(newPassword)
	if getPwdErr != nil {
		result.SimpleErrorMessage("系统错误")
		log.Println(getPwdErr)
		return
	}
	user.Password = string(password)
	_ = dao.UpdateUser(user)
	token, daoErr := utils.GenerateToken(user)
	if daoErr != nil {
		result.SimpleErrorMessage("登录失败")
		return
	}
	result.SuccessData(token)
}

func (u *userController) GetUserInfo(ctx *gin.Context) {
	result := r.NewResult(ctx)
	user := ctx.Keys["user"].(*models.User)
	result.SuccessData(user)
}
