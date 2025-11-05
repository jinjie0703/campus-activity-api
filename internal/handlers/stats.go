// 热门活动排名 和 组织方扇形图 筛选处理，主要使用 Activity 和 Registration 模型
package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 查询报名人数最多的前 5 个热门活动
func GetHotActivities(c *gin.Context) {
	// 定义热门活动的结果结构体
	type HotActivityResult struct {
		Title             string `json:"title"`
		Organizer         string `json:"organizer"`
		RegistrationCount int    `json:"registrationCount"`
	}
	// sql解析：
	// 1.`FROM activities AS a`：主表是 activities，别名为 a。
	// 2.`LEFT JOIN registrations AS r ON a.id = r.activity_id`：
	// 		registrations 表记录用户报名活动的情况，每条记录对应一个 activity_id。
	// 		LEFT JOIN 保证即使某个活动没有报名，也会显示（报名数为 0）。
	// 3.`COUNT(r.id) AS registration_count`：统计每个活动的报名人数。
	// 4.`GROUP BY a.id`：按照活动 ID 聚合数据，每个活动得到一行。
	// 5.`ORDER BY registration_count DESC`：按照报名人数降序排序，热门活动排在前面。
	// 6.`LIMIT 5`：只取前 5 个热门活动。
	query := `
				SELECT 
						a.title, 
						a.organizer, 
				COUNT(r.id) AS registration_count 
				FROM activities AS a LEFT 
				JOIN registrations AS r ON a.id = r.activity_id 
				GROUP BY a.id ORDER BY registration_count DESC 
				LIMIT 5;
		`
	rows, err := DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询热门活动失败"})
		return
	}
	defer rows.Close()

	// 创建一个切片保存热门活动数据，返回前端
	results := []HotActivityResult{}
	for rows.Next() {
		var res HotActivityResult
		if err := rows.Scan(&res.Title, &res.Organizer, &res.RegistrationCount); err != nil {
			log.Println("扫描热门活动数据失败:", err)
			continue
		}
		results = append(results, res)
	}
	c.JSON(http.StatusOK, results)
}

// 统计每个组织者举办活动的数量
func GetOrganizerStats(c *gin.Context) {
	// 保存组织者的统计数据结构体
	type OrganizerStat struct {
		Organizer     string `json:"organizer"`
		ActivityCount int    `json:"activityCount"`
	}
	// sql解析：
	// 1.`SELECT organizer, COUNT(id) AS activity_count`
	// 	  organizer：选择组织者列。
	// 		COUNT(id)：统计每个组织者的活动数量（每一行 activities 表的一条活动记录有一个唯一 id）。
	// 		AS activity_count：给统计结果起别名为 activity_count。
	// 2.`FROM activities`：查询的表是 activities。
	// 3.`GROUP BY organizer`：按 organizer 分组，每个组织者只会得到一行结果。
	// 4.`ORDER BY activity_count DESC`：按照活动数量从高到低排序，方便前端展示热门组织者。
	query := `
				SELECT 
						organizer, 
				COUNT(id) AS activity_count 
				FROM activities GROUP BY organizer 
				ORDER BY activity_count DESC;
		`
	rows, err := DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询组织方数据失败"})
		return
	}
	defer rows.Close()

	// 创建一个切片保存组织方数据，返回前端
	stats := []OrganizerStat{}
	for rows.Next() {
		var s OrganizerStat
		if err := rows.Scan(&s.Organizer, &s.ActivityCount); err != nil {
			log.Println("扫描组织方数据失败:", err)
			continue
		}
		stats = append(stats, s)
	}
	c.JSON(http.StatusOK, stats)
}
