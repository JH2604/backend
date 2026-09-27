# 后端部署文档

## 一、服务器信息

| 项 | 值 |
|---|---|
| 云服务商 | 阿里云 ECS |
| 公网 IP | `47.99.130.27` |
| 地域 | 华东1（杭州） |
| 规格 | 2 核 2 GB |
| 操作系统 | Ubuntu 22.04 64 位 |
| 后端访问地址 | `http://47.99.130.27:8000` |
| 代码目录（服务器上） | `/root/backend` |
| systemd 服务名 | `lostfound` |

**安全组放行的端口**

| 端口 | 用途 |
|---|---|
| 22 | SSH 登录 |
| 8000 | 后端服务 |

> ⚠️ 3306（MySQL）不要对公网开放。

---

## 二、日常更新流程

**改完代码、准备发布时，按这四步走。**

### ① 本地提交推送

```bash
git add .
git commit -m "feat: xxx"
git push
```

### ② 本地编译 Linux 版

**在项目根目录，PowerShell 里：**

```powershell
$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o backend .
```

| 变量 | 作用 |
|---|---|
| `GOOS=linux` | 编译出能在 Linux 上跑的程序 |
| `GOARCH=amd64` | 目标 CPU 架构 64 位 x86 |
| `CGO_ENABLED=0` | 不依赖 C 库（跨平台编译的必要条件） |

**产物**：项目根目录下多一个 `backend` 文件（**没有 `.exe` 后缀**，它不是 Windows 程序）。

> **为什么不直接在服务器上编译**：服务器只有 2G 内存，编译会内存不足。

### ③ 上传到服务器

```powershell
scp backend root@47.99.130.27:~/backend/
```

**会提示输入服务器密码。**

### ④ 重启服务

```powershell
ssh root@47.99.130.27
```

```bash
systemctl restart lostfound
systemctl status lostfound      # 确认是 active (running)
```

**在本地验证**：

```powershell
curl.exe http://47.99.130.27:8000/api/v1/lost-items
```
