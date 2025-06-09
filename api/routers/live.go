// FilePath: C:/LanshanClass1.3/api/routers/live.go
package routers

import (
	"LanshanClass1.3/api/controllers"
	"github.com/gin-gonic/gin"
)

// LiveRouter 定义直播相关的路由
func LiveRouter(r *gin.Engine) {
	live := r.Group("/live")
	{

		// 创建直播课
		live.POST("/create", controllers.CreateLiveClass)
		// 加入直播课
		live.POST("/join", controllers.JoinLiveClass)
		// 发送消息
		live.POST("/message/send", controllers.SendMessage)
		// 结束直播课
		live.POST("/end", controllers.EndLiveClass)
		// 发布题目
		live.POST("/question/publish", controllers.PublishQuestion)
		// 提交答案
		live.POST("/question/submit", controllers.SubmitAnswer)
		// 获取答题结果统计
		live.GET("/question/statistics", controllers.GetAnswerStatistics)
		live.GET("/message/get", controllers.GetMessages)
	}
}
