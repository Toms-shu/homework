package main

import (
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite" // 纯 Go SQLite 驱动
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
题目1：模型定义
假设你要开发一个博客系统，有以下几个实体： User （用户）、 Post （文章）、 Comment （评论）。
要求 ：
使用Gorm定义 User 、 Post 和 Comment 模型，
其中 User 与 Post 是一对多关系（一个用户可以发布多篇文章），
Post 与 Comment 也是一对多关系（一篇文章可以有多个评论）。
编写Go代码，使用Gorm创建这些模型对应的数据库表。
*/
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"uniqueIndex;size:255;not null"`
	Email     string    `gorm:"uniqueIndex;size:255;not null"`
	Age       uint      `gorm:"check:age>=0"`
	Posts     []Post    `gorm:"foreignKey:AuthorID"`
	PostCount uint      `gorm:"check:post_count>=0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Post struct {
	ID            uint      `gorm:"primaryKey"`
	Title         string    `gorm:"size:255;not null"`
	Content       string    `gorm:"type:text;not null"`
	AuthorID      uint      `gorm:"index;not null"`
	Author        User      `gorm:"foreignKey:AuthorID"`
	Comments      []Comment `gorm:"foreignKey:PostID;"`
	CommentStatus string    `gorm:"index;not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey"`
	PostID    uint      `gorm:"index;not null"`
	Post      Post      `gorm:"foreignKey:PostID;"`
	Content   string    `gorm:"size:1000;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type DataBase struct {
	db      *gorm.DB
	cleanUp func()
}

func insertUser(db *gorm.DB) {
	users := []User{
		{Name: "tom", Age: 33, Email: "694979892@qq.com"},
		//{Name: "mike", Age: 30, Email: "694979893@qq.com"},
		//{Name: "marrier", Age: 18, Email: "694979894@qq.com"},
	}
	if err := db.Create(&users).Error; err != nil {
		log.Fatal(err)
	}
}

func insertPost(db *gorm.DB) {
	posts := []Post{
		{Title: "post1", Content: "post content1", AuthorID: 1},
		//{Title: "post2", Content: "post content2", AuthorID: 1},
		//{Title: "post3", Content: "post content3", AuthorID: 1},
		//{Title: "post4", Content: "post content4", AuthorID: 2},
		//{Title: "post5", Content: "post content5", AuthorID: 2},
		//{Title: "post6", Content: "post content6", AuthorID: 2},
	}
	if err := db.Create(&posts).Error; err != nil {
		log.Fatal(err)
	}
}

func insertComment(db *gorm.DB) {
	comments := []Comment{
		{PostID: 1, Content: "comment 1"},
		//{PostID: 1, Content: "comment 2"},
		//{PostID: 2, Content: "comment 3"},
		//{PostID: 2, Content: "comment 4"},
	}

	if err := db.Create(&comments).Error; err != nil {
		log.Fatal(err)
	}
}

func deleteComment(db *gorm.DB, commentID uint) {
	//important,钩子函数执行需要完整的结构体数据，所以在执行钩子前，先查询一遍
	var comment Comment
	if err := db.First(&comment, commentID).Error; err != nil {
		log.Fatal(err)
	}

	if err := db.Delete(&comment).Error; err != nil {
		log.Fatal(err)
	}
	fmt.Println(commentID, "评论已删除")
}

func initData(db *gorm.DB) {
	//创建用户
	insertUser(db)

	//创建文章
	insertPost(db)

	//创建评论
	insertComment(db)
}

func ModelUse() *DataBase {
	dbPath := "./test.db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("获取数据库失败：", err)
	}

	//获取数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取数据库连接池失败：", err)
	}

	cleanUp := func() {
		if err := sqlDB.Close(); err != nil {
			log.Fatal("关闭连接失败：", err)
		}
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Printf("数据库连接成功%s\n", dbPath)
	log.Printf("开始自动迁移...")

	startTime := time.Now()
	err = db.AutoMigrate(
		&User{},
		&Post{},
		&Comment{},
	)
	if err != nil {
		log.Fatal("自动迁移失败：", err)
	}

	elapsed := time.Since(startTime)

	log.Printf("自动迁移结束！耗时：%v\n", elapsed)

	return &DataBase{
		db:      db,
		cleanUp: cleanUp,
	}

}

/*
题目2：关联查询
基于上述博客系统的模型定义。
要求 ：
编写Go代码，使用Gorm查询某个用户发布的所有文章及其对应的评论信息。
编写Go代码，使用Gorm查询评论数量最多的文章信息。
*/
func RelationQuery(db *gorm.DB) {
	//dbPath := "./test.db"
	//db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
	//	Logger: logger.Default.LogMode(logger.Info),
	//})
	//if err != nil {
	//	log.Fatal("数据库连接失败：", err)
	//}
	//
	//sqlDB, err := db.DB()
	//if err != nil {
	//	log.Fatal("连接池创建报错：", err)
	//}
	//defer sqlDB.Close()
	////最大空闲可用连接数
	//sqlDB.SetMaxIdleConns(10)
	////最大同时打开的连接数（包括空闲连接数）
	//sqlDB.SetMaxOpenConns(100)
	////定期刷新连接
	//sqlDB.SetConnMaxLifetime(time.Hour)
	////链接最大存活时间
	//sqlDB.SetConnMaxLifetime(time.Hour)
	//log.Printf("连接池设置完毕！")

	FindAllRelationInfos(db)
	FindMaxNumOfComments(db)

	////初始化测试数据
	//initData(db)
	//var users []User
	//var maxNum uint
	//maxNumPost := Post{}
	//maxNum = 0
	//db.Preload("Posts.Comments").Find(&users, "id<=3")
	//for _, user := range users {
	//	fmt.Printf("用户信息：姓名：%s, 邮箱：%s, 文章数:%d\n", user.Name, user.Email, len(user.Posts))
	//	for _, post := range user.Posts {
	//		//fmt.Printf("文章标题：%s, 评论：%d\n", post.Title, len(post.Comments))
	//		//fmt.Println(len(post.Comments))
	//		if int(maxNum) < len(post.Comments) {
	//			maxNum = uint(len(post.Comments))
	//			fmt.Printf("文章：%s,评论数：%d\n", post.Title, len(post.Comments))
	//			maxNumPost = post
	//		}
	//	}
	//}
	//fmt.Printf("最多评论的文章信息：评论数量%d; 文章标题:%s\n", maxNum, maxNumPost.Title)
}

func FindAllRelationInfos(db *gorm.DB) {
	var user User
	db.Preload("Posts.Comments").Find(&user, "id=5")
	fmt.Printf("用户信息：名称：%s, ID：%d, 年龄：%d, 邮箱：%s\n", user.Name, user.ID, user.Age, user.Email)
	for _, post := range user.Posts {
		fmt.Printf("文章%d: 标题：%s\n", post.ID, post.Title)
		for _, comment := range post.Comments {
			fmt.Printf("评论信息：%s\n", comment.Content)
		}
		fmt.Println("*******************")
	}
}

func FindMaxNumOfComments(db *gorm.DB) {
	type Result struct {
		PostID uint
		MaxNum uint
	}
	var r Result
	//用scan方法序列化到结构体的时候，select 后面字段需要对应上结构体字段顺序
	result := db.Table("comments").
		Select("post_id, COUNT(*) as maxNum").
		Group("post_id").
		Order("maxNum desc").
		Limit(1).
		Scan(&r)

	if result.Error != nil {
		fmt.Println(result.Error)
	} else if result.RowsAffected == 0 {
		fmt.Println("没查询到结果")
	} else {
		fmt.Printf("文章ID: %d， 评论数量：%d\n", r.PostID, r.MaxNum)
	}

	var p Post
	db.Table("posts").First(&p, r.PostID)
	fmt.Println("文章信息：", p)
}

/*
题目3：钩子函数
继续使用博客系统的模型。
要求 ：
为 Post 模型添加一个钩子函数，在文章创建时自动更新用户的文章数量统计字段。
为 Comment 模型添加一个钩子函数，在评论删除时检查文章的评论数量，如果评论数量为 0，则更新文章的评论状态为 "无评论"。
*/
func HookFunction(db *gorm.DB) {
	//insertUser(db)
	//insertPost(db)
	//insertComment(db)
	deleteComment(db, 2)
}

func (p *Post) AfterCreate(db *gorm.DB) error {
	//在钩子函数中，不会预加载关联对象，需要手动查询一遍才行，否则，是用的时候是空的
	//fmt.Println("p>>>>>", p.Author.ID)
	//fmt.Println(p.Author.Name, "的文章创建成功：", p.Title)
	var userName string
	result := db.Model(&User{}).Where("id=?", p.AuthorID).
		Pluck("name", &userName)
	if result.Error != nil {
		return result.Error
	}

	fmt.Println(userName, "的文章创建成功：", p.Title)
	return db.Transaction(func(db *gorm.DB) error {
		if err := db.Model(&User{}).
			Where("id = ?", p.AuthorID).
			Update("post_count", gorm.Expr("post_count + 1")).Error; err != nil {
			return err
		}

		return nil
	})

}

func (p *Post) AfterDelete(db *gorm.DB) error {
	fmt.Println(p.Author.Name, "的文章删除成功：", p.Title)

	if err := db.Model(&User{}).Where("id = ?", p.AuthorID).
		Update("post_count", gorm.Expr("post_count - 1")).Error; err != nil {
		return err
	}

	return nil
}

func (c *Comment) BeforeDelete(db *gorm.DB) error {
	var count int64
	result := db.Model(&Comment{}).
		Where("post_id=?", c.PostID).Count(&count)
	if result.Error != nil {
		return result.Error
	}

	fmt.Println("count:", count)
	//使用BeforeDelete钩子，事前检查和更新动作可以一起做
	if count == 1 {
		if err := db.Model(&Post{}).
			Where("id=?", c.PostID).
			Update("comment_status", "无评论").Error; err != nil {
			return err
		}
	}
	return nil
}

func (c *Comment) AfterCreate(db *gorm.DB) error {
	var count int64
	result := db.Model(&Comment{}).Where("post_id=?", c.PostID).Count(&count)
	if result.Error != nil {
		return result.Error
	}

	if count > 0 {
		if err := db.Model(&Post{}).Where("id=?", c.PostID).
			Update("comment_status", "有评论").Error; err != nil {
			return err
		}
	}
	return nil
}

func main() {
	db := ModelUse()
	//RelationQuery(db)
	gdb := db.db
	cleanUp := db.cleanUp
	defer cleanUp()
	HookFunction(gdb)
}
