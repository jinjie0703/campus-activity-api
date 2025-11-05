// 用户操作逻辑，使用 User 和 UserRegistration 模型
package handlers

import (
	"campus-activity-api/internal/config"
	"campus-activity-api/internal/models"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	stmt, err := DB.Prepare(
		"INSERT INTO users(username, password_hash, full_name, college, role) VALUES(?, ?, ?, ?, 'student')")
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
	// 定义匿名结构体来接受前端传回来的json
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// json结构不对或者类型不匹配返回 400 Bad Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	// 从数据库中查询用户，这次需要包含 password_hash
	var user models.User

	// 查询语句中增加了 password_hash
	err := DB.QueryRow(
		"SELECT id, username, password_hash, full_name, college, role FROM users WHERE username = ?", 
		req.Username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.FullName, &user.College, &user.Role)
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

	// 密码验证通过，生成JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
		// 过期时间设置为24小时
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	// 使用配置中的 secret 签名并获取完整的编码后的 token 字符串
	tokenString, err := token.SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法生成token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   tokenString,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"fullName": user.FullName,
			"college":  user.College,
			"role":     user.Role,
		},
	})
}
