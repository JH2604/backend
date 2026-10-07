# 失物招领后端

校园失物招领的 HTTP 服务。学号注册登录，发失物或拾物帖，用户之间发私信，图片存到本地目录。数据在 MySQL。

Go 模块名是 `gin-demo`。进程默认听 `8000` 端口。

## 跑起来

本机要有 Go 1.26 和 MySQL。先建好库 `lost_found_db`。启动时会按模型把表补齐，`students` 里的学号和姓名不会自动生成，注册前得自己写入。

PowerShell：

```powershell
$env:JWT_SECRET="换成至少32字节的随机字符串"
$env:DATABASE_DSN="root:123456@tcp(127.0.0.1:3306)/lost_found_db?charset=utf8mb4&parseTime=True&loc=Local"
go run .
```

`JWT_SECRET` 为空或短于 32 字节时，`main` 直接 panic，进程起不来。`DATABASE_DSN` 不设就用 `pkg/config` 里的本地默认串，账号 `root`，密码 `123456`，库 `lost_found_db`。

控制台打出「数据库连接成功，表结构已同步」之后：

- `GET /healthz` 返回 `{"ok":true}`，用来看进程是否还活着
- 业务接口都在 `/api/v1`

改代码想自动重编译，用 [air](https://github.com/air-verse/air)。配置在 `.air.toml`，Windows 下编出来的是 `tmp/main.exe`。

发验证码可以先不配邮箱。没配 SMTP 时，验证码打印在运行这个进程的终端里。

## 技术栈

| 做什么 | 用什么 |
|---|---|
| HTTP | gin v1.12 |
| 数据库 | GORM v1.31，MySQL 驱动 |
| 登录令牌 | `github.com/golang-jwt/jwt/v5`，HMAC-SHA256 |
| 密码 | `golang.org/x/crypto/bcrypt` |
| 入参校验 | `github.com/go-playground/validator/v10` |
| 读图片宽高 | `golang.org/x/image`，认 jpeg / png / webp |
| 发信 | 标准库 `net/smtp`，走 587 端口的 STARTTLS |

`go.sum` 锁依赖校验和，由 Go 工具链生成。

## 分层

一条请求的路径：

```
客户端
  → Gin
  → Recovery / Logger / CORS
  → 需要登录的路由再经过 Auth
  → handler    读参数，把业务错误换成错误码
  → service    规则：谁能删帖、令牌怎么换、验证码多久能再发
  → repository GORM
  → MySQL
```

`internal` 是业务代码，`pkg` 是配置、统一 JSON、错误码、学号和密码格式。handler 不写 SQL，repository 不决定 HTTP 状态码。

成功和失败都走同一套 JSON：

```json
{ "code": 0, "msg": "成功", "data": {} }
```

`code` 为 0 表示成功。失败时 HTTP 状态是业务码整除 100，例如 `40100` 对应 401，`40400` 对应 404。`/healthz` 单独返回，不套这层。

## 目录

```
.
├── main.go                      连库、装邮件、挂中间件、注册路由、静态目录
├── go.mod
├── go.sum
├── .air.toml                    热重载
├── .gitignore
├── docs/deploy.md               本机交叉编译后丢到阿里云、用 systemd 拉起
├── docs/gorm.md                 repository 里 GORM 的写法备忘
├── internal
│   ├── router/router.go         全部 URL，以及三个自定义校验规则
│   ├── handler                  每个接口一个函数
│   ├── service                  业务
│   ├── repository               读写数据库
│   ├── model                    表、请求体、返回体
│   ├── middleware               日志、panic 恢复、跨域、Bearer 校验
│   └── mail                     验证码邮件。开发默认只打印
└── pkg
    ├── config                   读环境变量
    ├── response                 写出 {code,msg,data}
    ├── errcode                  错误码和中文说明
    └── validate                 密码、学号
```

上传文件落在 `uploads/`，由 `r.Static("/uploads", ...)` 直接对外提供。这个目录在 `.gitignore` 里。

## 环境变量

| 变量 | 缺省 | 含义 |
|---|---|---|
| `JWT_SECRET` | 空，启动失败 | 签 access token 的密钥，至少 32 字节 |
| `DATABASE_DSN` | 本地 `root` / `lost_found_db` | MySQL DSN |
| `PORT` | `8000` | 监听端口。写 `8000` 或 `:8000` 都可以 |
| `UPLOAD_DIR` | `./uploads` | 图片落盘目录 |
| `BASE_URL` | `http://localhost:8000` | 读进内存。当前上传接口返回的是 `/uploads/...` 相对路径 |
| `SMTP_HOST` | 空 | 例如 `smtp.qq.com` |
| `SMTP_PORT` | `587` | |
| `SMTP_USER` | 空 | 发件邮箱，也是 SMTP 登录账号 |
| `SMTP_PASS` | 空 | 邮箱授权码 |

`SMTP_HOST`、`SMTP_USER`、`SMTP_PASS` 三个都有值才换成真发信。少一个就继续把验证码打在日志里。

## 账号和令牌

注册时用学号去 `students` 查姓名，查不到返回 `40401`。密码经 bcrypt 后入库。角色只能是 `student` 或 `admin`。格式校验里，学号 `admin` 被单独放行，其余学号必须是 3 到 20 位数字。

登录发出两枚令牌，会话写在 `sessions`：

- access token：JWT，2 小时。之后的请求带 `Authorization: Bearer <token>`。声明里有用户 id（`sub`）、角色、会话 id（`sid`）、类型 `access`。
- refresh token：32 字节随机数的 base64。7 天。只交给 `POST /api/v1/auth/refresh`。

库里存 SHA-256，不存令牌原文。刷新时旧的 refresh 作废，同时记下它的哈希。拿已经轮换掉的旧 refresh 再来一次，会把该用户全部会话吊销。

退出只吊销当前这一次会话。改密码吊销该用户全部会话。access token 还在有效期内但会话已被吊销时，返回 `40103`。JWT 过期返回 `40101`。

## 接口

注册、登录、刷新、发验证码不带 token。其余都要过 `middleware.Auth`。

| 方法 | 路径 | 作用 |
|---|---|---|
| POST | `/api/v1/auth/register` | 注册 |
| POST | `/api/v1/auth/login` | 登录，返回 access、refresh 和用户摘要 |
| POST | `/api/v1/auth/refresh` | 换一对新令牌 |
| POST | `/api/v1/auth/logout` | 退出 |
| GET | `/api/v1/users/me` | 当前用户资料，含发帖数 |
| PATCH | `/api/v1/users/me` | 把头像改成已上传的 `/uploads/...` |
| PUT | `/api/v1/users/me/password` | 改密码 |
| PUT | `/api/v1/users/me/contact` | 校验验证码后绑定邮箱 |
| GET | `/api/v1/users/admins` | 管理员列表 |
| GET | `/api/v1/users/:user_id` | 用户公开资料。`detail` 里的学号、手机、邮箱只在查看者是管理员时出现 |
| POST | `/api/v1/files` | `multipart` 上传。字段 `file`，`usage` 取 `avatar` 或 `post` |
| POST | `/api/v1/auth/post` | 发帖 |
| GET | `/api/v1/posts` | 列表。查询参数见下 |
| GET | `/api/v1/posts/:id` | 详情 |
| PATCH | `/api/v1/posts/:id/status` | 改 `open` / `closed`，仅作者 |
| DELETE | `/api/v1/posts/:id` | 软删除，作者或管理员。正文可带 `reason` |
| GET | `/api/v1/messages` | 我的消息 |
| GET | `/api/v1/messages/unread-count` | 未读条数 |
| PUT | `/api/v1/messages/read` | 标已读。`ids`、`peer_id`、`all` 按这个顺序认 |
| POST | `/api/v1/messages` | 发私信。可选 `remind` |
| GET | `/api/v1/messages/conversations/:peer_id` | 和某人的记录，按 `before_id` 往前翻 |
| POST | `/api/v1/verification-codes` | 发验证码。`channel` 为 `sms` 时直接拒绝 |

帖子列表的 query：`page`（默认 1）、`page_size`（默认 20，最大 50）、`type`（`lost` / `found` / `all`）、`status`（`open` / `closed` / `all`）、`keyword`（标题、正文、地点名）、`mine`、`order`（`asc` / `desc`，默认按 `created_at` 降序）。

帖子 `type` 只有 `lost` 和 `found`。`status` 只有 `open` 和 `closed`。关掉时写入 `closed_at`，重新打开时把该字段写成空。

图片按文件头判断类型，只收 JPEG、PNG、WebP。头像上限 2MB，帖子图上限 5MB。返回相对路径、字节数、宽、高。发帖时的 `images` 每一项都要以 `/uploads/` 开头。

私信的 `remind: true` 会看对方是否打开提醒、有没有手机或邮箱、同一对用户 12 小时内是否已经提醒过、对方当天是否已被提醒满 5 次。投递目前只写日志。私信本身照常入库，提醒失败不影响私信。短信通道没有实现。

## 表

`main.initDB` 对下面六个模型做 `AutoMigrate`：

| 表 | 模型 | 存什么 |
|---|---|---|
| `users` | `User` | 学号、密码哈希、角色、姓名、头像、手机、邮箱、是否接受提醒、主题 |
| `students` | `Student` | 学号主键、姓名。注册时只读 |
| `sessions` | `Session` | 会话 id、两枚令牌的哈希、上一枚 refresh 的哈希、过期时间、平台、IP、UA、吊销时间 |
| `posts` | `Post` | 帖子。`deleted_at` 有值的行，普通查询不会再出现 |
| `messages` | `Message` | 私信。可带帖子 id，有已读和是否提醒成功 |
| `verification_codes` | `VerificationCode` | 验证码。10 分钟有效，同一目标同一场景 60 秒内不重发 |
