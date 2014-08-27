package internals

import (
  "github.com/golang/glog"
  "github.com/jinzhu/gorm"

  // _ "github.com/lib/pq"
  // _ "github.com/go-sql-driver/mysql"
  _ "github.com/mattn/go-sqlite3"
)

type Page struct {
  Id         int64
  ImportId   int
  Title      string     `sql:"size:255"`
  Quotes     []Quote    `gorm:"many2many:pagei_quotei;"`
  Categories []Category `gorm:"many2many:pagei_categoryi;"`
}

func (p Page) TableName() string {
  return "page_import"
}

type Category struct {
  Id   int64
  Text string `sql:"size:255"`
}

func (c Category) TableName() string {
  return "category_import"
}

type Quote struct {
  Id        int64
  Text      string `sql:"size:1000"`
  Author    string `sql:"size:1000"`
  Booktitle string `sql:"size:1000"`
  Isbn      string `sql:"size:15"`
}

func (q Quote) TableName() string {
  return "quote_import"
}

func Connect() gorm.DB {
  //db, err := gorm.Open("postgres", "user=gorm dbname=gorm sslmode=disable")
  // db, err := gorm.Open("mysql", "gorm:gorm@/gorm?charset=utf8&parseTime=True")
  db, err := gorm.Open("sqlite3", "./gorm.db")

  if err != nil {
    panic(err)
  }
  glog.V(2).Infoln("Connected")

  // Get database connection handle [*sql.DB](http://golang.org/pkg/database/sql/#DB)
  db.DB()

  // Then you could invoke `*sql.DB`'s functions with it
  db.DB().Ping()
  db.DB().SetMaxIdleConns(10)
  db.DB().SetMaxOpenConns(100)

  db.DropTableIfExists(Page{})
  db.DropTableIfExists(Category{})
  db.DropTableIfExists(Quote{})

  db.AutoMigrate(Page{})
  db.AutoMigrate(Category{})
  db.AutoMigrate(Quote{})

  return db
}
