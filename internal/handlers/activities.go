// 活动管理模块，涵盖了活动的增删改查（CRUD）、查询、筛选、导出报名数据等功能，使用了 Activity 模型
package handlers

import (
	"campus-activity-api/internal/models"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 获取活动列表，支持条件筛选
func GetActivities(c *gin.Context) {
	category := c.Query("category")
	search := c.Query("search")
	query := "SELECT id, title, description, category, organizer, location, start_time, end_time, capacity, created_by_id FROM activities"
	conditions := []string{}
	args := []interface{}{}
	if category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, category)
	}
	if search != "" {
		conditions = append(conditions, "title LIKE ?")
		args = append(args, "%"+search+"%")
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY start_time DESC"
	rows, err := DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询活动失败: " + err.Error()})
		return
	}
	defer rows.Close()
	activities := []models.Activity{}
	for rows.Next() {
		var a models.Activity
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.Category, &a.Organizer, &a.Location, &a.StartTime, &a.EndTime, &a.Capacity, &a.CreatedByID); err != nil {
			log.Println("扫描活动数据失败:", err)
			continue
		}
		activities = append(activities, a)
	}
	c.JSON(http.StatusOK, activities)
}

// 根据活动 ID 获取单个活动的详细信息
func GetActivityByID(c *gin.Context) {
	id := c.Param("id")
	var a models.Activity
	err := DB.QueryRow("SELECT id, title, description, category, organizer, location, start_time, end_time, capacity, created_by_id FROM activities WHERE id = ?", id).Scan(&a.ID, &a.Title, &a.Description, &a.Category, &a.Organizer, &a.Location, &a.StartTime, &a.EndTime, &a.Capacity, &a.CreatedByID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "活动未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, a)
}

// 创建一个新活动
func CreateActivity(c *gin.Context) {
	var activity models.Activity

	// 1. 绑定 JSON 数据，如果失败则记录详细错误
	if err := c.ShouldBindJSON(&activity); err != nil {
		// 【关键】在服务器控制台打印出详细的错误，方便调试
		log.Printf("无法绑定活动数据, 错误: %v", err)
		// 返回给前端一个清晰的错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求的数据格式无效或不完整"})
		return
	}

	// 2. 后端数据验证
	if activity.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "活动标题不能为空"})
		return
	}
	if activity.EndTime.Before(activity.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间不能早于开始时间"})
		return
	}
	// 可以添加更多验证...

	// 3. 从 Gin 的上下文中获取当前登录用户ID (不再硬编码)
	//    这需要你的认证中间件在用户登录后，将用户ID存入 context
	//    例如: c.Set("userID", user.ID)
	userID, exists := c.Get("userID")
	if !exists {
		// 如果中间件没有设置 userID，说明用户未认证
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证，无法发布活动"})
		return
	}
	// 类型断言，将 any/interface{} 转换为 int
	if uid, ok := userID.(float64); ok {
		activity.CreatedByID = int(uid)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法解析用户ID"})
		return
	}

	// 4. 执行数据库插入操作
	// 使用 ExecContext 来支持请求的上下文，例如超时控制
	query := `
		INSERT INTO activities(title, description, category, organizer, location, start_time, end_time, capacity, created_by_id) 
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := DB.ExecContext(c.Request.Context(), query,
		activity.Title, activity.Description, activity.Category, activity.Organizer,
		activity.Location, activity.StartTime, activity.EndTime, activity.Capacity, activity.CreatedByID,
	)

	if err != nil {
		log.Printf("数据库插入活动失败: %v", err) // 记录详细的数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误，创建活动失败"})
		return
	}

	// 5. 获取新插入行的ID，并返回完整的活动对象
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("获取 LastInsertId 失败: %v", err)
		// 即使获取ID失败，活动也已创建成功，可以返回一个通用成功消息
		c.JSON(http.StatusCreated, gin.H{"message": "活动创建成功"})
		return
	}

	activity.ID = int(id) // 将新ID赋值给对象

	c.JSON(http.StatusCreated, activity) // 返回包含新ID的完整活动对象
}

// 删除一个活动
func DeleteActivity(c *gin.Context) {
	id := c.Param("id")
	stmt, err := DB.Prepare("DELETE FROM activities WHERE id = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除活动失败"})
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "执行删除活动失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "活动删除成功"})
}
