---
title: 失物招领系统 API
language_tabs:
  - shell: Shell
  - http: HTTP
  - javascript: JavaScript
  - ruby: Ruby
  - python: Python
  - php: PHP
  - java: Java
  - go: Go
toc_footers: []
includes: []
search: true
code_clipboard: true
highlight_theme: darkula
headingLevel: 2
generator: "@tarslib/widdershins v4.0.30"

---

# 失物招领系统 API

失物招领系统前后端接口。
- 统一响应结构：{ code, message, data }，code=0 表示成功
- 鉴权：登录后下发 access_token（JWT，2 小时）和 refresh_token（7 天），服务端只存二者的 SHA-256 哈希
- 除白名单（注册、登录、刷新令牌、忘记密码验证码）外，所有接口必须在请求头携带 Authorization: Bearer {access_token}
- 401 业务码：40100 未携带或无效，40101 已过期（调用 /auth/refresh），40103 会话已吊销，40104 refresh_token 无效
- v1.1：移除评论功能；发帖人信息仅管理员可见全部字段；私信支持短信 / 邮件提醒；新增帖子状态修改

Base URLs:

# Authentication

- HTTP Authentication, scheme: bearer<br/>请求头格式：Authorization: Bearer {access_token}
access_token 为 HS256 签名的 JWT，Payload 含 sub、role、sid、jti、type=access、iat、exp，有效期 2 小时。
服务端校验签名和过期时间后，再用 SHA-256(access_token) 查 user_sessions，确认会话未吊销。

# 认证

<a id="opIdregister"></a>

## POST 注册

POST /auth/register

学号+密码+角色注册，后端从实名库查出姓名并返回。（已完成，请与现有实现核对）

> Body 请求参数

```json
{
  "student_id": "202301010101",
  "password": "abc12345",
  "role": "student"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» student_id|body|string| 是 |none|
|» password|body|string| 是 |8~32 位，至少包含字母和数字|
|» role|body|[Role](#schemarole)| 是 |none|

#### 枚举值

|属性|值|
|---|---|
|» role|student|
|» role|admin|

> 返回示例

> 201 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1001,
    "student_id": "202301010101",
    "name": "张三",
    "role": "student"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|注册成功|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误（40000）|None|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|学号不在实名库中（code=40401）|[Error](#schemaerror)|
|409|[Conflict](https://tools.ietf.org/html/rfc7231#section-6.5.8)|学号已注册（code=40900）|[Error](#schemaerror)|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|role|student|
|role|admin|

状态码 **404**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

状态码 **409**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

<a id="opIdlogin"></a>

## POST 登录

POST /auth/login

（已完成，需按新鉴权规范调整：返回 access_token + refresh_token，服务端写入 user_sessions，只存 SHA-256 哈希）

> Body 请求参数

```json
{
  "student_id": "202301010101",
  "password": "abc12345"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» student_id|body|string| 是 |none|
|» password|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 7200,
    "refresh_token": "Qm9vZ2llV29vZ2llLXJhbmRvbS0zMi1ieXRlcw",
    "refresh_expires_in": 604800,
    "user": {
      "id": null,
      "student_id": null,
      "name": null,
      "avatar_url": null,
      "role": null,
      "phone": null,
      "email": null,
      "allow_remind": null,
      "theme": null,
      "post_count": null,
      "created_at": null
    }
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|登录成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|学号或密码错误（code=40102）|[Error](#schemaerror)|
|429|[Too Many Requests](https://tools.ietf.org/html/rfc6585#section-4)|连续失败过多，暂时锁定（code=42900）|[Error](#schemaerror)|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|role|student|
|role|admin|
|theme|light|
|theme|dark|
|theme|system|

状态码 **401**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

状态码 **429**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

<a id="opIdrefreshToken"></a>

## POST 刷新令牌

POST /auth/refresh

鉴权白名单接口，不需要 Authorization 头。
每次刷新都会轮换 refresh_token，旧值立即作废；旧值再次使用视为重放，吊销该用户全部会话。

> Body 请求参数

```json
{
  "refresh_token": "Qm9vZ2llV29vZ2llLXJhbmRvbS0zMi1ieXRlcw"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» refresh_token|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 7200,
    "refresh_token": "Qm9vZ2llV29vZ2llLXJhbmRvbS0zMi1ieXRlcw",
    "refresh_expires_in": 604800
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|刷新成功，返回新的令牌对|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|refresh_token 无效、过期或被重放（code=40104）|[Error](#schemaerror)|

### 返回数据结构

状态码 **401**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

<a id="opIdlogout"></a>

## POST 退出登录

POST /auth/logout

吊销当前会话（user_sessions.revoked_at），对应的 access_token 和 refresh_token 立即失效。

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|None|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

# 用户

<a id="opIdgetMe"></a>

## GET 获取当前用户信息

GET /users/me

设置页、首页头像使用。不返回任何密码字段。

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1001,
    "student_id": "202301010101",
    "name": "张三",
    "avatar_url": "https://cdn.example.com/avatar/1001.png",
    "role": "student",
    "phone": "13812345678",
    "email": "zhangsan@example.com",
    "allow_remind": true,
    "theme": "light",
    "post_count": 5,
    "created_at": "2019-08-24T14:15:22Z"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|role|student|
|role|admin|
|theme|light|
|theme|dark|
|theme|system|

<a id="opIdupdateMe"></a>

## PATCH 修改个人资料（头像 / 主题 / 私信提醒开关）

PATCH /users/me

只传需要修改的字段；name、student_id、role 不可修改；手机号和邮箱走 /users/me/contact。

> Body 请求参数

```json
{
  "avatar_url": "https://cdn.example.com/avatar/1001_v2.png",
  "theme": "light",
  "allow_remind": true
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» avatar_url|body|string| 否 |先调用上传接口（usage=avatar）获取|
|» theme|body|[Theme](#schematheme)| 否 |none|
|» allow_remind|body|boolean| 否 |是否接收私信的短信 / 邮件提醒|

#### 枚举值

|属性|值|
|---|---|
|» theme|light|
|» theme|dark|
|» theme|system|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1001,
    "student_id": "202301010101",
    "name": "张三",
    "avatar_url": "https://cdn.example.com/avatar/1001.png",
    "role": "student",
    "phone": "13812345678",
    "email": "zhangsan@example.com",
    "allow_remind": true,
    "theme": "light",
    "post_count": 5,
    "created_at": "2019-08-24T14:15:22Z"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功，返回修改后的用户信息|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误（40000）|None|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|role|student|
|role|admin|
|theme|light|
|theme|dark|
|theme|system|

<a id="opIdchangePassword"></a>

## PUT 修改密码

PUT /users/me/password

成功后吊销该用户全部会话（所有设备下线），前端清除令牌并跳转登录页。

> Body 请求参数

```json
{
  "old_password": "abc12345",
  "new_password": "newPass678"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» old_password|body|string| 是 |none|
|» new_password|body|string| 是 |none|

> 返回示例

> 400 Response

```json
{
  "code": 40300,
  "message": "无权删除该帖子",
  "data": null
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|None|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|原密码错误（40002）或新密码格式不合法（40000）|[Error](#schemaerror)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

<a id="opIdsendVerificationCode"></a>

## POST 发送验证码（短信 / 邮件）

POST /verification-codes

channel=sms 发短信，channel=email 发邮件。
scene=bind_contact 绑定 / 修改手机号或邮箱（需携带 access_token）；scene=reset_password 忘记密码（预留，白名单）。

> Body 请求参数

```json
{
  "channel": "sms",
  "target": "13812345678",
  "scene": "bind_contact"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» channel|body|[Channel](#schemachannel)| 是 |sms 手机短信 / email 邮件|
|» target|body|string| 是 |手机号或邮箱|
|» scene|body|string| 是 |none|

#### 枚举值

|属性|值|
|---|---|
|» channel|sms|
|» channel|email|
|» scene|bind_contact|
|» scene|reset_password|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "expires_in": 300,
    "resend_after": 60
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|发送成功|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误（40000）|None|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|
|409|[Conflict](https://tools.ietf.org/html/rfc7231#section-6.5.8)|已被其他账号绑定（40900）|[Error](#schemaerror)|
|429|[Too Many Requests](https://tools.ietf.org/html/rfc6585#section-4)|发送过于频繁（42900）|[Error](#schemaerror)|

### 返回数据结构

状态码 **409**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

状态码 **429**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

<a id="opIdbindContact"></a>

## PUT 绑定 / 修改手机号或邮箱

PUT /users/me/contact

> Body 请求参数

```json
{
  "channel": "sms",
  "target": "zhangsan@example.com",
  "code": "123456"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» channel|body|[Channel](#schemachannel)| 是 |sms 手机短信 / email 邮件|
|» target|body|string| 是 |必须与发送验证码时一致|
|» code|body|string| 是 |none|

#### 枚举值

|属性|值|
|---|---|
|» channel|sms|
|» channel|email|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "phone": "13812345678",
    "email": "zhangsan@example.com"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|绑定成功|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|验证码错误或已过期（40001）|[Error](#schemaerror)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|
|409|[Conflict](https://tools.ietf.org/html/rfc7231#section-6.5.8)|已被其他账号绑定（40900）|[Error](#schemaerror)|

### 返回数据结构

状态码 **400**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

状态码 **409**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

<a id="opIdlistAdmins"></a>

## GET 管理员列表（联系管理员）

GET /users/admins

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "王老师",
      "avatar_url": "string",
      "role": "admin",
      "email": "admin@example.com",
      "can_message": true
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

### 返回数据结构

<a id="opIdgetUser"></a>

## GET 查看发帖人信息

GET /users/{user_id}

按查看者角色返回不同字段（后端控制）：
- 普通用户：只返回姓名、头像、角色、发帖数、能否私信 / 提醒，detail 为 null
- 管理员：额外返回 detail（学号、手机号、邮箱、提醒开关、注册时间、最近登录时间）

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|user_id|path|integer(int64)| 是 |none|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1002,
    "name": "李四",
    "avatar_url": "string",
    "role": "student",
    "post_count": 3,
    "can_message": true,
    "can_remind": true,
    "detail": {
      "student_id": "202301010202",
      "phone": "13900001111",
      "email": "lisi@example.com",
      "allow_remind": true,
      "created_at": "2019-08-24T14:15:22Z",
      "last_login_at": "2019-08-24T14:15:22Z"
    }
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|资源不存在（40400）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|role|student|
|role|admin|

# 文件

<a id="opIduploadFile"></a>

## POST 上传图片

POST /files

usage=avatar 头像（≤2MB）；usage=post 帖子图片（≤5MB）。支持 jpg/png/webp。

> Body 请求参数

```yaml
file: ""
usage: ""

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» file|body|string(binary)| 是 |none|
|» usage|body|string| 是 |none|

#### 枚举值

|属性|值|
|---|---|
|» usage|avatar|
|» usage|post|

> 返回示例

> 201 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "url": "https://cdn.example.com/post/2026/10/01/abc.webp",
    "width": 1080,
    "height": 1440,
    "size": 345678
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|上传成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|
|413|[Payload Too Large](https://tools.ietf.org/html/rfc7231#section-6.5.11)|文件过大（41300）|[Error](#schemaerror)|
|415|[Unsupported Media Type](https://tools.ietf.org/html/rfc7231#section-6.5.13)|文件类型不支持（41500）|[Error](#schemaerror)|

### 返回数据结构

状态码 **413**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

状态码 **415**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» code|integer|false|none||none|
|» message|string|false|none||none|
|» data|object¦null|false|none||none|

# 帖子

<a id="opIdlistPosts"></a>

## GET 帖子列表（首页 / 分类 / 搜索 / 我的帖子）

GET /posts

- 首页全部：type=all
- 首页失物 / 招领：type=lost / type=found
- 搜索：keyword=xxx（可叠加 type）
- 我发布的帖子：mine=true，可叠加 status、order

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|type|query|string| 否 |none|
|keyword|query|string| 否 |模糊匹配标题、正文、地点名称|
|mine|query|boolean| 否 |true 时只返回当前用户的帖子|
|status|query|string| 否 |none|
|order|query|string| 否 |按发布时间排序|
|page|query|integer| 否 |none|
|page_size|query|integer| 否 |none|

#### 枚举值

|属性|值|
|---|---|
|type|all|
|type|lost|
|type|found|
|status|all|
|status|open|
|status|closed|
|order|desc|
|order|asc|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 128,
    "page": 1,
    "page_size": 20,
    "list": [
      {}
    ]
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|type|lost|
|type|found|
|status|open|
|status|closed|
|role|student|
|role|admin|

<a id="opIdcreatePost"></a>

## POST 发布帖子

POST /posts

> Body 请求参数

```json
{
  "type": "lost",
  "title": "图书馆捡到一副耳机",
  "content": "白色无线耳机，放在三楼服务台了。",
  "images": [
    "https://cdn.example.com/post/2026/10/01/abc.webp"
  ],
  "location": {
    "name": "图书馆三楼",
    "latitude": 30.2301,
    "longitude": 120.042
  },
  "event_time": "2026-10-01T09:30:00+08:00"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» type|body|[PostType](#schemaposttype)| 是 |lost 失物 / found 招领|
|» title|body|string| 是 |none|
|» content|body|string| 是 |none|
|» images|body|[string]| 否 |none|
|» location|body|[Location](#schemalocation)| 是 |none|
|»» name|body|string| 是 |none|
|»» latitude|body|number¦null| 否 |none|
|»» longitude|body|number¦null| 否 |none|
|» event_time|body|string(date-time)| 否 |丢失 / 拾到时间|

#### 枚举值

|属性|值|
|---|---|
|» type|lost|
|» type|found|

> 返回示例

> 201 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 501,
    "type": "lost",
    "title": "丢失一张校园卡",
    "content": "string",
    "images": [
      "string"
    ],
    "location": {
      "name": "图书馆三楼",
      "latitude": 30.2301,
      "longitude": 120.042
    },
    "event_time": "2019-08-24T14:15:22Z",
    "status": "open",
    "author": {
      "id": 1002,
      "name": "李四",
      "avatar_url": "https://cdn.example.com/avatar/1002.png",
      "role": "student"
    },
    "is_mine": true,
    "can_delete": true,
    "can_change_status": true,
    "created_at": "2019-08-24T14:15:22Z",
    "closed_at": "2019-08-24T14:15:22Z"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|发布成功|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误（40000）|None|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|type|lost|
|type|found|
|status|open|
|status|closed|
|role|student|
|role|admin|

<a id="opIdgetPost"></a>

## GET 帖子详情

GET /posts/{post_id}

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|post_id|path|integer(int64)| 是 |none|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 501,
    "type": "lost",
    "title": "丢失一张校园卡",
    "content": "string",
    "images": [
      "string"
    ],
    "location": {
      "name": "图书馆三楼",
      "latitude": 30.2301,
      "longitude": 120.042
    },
    "event_time": "2019-08-24T14:15:22Z",
    "status": "open",
    "author": {
      "id": 1002,
      "name": "李四",
      "avatar_url": "https://cdn.example.com/avatar/1002.png",
      "role": "student"
    },
    "is_mine": true,
    "can_delete": true,
    "can_change_status": true,
    "created_at": "2019-08-24T14:15:22Z",
    "closed_at": "2019-08-24T14:15:22Z"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|资源不存在（40400）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|type|lost|
|type|found|
|status|open|
|status|closed|
|role|student|
|role|admin|

<a id="opIddeletePost"></a>

## DELETE 删除帖子

DELETE /posts/{post_id}

发帖人本人或管理员可删除。管理员删除他人帖子时可附带原因，后端以站内私信通知发帖人。

> Body 请求参数

```json
{
  "reason": "内容与失物招领无关"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|post_id|path|integer(int64)| 是 |none|
|body|body|object| 否 |none|
|» reason|body|string| 否 |none|

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|None|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|无权限（40300）|None|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|资源不存在（40400）|None|

<a id="opIdupdatePostStatus"></a>

## PATCH 修改帖子状态（已找到 / 已认领）

PATCH /posts/{post_id}/status

仅发帖人本人可操作，在"我发布的帖子"页面使用。
closed：失物帖标记为已找到、招领帖标记为已认领；open：撤回为进行中。状态相同时直接返回成功。

> Body 请求参数

```json
{
  "status": "open"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|post_id|path|integer(int64)| 是 |none|
|body|body|object| 是 |none|
|» status|body|[PostStatus](#schemapoststatus)| 是 |open 进行中（未找到 / 待认领）；closed 已完结（已找到 / 已认领）|

#### 枚举值

|属性|值|
|---|---|
|» status|open|
|» status|closed|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 501,
    "status": "open",
    "closed_at": "2019-08-24T14:15:22Z"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误（40000）|None|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|无权限（40300）|None|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|资源不存在（40400）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|status|open|
|status|closed|

# 消息

<a id="opIdgetUnreadCount"></a>

## GET 未读私信数（首页头像小红点）

GET /messages/unread-count

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 2
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

### 返回数据结构

<a id="opIdlistMessages"></a>

## GET 我的消息（发出 + 收到）

GET /messages

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|box|query|string| 否 |none|
|is_read|query|boolean| 否 |不传返回全部|
|page|query|integer| 否 |none|
|page_size|query|integer| 否 |none|

#### 枚举值

|属性|值|
|---|---|
|box|all|
|box|sent|
|box|received|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 128,
    "page": 1,
    "page_size": 20,
    "list": [
      {}
    ]
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|direction|sent|
|direction|received|
|role|student|
|role|admin|

<a id="opIdsendMessage"></a>

## POST 发送私信（可选短信 / 邮件提醒）

POST /messages

remind=true 时系统通过对方绑定的手机号（优先）或邮箱发送提醒，提醒内容不含私信正文，也不向发送人透露对方联系方式。
频率限制：同一发送人对同一接收人 12 小时 1 次；每个接收人每天最多 5 次。
提醒失败不影响私信本身，私信始终返回 201，结果见 data.remind。

> Body 请求参数

```json
{
  "receiver_id": 1002,
  "content": "你好，耳机是我的，请问在哪里拿？",
  "post_id": 0,
  "remind": false
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» receiver_id|body|integer(int64)| 是 |none|
|» content|body|string| 是 |none|
|» post_id|body|integer(int64)¦null| 否 |从帖子详情页发起私信时携带|
|» remind|body|boolean| 否 |是否通过短信 / 邮件提醒对方查看私信|

> 返回示例

> 201 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": {
      "id": 7011,
      "direction": "sent",
      "content": "string",
      "is_read": true,
      "reminded": true,
      "created_at": "2019-08-24T14:15:22Z"
    },
    "remind": {
      "status": "sent",
      "channel": "sms",
      "reason": "not_requested"
    }
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|发送成功|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误（40000）|None|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|资源不存在（40400）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|direction|sent|
|direction|received|
|status|sent|
|status|skipped|
|status|failed|
|channel|sms|
|channel|email|
|reason|not_requested|
|reason|no_contact|
|reason|disabled|
|reason|rate_limited|
|reason|provider_error|

<a id="opIdgetConversation"></a>

## GET 与某用户的私信记录

GET /messages/conversations/{peer_id}

游标分页，list 按时间正序。

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|peer_id|path|integer(int64)| 是 |none|
|before_id|query|integer(int64)| 否 |加载这条消息之前的记录，不传表示最新|
|limit|query|integer| 否 |none|
|mark_read|query|boolean| 否 |是否把对方发给我的消息标记为已读|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "peer": {
      "id": 1002,
      "name": "李四",
      "avatar_url": "https://cdn.example.com/avatar/1002.png",
      "role": "student"
    },
    "can_remind": true,
    "list": [
      {
        "id": 7011,
        "direction": "[",
        "content": "string",
        "is_read": true,
        "reminded": true,
        "created_at": "2019-08-24T14:15:22Z"
      }
    ],
    "has_more": true
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|资源不存在（40400）|None|

### 返回数据结构

#### 枚举值

|属性|值|
|---|---|
|role|student|
|role|admin|
|direction|sent|
|direction|received|

<a id="opIdmarkRead"></a>

## PUT 标记已读

PUT /messages/read

范围优先级 ids > peer_id > all。

> Body 请求参数

```json
{
  "ids": [
    7001,
    7003
  ],
  "peer_id": 0,
  "all": true
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» ids|body|[integer]| 否 |none|
|» peer_id|body|integer(int64)| 否 |none|
|» all|body|boolean| 否 |true 表示全部已读|

> 返回示例

> 200 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "updated": 3,
    "unread_total": 0
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|成功|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|参数错误（40000）|None|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|未携带或无效（40100）/ 已过期（40101）/ 会话已吊销（40103）|None|

### 返回数据结构

# 数据模型

<h2 id="tocS_Result">Result</h2>

<a id="schemaresult"></a>
<a id="schema_Result"></a>
<a id="tocSresult"></a>
<a id="tocsresult"></a>

```json
{
  "code": 0,
  "message": "success"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|code|integer|true|none||none|
|message|string|true|none||none|

<h2 id="tocS_Error">Error</h2>

<a id="schemaerror"></a>
<a id="schema_Error"></a>
<a id="tocSerror"></a>
<a id="tocserror"></a>

```json
{
  "code": 40300,
  "message": "无权删除该帖子",
  "data": null
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|code|integer|false|none||none|
|message|string|false|none||none|
|data|object¦null|false|none||none|

<h2 id="tocS_PageMeta">PageMeta</h2>

<a id="schemapagemeta"></a>
<a id="schema_PageMeta"></a>
<a id="tocSpagemeta"></a>
<a id="tocspagemeta"></a>

```json
{
  "total": 128,
  "page": 1,
  "page_size": 20
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|total|integer|false|none||none|
|page|integer|false|none||none|
|page_size|integer|false|none||none|

<h2 id="tocS_TokenPair">TokenPair</h2>

<a id="schematokenpair"></a>
<a id="schema_TokenPair"></a>
<a id="tocStokenpair"></a>
<a id="tocstokenpair"></a>

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 7200,
  "refresh_token": "Qm9vZ2llV29vZ2llLXJhbmRvbS0zMi1ieXRlcw",
  "refresh_expires_in": 604800
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|access_token|string|false|none||JWT，放入 Authorization 头|
|token_type|string|false|none||none|
|expires_in|integer|false|none||access_token 有效期（秒）|
|refresh_token|string|false|none||仅用于 /auth/refresh|
|refresh_expires_in|integer|false|none||refresh_token 有效期（秒）|

<h2 id="tocS_Role">Role</h2>

<a id="schemarole"></a>
<a id="schema_Role"></a>
<a id="tocSrole"></a>
<a id="tocsrole"></a>

```json
"student"

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|string|false|none||none|

#### 枚举值

|属性|值|
|---|---|
|*anonymous*|student|
|*anonymous*|admin|

<h2 id="tocS_Theme">Theme</h2>

<a id="schematheme"></a>
<a id="schema_Theme"></a>
<a id="tocStheme"></a>
<a id="tocstheme"></a>

```json
"light"

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|string|false|none||none|

#### 枚举值

|属性|值|
|---|---|
|*anonymous*|light|
|*anonymous*|dark|
|*anonymous*|system|

<h2 id="tocS_Channel">Channel</h2>

<a id="schemachannel"></a>
<a id="schema_Channel"></a>
<a id="tocSchannel"></a>
<a id="tocschannel"></a>

```json
"sms"

```

sms 手机短信 / email 邮件

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|string|false|none||sms 手机短信 / email 邮件|

#### 枚举值

|属性|值|
|---|---|
|*anonymous*|sms|
|*anonymous*|email|

<h2 id="tocS_PostType">PostType</h2>

<a id="schemaposttype"></a>
<a id="schema_PostType"></a>
<a id="tocSposttype"></a>
<a id="tocsposttype"></a>

```json
"lost"

```

lost 失物 / found 招领

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|string|false|none||lost 失物 / found 招领|

#### 枚举值

|属性|值|
|---|---|
|*anonymous*|lost|
|*anonymous*|found|

<h2 id="tocS_PostStatus">PostStatus</h2>

<a id="schemapoststatus"></a>
<a id="schema_PostStatus"></a>
<a id="tocSpoststatus"></a>
<a id="tocspoststatus"></a>

```json
"open"

```

open 进行中（未找到 / 待认领）；closed 已完结（已找到 / 已认领）

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|string|false|none||open 进行中（未找到 / 待认领）；closed 已完结（已找到 / 已认领）|

#### 枚举值

|属性|值|
|---|---|
|*anonymous*|open|
|*anonymous*|closed|

<h2 id="tocS_UserBrief">UserBrief</h2>

<a id="schemauserbrief"></a>
<a id="schema_UserBrief"></a>
<a id="tocSuserbrief"></a>
<a id="tocsuserbrief"></a>

```json
{
  "id": 1002,
  "name": "李四",
  "avatar_url": "https://cdn.example.com/avatar/1002.png",
  "role": "student"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|id|integer(int64)|false|none||none|
|name|string|false|none||none|
|avatar_url|string|false|none||none|
|role|[Role](#schemarole)|false|none||none|

<h2 id="tocS_UserProfile">UserProfile</h2>

<a id="schemauserprofile"></a>
<a id="schema_UserProfile"></a>
<a id="tocSuserprofile"></a>
<a id="tocsuserprofile"></a>

```json
{
  "id": 1001,
  "student_id": "202301010101",
  "name": "张三",
  "avatar_url": "https://cdn.example.com/avatar/1001.png",
  "role": "student",
  "phone": "13812345678",
  "email": "zhangsan@example.com",
  "allow_remind": true,
  "theme": "light",
  "post_count": 5,
  "created_at": "2019-08-24T14:15:22Z"
}

```

当前登录用户的完整信息

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|id|integer(int64)|false|none||none|
|student_id|string|false|none||none|
|name|string|false|none||none|
|avatar_url|string|false|none||none|
|role|[Role](#schemarole)|false|none||none|
|phone|string¦null|false|none||none|
|email|string¦null|false|none||none|
|allow_remind|boolean|false|none||是否接收私信的短信 / 邮件提醒|
|theme|[Theme](#schematheme)|false|none||none|
|post_count|integer|false|none||none|
|created_at|string(date-time)|false|none||none|

<h2 id="tocS_UserPublic">UserPublic</h2>

<a id="schemauserpublic"></a>
<a id="schema_UserPublic"></a>
<a id="tocSuserpublic"></a>
<a id="tocsuserpublic"></a>

```json
{
  "id": 1002,
  "name": "李四",
  "avatar_url": "string",
  "role": "student",
  "post_count": 3,
  "can_message": true,
  "can_remind": true,
  "detail": {
    "student_id": "202301010202",
    "phone": "13900001111",
    "email": "lisi@example.com",
    "allow_remind": true,
    "created_at": "2019-08-24T14:15:22Z",
    "last_login_at": "2019-08-24T14:15:22Z"
  }
}

```

查看他人信息。detail 仅管理员查看时有值。

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|id|integer(int64)|false|none||none|
|name|string|false|none||none|
|avatar_url|string|false|none||none|
|role|[Role](#schemarole)|false|none||none|
|post_count|integer|false|none||none|
|can_message|boolean|false|none||能否私信对方|
|can_remind|boolean|false|none||发私信时能否勾选短信 / 邮件提醒（不暴露具体联系方式）|
|detail|object¦null|false|none||仅管理员可见，普通用户查看时为 null|
|» student_id|string|false|none||none|
|» phone|string¦null|false|none||none|
|» email|string¦null|false|none||none|
|» allow_remind|boolean|false|none||none|
|» created_at|string(date-time)|false|none||none|
|» last_login_at|string(date-time)¦null|false|none||none|

<h2 id="tocS_AdminContact">AdminContact</h2>

<a id="schemaadmincontact"></a>
<a id="schema_AdminContact"></a>
<a id="tocSadmincontact"></a>
<a id="tocsadmincontact"></a>

```json
{
  "id": 1,
  "name": "王老师",
  "avatar_url": "string",
  "role": "admin",
  "email": "admin@example.com",
  "can_message": true
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|id|integer(int64)|false|none||none|
|name|string|false|none||none|
|avatar_url|string|false|none||none|
|role|string|false|none||none|
|email|string¦null|false|none||管理员公开的工作邮箱|
|can_message|boolean|false|none||none|

<h2 id="tocS_Location">Location</h2>

<a id="schemalocation"></a>
<a id="schema_Location"></a>
<a id="tocSlocation"></a>
<a id="tocslocation"></a>

```json
{
  "name": "图书馆三楼",
  "latitude": 30.2301,
  "longitude": 120.042
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|name|string|true|none||none|
|latitude|number¦null|false|none||none|
|longitude|number¦null|false|none||none|

<h2 id="tocS_PostListItem">PostListItem</h2>

<a id="schemapostlistitem"></a>
<a id="schema_PostListItem"></a>
<a id="tocSpostlistitem"></a>
<a id="tocspostlistitem"></a>

```json
{
  "id": 501,
  "type": "lost",
  "title": "丢失一张校园卡",
  "content_preview": "string",
  "cover_url": "string",
  "image_count": 2,
  "location": {
    "name": "图书馆三楼",
    "latitude": 30.2301,
    "longitude": 120.042
  },
  "status": "open",
  "author": {
    "id": 1002,
    "name": "李四",
    "avatar_url": "https://cdn.example.com/avatar/1002.png",
    "role": "student"
  },
  "created_at": "2019-08-24T14:15:22Z",
  "closed_at": "2019-08-24T14:15:22Z"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|id|integer(int64)|false|none||none|
|type|[PostType](#schemaposttype)|false|none||lost 失物 / found 招领|
|title|string|false|none||none|
|content_preview|string|false|none||正文前 60 字|
|cover_url|string¦null|false|none||none|
|image_count|integer|false|none||none|
|location|[Location](#schemalocation)|false|none||none|
|status|[PostStatus](#schemapoststatus)|false|none||open 进行中（未找到 / 待认领）；closed 已完结（已找到 / 已认领）|
|author|[UserBrief](#schemauserbrief)|false|none||none|
|created_at|string(date-time)|false|none||none|
|closed_at|string(date-time)¦null|false|none||none|

<h2 id="tocS_PostDetail">PostDetail</h2>

<a id="schemapostdetail"></a>
<a id="schema_PostDetail"></a>
<a id="tocSpostdetail"></a>
<a id="tocspostdetail"></a>

```json
{
  "id": 501,
  "type": "lost",
  "title": "丢失一张校园卡",
  "content": "string",
  "images": [
    "string"
  ],
  "location": {
    "name": "图书馆三楼",
    "latitude": 30.2301,
    "longitude": 120.042
  },
  "event_time": "2019-08-24T14:15:22Z",
  "status": "open",
  "author": {
    "id": 1002,
    "name": "李四",
    "avatar_url": "https://cdn.example.com/avatar/1002.png",
    "role": "student"
  },
  "is_mine": true,
  "can_delete": true,
  "can_change_status": true,
  "created_at": "2019-08-24T14:15:22Z",
  "closed_at": "2019-08-24T14:15:22Z"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|id|integer(int64)|false|none||none|
|type|[PostType](#schemaposttype)|false|none||lost 失物 / found 招领|
|title|string|false|none||none|
|content|string|false|none||none|
|images|[string]|false|none||none|
|location|[Location](#schemalocation)|false|none||none|
|event_time|string(date-time)¦null|false|none||none|
|status|[PostStatus](#schemapoststatus)|false|none||open 进行中（未找到 / 待认领）；closed 已完结（已找到 / 已认领）|
|author|[UserBrief](#schemauserbrief)|false|none||none|
|is_mine|boolean|false|none||是否是当前用户发布的|
|can_delete|boolean|false|none||本人或管理员为 true|
|can_change_status|boolean|false|none||仅本人为 true|
|created_at|string(date-time)|false|none||none|
|closed_at|string(date-time)¦null|false|none||none|

<h2 id="tocS_Message">Message</h2>

<a id="schemamessage"></a>
<a id="schema_Message"></a>
<a id="tocSmessage"></a>
<a id="tocsmessage"></a>

```json
{
  "id": 7010,
  "direction": "sent",
  "peer": {
    "id": 1002,
    "name": "李四",
    "avatar_url": "https://cdn.example.com/avatar/1002.png",
    "role": "student"
  },
  "content": "string",
  "post": {
    "id": 0,
    "title": "string"
  },
  "is_read": true,
  "reminded": true,
  "created_at": "2019-08-24T14:15:22Z"
}

```

我的消息列表项

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|id|integer(int64)|false|none||none|
|direction|string|false|none||none|
|peer|[UserBrief](#schemauserbrief)|false|none||none|
|content|string|false|none||none|
|post|object¦null|false|none||从帖子详情发起私信时的关联帖子|
|» id|integer(int64)|false|none||none|
|» title|string|false|none||none|
|is_read|boolean|false|none||收到的消息：我是否已读；发出的消息：对方是否已读|
|reminded|boolean|false|none||是否触发过短信 / 邮件提醒|
|created_at|string(date-time)|false|none||none|

#### 枚举值

|属性|值|
|---|---|
|direction|sent|
|direction|received|

<h2 id="tocS_ChatMessage">ChatMessage</h2>

<a id="schemachatmessage"></a>
<a id="schema_ChatMessage"></a>
<a id="tocSchatmessage"></a>
<a id="tocschatmessage"></a>

```json
{
  "id": 7011,
  "direction": "sent",
  "content": "string",
  "is_read": true,
  "reminded": true,
  "created_at": "2019-08-24T14:15:22Z"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|id|integer(int64)|false|none||none|
|direction|string|false|none||none|
|content|string|false|none||none|
|is_read|boolean|false|none||none|
|reminded|boolean|false|none||none|
|created_at|string(date-time)|false|none||none|

#### 枚举值

|属性|值|
|---|---|
|direction|sent|
|direction|received|

<h2 id="tocS_RemindResult">RemindResult</h2>

<a id="schemaremindresult"></a>
<a id="schema_RemindResult"></a>
<a id="tocSremindresult"></a>
<a id="tocsremindresult"></a>

```json
{
  "status": "sent",
  "channel": "sms",
  "reason": "not_requested"
}

```

短信 / 邮件提醒结果

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|status|string|false|none||none|
|channel|string¦null|false|none||none|
|reason|string¦null|false|none||not_requested 未要求提醒；no_contact 对方未绑定手机号和邮箱；disabled 对方关闭了提醒；<br />rate_limited 触发频率限制；provider_error 服务商发送失败|

#### 枚举值

|属性|值|
|---|---|
|status|sent|
|status|skipped|
|status|failed|
|channel|sms|
|channel|email|
|reason|not_requested|
|reason|no_contact|
|reason|disabled|
|reason|rate_limited|
|reason|provider_error|

<h2 id="tocS_ResultEmpty">ResultEmpty</h2>

<a id="schemaresultempty"></a>
<a id="schema_ResultEmpty"></a>
<a id="tocSresultempty"></a>
<a id="tocsresultempty"></a>

```json
{
  "code": 0,
  "message": "success",
  "data": null
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|code|integer|false|none||none|
|message|string|false|none||none|
|data|object¦null|false|none||none|

