# GORM 常用写法速查（repository 层）

> 这份文档只讲"每一步做完是什么效果"，不涉及框架内部实现。
> 写 repository 层时遇到拿不准的写法，先来这里对一眼。

---

## 一、先记住一个比喻：db 是一张点菜单

`db` 不是"一条 SQL"，它更像**一张点菜单**。

你一路点下去（要哪几列、什么条件、怎么排、第几页），菜单越写越满，**但后厨什么都没干**；
直到碰上 `Find`、`Count` 这种"我要结果"的动作，才把菜单递进后厨，做完把菜端回来。

菜单上的每一步都返回**同一张菜单**，所以可以一直用点号接下去。

方法分三类：

| 类别 | 特征 | 常见方法 |
|---|---|---|
| 拼条件（不查库） | 往菜单上加话 | `Model` `Select` `Where` `Order` `Limit` `Offset` `Session` `Unscoped` |
| 要结果（真查库） | 把菜单交给后厨 | `Find` `First` `Count` `Create` `Updates` `Delete` `Scan` |
| 看结果（字段，不是方法） | 读这一趟的结果 | `.Error` `.RowsAffected` `.Statement` |

---

## 二、拼条件：常用写法

| 我想… | 这么写 | 效果 |
|---|---|---|
| 说查哪张表 | `db.Model(&model.Message{})` | 用 messages 表 |
| 只要某几列 | `.Select("id", "title")` | 只查这两列；别动不动 `SELECT *`（TEXT 这种大字段会白拖） |
| 加条件 | `.Where("sender_id = ?", id)` | 多写几条 `.Where`，之间是 AND |
| 加"或者" | `.Where("a = ? OR b = ?", x, y)` | 两个 OR 必须写在**同一句**里（原因见第六节） |
| 批量匹配 | `.Where("id IN ?", ids)` | 等于 `id IN (1,2,3)`，切片直接传，不用自己加括号 |
| 模糊查 | `.Where("title LIKE ?", "%词%")` | 就是 LIKE |
| 排序 | `.Order("created_at DESC")` | ⚠️ 里面写的字**原样拼进 SQL**，千万别放用户输入 |
| 分页 | `.Limit(n)` `.Offset(n)` | LIMIT / OFFSET（边界见第五节） |
| 要一份独立副本 | `.Session(&gorm.Session{})` | 把当前菜单复印一份，两份各改各的 |
| 连被软删的也查 | `.Unscoped()` | 平时查不到被删的行，加了就能查到 |
| 这条打印 SQL | `.Debug()` | 调试用 |

`?` 是**给值留的空位**：值单独传给数据库，不拼进句子 —— 这就是防注入的原理。
所以值永远用 `?` 传，不要自己拼字符串。

---

## 三、要结果：常用写法

| 我想… | 这么写 | 结果放哪 | 要注意 |
|---|---|---|---|
| 查多行 | `Find(&list)` | `list` | 查不到**不算错**，要自己看 `len(list) == 0` |
| 查一行 | `First(&row)` | `row` | 查不到会返回"没找到"的错误（`gorm.ErrRecordNotFound`） |
| 取一条（不排序） | `Take(&row)` | `row` | 比 First 快，但不保证取到哪一条 |
| 数总数 | `Count(&n)` | `n` | 别在同一个 query 上数完接着 Find，用 `Session` 复印一份 |
| 插入 | `Create(&v)` | 自增 ID 会写回 `v.ID` | — |
| 更新（零值也更新） | `Updates(map[string]any{"is_read": true})` | — | **必须带 Where**，否则直接报错拦住你 |
| 更新（自动跳过零值） | `Updates(struct{...})` | — | `""`、`0`、`false` 不会被写进去 |
| 删一行 | `Delete(&model.Message{}, id)` | — | 表里有 `DeletedAt` 字段就是软删除（打标记，不真删） |
| 结果装进自己的结构体 | `Scan(&dto)` | `dto` | 常和 `Select` 搭配做统计 |

> **`&` 不能省**：Go 的参数是"复制一份递过去"。
> 写 `Find(posts)` 等于把 `posts` 的**复印件**递过去，GORM 往复印件里装东西，你手上那份还是空的；
> 写 `Find(&posts)` 是把 `posts` 的**门牌号**递过去，它才知道往哪放东西。
> `Count(&n)`、`First(&user)` 全是同一个道理。

---

## 四、看结果：三个字段

| 字段 | 是什么 |
|---|---|
| `.Error` | 这一趟有没有出错；没出错是 `nil`。判断失败一律 `if err != nil` |
| `.RowsAffected` | 真正改了几行（`MarkRead` 返回的 `updated` 就是它） |
| `.Statement` | 调试用：看最终拼出来的语句 |

---

## 五、分页怎么写

```go
query.
    Order("created_at DESC, id DESC").   // 最新在前；再加 id 防同一秒的消息乱序
    Offset((page - 1) * pageSize).       // 第 1 页跳 0 条，第 2 页跳 pageSize 条
    Limit(pageSize).                     // 最多要几条
    Find(&list)
```

| 写法 | 实际效果 |
|---|---|
| `Limit(20)` | `LIMIT 20`，正常 |
| `Limit(0)` | **`LIMIT 0` → 一条都不要**（列表永远是空的） |
| `Limit(-1)` | 不写 LIMIT（不限制） |
| `Offset(0)` | 不写 OFFSET，第 1 页正常 |
| `Offset(-1)` | 不写 OFFSET（取消偏移） |

> ⚠️ 所以 `page` / `pageSize` **一定要在 service 层补好默认值**（本项目用 `model.DefaultPage` / `model.DefaultPageSize`）。
> 忘了填就是空列表，这类 bug 很难一眼看出来。

---

## 六、四个容易踩的坑

### 1. `OR` 必须写在同一条 Where 字符串里

```go
// ✅ 对：GORM 会自动加括号 → (sender_id = ? OR receiver_id = ?) AND is_read = ?
query.Where("sender_id = ? OR receiver_id = ?", uid, uid)

// ❌ 错：拆成两条，SQL 里 AND 的优先级比 OR 高，等于
//    sender_id = ? OR (receiver_id = ? AND is_read = ?)
//    → 我发出去的消息会绕过 is_read 筛选，全都跑出来
query.Where("sender_id = ?", uid).Where("receiver_id = ?", uid)
```

### 2. `Limit(0)` 是"一条都不要"，不是"不限制"

补默认值，别让它变成 0。

### 3. 空切片 `IN` 查不到东西，但不会报错

`ids = []` → SQL 变成 `id IN (NULL)` → 一行也匹配不到，**但它不报错**。
所以用之前先 `if len(ids) == 0 { return nil, nil }`，意义是"别白跑一趟"。

### 4. `&` 不能省

见第三节最后的说明。

---

## 七、实战对照：M2 的 `ListMessages`

```go
query := db.Model(&model.Message{})      // 菜单：查 messages 表
switch q.Box {                            // 菜单：哪些消息算我的
case model.MessageBoxSent:                //   我发的
case model.MessageBoxReceived:            //   我收的
default:                                  //   都要（两个 OR 写在同一句里）
}
if q.IsRead != nil { /* 菜单：只看已读 / 只看未读 */ } // nil = 没传，不筛

query.Session(&gorm.Session{}).Count(&total)   // 复印一份去数总数，原件留着
query.Order(...).Offset(...).Limit(...).Find(&messages)  // 用原件查当页
```

一句话流程：**service 补默认值 → repository 拼条件 → 判 `.Error` → service 组装前端要的字段。**

---

## 八、事务（以后用得上）

```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&a).Error; err != nil {
        return err                 // 返回 error → 整笔回滚
    }
    return tx.Create(&b).Error     // 返回 nil → 提交
})
```

> ⚠️ 事务里面所有操作都要用函数参数里的 `tx`，不要用包级的 `db`，否则那一条不在事务里。
