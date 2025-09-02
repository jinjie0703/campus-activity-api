// 用户报名与管理员管理报名状态的 Handler 层，主要处理学生报名、取消报名，以及管理员审核报名，使用了 RegistrationDetails 和 RegistrationDetailsForActivity模型
package handlers

import (
	"campus-activity-api/internal/models"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 获取所有用户的报名信息
func GetRegistrationsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		registrations, err := GetAllRegistrations(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve registrations"})
			return
		}
		c.JSON(http.StatusOK, registrations)
	}
}

// 处理“用户报名活动”的请求
// func RegisterForActivityHandler(db *sql.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// 从 URL 获取活动 ID
// 		activityID, err := strconv.Atoi(c.Param("id"))
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的活动ID"})
// 			return
// 		}

// 		// 从认证中间件获取用户ID (这是标准做法)
// 		userID, exists := c.Get("userID")
// 		if !exists {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
// 			return
// 		}

// 		// --- 核心修改在这里 ---
// 		// 准备插入语句，明确包含 status 字段
// 		query := "INSERT INTO registrations (user_id, activity_id, status) VALUES (?, ?, 'pending')"

// 		_, err = db.Exec(query, userID, activityID)
// 		if err != nil {
// 			// 处理可能的错误，比如重复报名
// 			if strings.Contains(err.Error(), "Duplicate entry") {
// 				c.JSON(http.StatusConflict, gin.H{"error": "你已经报名过该活动"})
// 				return
// 			}
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "报名失败，服务器错误"})
// 			return
// 		}

// 		c.JSON(http.StatusCreated, gin.H{"message": "报名成功，请等待管理员审核"})
// 	}
// }

// 管理员更新报名状态的 Gin 处理函数
func AdminUpdateRegistrationStatusHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 URL 获取报名ID
		registrationID, err := strconv.Atoi(c.Param("registrationId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的报名ID"})
			return
		}

		// 从请求体获取新的状态
		var req struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求体无效, 需要 'status' 字段"})
			return
		}

		// 验证 status 值是否合法
		if req.Status != "approved" && req.Status != "rejected" && req.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "状态值必须是 'approved', 'rejected', 或 'pending'"})
			return
		}

		// 调用数据库函数更新状态
		if err := UpdateRegistrationStatus(db, registrationID, req.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新报名状态失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "报名状态更新成功"})
	}
}

// 学生自己取消自己报名的活动
// func DeleteRegistrationHandler(db *sql.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// 从 URL 获取要删除的报名记录 ID
// 		registrationID, err := strconv.Atoi(c.Param("id"))
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的报名ID"})
// 			return
// 		}

// 		// 从认证中间件获取当前登录用户的 ID
// 		currentUserID, exists := c.Get("userID")
// 		if !exists {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
// 			return
// 		}

// 		// --- 这是最关键的安全校验 ---
// 		// 准备一条带有双重验证的 DELETE 语句
// 		// 只有当 id 和 user_id 同时匹配时，才会执行删除
// 		query := "DELETE FROM registrations WHERE id = ? AND user_id = ?"

// 		result, err := db.Exec(query, registrationID, currentUserID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消报名失败，服务器错误"})
// 			return
// 		}

// 		// 检查是否有行被影响。如果没有，说明这条记录不属于该用户，或者不存在
// 		rowsAffected, err := result.RowsAffected()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "取消报名失败，无法确认操作"})
// 			return
// 		}

// 		if rowsAffected == 0 {
// 			// 这通常意味着用户在尝试删除不属于自己的记录
// 			c.JSON(http.StatusForbidden, gin.H{"error": "无权操作或该报名不存在"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "取消报名成功"})
// 	}
// }

// 根据活动 ID 获取该活动的所有报名者信息，并返回给前端
func GetRegistrationsByActivityIDHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 URL 中获取活动 ID
		activityID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
			return
		}

		// 调用新的数据库函数
		registrants, err := GetRegistrationsByActivityID(db, activityID)
		if err != nil {
			// 可以在这里增加日志记录 err 的具体信息
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve registrants"})
			return
		}

		// 如果没有报名者，返回一个空数组而不是错误
		if registrants == nil {
			c.JSON(http.StatusOK, []models.RegistrationDetailsForActivity{})
			return
		}

		c.JSON(http.StatusOK, registrants)
	}
}
