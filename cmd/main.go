package main

import (
	"campus-activity-api/internal/config"
	"campus-activity-api/internal/database"
	"campus-activity-api/internal/handlers"
	"campus-activity-api/internal/middleware"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	// 2. 初始化数据库连接
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}
	defer db.Close()
	
	// 3. 将数据库连接实例注入到handlers包
	handlers.DB = db
	log.Println("数据库连接成功!")

	// 4. Gin 路由
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://jinjie1101.z23.web.core.windows.net"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Authorization"},
		AllowCredentials: true,
		// 对于相同的请求，无需再发送 OPTIONS 预检请求
		MaxAge: 12 * time.Hour,
	}))

	api := router.Group("/api")
	{
		// auth
		api.POST("/register", handlers.Register) // 注册
		api.POST("/login", handlers.Login)       // 登录
		// user
		api.GET("/users/:id/registrations", handlers.GetMyActivities)
		api.POST("/activities/:id/register", middleware.AuthMiddleware(), handlers.RegisterForActivityHandler(db))
		api.DELETE("/registrations/:id", handlers.CancelRegistration)
		// activity
		api.GET("/activities", handlers.GetActivities)
		api.GET("/activities/:id", handlers.GetActivityByID)
		api.POST("/activities", middleware.AuthMiddleware(), handlers.CreateActivity)
		api.DELETE("/activities/:id", handlers.DeleteActivity)
		// stats
		api.GET("/stats/hot-activities", handlers.GetHotActivities)
		api.GET("/stats/organizer-activity-counts", handlers.GetOrganizerStats)
		// admin
		api.GET("/activities/:id/registrations", handlers.GetRegistrationsByActivityIDHandler(db))
		admin := api.Group("/admin")
		{
			admin.GET("/registrations", handlers.GetRegistrationsHandler(db))
			admin.PUT("/registrations/:registrationId/status", handlers.AdminUpdateRegistrationStatusHandler(db))
			admin.DELETE("/registrations/:id", middleware.AuthMiddleware(), handlers.AdminDeleteRegistrationHandler(db))
		}
	}

	router.Run(":8080")
}
