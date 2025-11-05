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
	// 从查询参数获取筛选条件
	category := c.Query("category")
	// 从查询参数获取搜索关键词
	search := c.Query("search")
	// 构建基础查询语句
	query := "SELECT id, title, description, category, organizer, location, start_time, end_time, capacity, created_by_id FROM activities"
	// 动态添加筛选条件
	conditions := []string{}
	// 动态添加查询参数
	args := []interface{}{}
	// 根据是否提供筛选条件，构建 WHERE 子句
	if category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, category)
	}
	// 根据搜索关键词模糊匹配标题
	if search != "" {
		conditions = append(conditions, "title LIKE ?")
		args = append(args, "%"+search+"%")
	}
	// 拼接最终查询语句
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	// 默认按开始时间降序排列
	query += " ORDER BY start_time DESC"
	// 执行查询
	rows, err := DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询活动失败: " + err.Error()})
		return
	}
	defer rows.Close()

	// 遍历结果集，构建活动切片，返回给前端
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
	// 定义一个 Activity 变量来接收前端传来的数据，使用 models 包中的 Activity 结构体
	var activity models.Activity

	// 1. 绑定 JSON 数据
	if err := c.ShouldBindJSON(&activity); err != nil {
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

	// 从auth中间件获取用户ID，作为活动的创建者
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
	query := `
		INSERT INTO activities(title, description, category, organizer, location, start_time, end_time, capacity, created_by_id) 
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 5. 使用 ExecContext 执行插入，并获取结果
	result, err := DB.ExecContext(c.Request.Context(), query,
		activity.Title, activity.Description, activity.Category, activity.Organizer,
		activity.Location, activity.StartTime, activity.EndTime, activity.Capacity, activity.CreatedByID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误，创建活动失败"})
		return
	}

	// 5. 获取新插入行的ID，并返回完整的活动对象
	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{"message": "活动创建成功"})
		return
	}

	activity.ID = int(id) // 将新ID赋值给对象

	// 返回包含新活动的详细信息，减小开销
	c.JSON(http.StatusCreated, activity)
}

// 删除一个活动
func DeleteActivity(c *gin.Context) {
	// 从URL参数中获取活动ID
	id := c.Param("id")

	// 准备删除语句
	stmt, err := DB.Prepare("DELETE FROM activities WHERE id = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除活动失败"})
		return
	}
	defer stmt.Close()

	// 执行删除
	_, err = stmt.Exec(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "执行删除活动失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "活动删除成功"})
}
