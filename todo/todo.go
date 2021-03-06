package todo

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	// because we have to connect db here
	_ "github.com/mattn/go-sqlite3"
)

// Todos ...
type Todos struct {
	ID          int    `gorm:"AUTO_INCREMENT" form:"id" json:"id"`
	Title       string `gorm:"not null" form:"title" json:"title"`
	Time        int64  `gorm:"not null" form:"time" json:"time"`
	Priority    int8   `gorm:"not null" form:"priority" json:"priority"`
	CreatedDate int64  `gorm:"not null" form:"created_date" json:"created_date"`
	UpdatedDate int64  `gorm:"not null" form:"updated_date" json:"updated_date"`
	UserID      string `gorm:"not null" form:"user_id" json:"user_id"`
	Status      int8   `gorm:"not null" form:"status" json:"status"`
}

// Item ...
type Item struct {
	ID          int
	Title       string
	Time        time.Time
	Priority    string
	CreatedDate time.Time
	UpdatedDate time.Time
	Status      string
}

var status = map[int8]string{
	0: "Complete",
	1: "In complete",
	3: "Cancel",
}
var prioritys = map[int8]string{
	0: "important",
	1: "high",
	2: "medium",
	3: "low",
}

func dropDB() {
	db, err := gorm.Open("sqlite3", "./todo.db")
	if err != nil {
		panic(err)
	}
	db.DropTable(&Todos{})
}

func initDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "./todo.db")
	if err != nil {
		log.Fatalln("connect db sqlite error")
	}

	if !db.HasTable(&Todos{}) {
		db.CreateTable(&Todos{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&Todos{})
		log.Println("Create table success")
	}

	return db
}

// Save todo
func Save(userID, msg string) error {
	todo, err := parseText(msg)
	todo.UserID = userID
	db := initDB()
	if err != nil {
		return err
	}
	db.Create(&todo)
	return nil
}

// GetAll get all todo
func GetAll() []Todos {
	db := initDB()

	var todos []Todos
	db.Find(&todos)

	return todos
}

// Get todo by id
func Get(id int) Todos {
	db := initDB()

	var todo Todos
	db.Where("id = ?", id).First(&todo)

	return todo
}

// GetByUserID ...
func GetByUserID(userID string) []Item {
	db := initDB()

	var todos []Todos
	db.Where("user_id = ?", userID).Order("priority asc").Find(&todos)

	newTodos := []Item{}

	for _, t := range todos {
		newTodo := Item{
			ID:          t.ID,
			Title:       t.Title,
			Priority:    prioritys[t.Priority],
			Time:        time.Unix(t.Time, 0),
			Status:      status[t.Status],
			CreatedDate: time.Unix(t.CreatedDate, 0),
			UpdatedDate: time.Unix(t.UpdatedDate, 0),
		}
		newTodos = append(newTodos, newTodo)
	}

	return newTodos
}

// Delete todo by id
func Delete(id int) {
	db := initDB()
	var todo Todos
	db.First(&todo, id)
	db.Delete(&todo)
}

// DropDB for remove all record
func DropDB() {
	dropDB()
}

func parseText(text string) (Todos, error) {

	msgs := strings.Split(text, ":")
	newMsgs := []string{}

	for _, t := range msgs {
		v := strings.Trim(t, " ")
		if len(v) > 0 {
			newMsgs = append(newMsgs, v)
		}
	}

	if len(newMsgs) < 2 {
		return Todos{}, errors.New("invalid format eg. task : eat : 12:00")
	}
	if len(newMsgs) > 4 {
		return Todos{}, errors.New("invalid format eg. task : eat : 12:00")
	}

	if len(newMsgs) == 2 {
		newMsgs = append(newMsgs, "12", "00")
	}

	t := time.Now()

	taskHour, _ := strconv.Atoi(newMsgs[2])
	taskMinute, _ := strconv.Atoi(newMsgs[3])

	taskTime := time.Date(t.Year(), t.Month(), t.Day(), taskHour, taskMinute, 0, 0, t.Location())

	todo := Todos{
		Title:       newMsgs[1],
		Priority:    2,
		Time:        taskTime.Unix(),
		CreatedDate: time.Now().Unix(),
		UpdatedDate: time.Now().Unix(),
		Status:      1,
	}

	return todo, nil

}
