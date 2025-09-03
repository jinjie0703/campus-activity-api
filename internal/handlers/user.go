// 用户对于自己报名活动的三个api
package handlers

import (
	"campus-activity-api/internal/models"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// 获取用户报名的所有活动
func GetMyActivities(c *gin.Context) {
	// 从 url 参数中获取用户id
	userID := c.Param("id")
	// sql查询
	query := `SELECT r.id, a.id, a.title, a.location, a.start_time FROM registrations r JOIN activities a ON r.activity_id = a.id WHERE r.user_id = ? ORDER BY a.start_time DESC`
	rows, err := DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询我的活动失败"})
		return
	}
	defer rows.Close()

	// 创建一个 []models.UserRegistration 切片存储结果，返回到前端页面
	registrations := []models.UserRegistration{}
	for rows.Next() {
		var reg models.UserRegistration
		if err := rows.Scan(&reg.RegistrationID, &reg.ActivityID, &reg.Title, &reg.Location, &reg.StartTime); err != nil {
			log.Println("扫描我的活动数据失败:", err)
			continue
		}
		registrations = append(registrations, reg)
	}
	c.JSON(http.StatusOK, registrations)
}

// 处理“用户报名活动”的请求
func RegisterForActivityHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 URL 获取活动 ID
		activityID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的活动ID"})
			return
		}

		// 从认证中间件获取用户ID (这是标准做法)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			return
		}

		// 类型断言
		uid, ok := userID.(float64)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法解析用户ID"})
			return
		}

		// --- 核心修改在这里 ---
		// 准备插入语句，明确包含 status 字段
		query := "INSERT INTO registrations (user_id, activity_id, status) VALUES (?, ?, 'pending')"

		_, err = db.Exec(query, int(uid), activityID)
		if err != nil {
			// 处理可能的错误，比如重复报名
			if strings.Contains(err.Error(), "Duplicate entry") {
				c.JSON(http.StatusConflict, gin.H{"error": "你已经报名过该活动"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "报名失败，服务器错误"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "报名成功，请等待管理员审核"})
	}
}

// 用户取消活动报名处理
func CancelRegistration(c *gin.Context) {
	registrationID := c.Param("id")
	// sql删除
	stmt, err := DB.Prepare("DELETE FROM registrations WHERE id = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消报名失败"})
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(registrationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "执行取消报名失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "取消报名成功"})
}
