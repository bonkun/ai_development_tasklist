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
	Priority_name string    `json:"priority_name,omitempty"`
	Progress_name string    `json:"progress_name,omitempty"`
	Position      float64   `json:"position,omitempty"`
}

type User struct {
	Id            *int   `json:"id" binding:"omitempty"`
	Username      string `json:"username" binding:"required,max=255"`
	Password      string `json:"password" binding:"required,max=255"`
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

	rows, err := db.Query("SELECT tasklist.id, title, content, due, priority, progress_name, position FROM tasklist JOIN task_progress ON tasklist.progress_id = task_progress.id")
	if err != nil {
		log.Printf("データベースクエリエラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tasks []gin.H
	for rows.Next() {
		var task Tasklist
		err := rows.Scan(&task.Id, &task.Title, &task.Content, &task.Due, &task.Priority, &task.Progress_name, &task.Position)
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
			"priority_name": task.GetPriorityMessage(),
			"priority":      task.Priority,
			"progress_name": task.Progress_name,
			"position":      task.Position,
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

func convertPriorityLabelToId(priority string) int {
    switch priority {
    case "高":
        return 1
    case "中":
        return 2
    case "低":
        return 3
    default:
        return 0 // 不明な優先度の場合のデフォルト値
    }
}

func insertTasklist(c *gin.Context) {
    var request struct {
        Tasks []struct {
            Title       string `json:"title" binding:"required,max=255"`
            Content     string `json:"content" binding:"required,max=255"`
            Due         string `json:"due" binding:"required"`
			Priority    int    `json:"priority"`
            Priority_name      string    `json:"priority_name" binding:"required"`
            Progress_id   int    `json:"progress_id" binding:"required,min=0,max=100"`
			Position      float64 `json:"position"`
        } `json:"tasks" binding:"required,dive"`
    }


	// リクエストボディをログに
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
		log.Printf("JSONバインドエラー: %v", err)
		var errorMessages []string
		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, e := range ve {
				message := getValidationMessage(e.Field(), e.Tag(), validationMessages)
				errorMessages = append(errorMessages, message)
			}
		} else {
			errorMessages = append(errorMessages, "不正なリクエスト形式です")
		}

		log.Printf("バインドエラー詳細: %v", errorMessages)
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
		log.Printf("優先度: %d", task.Priority)
		log.Printf("進捗ID: %d", task.Progress_id)
		log.Printf("位置: %f", task.Position)
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
		// 文字列の優先順位を整数に変換
		task.Priority = convertPriorityLabelToId(task.Priority_name)

		log.Printf("挿入するタスク: タイトル=%s, 内=%s, 期限=%v, 優先度=%d, 進捗ID=%d, 位置=%f", task.Title, task.Content, task.Due, task.Priority, task.Progress_id, task.Position)
		result, err := tx.Exec("INSERT INTO tasklist (title, content, due, priority, progress_id, position) VALUES (?, ?, ?, ?, ?, ?)",
			task.Title, task.Content, task.Due, task.Priority, task.Progress_id, task.Position)
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
		log.Printf("挿入失敗: %v", failedInserts)
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

	log.Println("タスクの挿入に成功しました")
	c.JSON(http.StatusOK, gin.H{
		"successInserts": successInserts,
	})
}

func updateTasklist(c *gin.Context) {
    var request struct {
        Tasks []struct {
            Id           *int   `json:"id" binding:"required"`
            Title        string `json:"title" binding:"required,max=255"`
            Content      string `json:"content" binding:"required,max=255"`
            Due          string `json:"due" binding:"required"`
            PriorityName string `json:"priority_name"`
            Priority     int    `json:"priority"`
            Progress_id  int    `json:"progress_id" binding:"required,min=0,max=100"`
            Position     float64 `json:"position"`
        } `json:"tasks" binding:"required,dive"`
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
        log.Printf("JSONバインドエラー: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "不正なリクエスト形式です"})
        return
    }

    var successUpdates []gin.H
    var failedUpdates []gin.H

    for _, task := range request.Tasks {
        if task.Id == nil {
            failedUpdates = append(failedUpdates, gin.H{"id": nil, "error": "IDは必須です"})
            continue
        }

        // Convert priority_name to priority
        task.Priority = convertPriorityLabelToId(task.PriorityName)

        log.Printf("実行するSQL: UPDATE tasklist SET title=%s, content=%s, due=%v, priority=%d, progress_id=%d, position=%f WHERE id=%d", task.Title, task.Content, task.Due, task.Priority, task.Progress_id, task.Position, *task.Id)

        result, err := db.Exec(
            "UPDATE tasklist SET title=?, content=?, due=?, priority=?, progress_id=?, position=? WHERE id=?",
            task.Title,
            task.Content,
            task.Due,
            task.Priority,
            task.Progress_id,
            task.Position,
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

        successUpdates = append(successUpdates, gin.H{
            "id":          task.Id,
            "title":       task.Title,
            "content":     task.Content,
            "due":         task.Due,
            "priority":    task.Priority,
            "progress_id": task.Progress_id,
        })
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

func login(c *gin.Context) {
	var loginData User

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user User
	err := db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", loginData.Username).Scan(&user.Id, &user.Username, &user.Password)
	if err != nil || user.Password != loginData.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 簡単なトークンを生成（ここではユーザー名をそのまま使用）
	token := user.Username

	c.JSON(http.StatusOK, gin.H{"success": true, "token": token})
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

	r.GET("/tasklist", getTasklist)
	r.POST("/insert", insertTasklist)
	r.POST("/update", updateTasklist)
	r.GET("/delete/:id", deleteTasklist)
	r.POST("/login", login)
	r.Run(":8080")

}
