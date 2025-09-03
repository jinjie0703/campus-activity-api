// 用户报名与管理员管理报名状态的 Handler 层，主要处理学生报名、取消报名，以及管理员审核报名，使用了 RegistrationDetails 和 RegistrationDetailsForActivity模型
package handlers

import (
	"campus-activity-api/internal/models"
	"database/sql"
	"log"
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

// 		// 类型断言
// 		uid, ok := userID.(float64)
// 		if !ok {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法解析用户ID"})
// 			return
// 		}

// 		// --- 核心修改在这里 ---
// 		// 准备插入语句，明确包含 status 字段
// 		query := "INSERT INTO registrations (user_id, activity_id, status) VALUES (?, ?, 'pending')"

// 		_, err = db.Exec(query, int(uid), activityID)
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

// AdminDeleteRegistrationHandler (最终正确版本 - 实时计算模式)
// 管理员删除用户的报名记录。
// 在这种模式下，函数唯一的职责就是从 registrations 表删除记录。
// 报名人数会在需要时通过查询 registrations 表重新计算。
func AdminDeleteRegistrationHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 URL 获取要删除的报名记录 ID
		registrationID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的报名ID"})
			return
		}

		// 【核心修改】移除了所有关于 activities 表和 registered_count 的操作
		// 我们不再需要事务，因为现在只执行一个单一的数据库操作。

		// 2. 直接根据 registrationID 执行删除操作
		query := "DELETE FROM registrations WHERE id = ?"
		result, err := db.Exec(query, registrationID)
		if err != nil {
			log.Printf("删除报名记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除报名记录失败"})
			return
		}

		// 3. （可选但推荐）检查是否真的删除了记录
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("获取影响行数失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，无法确认结果"})
			return
		}
		if rowsAffected == 0 {
			// 如果没有行被影响，说明这个ID可能一开始就不存在
			c.JSON(http.StatusNotFound, gin.H{"error": "该报名记录不存在"})
			return
		}

		// 4. 返回成功响应
		c.JSON(http.StatusOK, gin.H{"message": "删除报名成功"})
	}
}

// AdminDeleteRegistrationHandler 是一个正确的、供管理员使用的删除函数
// 它在一个事务中处理删除操作并同步更新活动人数
// func AdminDeleteRegistrationHandler(db *sql.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// 从 URL 获取要删除的报名记录 ID
// 		registrationID, err := strconv.Atoi(c.Param("id")) // 假设路由是 /registrations/:id
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的报名ID"})
// 			return
// 		}

// 		// --- 核心逻辑：使用事务 ---
// 		tx, err := db.Begin()
// 		if err != nil {
// 			log.Printf("开启事务失败: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
// 			return
// 		}
// 		defer tx.Rollback() // 保证出错时回滚

// 		// 1. 在删除前，获取该报名的状态和关联的活动ID
// 		var status string
// 		var activityID int
// 		querySelect := "SELECT status, activity_id FROM registrations WHERE id = ? FOR UPDATE"
// 		err = tx.QueryRow(querySelect, registrationID).Scan(&status, &activityID)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				c.JSON(http.StatusNotFound, gin.H{"error": "该报名记录不存在"})
// 				return
// 			}
// 			log.Printf("查询报名信息失败: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
// 			return
// 		}

// 		// 2. 执行删除操作 (没有 user_id 校验)
// 		_, err = tx.Exec("DELETE FROM registrations WHERE id = ?", registrationID)
// 		if err != nil {
// 			log.Printf("删除报名记录失败: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
// 			return
// 		}

// 		// 3. 如果被删除的记录是 "approved" 状态，则需要将活动人数减 1
// 		if status == "approved" {
// 			_, err = tx.Exec("UPDATE activities SET registered_count = registered_count - 1 WHERE id = ? AND registered_count > 0", activityID)
// 			if err != nil {
// 				log.Printf("更新活动人数失败: %v", err)
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "更新活动人数失败"})
// 				return
// 			}
// 		}

// 		// 4. 提交事务
// 		if err := tx.Commit(); err != nil {
// 			log.Printf("提交事务失败: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "删除报名成功"})
// 	}
// }
