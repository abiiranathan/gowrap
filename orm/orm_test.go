package orm_test

import (
	"errors"
	"testing"
	"time"

	"github.com/abiiranathan/gowrap/orm"
)

type Post struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	Title    string    `json:"title" validate:"required,max=255"`
	CreateAt time.Time `json:"created_at" validate:"omitempty,datetime"`
	Comments []Comment `json:"comments"`
}

type Comment struct {
	ID       uint      `json:"id"`
	PostID   uint      `json:"post_id"`
	CreateAt time.Time `json:"created_at"`
}

func TestORM(t *testing.T) {
	t.Parallel()

	db := orm.ConnectToSqlite3(orm.MemorySQLiteDB, true)
	err := db.AutoMigrate(&Post{}, &Comment{})

	if err != nil {
		t.Fatalf("unable to to run gorm automigrate: %v", err)
	}

	dborm := orm.New(db)

	// Test database insert
	p := &Post{Title: "My first post", CreateAt: db.NowFunc()}
	err = dborm.Insert(p)

	if err != nil {
		t.Errorf("db insert failed with error: %v", err)
	}

	if p.ID == 0 {
		t.Error("primaryKey ID not populated")
	}

	// Test update
	p.Title = "My first updated post"
	err = dborm.Update(&p)
	if err != nil {
		t.Errorf("db update failed with error: %v", err)
	}

	// Test partial update
	partialUpdateTitle := "My new partially updated post"
	err = dborm.PartialUpdate(&p, Post{Title: partialUpdateTitle}, orm.Where{
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
	err = dborm.PartialUpdate(&Post{ID: 20}, Post{Title: "new title"}, orm.Where{Query: "title=?", Args: []any{"Post"}})
	if !errors.Is(err, orm.ErrNoRecordsUpdated) {
		t.Error("expected error of type:ErrNoRecordsUpdated ")
	}

	// Get a single record by id
	post := &Post{}
	err = dborm.First(post, p.ID, orm.Limit{L: 1}, orm.Select{Fields: []any{"id", "title"}})

	if err != nil {
		t.Errorf("db first failed with error: %v", err)
	}

	if post.ID != p.ID {
		t.Errorf("db first returned a post not does not match id provided: %#v and %#v", p.ID, post.ID)
	}

	// Get a single record by custom column name
	foundPost := &Post{}
	err = dborm.FindOne(foundPost, orm.Where{Query: "title LIKE ?", Args: []any{"%%My new partially%%"}}, orm.Select{Fields: []any{"id"}})
	if err != nil {
		t.Errorf("db FindOne failed with error: %v", err)
	}

	if foundPost.ID != post.ID {
		t.Errorf("findOne returned a wrong post")
	}

	// Fetch all records from the database
	posts := []Post{}

	err = dborm.FindAll(&posts, orm.Select{}, orm.Order{Name: "id DESC"})
	if err != nil {
		t.Errorf("db findAll failed with error: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("expected len(posts) to be 1, got: %d", len(posts))
	}

	// Fetch paginated
	results, err := orm.Paginate(&Post{}, 1, 10, db)
	if err != nil {
		t.Errorf("Paginate failed with error: %v", err)
	}

	// Joins
	err = dborm.FindAll(&posts, orm.Preload{Query: "Comments"})
	if err != nil {
		t.Error(err)
	}

	// If page is zero, it's increased to 1
	results, _ = orm.Paginate(&Post{}, 0, 1, db)
	if results.Page != 1 {
		t.Fail()
	}

	if results.Count != 1 || results.Results[0].ID != post.ID {
		t.Errorf("Paginate returned inconsistent count or object id")
	}

	// Test delete
	err = dborm.Delete(post)
	if err != nil {
		t.Errorf("orm Delete failed with error: %v", err)
	}

	checkError := func(err error) {
		if !errors.Is(err, orm.ErrNotPointer) {
			t.Errorf("expected ErrNotPointer, got %v", err)
		}
	}

	err = dborm.Insert(Post{})
	checkError(err)

	err = dborm.Update(Post{})
	checkError(err)

	err = dborm.PartialUpdate(Post{}, Post{}, orm.Where{})
	checkError(err)

	err = dborm.First(Post{}, 1)
	checkError(err)

	err = dborm.FindOne(Post{}, orm.Where{Query: "id=?", Args: []any{1}})
	checkError(err)

	err = dborm.FindAll(Post{})
	checkError(err)

	err = dborm.Delete(Post{})
	checkError(err)

	// Save new comment
	c := Comment{CreateAt: db.NowFunc()}
	dborm.Insert(&c)
}
