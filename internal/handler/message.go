package handler

import (
	"errors"
	"fmt"
	"gin-demo/internal/middleware"
	"gin-demo/internal/model"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)

	count, err := service.GetUnreadCount(userID)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 查询未读数失败:", err)
		return
	}
	response.Success(c, gin.H{"total": count})

}

// MarkRead：M5 标记已读（PUT /messages/read）
// handler 只做三件事：① 从 token 拿"我是谁" ② 把 JSON 装进 req ③ 把结果/错误翻译成响应
func MarkRead(c *gin.Context) {
	// ① 身份从 token 来，不从请求体来 —— 请求体里的东西用户能随便改
	userID := middleware.GetUserID(c)

	// ② 解析请求体：JSON → model.MarkReadReq
	//    绑定失败（比如 ids 超过 500 条、或根本不是合法 JSON）就是参数错误
	var req model.MarkReadReq
	//`c.ShouldBindJSON(&req)` 是干嘛的：把前端发来的 JSON 正文，
	// 读出来、对号入座填进你给的结构体里。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	// ③ 交给 service，拿回 (改了几条, 剩几条未读)
	updated, unreadTotal, err := service.MarkRead(userID, req)
	if err != nil {
		// 认错：把"service 的错误"翻译成"给前端的错误码"
		// 三空 → 参数错误 40000；其它（数据库挂了之类）→ 服务器错误 50000
		//`errors.Is(拿到的错误, 想对照的那个错误)`
		//  → 是同一个吗？返回 `true` / `false`。
		if errors.Is(err, service.ErrNoMarkReadScope) {
			response.Fail(c, errcode.ErrInvalidParams)
		} else {
			response.Fail(c, errcode.ErrServer)
			fmt.Println("❌ 标记已读失败:", err)
		}
		return
	}

	// ④ 成功：字段名必须跟 Apifox 对齐（注意和 M1 的 total 不一样）
	response.Success(c, gin.H{
		"updated":      updated,
		"unread_total": unreadTotal,
	})
}
