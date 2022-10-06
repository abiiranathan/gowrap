package gowrap_test

import (
	"errors"
	"testing"
	"time"

	"github.com/abiiranathan/gowrap"
)

type Post struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	Title    string    `json:"title" validate:"required,max=255"`
	CreateAt time.Time `json:"created_at" validate:"omitempty,datetime"`
}

func TestORM(t *testing.T) {
	t.Parallel()

	db := gowrap.ConnectToSqlite3(gowrap.MemorySQLiteDB, true)
	err := db.AutoMigrate(&Post{})

	if err != nil {
		t.Fatal("unable on gorm automigrate")
	}

	orm := gowrap.New(db)

	// Test database insert
	p := &Post{Title: "My first post", CreateAt: time.Now()}
	err = orm.Insert(p)

	if err != nil {
		t.Errorf("db insert failed with error: %v", err)
	}

	if p.ID == 0 {
		t.Error("primaryKey ID not populated")
	}

	// Test update
	p.Title = "My first updated post"
	err = orm.Update(&p)
	if err != nil {
		t.Errorf("db update failed with error: %v", err)
	}

	// Test partial update
	partialUpdateTitle := "My new partially updated post"
	err = orm.PartialUpdate(&p, Post{Title: partialUpdateTitle}, gowrap.Where{
		Query: "id=?",
		Args:  []any{p.ID},
	})

	if err != nil {
		t.Errorf("db partial update failed with error: %v", err)
	}

	if p.Title != partialUpdateTitle {
		t.Errorf("failed to update post title with new title")
	}

	// test record that do not exists should raise errors
	err = orm.PartialUpdate(&Post{ID: 20}, Post{Title: "new title"}, gowrap.Where{Query: "title=?", Args: []any{"Post"}})
	if !errors.Is(err, gowrap.ErrNoRecordsUpdated) {
		t.Error("expected error of type:ErrNoRecordsUpdated ")
	}

	// Get a single record by id
	post := &Post{}
	err = orm.First(post, p.ID)
	if err != nil {
		t.Errorf("db first failed with error: %v", err)
	}

	if post.ID != p.ID {
		t.Errorf("db first returned a post not does not match id provided: %#v and %#v", p.ID, post.ID)
	}

	// Get a single record by custom column name
	foundPost := &Post{}
	err = orm.FindOne(foundPost, gowrap.Where{Query: "title LIKE ?", Args: []any{"%%My new partially%%"}})
	if err != nil {
		t.Errorf("db FindOne failed with error: %v", err)
	}

	if foundPost.ID != post.ID {
		t.Errorf("findOne returned a wrong post")
	}

	// Fetch all records from the database
	posts := []Post{}

	err = orm.FindAll(&posts)
	if err != nil {
		t.Errorf("db findAll failed with error: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("expected len(posts) to be 1, got: %d", len(posts))
	}

	// Fetch paginated
	results, err := gowrap.Paginate(&Post{}, 1, 10, db)
	if err != nil {
		t.Errorf("Paginate failed with error: %v", err)
	}

	// If page is zero, it's increased to 1
	results, _ = gowrap.Paginate(&Post{}, 0, 1, db)
	if results.Page != 1 {
		t.Fail()
	}

	if results.Count != 1 || results.Results[0].ID != post.ID {
		t.Errorf("Paginate returned inconsistent count or object id")
	}

	// Test delete
	err = orm.Delete(post)
	if err != nil {
		t.Errorf("orm Delete failed with error: %v", err)
	}

	checkError := func(err error) {
		if !errors.Is(err, gowrap.ErrNotPointer) {
			t.Errorf("expected ErrNotPointer, got %v", err)
		}
	}

	err = orm.Insert(Post{})
	checkError(err)

	err = orm.Update(Post{})
	checkError(err)

	err = orm.PartialUpdate(Post{}, Post{}, gowrap.Where{})
	checkError(err)

	err = orm.First(Post{}, 1)
	checkError(err)

	err = orm.FindOne(Post{}, gowrap.Where{Query: "id=?", Args: []any{1}})
	checkError(err)

	err = orm.FindAll(Post{})
	checkError(err)

	err = orm.Delete(Post{})
	checkError(err)
}
