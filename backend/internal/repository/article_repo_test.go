package repository

import (
	"testing"
	"time"

	"WeMediaSpider/backend/internal/database/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// 每个测试使用独立的内存数据库，避免跨测试数据污染
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	if err := db.AutoMigrate(&models.Account{}, &models.Article{}, &models.AppStats{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestArticleRepo_CreateAndFind(t *testing.T) {
	db := newTestDB(t)
	repo := NewArticleRepository(db)

	article := &models.Article{
		ArticleID:        "art001",
		Title:            "测试文章",
		AccountFakeid:    "fakeid001",
		AccountName:      "测试公众号",
		PublishTimestamp: time.Now().Unix(),
	}

	if err := repo.Create(article); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID("art001")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Title != "测试文章" {
		t.Errorf("got title %q, want %q", found.Title, "测试文章")
	}
}

func TestArticleRepo_BatchCreate_OnConflictUpdate(t *testing.T) {
	db := newTestDB(t)
	repo := NewArticleRepository(db)

	articles := []*models.Article{
		{ArticleID: "a1", Title: "原始标题", AccountFakeid: "fid", AccountName: "acc"},
		{ArticleID: "a2", Title: "文章2", AccountFakeid: "fid", AccountName: "acc"},
	}
	if err := repo.BatchCreate(articles); err != nil {
		t.Fatalf("BatchCreate failed: %v", err)
	}

	// 重复插入同一 article_id，标题应被更新
	updated := []*models.Article{
		{ArticleID: "a1", Title: "更新后标题", AccountFakeid: "fid", AccountName: "acc"},
	}
	if err := repo.BatchCreate(updated); err != nil {
		t.Fatalf("BatchCreate upsert failed: %v", err)
	}

	found, err := repo.FindByID("a1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Title != "更新后标题" {
		t.Errorf("upsert: got %q, want %q", found.Title, "更新后标题")
	}
}

func TestArticleRepo_Count(t *testing.T) {
	db := newTestDB(t)
	repo := NewArticleRepository(db)

	count, err := repo.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	repo.Create(&models.Article{ArticleID: "x1", AccountFakeid: "f", AccountName: "n"})
	repo.Create(&models.Article{ArticleID: "x2", AccountFakeid: "f", AccountName: "n"})

	count, err = repo.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestArticleRepo_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := NewArticleRepository(db)

	repo.Create(&models.Article{ArticleID: "del1", AccountFakeid: "f", AccountName: "n"})

	if err := repo.Delete("del1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.FindByID("del1")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestArticleRepo_Search(t *testing.T) {
	db := newTestDB(t)
	repo := NewArticleRepository(db)

	repo.Create(&models.Article{ArticleID: "s1", Title: "Go语言并发编程", AccountFakeid: "f", AccountName: "n"})
	repo.Create(&models.Article{ArticleID: "s2", Title: "Python入门教程", AccountFakeid: "f", AccountName: "n"})
	repo.Create(&models.Article{ArticleID: "s3", Title: "Go微服务架构", AccountFakeid: "f", AccountName: "n"})

	results, err := repo.Search("Go", 10, 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestArticleRepo_BatchCreate_Empty(t *testing.T) {
	db := newTestDB(t)
	repo := NewArticleRepository(db)

	if err := repo.BatchCreate(nil); err != nil {
		t.Errorf("BatchCreate(nil) should return nil, got %v", err)
	}
	if err := repo.BatchCreate([]*models.Article{}); err != nil {
		t.Errorf("BatchCreate([]) should return nil, got %v", err)
	}
}
