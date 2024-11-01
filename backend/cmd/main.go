package main

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
)

type Tasklist struct {
	Id            *int      `json:"id" binding:"omitempty"`
	Title         string    `json:"title" binding:"required,max=255"`
	Content       string    `json:"content" binding:"required,max=255"`
	Due           time.Time `json:"due" binding:"required"`
	Priority      int       `json:"priority" binding:"required,min=1,max=3"`
	Progress_id   int       `json:"progress_id" binding:"required,min=0,max=100"`
	Progress_name string    `json:"progress_name,omitempty"`
}

type Progress_status struct {
	Id            *int   `json:"id" binding:"omitempty"`
	Progress_name string `json:"progress_name" binding:"required,max=255"`
}

var db *sql.DB

// 優先順位の定義
var priorityMessages = map[int]string{
	1: "高",
	2: "中",
	3: "低",
}

func (t Tasklist) GetPriorityMessage() string {
	if msg, ok := priorityMessages[t.Priority]; ok {
		return msg
	}
	return "不明"
}

// バリデーションメッセージのテンプレート
var validationMessages = map[string]string{
	"required": ":fieldは必須です",
	"max":      ":fieldは255文字以内で入力してください",
	"number":   ":fieldは数値で入力してください",
}

// getValidationMessage フィールド名とタグからメッセージを生成
func getValidationMessage(field, tag string, validationMessages map[string]string) string {
	if msg, ok := validationMessages[tag]; ok {
		return strings.Replace(msg, ":field", field, 1)
	}
	return "不正な入力です"
}

// 優先順位メッセージを取得するためのgetter関数
func GetPriorityMessages() map[int]string {
	// マップのコピーを返すことで、外部からの変更を防ぐ
	copy := make(map[int]string)
	for k, v := range priorityMessages {
		copy[k] = v
	}
	return copy
}
func getTasklist(c *gin.Context) {
	log.Println("getTasklist関数が呼び出されました")

	rows, err := db.Query("SELECT tasklist.id, title, content, due, priority, progress_name FROM tasklist JOIN task_progress ON tasklist.progress_id = task_progress.id")
	if err != nil {
		log.Printf("データベースクエリエラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tasks []gin.H
	for rows.Next() {
		var task Tasklist
		err := rows.Scan(&task.Id, &task.Title, &task.Content, &task.Due, &task.Priority, &task.Progress_name)
		if err != nil {
			log.Printf("データベーススキャンエラー: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		tasks = append(tasks, gin.H{
			"id":            task.Id,
			"title":         task.Title,
			"content":       task.Content,
			"due":           task.Due,
			"priority":      task.GetPriorityMessage(),
			"progress_name": task.Progress_name,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("行の処理中のエラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("タスクリストの取得に成功しました")
	c.JSON(http.StatusOK, tasks)
}

func insertTasklist(c *gin.Context) {
	var request struct {
		Tasks []Tasklist `json:"tasks" binding:"required,dive"`
	}

	// リクエストボディをログに出力
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("リクエストボディの読み取りエラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部サーバーエラー"})
		return
	}
	log.Printf("受信したJSON: %s", string(body))

	// JSONを再度バインドするためにリクエストボディをリセット
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err := c.ShouldBindJSON(&request); err != nil {
		var errorMessages []string
		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, e := range ve {
				message := getValidationMessage(e.Field(), e.Tag(), validationMessages)
				errorMessages = append(errorMessages, message)
			}
		} else {
			errorMessages = append(errorMessages, "不正なリクエスト形式です")
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"errors": errorMessages,
		})
		return
	}

	log.Printf("受信したタスク数: %d", len(request.Tasks))
	for _, task := range request.Tasks {
		log.Printf("タイトル: %s", task.Title)
		log.Printf("内容: %s", task.Content)
		log.Printf("期限: %v", task.Due)
	}

	var successInserts []gin.H
	var failedInserts []gin.H

	tx, err := db.Begin()
	if err != nil {
		log.Printf("トランザクション開始エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部サーバーエラー"})
		return
	}

	for _, task := range request.Tasks {
		result, err := tx.Exec("INSERT INTO tasklist (title, content, due, priority, progress_id) VALUES (?, ?, ?, ?, ?)",
			task.Title, task.Content, task.Due, task.Priority, task.Progress_id)
		if err != nil {
			log.Printf("データベース挿入エラー: %v", err)
			failedInserts = append(failedInserts, gin.H{"title": task.Title, "error": err.Error()})
			continue
		}

		id, err := result.LastInsertId()
		if err != nil {
			log.Printf("LastInsertId取得エラー: %v", err)
			failedInserts = append(failedInserts, gin.H{"title": task.Title, "error": "ID取得に失敗しました"})
			continue
		}

		successInserts = append(successInserts, gin.H{"id": id})
	}

	if len(failedInserts) > 0 {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"failedInserts":  failedInserts,
			"successInserts": successInserts,
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("トランザクションコミットエラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "トランザクションコミットに失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"successInserts": successInserts,
	})
}

func updateTasklist(c *gin.Context) {
	var request struct {
		Tasks []Tasklist `json:"tasks" binding:"required,dive"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		var errorMessages []string
		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, e := range ve {
				message := getValidationMessage(e.Field(), e.Tag(), validationMessages)
				errorMessages = append(errorMessages, message)
			}
		} else {
			errorMessages = append(errorMessages, "不正なリクエスト形式です")
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"errors": errorMessages,
		})
		return
	}

	log.Printf("受信したタスク数: %d", len(request.Tasks))
	for _, task := range request.Tasks {
		log.Printf("タスクID: %v, タイトル: %s", *task.Id, task.Title)
	}

	var successUpdates []Tasklist
	var failedUpdates []gin.H

	for _, task := range request.Tasks {
		if task.Id == nil {
			failedUpdates = append(failedUpdates, gin.H{"id": nil, "error": "IDは必須です"})
			continue
		}
		log.Printf("実行するSQL: UPDATE tasklist SET title=%s, content=%s, due=%v, priority=%d WHERE id=%d", task.Title, task.Content, task.Due, task.Priority, *task.Id)
		// データベースにタスクを更新
		result, err := db.Exec(
			"UPDATE tasklist SET title=?, content=?, due=?, priority=?, progress_id=? WHERE id=?",
			task.Title,
			task.Content,
			task.Due,
			task.Priority,
			task.Progress_id,
			task.Id,
		)
		if err != nil {
			log.Printf("データベース更新エラー: %v", err)
			failedUpdates = append(failedUpdates, gin.H{"id": *task.Id, "error": err.Error()})
			continue
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("RowsAffected取得エラー: %v", err)
			failedUpdates = append(failedUpdates, gin.H{"id": *task.Id, "error": "内部サーバーエラー"})
			continue
		}

		if rowsAffected == 0 {
			log.Printf("タスクが見つからないか、変更がありません: ID %d", *task.Id)
			failedUpdates = append(failedUpdates, gin.H{"id": *task.Id, "error": "指定されたIDのタスクが見つからないか、変更がありません"})
			continue
		}

		successUpdates = append(successUpdates, task)
	}

	response := gin.H{
		"successUpdates": successUpdates,
		"failedUpdates":  failedUpdates,
	}

	if len(failedUpdates) > 0 && len(successUpdates) > 0 {
		c.JSON(http.StatusPartialContent, response)
	} else if len(failedUpdates) > 0 {
		c.JSON(http.StatusBadRequest, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

func deleteTasklist(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec(
		"DELETE FROM tasklist WHERE id=?",
		id,
	)
	if err != nil {
		log.Printf("データベース削除エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("RowsAffected取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部サーバーエラー"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "指定されたIDのタスクが見つかりません"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "タスクが正常に削除されました",
	})
}

func main() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(db:3306)/myapp?parseTime=true&loc=Local")
	if err != nil {
		log.Printf("データベース接続エラー: %v", err)
		return
	}
	defer db.Close()

	r := gin.Default()

	// CORSミドルウェアを追加
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.POST("/cat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello Worl from Gin!",
		})
	})

	r.GET("/tasklist", getTasklist)
	r.POST("/insert", insertTasklist)
	r.POST("/update", updateTasklist)
	r.GET("/delete/:id", deleteTasklist)
	r.Run(":8080")

}
