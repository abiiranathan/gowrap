# gowrap

Wrapper around gorm ORM to simplify queries, pagination 
and struct validation.

## Installation

```bash
go get github.com/abiiranathan/gowrap
```


## Usage

```go
package main

import (
    "github.com/abiiranathan/gowrap/orm"
)

type Post struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	Title    string    `json:"title" gorm:"not null; unique" validate:"required,max=255"`
	CreateAt time.Time `json:"created_at" validate:"omitempty"`
}


func main(){
    db := orm.ConnectToSqlite3(orm.MemorySQLiteDB, true)
	err := db.AutoMigrate(&Post{})

	if err != nil {
		t.Fatalf("unable to run gorm automigrate: %v", err)
	}
  
	dborm := orm.New(db)
    p := &Post{Title: "My first post", CreateAt: db.NowFunc()}
	err = dborm.Insert(p)
    ...
}

```