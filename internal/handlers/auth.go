// 用户操作逻辑，使用 User 和 UserRegistration 模型
package handlers

import (
	"campus-activity-api/internal/models"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB // 全局变量，由main.go注入，handlers整个包共享

// 注册功能处理
func Register(c *gin.Context) {
	// 定义匿名结构体来接受前端传回来的json
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
		College  string `json:"college"`
	}

	// json结构不对或者类型不匹配返回 400 Bad Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的注册信息"})
		return
	}

	// 基本验证
	if len(req.Username) < 4 || len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名至少4位，密码至少6位"})
		return
	}

	// 使用 bcrypt 库对密码进行不可逆哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 将新用户插入数据库
	stmt, err := DB.Prepare("INSERT INTO users(username, password_hash, full_name, college, role) VALUES(?, ?, ?, ?, 'student')")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库准备失败"})
		return
	}
	defer stmt.Close()

	// 将哈希之后的密码插入数据库
	_, err = stmt.Exec(req.Username, string(hashedPassword), req.FullName, req.College)
	if err != nil {
		// 检查是否是唯一键冲突错误 (用户名已存在)
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "注册成功"})
}

// 登录处理
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	// 从数据库中查询用户，这次需要包含 password_hash
	var user models.User
	// 查询语句中增加了 password_hash
	err := DB.QueryRow("SELECT id, username, password_hash, full_name, college, role FROM users WHERE username = ?", req.Username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.FullName, &user.College, &user.Role)
	if err != nil {
		// 无论是用户不存在还是其他数据库错误，都返回统一的错误信息
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 使用 bcrypt 比较哈希值和用户输入的明文密码
	// 第一个参数是从数据库取出的哈希值
	// 第二个参数是用户在登录框输入的原始密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		// 如果 err 不为 nil，说明密码不匹配
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 密码验证通过，返回成功信息 (不包含密码哈希)
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"fullName": user.FullName,
			"college":  user.College,
			"role":     user.Role,
		},
	})
}

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

// 用户报名活动处理
func RegisterForActivity(c *gin.Context) {
	var req struct {
		UserID int `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求"})
		return
	}
	activityID := c.Param("id")
	// 将报名信息插入注册信息表
	stmt, err := DB.Prepare("INSERT INTO registrations(user_id, activity_id) VALUES(?, ?)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "报名失败"})
		return
	}
	defer stmt.Close()
	// 检查约束
	_, err = stmt.Exec(req.UserID, activityID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "你已经报过名了"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "报名成功"})
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
