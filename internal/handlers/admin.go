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

// 从数据库中获取所有用户报名信息
func GetAllRegistrations(db *sql.DB) ([]models.RegistrationDetails, error) {
	query := `
        SELECT
            r.id,
            r.activity_id,
            a.title,
            r.user_id,
            u.full_name,
            u.college,
            r.registration_time,
            r.status
        FROM registrations r
        JOIN users u ON r.user_id = u.id
        JOIN activities a ON r.activity_id = a.id
        ORDER BY r.registration_time DESC
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 创建空切片存放结果，遍历查询结果，放入切片
	var registrations []models.RegistrationDetails
	for rows.Next() {
		var reg models.RegistrationDetails
		if err := rows.Scan(
			&reg.RegistrationID,
			&reg.ActivityID,
			&reg.ActivityTitle,
			&reg.UserID,
			&reg.UserFullName,
			&reg.UserCollege,
			&reg.RegistrationTime,
			&reg.Status,
		); err != nil {
			return nil, err
		}
		registrations = append(registrations, reg)
	}
	return registrations, nil
}

// 获取系统中所有用户的报名信息
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

// 更新某个报名记录的 status 字段
func UpdateRegistrationStatus(db *sql.DB, registrationID int, status string) error {
	query := "UPDATE registrations SET status = ? WHERE id = ?"
	_, err := db.Exec(query, status, registrationID)
	return err
}

// 管理员更新报名状态的 Gin 处理函数
// 管理员修改某个报名的状态
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
		if req.Status != "approved" && req.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "状态值必须是 'approved' 或 'pending'"})
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

// 根据 activityID 查询该活动下的所有报名用户信息
func GetRegistrationsByActivityID(db *sql.DB, activityID int) ([]models.RegistrationDetailsForActivity, error) {
	query := `
        SELECT
            r.id, 
            u.id,
            u.username,
            u.full_name,
            u.college,
            r.registration_time,
            r.status 
        FROM registrations r
        JOIN users u ON r.user_id = u.id
        WHERE r.activity_id = ?
        ORDER BY r.registration_time ASC
    `
	rows, err := db.Query(query, activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 创建空切片存放结果，遍历查询结果，放入切片
	var registrants []models.RegistrationDetailsForActivity
	for rows.Next() {
		var reg models.RegistrationDetailsForActivity
		if err := rows.Scan(
			&reg.RegistrationID,
			&reg.UserID,
			&reg.Username,
			&reg.UserFullName,
			&reg.UserCollege,
			&reg.RegistrationTime,
			&reg.Status,
		); err != nil {
			return nil, err
		}
		registrants = append(registrants, reg)
	}
	return registrants, nil
}

// 根据活动 ID 获取该活动的所有报名者信息，并返回给前端
// 根据活动 ID 获取该活动的所有报名者信息
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

// 管理员删除某个用户的报名记录
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
