package main

import (
	//基本的にここに書いたものは使わないといけない...らしい...
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	//追加(ログイン)
	"golang.org/x/crypto/bcrypt"
)

//databaseの定義
type Books struct {
	gorm.Model
	Contents string
	Status   string
	Number   int
	Author   string
}

//ユーザー情報
type User struct {
	gorm.Model
	Username string
	Password string
}

// PasswordEncrypt パスワードをhash化
func PasswordEncrypt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CompareHashAndPassword hashと非hashパスワード比較
func CompareHashAndPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

//データベース初期化
func dbInit() {
	db, err := gorm.Open("sqlite3", "books.sqlite3")
	if err != nil {
		panic("データベース開けず！（dbInit）")
	}
	db.AutoMigrate(&Books{})
	defer db.Close()
}

//追加
func dbInsert(name string, status string, number int, author string) {
	db, err := gorm.Open("sqlite3", "books.sqlite3")
	if err != nil {
		panic("データベース開けず！（dbInsert)")
	}
	db.Create(&Books{Contents: name, Status: status, Number: number, Author: author})
	defer db.Close()
}

//全取得
func dbGetAll() []Books {
	db, err := gorm.Open("sqlite3", "books.sqlite3")
	if err != nil {
		panic("データベース開けず！(dbGetAll())")
	}
	var books []Books
	db.Order("created_at desc").Find(&books)
	db.Close()
	return books
}

//一つ取得
func dbGetOne(id int) Books {
	db, err := gorm.Open("sqlite3", "books.sqlite3")
	if err != nil {
		panic("データベース開けず！(dbGetOne())")
	}
	var books Books
	db.First(&books, id)
	db.Close()
	return books
}

//更新
func dbUpdate(id int, name string, status string, number int, author string) {
	db, err := gorm.Open("sqlite3", "books.sqlite3")
	if err != nil {
		panic("データベース開けず！（dbUpdate)")
	}
	var books Books
	db.First(&books, id)
	books.Contents = name
	books.Status = status
	books.Number = number
	books.Author = author
	db.Save(&books)
	db.Close()
}

//削除
func dbDelete(id int) {
	db, err := gorm.Open("sqlite3", "books.sqlite3")
	if err != nil {
		panic("データベース開けず！（dbDelete)")
	}
	var books Books
	db.First(&books, id)
	db.Delete(&books)
	db.Close()
}

//作者列挙
func dbGetAuthor() []Books {
	db, err := gorm.Open("sqlite3", "books.sqlite3")
	if err != nil {
		panic("データベース開けず！(dbGetAll())")
	}
	var authors []Books
	db.Select("Distinct(Author)").Find(&authors) //DistinctはSelect内で行う.
	db.Close()
	return authors
}

//作者1人の取得
func dbGetAuthor1(name string) Books {
	db, err := gorm.Open("sqlite3", "books.sqlite3")
	if err != nil {
		panic("データベース開けず！(dbGetOne())")
	}
	var author Books
	db.Where("Author = ?", name).First(&author)
	db.Close()
	return author
}

//ある作者が書いた漫画全部を取得
func dbAuthorDetail(name string) []Books {
	db, err := gorm.Open("sqlite3", "books.sqlite3")
	if err != nil {
		panic("データベースが開けません")
	}

	var author_detail []Books
	db.Where("Author = ?", name).Find(&author_detail)
	db.Close()
	return author_detail
}

//ユーザー情報取得
func getUser(username string) User {
	db, err := gorm.Open("sqlite3", "user.sqlite3")
	if err != nil {
		panic("データベース開けず！(dbGetOne())")
	}
	var user User
	db.Where("username = ?", username).First(&user)
	db.Close()
	return user
}

func main() {
	// router. でとにかく書けばルーティングはできる。
	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html") // htmlの適応許可
	router.Static("/assets", "./assets")    // cssを適応許可

	dbInit()

	//追加
	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "login.html", gin.H{})
	})

	//ログイン認証
	router.POST("/login", func(ctx *gin.Context) {
		//username := ctx.PostForm("username")
		DBpassword := getUser(ctx.PostForm("username")).Password
		FORMpassword := ctx.PostForm("password")

		if err := CompareHashAndPassword(DBpassword, FORMpassword); err != nil {
			ctx.HTML(200, "fail.html", gin.H{})
		} else {
			ctx.Redirect(302, "/main")
		}
	})

	router.GET("/main", func(ctx *gin.Context) {
		books := dbGetAll()
		ctx.HTML(200, "index.html", gin.H{
			"books": books,
		})
	})

	router.POST("/new", func(ctx *gin.Context) {
		name := ctx.PostForm("name")
		status := ctx.PostForm("status")
		var num = ctx.PostForm("number")
		number, _ := strconv.Atoi(num)
		author := ctx.PostForm("author")

		dbInsert(name, status, number, author)
		ctx.Redirect(302, "/")
	})

	router.GET("/detail/:id", func(ctx *gin.Context) {
		n := ctx.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic(err)
		}
		books := dbGetOne(id)
		ctx.HTML(200, "detail.html", gin.H{"books": books})
	})

	router.POST("/update/:id", func(ctx *gin.Context) {
		n := ctx.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic("ERROR")
		}
		name := ctx.PostForm("name")
		status := ctx.PostForm("status")
		var num = ctx.PostForm("number")
		number, _ := strconv.Atoi(num)
		author := ctx.PostForm("author")
		dbUpdate(id, name, status, number, author)
		ctx.Redirect(302, "/main")
	})

	router.GET("/delete_check/:id", func(ctx *gin.Context) {
		n := ctx.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic("ERROR")
		}
		books := dbGetOne(id)
		ctx.HTML(200, "delete.html", gin.H{
			"books": books,
		})
	})

	//削除
	router.POST("/delete/:id", func(ctx *gin.Context) {
		n := ctx.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic("ERROR")
		}
		dbDelete(id)

		ctx.Redirect(302, "/")
	})

	//ここから先は追加
	router.GET("/author", func(ctx *gin.Context) {
		authors := dbGetAuthor()
		ctx.HTML(200, "author.html", gin.H{"author": authors})
	})

	router.POST("/author_detail", func(ctx *gin.Context) {
		var name = ctx.PostForm("authorname")
		authorname := dbGetAuthor1(name)
		authorinfo := dbAuthorDetail(name)
		ctx.HTML(200, "author_detail.html", gin.H{
			"author": authorname,
			"info":   authorinfo,
		})
	})

	router.Run() //基本的には最後には [Run()] する
	//時々、ポート番号変えるとデバックしやすい。
}
