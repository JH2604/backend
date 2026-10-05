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

// ListMessages：M2 我的消息列表（GET /messages）
// handler 只做两件事：① 把 query 参数装进 q ② 把 (列表, 总数) 塞进统一分页响应
func ListMessages(c *gin.Context) {
	// ① 身份从 token 来
	userID := middleware.GetUserID(c)

	// ② 解析 query 参数：?box=&is_read=&page=&page_size=
	//    结构体 binding 标签里的 oneof / gte / lte 在这里生效，传了非法值直接 40000
	var q model.MessageListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	// ③ 交给 service。它会把 q.Page / q.PageSize 补成"实际生效的值"，
	//    所以下面回显的必须是 q 里的值（前端才知道真实生效的是第几页几条）
	messages, total, err := service.ListMessages(userID, &q)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 查询消息列表失败:", err)
		return
	}

	// ④ 分页外壳和 P1 的 /posts 共用同一个 response.Page
	response.Success(c, response.Page{
		List:     messages,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

// SendMessage：M3 发送私信（POST /messages）
// handler 只做三件事：① 从 token 拿"我是谁" ② 把 JSON 装进 req ③ 把结果翻译成响应
func SendMessage(c *gin.Context) {
	// ① 身份从 token 来 —— 请求体里没有 sender_id，前端伪造不了
	userID := middleware.GetUserID(c)

	// ② 解析请求体：receiver_id 必填、content 必填且不能是纯空白，全靠 binding 标签管
	var req model.SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	// ③ 交给 service，拿回 (落库的私信, 提醒结果)
	m, remind, err := service.SendMessage(userID, req)
	if err != nil {
		// 认错：把 service 的业务错误翻译成给前端的错误码
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			// 收信人不存在
			response.Fail(c, errcode.ErrNotFound)
		case errors.Is(err, service.ErrSelfMessage):
			response.FailReason(c, errcode.ErrInvalidParams, "不能给自己发私信")
		default:
			response.Fail(c, errcode.ErrServer)
			fmt.Println("❌ 发送私信失败:", err)
		}
		return
	}

	// ④ 成功：私信永远 201 —— 提醒成没成都不影响这一条，只看 data.remind
	response.SuccessCreated(c, model.SendMessageData{
		Message: model.MessageView{
			ID:        m.ID,
			Direction: model.MessageDirectionSent, // 刚发出去的，方向恒为 sent
			Content:   m.Content,
			IsRead:    m.IsRead, // 刚发出来必然是 false，对方还没看
			Reminded:  m.Reminded,
			CreatedAt: m.CreatedAt,
		},
		Remind: remind,
	})
}

// GetConversation：M4 与某用户的私信记录（GET /messages/conversations/:peer_id）
func GetConversation(c *gin.Context) {
	// ① 身份从 token 来
	userID := middleware.GetUserID(c)

	// ② 路径参数：字符串 → 数字，转不动或者传了 0 都算参数错误
	peerID, err := strconv.ParseUint(c.Param("peer_id"), 10, 64)
	if err != nil || peerID == 0 {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	// ③ 解析 query 参数：?before_id=&limit=&mark_read=
	//    gte / lte 在这里生效，limit 传 999 直接 40000，不用进 service 再拦
	var q model.ConversationQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	// ④ 交给 service。返回值直接就是契约里的 data（peer / can_remind / list / has_more）
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
