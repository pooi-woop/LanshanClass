



# LanshanClass v1.5

## 项目简介
LanshanClass 是一个功能丰富的在线课堂系统，旨在为用户提供便捷的直播课程体验。系统支持直播课程的创建、加入、实时消息发送、题目发布与答题统计等功能。此外，系统还提供了用户注册与登录功能，并通过 JWT Token 实现鉴权，确保通信安全。

## 项目结构
- **`api`**：包含 HTTP API 的路由定义与控制器逻辑。
- **`global`**：存放全局配置、数据库操作等公共模块。
- **`proto`**：存放 gRPC 服务的协议定义文件及生成的 Go 代码。
- **`service`**：实现 gRPC 服务端的业务逻辑。
- **`utils`**：存放工具函数，如 JWT 生成和验证。

## 环境依赖
- Go 1.18+
- MySQL
- Redis
- Gin 框架
- gRPC
- Protobuf
- LiveGo

## 启动项目
### 1. 初始化数据库与 Redis
确保数据库和 Redis 已正确配置，并且配置文件`database.yaml`中的连接信息正确。运行以下命令初始化数据库（假设你已经安装了 MySQL 和 Redis）：
```bash
mysql -u root -p < database/init.sql
```

### 2. 启动 LiveGo 服务器
```bash
go run C:/LanshanClass1.3/global/livegooooo/main.go
```

### 3. 启动 gRPC 服务
```bash
go run C:/LanshanClass1.3/service/auth/main.go
go run C:/LanshanClass1.3/service/LIVEGO/main.go
```

### 4. 启动 HTTP API 服务
```bash
go run C:/LanshanClass1.3/api/main.go
```

## 功能测试用例

### 用户注册
- **请求**
  - **URL**：`POST /register`
  - **Body**：
    ```json
    {
      "username": "testuser",
      "password": "testpass"
    }
    ```
- **预期响应**
  - **状态码**：`200 OK`
  - **Body**：
    ```json
    {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "message": "注册成功"
    }
    ```

### 用户登录
- **请求**
  - **URL**：`POST /login`
  - **Body**：
    ```json
    {
      "username": "testuser",
      "password": "testpass"
    }
    ```
- **预期响应**
  - **状态码**：`200 OK`
  - **Body**：
    ```json
    {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "message": "登录成功"
    }
    ```

### 创建直播课
- **请求**
  - **URL**：`POST /live/create`
  - **Header**：
    ```
    Authorization: Bearer <token>
    ```
  - **Body**：
    ```json
    {
      "teacher_name": "张老师",
      "class_name": "数学课",
      "room_name": "math_class_room"
    }
    ```
- **预期响应**
  - **状态码**：`201 Created`
  - **Body**：
    ```json
    {
      "classID": "数学课",
      "status": "success",
      "streamKey": "推流密钥"
    }
    ```

### 加入直播课
- **请求**
  - **URL**：`POST /live/join`
  - **Header**：
    ```
    Authorization: Bearer <token>
    ```
  - **Body**：
    ```json
    {
      "class_id": "数学课",
      "student_name": "小明"
    }
    ```
- **预期响应**
  - **状态码**：`200 OK`
  - **Body**：
    ```json
    {
      "status": "success",
      "streamUrl": "http://example.com/stream"
    }
    ```

### 发送消息
- **请求**
  - **URL**：`POST /live/message/send`
  - **Header**：
    ```
    Authorization: Bearer <token>
    ```
  - **Body**：
    ```json
    {
      "class_id": "数学课",
      "message": {
        "message_content": "老师好！"
      }
    }
    ```
- **预期响应**
  - **状态码**：`200 OK`
  - **Body**：
    ```json
    {
      "status": "success"
    }
    ```

### 获取直播课消息
- **请求**
  - **URL**：`GET /live/message/get`
  - **Header**：
    ```
    Authorization: Bearer <token>
    ```
  - **Query Parameters**：
    ```
    class_id=数学课&last_timestamp=1680307200
    ```
- **预期响应**
  - **状态码**：`200 OK`
  - **Body**：
    ```json
    {
      "messages": [
        {
          "sender_name": "张老师",
          "message_content": "欢迎来到数学课！",
          "timestamp": 1680307201
        },
        {
          "sender_name": "小明",
          "message_content": "老师好！",
          "timestamp": 1680307202
        }
      ]
    }
    ```

### 结束直播课
- **请求**
  - **URL**：`POST /live/end`
  - **Header**：
    ```
    Authorization: Bearer <token>
    ```
  - **Body**：
    ```json
    {
      "class_id": "数学课"
    }
    ```
- **预期响应**
  - **状态码**：`200 OK`
  - **Body**：
    ```json
    {
      "status": "success"
    }
    ```

### 发布题目
- **请求**
  - **URL**：`POST /live/question/publish`
  - **Header**：
    ```
    Authorization: Bearer <token>
    ```
  - **Body**：
    ```json
    {
      "class_id": "数学课",
      "question": "1+1=？"
    }
    ```
- **预期响应**
  - **状态码**：`200 OK`
  - **Body**：
    ```json
    {
      "status": "success"
    }
    ```

### 提交答案
- **请求**
  - **URL**：`POST /live/question/submit`
  - **Header**：
    ```
    Authorization: Bearer <token>
    ```
  - **Body**：
    ```json
    {
      "class_id": "数学课",
      "student_name": "小明",
      "answer": "2",
      "question_id": "20240601123456"
    }
    ```
- **预期响应**
  - **状态码**：`200 OK`
  - **Body**：
    ```json
    {
      "status": "success"
    }
    ```

### 获取答题结果统计
- **请求**
  - **URL**：`GET /live/question/statistics?class_id=数学课&question_id=20240601123456`
  - **Header**：
    ```
    Authorization: Bearer <token>
    ```
- **预期响应**
  - **状态码**：`200 OK`
  - **Body**：
    ```json
    {
      "question_id": "20240601123456",
      "answer_counts": {
        "2": 1
      }
    }
    ```

## 注意事项
- 在测试过程中，确保 LiveGo 服务器、gRPC 服务和 HTTP API 服务均已正常启动。
- 测试时可使用 Postman 或 curl 等工具发送请求。
- 对于涉及数据库操作的接口，需确保数据库已正确初始化且数据表结构与项目代码一致。

## 贡献指南
欢迎对 LanshanClass 项目进行贡献！你可以通过以下方式参与：
- 提交 Issue 报告问题或提出改进建议。
- 提交 Pull Request 修复问题或添加新功能。
