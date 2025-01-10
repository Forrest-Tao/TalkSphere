## 论坛项目

### **1. 用户模块**

- **注册**：用户可以创建账号。
- **登录**：用户可以通过账号密码进行登录。
- **登出**：用户可以退出当前会话。
- **RBAC权限分配**：不同角色的用户具有不同的权限，如普通用户、管理员等，使用角色基于访问控制（RBAC）来管理权限。

### **2. 板块管理模块**

- **板块管理**：管理论坛的不同板块，每个板块可以承载不同主题的帖子，管理员可以新增、修改或删除板块。

### **3. 帖子模块**

- **发表帖子**：用户可以发表带有图片和文字的帖子，内容可以包括图片、文字、链接等多种格式。

### **4. 互动模块**

- **评论**：帖子下方允许用户进行评论，支持树形结构的评论展示，即评论可以嵌套和回复。
- **点赞**：用户可以对帖子或评论进行点赞。
- **收藏**：用户可以收藏自己喜欢的帖子，以便以后查看。

### **5. 榜单模块**

- **排行榜**：展示基于某些数据（如点赞数、评论数等）和算法计算的实时榜单（例如 Top 10 最受欢迎的帖子）。

### **6. 模糊搜索模块**

- **帖子搜索**：用户可以通过关键字搜索帖子，支持模糊搜索，返回相关帖子列表。

### **7. 数据可视化模块**

- **数据展示**：将论坛的各种数据（如用户活跃度、帖子数量、点赞数等）通过图表或其他方式进行可视化，帮助管理员或用户分析数据趋势。



---

## timeLine

- 2025.1.3 
  - 项目前端、后端骨架搭建
  - 登陆、注册前后端连调
- 2025.1.4
  - 实现JWT
  - 对接腾讯OSS，实现bucket创建和object的获取、删除、更新操作
  - 实现用户头像上传，bio更新
  - 表结构设计；简化业务
  - fix 登陆后，跳转forum失败问题
- 2025.1.5
  - 表结构设计
  - 业务功能设计
  - 板块模块的CRUD
  - 使用casbin完成 用户RBAC
- 2025.1.8
  - fix JWT
  - 帖子模块
    - 用户发表帖子
    - 根据id获取帖子详情
    - 根据id删除帖子
    - 根据id更新帖子
    - 获取某用户的所有帖子
    - 获取某板块下的所有帖子
- 2025.1.9
  - 交互模块
    - 评论
      - 根评论
      - 子评论
    - 点赞
      - 给贴子点赞
      - 给用户评论点赞
    - 收藏帖子
- 2025.1.10
  - 前端
    -用户发表帖子
    - 按照板块查看帖子列表
    - 查看帖子详情
    - 给帖子点赞、收藏、阅读量增加
- 2025.1.11
  - 前端
    - 支持富文本的格式创建帖子
    - 支持嵌套评论
    - 支持用户查看自己点赞的帖子
    - 支持用户查看自己收藏的帖子
    - 支持用户查看自己发表的帖子
    - 支持用户查看自己评论的帖子
  - 后端
    - 榜单模块（展示基于某些数据（如点赞数、评论数等）和算法计算的实时榜单（例如 Top 10 最受欢迎的帖子））