// FilePath: C:/LanshanClass1.3/api/routers\auth.go
package routers

import (
	"LanshanClass1.3/api/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRouter(r *gin.Engine) {
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
}
