package models

import "time"

// 用户模型
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	FullName     string `json:"fullName"`
	College      string `json:"college"`
	Role         string `json:"role"`
}

// 活动模型
type Activity struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Organizer   string    `json:"organizer"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	Capacity    int       `json:"capacity"`
	CreatedByID int       `json:"createdById"`
}

// 用户注册信息模型
type UserRegistration struct {
	RegistrationID int       `json:"registrationId"`
	ActivityID     int       `json:"activityId"`
	Title          string    `json:"title"`
	Location       string    `json:"location"`
	StartTime      time.Time `json:"startTime"`
}

// 活动报名信息模型
type Registration struct {
	ID               int       `json:"id"`
	UserID           int       `json:"userId"`
	ActivityID       int       `json:"activityId"`
	RegistrationTime time.Time `json:"registrationTime"`
	Status           string    `json:"status"` // "pending", "approved", "rejected"
}

// 报名信息
type RegistrationDetails struct {
	RegistrationID   int       `json:"registrationId"`
	ActivityID       int       `json:"activityId"`
	ActivityTitle    string    `json:"activityTitle"`
	UserID           int       `json:"userId"`
	UserFullName     string    `json:"userFullName"`
	UserCollege      string    `json:"userCollege"`
	RegistrationTime time.Time `json:"registrationTime"`
	Status           string    `json:"status"`
}

// 显示特定活动报名者信息的视图模型
type RegistrationDetailsForActivity struct {
	RegistrationID   int       `json:"registrationId"` // 新增：报名记录ID
	UserID           int       `json:"userId"`
	Username         string    `json:"username"`
	UserFullName     string    `json:"fullName"`
	UserCollege      string    `json:"college"`
	RegistrationTime time.Time `json:"registrationTime"`
	Status           string    `json:"status"` // 新增：报名状态
}
