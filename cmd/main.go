package main

import (
	"campus-activity-api/internal/database"
	"campus-activity-api/internal/handlers"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化数据库连接
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}
	defer db.Close()
	// 2. 将数据库连接实例注入到handlers包
	handlers.DB = db
	log.Println("数据库连接成功!")

	// 3. Gin 路由
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := router.Group("/api")
	{
		api.POST("/register", handlers.Register) // 新增注册路由
		api.POST("/login", handlers.Login)
		api.GET("/users/:id/registrations", handlers.GetMyActivities)
		api.GET("/activities", handlers.GetActivities)
		api.GET("/activities/:id", handlers.GetActivityByID)
		api.POST("/activities", handlers.CreateActivity)
		api.PUT("/activities/:id", handlers.UpdateActivity)
		api.DELETE("/activities/:id", handlers.DeleteActivity)
		api.DELETE("/registrations/:id", handlers.CancelRegistration)
		api.GET("/stats/hot-activities", handlers.GetHotActivities)
		api.GET("/stats/organizer-activity-counts", handlers.GetOrganizerStats)
		api.GET("/activities/:id/export", handlers.ExportRegistrations)

		api.GET("/activities/:id/registrations", handlers.GetRegistrationsByActivityIDHandler(db))
		api.POST("/activities/:id/register", handlers.RegisterForActivityHandler(db))
		// Admin routes
		admin := api.Group("/admin")
		{
			admin.GET("/registrations", handlers.GetRegistrationsHandler(db))
			admin.PUT("/registrations/:registrationId/status", handlers.AdminUpdateRegistrationStatusHandler(db))
			admin.DELETE("/registrations/:id", handlers.DeleteRegistrationHandler(db))
		}
	}

	router.Run(":8080")
}
