package handler

import (
	"errors"
	"fmt"
	"gin-demo/internal/middleware"
	"gin-demo/internal/model"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"
	"strconv"

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

// MarkRead M5 标记已读 PUT /messages/read
func MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req model.MarkReadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	updated, unreadTotal, err := service.MarkRead(userID, req)
	if err != nil {
		// 三空 → 参数错误；其它（数据库挂了之类）→ 服务器错误
		if errors.Is(err, service.ErrNoMarkReadScope) {
			response.Fail(c, errcode.ErrInvalidParams)
		} else {
			response.Fail(c, errcode.ErrServer)
			fmt.Println("❌ 标记已读失败:", err)
		}
		return
	}

	// 字段名必须跟 Apifox 对齐（注意和 M1 的 total 不一样）
	response.Success(c, gin.H{
		"updated":      updated,
		"unread_total": unreadTotal,
	})
}

// ListMessages M2 我的消息列表 GET /messages
func ListMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var q model.MessageListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	// service 会把 q.Page / q.PageSize 补成实际生效的值，所以回显要用 q 里的值
	messages, total, err := service.ListMessages(userID, &q)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 查询消息列表失败:", err)
		return
	}

	// 分页外壳和 P1 的 /posts 共用同一个 response.Page
	response.Success(c, response.Page{
		List:     messages,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

// SendMessage 发送私信 POST /messages
func SendMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req model.SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	m, remind, err := service.SendMessage(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.Fail(c, errcode.ErrNotFound)
		case errors.Is(err, service.ErrSelfMessage):
			response.FailReason(c, errcode.ErrInvalidParams, "不能给自己发私信")
		default:
			response.Fail(c, errcode.ErrServer)
			fmt.Println("❌ 发送私信失败:", err)
		}
		return
	}

	// 私信恒 201，提醒是否成功只看 data.remind
	response.SuccessCreated(c, model.SendMessageData{
		Message: model.MessageView{
			ID:        m.ID,
			Direction: model.MessageDirectionSent,
			Content:   m.Content,
			IsRead:    m.IsRead,
			Reminded:  m.Reminded,
			CreatedAt: m.CreatedAt,
		},
		Remind: remind,
	})
}

// GetConversation M4 与某用户的私信记录 GET /messages/conversations/:peer_id
func GetConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	peerID, err := strconv.ParseUint(c.Param("peer_id"), 10, 64)
	if err != nil || peerID == 0 {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	// limit 的 gte/lte 边界由 binding 拦下
	var q model.ConversationQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	result, err := service.ListConversation(userID, uint(peerID), &q)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Fail(c, errcode.ErrNotFound)
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 查询会话失败:", err)
		return
	}
	response.Success(c, result)
}
