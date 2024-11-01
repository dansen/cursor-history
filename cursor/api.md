# Api List:
{"/api/page/count", WithAuth(pageCountHandler)},
{"/api/games", WithLog(gameListHandler)},
{"/api/game/create", WithAuth(createGameHandler)},
{"/api/game", WithAuth(getGameHandler)},
{"/api/game/update", WithAuth(updateGameHandler)},
{"/api/game/delete", WithAuth(deleteGameHandler)},
{"/api/login", WithLog(loginHandler)},
{"/api/register", WithLog(registerHandler)},
{"/api/prompt/upload", WithLog(uploadPromptHandler)},
{"/api/prompt/set-public", WithAuth(setPromptPublicHandler)},
{"/api/prompt/set-private", WithAuth(setPromptPrivateHandler)},
{"/api/api-key/generate", WithAuth(generateAPIKeyHandler)},
{"/api/api-key/list", WithAuth(listAPIKeysHandler)},
{"/api/api-key/disable", WithAuth(disableAPIKeyHandler)},
{"/api/prompts", WithLog(getPublicPromptsHandler)},
{"/api/user/prompts", WithLog(getUserPromptsHandler)}, // 添加这一行
{"/api/project/create", WithAuth(createProjectHandler)},
{"/api/project", WithAuth(getProjectHandler)},
{"/api/project/update", WithAuth(updateProjectHandler)},
{"/api/project/delete", WithAuth(deleteProjectHandler)},
{"/api/user/projects", WithAuth(getUserProjectsHandler)}, // 新增：获取用户所有项目
{"/api/workspaces", WithAuth(getWorkspacesHandler)},
{"/api/email/verify/send", WithLog(sendEmailVerifyHandler, true)},
{"/api/email/verify", WithLog(verifyEmailHandler, true)},
{"/api/api-key/valid", WithLog(validateApiKeyHandler, true)},

http api server:
http://localhost:7600/api/xxx

# API 文档

## 认证相关接口

### 1. 用户登录
路径: /api/login
方法: POST
请求数据:
{
    "username": "string",  // 用户名（可选，与email二选一）
    "email": "string",     // 邮箱（可选，与username二选一）
    "password": "string"   // 密码
}
响应数据:
{
    "error_code": 0,       // 0表示成功，非0表示失败
    "message": "string",   // 响应消息
    "token": "string",     // JWT token
    "username": "string"   // 用户名
}

### 2. 用户注册
路径: /api/register
方法: POST
请求数据:
{
    "username": "string",  // 用户名
    "password": "string",  // 密码
    "email": "string",     // 邮箱
    "code": "string"       // 邮箱验证码
}
响应数据:
{
    "error_code": 0,       // 0表示成功，非0表示失败
    "message": "string",   // 响应消息
    "data": {
        "token": "string",     // JWT token
        "user_id": 123,        // 用户ID
        "username": "string",  // 用户名
        "email": "string"      // 邮箱
    }
}

错误码说明：
1: 方法不允许
2: 无效的请求数据
3: 必填字段为空
4: 验证码验证失败
5: 用户名已存在
6: 邮箱已存在
7: 创建用户失败
8: 生成 token 失败

注意事项：
1. 注册前需要先调用发送验证码接口获取邮箱验证码
2. 验证码有效期为15分钟
3. 验证码���能使用一次
4. 用户名和邮箱都必须是唯一的

## API 密钥管理

### 3. 生成 API 密钥
路径: /api/api-key/generate
方法: POST
认证: 需要 JWT token
请求数据:
{
    "description": "string"  // API 密钥描述
}
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "id": 123,
        "user_id": 456,
        "api_key": "string",
        "description": "string",
        "created_at": "string",
        "last_used_at": "string",
        "is_active": true
    }
}

### 4. 获取 API 密钥列表
路径: /api/api-key/list
方法: GET
认证: 需要 JWT token
响应数据:
{
    "error_code": 0,
    "message": "string",
    "keys": [
        {
            "id": 123,
            "user_id": 456,
            "api_key": "string",
            "description": "string",
            "created_at": "string",
            "last_used_at": "string",
            "is_active": true
        }
    ]
}

### 5. 禁用 API 密钥
路径: /api/api-key/disable
方法: POST
认证: 需要 JWT token
请求参数: ?id=123 (API 密钥 ID)
响应数据:
{
    "error_code": 0,
    "message": "string"
}

## Prompt 相关接口

### 6. 上传 Prompt
路径: /api/prompt/upload  // 修改路径
方法: POST
认证: 需要 API 密钥
请求数据:
{
    "token": "string",
    "value": "string",
    "commandType": "string",
    "md5": "string",
    "timestamp": 1234567890,
    "workspace": "string",
    "uploadTime": 1234567890
}
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 7. 获取公共 Prompts
路径: /api/prompts
方法: GET
响应数据:
{
    "error_code": 0,
    "message": "string",
    "prompts": [
        {
            "id": 123,
            "user_id": 456,
            "value": "string",
            "command_type": "string",
            "md5": "string",
            "timestamp": "string",
            "created_at": "string",
            "is_public": true
        }
    ]
}

### 8. 获取用户 Prompts
路径: /api/user/prompts
方法: GET
认证: 需要 JWT token
响应数据:
{
    "error_code": 0,
    "message": "string",
    "prompts": [
        {
            "id": 123,
            "user_id": 456,
            "value": "string",
            "command_type": "string",
            "md5": "string",
            "timestamp": "string",
            "created_at": "string",
            "is_public": true
        }
    ]
}

### 20. 设置 Prompt 为公开
路径: /api/prompt/set-public
方法: POST
认证: 需要 JWT token
请求数据:
{
    "prompt_id": 123    // Prompt ID
}
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 21. 设置 Prompt 为私有
路径: /api/prompt/set-private
方法: POST
认证: 需要 JWT token
请求数据:
{
    "prompt_id": 123    // Prompt ID
}
响应数据:
{
    "error_code": 0,
    "message": "string"
}

## 游戏相关接口

### 9. 创建游戏
路径: /api/game/create
方法: POST
认证: 需要 JWT token
请求数据:
{
    "name": "string",           // 游戏名称
    "featureCount": 123,        // 特性数量
    "platform": "string",       // 平台
    "lastUpdate": "string",     // 最后更新时间
    "status": "string",         // 状态
    "image": "string",          // 图片链接
    "downloads": 123,           // 下载次数
    "downloadLink": "string",   // 下载链接
    "apkLink": "string"         // APK 下载链接
}
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "id": 123,
        "name": "string",
        "featureCount": 123,
        "platform": "string",
        "lastUpdate": "string",
        "status": "string",
        "image": "string",
        "downloads": 123,
        "downloadLink": "string",
        "apkLink": "string"
    }
}

### 10. 获取游戏详情
路径: /api/game
方法: GET
认证: 需要 JWT token
请求参数: ?id=123 (游戏 ID)
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "id": 123,
        "name": "string",
        "featureCount": 123,
        "platform": "string",
        "lastUpdate": "string",
        "status": "string",
        "image": "string",
        "downloads": 123,
        "downloadLink": "string",
        "apkLink": "string"
    }
}

### 11. 更新游戏
路径: /api/game/update
方��: PUT
认证: 需要 JWT token
请求数据:
{
    "id": 123,
    "name": "string",
    "featureCount": 123,
    "platform": "string",
    "lastUpdate": "string",
    "status": "string",
    "image": "string",
    "downloads": 123,
    "downloadLink": "string",
    "apkLink": "string"
}
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "id": 123,
        "name": "string",
        "featureCount": 123,
        "platform": "string",
        "lastUpdate": "string",
        "status": "string",
        "image": "string",
        "downloads": 123,
        "downloadLink": "string",
        "apkLink": "string"
    }
}

### 12. 删除游戏
路径: /api/game/delete
方法: DELETE
认证: 需要 JWT token
请求参数: ?id=123 (游戏 ID)
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 13. 获取游戏列表
路径: /api/games
方法: GET
响应数据:
{
    "error_code": 0,
    "message": "string",
    "games": [
        {
            "id": 123,
            "name": "string",
            "featureCount": 123,
            "platform": "string",
            "lastUpdate": "string",
            "status": "string",
            "image": "string",
            "downloads": 123,
            "downloadLink": "string",
            "apkLink": "string"
        }
    ]
}

## Project 相关接口

### 14. 创建项目
路径: /api/project/create
方法: POST
认证: 需要 JWT token
请求数据:
{
    "name": "string",        // 项目名称
    "description": "string", // 项目描述
    "status": 1             // 项目状态
}
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "id": 123,
        "user_id": 456,
        "name": "string",
        "description": "string",
        "create_time": "2024-03-21T10:00:00Z",
        "update_time": "2024-03-21T10:00:00Z",
        "status": 1
    }
}

### 15. 获取项目详情
路径: /api/project
方法: GET
认证: 需要 JWT token
请求参数: ?id=123 (项目 ID)
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "id": 123,
        "user_id": 456,
        "name": "string",
        "description": "string",
        "create_time": "2024-03-21T10:00:00Z",
        "update_time": "2024-03-21T10:00:00Z",
        "status": 1
    }
}

### 16. 更新项目
路径: /api/project/update
方法: PUT
认证: 需要 JWT token
请求数据:
{
    "id": 123,              // 项目ID
    "name": "string",       // 项目名称
    "description": "string",// 项目描述
    "status": 1            // 项目状态
}
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 17. 删除项目
路径: /api/project/delete
方法: DELETE
认证: 需要 JWT token
请求参数: ?id=123 (项目 ID)
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 18. 获取用户的所有项目
路径: /api/user/projects
方法: GET
认证: 需要 JWT token
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": [
        {
            "id": 123,
            "user_id": 456,
            "name": "string",
            "description": "string",
            "create_time": "2024-03-21T10:00:00Z",
            "update_time": "2024-03-21T10:00:00Z",
            "status": 1
        }
    ]
}

### 26. 获取项目详细信息
路径: /api/project/detail
方法: GET
认证: 需要 JWT token
请求参数: ?id=123 (项目 ID)
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "id": 123,
        "user_id": 456,
        "name": "string",
        "description": "string",
        "create_time": "2024-03-21T10:00:00Z",
        "update_time": "2024-03-21T10:00:00Z",
        "status": 1
    }
}

### 27. 获取项目关联的 Workspaces
路径: /api/project/workspaces
方法: GET
认证: 需要 JWT token
请求参数: ?id=123 (项目 ID)
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": [
        {
            "id": 123,
            "user_id": 456,
            "workspace": "string",
            "label": "string",
            "create_time": "2024-03-21T10:00:00Z",
            "update_time": "2024-03-21T10:00:00Z"
        }
    ]
}

### 28. 添加项目关联的 Workspace
路径: /api/project/add-workspace
方法: POST
认证: 需要 JWT token
请求参数: 
- projectID: 项目 ID
- workspaceID: Workspace ID
示例: /api/project/add-workspace?projectID=6&workspaceID=7
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 29. 删除项目关联的 Workspace
路径: /api/project/remove-workspace
方法: DELETE
认证: 需要 JWT token
请求参数: 
- projectID: 项目 ID
- workspaceID: Workspace ID
示例: /api/project/remove-workspace?projectID=6&workspaceID=7
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 31. 获取项目关联的 Prompts
路径: /api/project/prompts
方法: GET
认证: 需要 JWT token
请求参数: ?id=123 (项目 ID)
响应数据:
{
    "error_code": 0,
    "message": "string",
    "prompts": [
        {
            "id": 123,
            "user_id": 456,
            "value": "string",
            "command_type": "string",
            "md5": "string",
            "timestamp": 1234567890,
            "workspace": "string",
            "upload_time": 1234567890,
            "created_at": "2024-03-21T10:00:00Z",
            "is_public": true
        }
    ]
}

## 其他接口

### 19. 页面计数
路径: /api/page/count
方法: GET
认证: 需要 API 密钥
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "count": 123
    }
}

### 22. 获取用户信息
路径: /api/user/info
方法: GET
认证: 需要 JWT token
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "id": 123,
        "username": "string",
        "email": "string",
        "nickname": "string",
        "avatar": "string",
        "phone": "string"
    }
}

### 23. 更新用户信息
路径: /api/user/update
方法: POST
认证: 需要 JWT token
请求数据:
{
    "nickname": "string",
    "email": "string",
    "phone": "string"
}
响数据:
{
    "error_code": 0,
    "message": "string"
}

### 24. 修改密码
路径: /api/user/update-password
方法: POST
认证: 需要 JWT token
请求数据:
{
    "old_password": "string",
    "new_password": "string"
}
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 25. 上传头像
路径: /api/user/upload-avatar
方法: POST
认证: 需要 JWT token
请求数据: multipart/form-data
- avatar: 文件（图片）
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 30. 获取 Workspace 列表
路径: /api/workspaces
方法: GET
认证: 需要 JWT token
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": [
        {
            "id": 123,
            "user_id": 456,
            "workspace": "string",
            "label": "string",
            "create_time": "2024-03-21T10:00:00Z",
            "update_time": "2024-03-21T10:00:00Z"
        }
    ]
}

### 32. 发送邮箱验证码
路径: /api/email/verify/send
��法: POST
请求数据:
{
    "email": "string"
}
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 33. 验证邮箱验证码
路径: /api/email/verify
方法: POST
请求数据:
{
    "email": "string",
    "code": "string"
}
响应数据:
{
    "error_code": 0,
    "message": "string"
}

### 34. 验证 API Key 是否有效
路径: /api/api-key/valid
方法: GET
请求参数: ?key=your_api_key
响应数据:
{
    "error_code": 0,
    "message": "string",
    "data": {
        "valid": true,
        "username": "string",
        "expires": "2024-03-21T10:00:00Z"
    }
}

错误码说明：
1: 方法不允许
2: token 参数不能为空
3: 无效的 token 签名
4: token 解析失败
5: token 已过期或无效
6: 用户不存在

注意事项：
1. token 需要通过 URL 参数传递
2. 返回的 expires 时间使用 ISO 8601 格式
3. 验证包括：
   - token 格式是否正确
   - token 签名是否有效
   - token 是否过期
   - token 对应的用户是否存在
