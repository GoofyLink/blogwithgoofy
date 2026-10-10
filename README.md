# 小天 Goofy 博客（blogwitgoofy）

Vue3 + Go 的个人博客，带 **英语/德语学习模块**（学习文章 + 段落批注笔记）。

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3.5 + TypeScript + Vite 6 + Vue Router + Pinia + Element Plus |
| Markdown | 后台编辑 md-editor-v3，前台渲染 markdown-it + highlight.js + DOMPurify |
| 后端 | Go 1.26 + Gin + GORM |
| 数据库 | MySQL 5.7（utf8mb4）|
| 认证 | JWT（Bearer Token，7 天有效期），密码 bcrypt |

## 目录结构

```
blogwitgoofy/
├── server/               # Go 后端（:8080）
│   ├── cmd/server/       # 入口
│   ├── internal/         # config / model / handler / middleware / router
│   ├── pkg/              # response / utils（jwt、bcrypt）
│   ├── uploads/          # 上传图片存放处
│   └── config.yaml       # 数据库与 JWT 配置
├── web/                  # Vue3 前端（dev :5173，/api 代理到 8080）
├── scripts/init.sql      # 建库脚本
```

## 快速启动

### 1. 准备数据库

启动 phpstudy 的 MySQL 后执行一次：

```bash
mysql -h127.0.0.1 -uroot -p < scripts/init.sql
```

（只建 `blog` 库；表结构由后端启动时 AutoMigrate 自动创建，并写入种子数据）

首次运行先将 `server/config.example.yaml` 复制为 `server/config.yaml`，填写本地 MySQL 密码，并将 `jwt.secret` 替换为足够长的随机密钥。`server/config.yaml` 仅保存在本地，不提交到 Git。

### 2. 启动后端

```bash
cd server
go run ./cmd/server
# 输出: Blog API 已启动: http://localhost:8080
```

### 3. 启动前端

```bash
cd web
npm install        # 首次
npm run dev
# 打开 http://localhost:5173
```

## 默认账号

- 后台地址：`http://localhost:5173/admin/login`
- 用户名 `admin` / 密码 `admin123`
- **登录后请在右上角头像菜单里修改密码**

## 功能清单

**博客**
- 文章（Markdown、草稿/发布、封面、浏览量）、分类、标签
- 评论（访客填写昵称直接评论，后台可删）
- 归档时间线 + 标题/内容关键词搜索
- 友链管理、自定义单页（`/about`、`/p/:slug`）

**学习模块（双语书房，特色）**
- 书架页（`/learn`）：英/德语书籍卡片，可按语言筛选
- 书籍页（`/learn/book/:id`）：书籍信息 + 章节目录 + 开始阅读
- **章节阅读页（`/learn/chapter/:id`）**：
  - **鼠标悬停任意单词** → 浮层显示中文释义
  - **点击任意句子** → 句尾显示整句中文翻译，**再点一次隐藏**
  - 字号调节（A−/A＋）；章节若录有全文对照，可开关「中文对照」逐段显示
  - 上一章/下一章导航
- 翻译由后端代理（MyMemory 为主、Google 为备，服务端缓存），前台无需配置密钥
- 后台「书籍管理」维护书籍与章节（原文 + 可选整章中文对照）
- **批量导入章节**：整本书 TXT 一次导入——粘贴全文或上传文件（自动识别 UTF-8/GBK），
  按「Chapter 1 / 第1章 / Kapitel 1」标题自动切分，或用自定义分隔符；
  可同时导入中文译文并按章节顺序对齐；导入前有预览表格（改标题/删章）

**后台**（`/admin`，JWT 鉴权）
- 仪表盘统计、文章管理（md-editor-v3 + 图片上传）、分类/标签/评论/友链/单页管理、修改密码

## 生产部署（简述）

1. `cd web && npm run build` 产出 `web/dist`
2. 用 nginx 托管 `dist`，`/api`、`/uploads` 反代到 Go 服务（8080）
3. `cd server && go build -o blog ./cmd/server` 编译单文件运行
4. 生产环境请修改 `config.yaml` 中的 JWT secret

## 常见问题

访问统计的指标口径、本地调试方法和数据保留规则见 [访问统计说明](docs/analytics.md)。

- **后端起不来提示连接失败**：确认 phpstudy MySQL 已启动、`blog` 库已建、`config.yaml` 密码正确
- **翻译不出/变慢**：翻译走后端代理（`/api/v1/translate`），依赖外网 MyMemory 服务；同一词句翻译过一次后会命中缓存。国内网络 MyMemory 可直连；若部署在海外可自动回退 Google 翻译
- **图片上传失败**：图片 ≤5MB，仅支持 jpg/png/gif/webp；文件存在 `server/uploads/`
- **MySQL 5.7 兼容**：GORM 默认字符串长度 191，utf8mb4 唯一索引安全
