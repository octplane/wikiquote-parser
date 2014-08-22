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
  Title      string     `sql:"size:255"`
  Quotes     []Quote    `gorm:"many2many:page_quote;"`
  Categories []Category `gorm:"many2many:page_category;"`
}

type Category struct {
  Id     int64
  Text   string  `sql:"size:255"`
  Quotes []Quote `gorm:"many2many:category_quote;"`
}

type Quote struct {
  Id   int64
  Text string `sql:"size:1000"`
}

func Connect() gorm.DB {
  //db, err := gorm.Open("postgres", "user=gorm dbname=gorm sslmode=disable")
  // db, err := gorm.Open("mysql", "gorm:gorm@/gorm?charset=utf8&parseTime=True")
  db, err := gorm.Open("sqlite3", "/tmp/gorm.db")

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

  db.AutoMigrate(Page{})
  db.AutoMigrate(Category{})
  db.AutoMigrate(Quote{})

  return db
}

// type User struct {
//     Id           int64
//     Birthday     time.Time
//     Age          int64
//     Name         string  `sql:"size:255"`
//     CreatedAt    time.Time
//     UpdatedAt    time.Time
//     DeletedAt    time.Time

//     Emails            []Email         // Embedded structs (has many)
//     BillingAddress    Address         // Embedded struct (has one)
//     BillingAddressId  sql.NullInt64   // Foreign key of BillingAddress
//     ShippingAddress   Address         // Embedded struct (has one)
//     ShippingAddressId int64           // Foreign key of ShippingAddress
//     IgnoreMe          int64 `sql:"-"` // Ignore this field
//     Languages         []Language `gorm:"many2many:user_languages;"` // Many To Many, user_languages is the join table
// }

// type Email struct {
//     Id         int64
//     UserId     int64   // Foreign key for User (belongs to)
//     Email      string  `sql:"type:varchar(100);"` // Set field's type
//     Subscribed bool
// }

// type Address struct {
//     Id       int64
//     Address1 string         `sql:"not null;unique"` // Set field as not nullable and unique
//     Address2 string         `sql:"type:varchar(100);unique"`
//     Post     sql.NullString `sql:not null`
// }

// type Language struct {
//     Id   int64
//     Name string
// }
