package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupSQLMock() (*sql.DB, sqlmock.Sqlmock, func(), error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, nil, err
	}

	// グローバル変数dbにモックDBを設定
	originalDB := db
	db = mockDB

	// クリーンアップ関数を返す
	cleanup := func() {
		db = originalDB
		mockDB.Close()
	}

	return mockDB, mock, cleanup, nil
}
func TestUpdateOneTask(t *testing.T) {

	mockDB, mock, cleanup, err := setupSQLMock()
	if err != nil {
		t.Fatalf("モックDBの作成に失敗しました: %s", err)
	}
	defer cleanup()

	// グローバル変数dbにモックDBを設定
	originalDB := db
	db = mockDB
	defer func() { db = originalDB }()

	// モックの期待値を設定
	mock.ExpectExec(".*").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// ginのテスト用コンテキストを作成
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// テスト用のJSONデータを作成
	input := gin.H{
		"tasks": []gin.H{
			{
				"id":       4,
				"title":    "cuuat",
				"content":  "内容1",
				"due":      "2024-02-23T12:00:00Z",
				"priority": 1,
			},
		},
	}
	jsonData, _ := json.Marshal(input)

	// リクエストを作成
	req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	// updateTasklist関数を呼び出し
	updateTasklist(ctx)

	// 期待されるJSON出力
	expectedOutput := `{
		"failedUpdates": null,
		"successUpdates": [
			{
				"id": 4,
				"title": "cuuat",
				"content": "内容1",
				"due": "2024-02-23T12:00:00Z",
				"priority": 1
			}
		]
	}`

	// レスポンスの検証
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, expectedOutput, w.Body.String())

	// モックの期待値を検証
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("モックの期待値が満たされていません: %s", err)
	}
}

func TestUpdateMultipleTask(t *testing.T) {
	mockDB, mock, cleanup, err := setupSQLMock()
	if err != nil {
		t.Fatalf("モックDBの作成に失敗しました: %s", err)
	}
	// グローバル変数dbにモックDBを設定
	originalDB := db
	db = mockDB
	defer func() { db = originalDB }()
	defer cleanup()

	// モックの期待値を設定
	mock.ExpectExec(".*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(".*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// ginのテスト用コンテキストを作成
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// テスト用のJSONデータを作成
	input := gin.H{
		"tasks": []gin.H{
			{
				"id":       4,
				"title":    "cuuat",
				"content":  "内容1",
				"due":      "2024-02-23T12:00:00Z",
				"priority": 1,
			},
			{
				"id":       5,
				"title":    "更新内容5",
				"content":  "内容5",
				"due":      "2024-02-23T12:00:00Z",
				"priority": 1,
			},
		},
	}
	jsonData, _ := json.Marshal(input)

	// リクエストを作成
	req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	// updateTasklist関数を呼び出し
	updateTasklist(ctx)

	// 期待されるJSON出力
	expectedOutput := `{
		"failedUpdates": null,
		"successUpdates": [
			{
				"id": 4,
				"title": "cuuat",
				"content": "内容1",
				"due": "2024-02-23T12:00:00Z",
				"priority": 1
			},
			{
				"id": 5,
				"title": "更新内容5",
				"content": "内容5",
				"due": "2024-02-23T12:00:00Z",
				"priority": 1
			}
		]
	}`

	// レスポンスの検証
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, expectedOutput, w.Body.String())

	// モックの期待値を検証
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("モックの期待値が満たされていません: %s", err)
	}
}

func TestGetValidationMessage(t *testing.T) {
	// main.goのvalidationMessagesを使用
	tests := []struct {
		field    string
		tag      string
		expected string
	}{
		{"Title", "required", "Titleは必須です"},
		{"Content", "max", "Contentは255文字以内で入力してください"},
		{"Due", "unknown", "不正な入力です"},
		{"Priority", "number", "Priorityは数値で入力してください"},
	}

	for _, test := range tests {
		t.Run(test.field+"_"+test.tag, func(t *testing.T) {
			actual := getValidationMessage(test.field, test.tag, validationMessages)
			assert.Equal(t, test.expected, actual)
		})
	}
}
