package users

import (
	"log"
	"time"

	"github.com/dchest/uniuri"
	"github.com/jinzhu/gorm"
	// because we have to connect db here
	_ "github.com/mattn/go-sqlite3"
)

// Users ...
type Users struct {
	ID           int    `gorm:"AUTO_INCREMENT" form:"id" json:"id"`
	UserID       string `gorm:"not null" form:"user_id" json:"user_id"`
	AccessToken  string `gorm:"not null" form:"access_token" json:"access_token"`
	ExpireIn     int64  `gorm:"not null" form:"expire_in" json:"expire_in"`
	RefreshToken string `gorm:"not null" form:"refresh_token" json:"refresh_token"`
	CreatedDate  int64  `gorm:"not null" form:"created_date" json:"created_date"`
	UpdatedDate  int64  `gorm:"not null" form:"updated_date" json:"updated_date"`
}

func dropDB() {
	db, err := gorm.Open("sqlite3", "./user.db")
	if err != nil {
		panic(err)
	}
	db.DropTable(&Users{})
}

func initDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "./user.db")
	if err != nil {
		log.Fatalln("connect db sqlite error")
	}

	if !db.HasTable(&Users{}) {
		db.CreateTable(&Users{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&Users{})
		log.Println("Create table success")
	}

	return db
}

// Save todo
func Save(accessToken, refreshToken string, expireIn int64) string {
	db := initDB()

	userID := getUserID()

	user := Users{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpireIn:     time.Now().Unix() + expireIn,
		CreatedDate:  time.Now().Unix(),
		UpdatedDate:  time.Now().Unix(),
	}
	db.Create(&user)
	return userID
}

// Get todo by id
func Get(userID string) Users {
	db := initDB()
	var user Users
	db.Where("user_id = ?", userID).First(&user)
	return user
}

// GetAll users
func GetAll() []Users {
	db := initDB()
	var users []Users
	db.Find(&users)

	return users
}

// DropDB for remove all record
func DropDB() {
	dropDB()
}

func getUserID() string {
	id := uniuri.New()
	return id
}
