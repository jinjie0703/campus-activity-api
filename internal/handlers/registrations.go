// 报名信息管理模块，用于管理员界面，使用了 RegistrationDetails 和 RegistrationDetailsForActivity 模型
package handlers

import (
	"campus-activity-api/internal/models"
	"database/sql"
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

// 更新某个报名记录的 status 字段（比如：已通过、已拒绝、待审核等）
func UpdateRegistrationStatus(db *sql.DB, registrationID int, status string) error {
	query := "UPDATE registrations SET status = ? WHERE id = ?"
	_, err := db.Exec(query, status, registrationID)
	return err
}

// 根据 registrationID 删除报名记录
func DeleteRegistration(db *sql.DB, registrationID int) error {
	query := "DELETE FROM registrations WHERE id = ?"
	_, err := db.Exec(query, registrationID)
	return err
}
