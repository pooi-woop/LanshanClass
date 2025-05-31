# LanshanClass

## 项目简介
LanshanClass 是一个在线课堂系统，提供直播课程创建、加入、消息发送、题目发布与统计等功能，支持用户注册与登录。

## 项目结构
- `api`：包含 HTTP API 的路由定义与控制器逻辑。
- `global`：存放全局配置、数据库操作等公共模块。
- `proto`：存放 gRPC 服务的协议定义文件及生成的 Go 代码。
- `service`：实现 gRPC 服务端的业务逻辑。

## 环境依赖
- Go 1.18+
- MySQL
- Redis
- Gin 框架
- gRPC
- Protobuf
- livego
## 启动项目
1. 初始化数据库与 Redis，确保配置文件`databse.yaml`中的连接信息正确。
2. 启动 LiveGo 服务器（位于`global/livegooooo/main.go`）。
3. 启动 gRPC 服务（位于`service/LIVEGO/main.go`和`service/auth/main.go`）。
4. 启动 HTTP API 服务（位于`api/main.go`）。

## 路由测试用例

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
  - **Body**：
    ```json
    {
      "teacher_name": "张老师",
      "class_name": "数学课",
      "stream_url": "http://example.com/stream"
    }
    ```
- **预期响应**
  - **状态码**：`201 Created`
  - **Body**：
    ```json
    {
      "classID": "数学课",
      "status": "success"
    }
    ```

### 加入直播课
- **请求**
  - **URL**：`POST /live/join`
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
  - **Body**：
    ```json
    {
      "class_id": "数学课",
      "sender_name": "小明",
      "message_content": "老师好！"
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

### 结束直播课
- **请求**
  - **URL**：`POST /live/end`
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

## 联系方式
如果有任何问题或建议，可以通过以下方式联系我们：
- 邮箱：[your-email@example.com](mailto:your-email@example.com)
- GitHub：[https://github.com/your-username/LanshanClass](https://github.com/your-username/LanshanClass)

希望这个`README.md`文件对你有帮助，你可以根据实际情况对内容进行调整和补充。
