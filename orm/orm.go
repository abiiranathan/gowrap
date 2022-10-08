package orm

import (
	"errors"
	"reflect"

	"gorm.io/gorm"
)

var (
	// The model passed in is not a pointer
	ErrNotPointer = errors.New("not a valid pointer")

	// An update query affected 0 rows
	ErrNoRecordsUpdated = errors.New("no records updated")
)

type ORM interface {
	Insert(v any) error
	Update(v any) error
	PartialUpdate(model any, updates any, where Where) error
	Delete(v any, conditions ...Condition) error
	First(v any, id uint, conditions ...Condition) error
	FindOne(v any, where Where, conditions ...Condition) error
	FindAll(slicePtr any, conditions ...Condition) error
	DB() *gorm.DB
}

type orm struct {
	db *gorm.DB
}

func New(db *gorm.DB) ORM {
	return &orm{db: db}
}

func (o *orm) DB() *gorm.DB {
	return o.db
}

// validates that v is a pointer
func IsPointer(v any) bool {
	return reflect.ValueOf(v).Kind() == reflect.Pointer
}

// Insert v into the database.
// Note that relationships are not preloaded after insert.
// Requery the database with Preload conditions to load the records with relationships.
func (o *orm) Insert(v any) error {
	if !IsPointer(v) {
		return ErrNotPointer
	}

	return o.db.Create(v).Error
}

// Update v in the database. v must have a primary key field(id) set
func (o *orm) Update(v any) error {
	if !IsPointer(v) {
		return ErrNotPointer
	}

	return o.db.Save(v).Error

}

// Partial update of model(pointer) with updates struct.
// Where condition specified the select where condition is must be provided.
func (o *orm) PartialUpdate(model any, updates any, where Where) error {
	if !IsPointer(model) {
		return ErrNotPointer
	}

	ret := o.db.Model(model).Where(where.Query, where.Args...).Updates(updates)
	if ret.Error != nil {
		return ret.Error
	}

	if ret.RowsAffected < 1 {
		return ErrNoRecordsUpdated
	}

	return nil
}

// Delete the record from the database for the given where condition
// v must be a pointer
func (o *orm) Delete(v any, conditions ...Condition) error {
	if !IsPointer(v) {
		return ErrNotPointer
	}
	model := applyConditions(o.db, conditions...)
	return model.Unscoped().Delete(v).Error
}

// Get record by ID
// v pointer is populated by the query
func (o *orm) First(v any, id uint, conditions ...Condition) error {
	if !IsPointer(v) {
		return ErrNotPointer
	}
	model := applyConditions(o.db, conditions...)
	return model.First(v, id).Error
}

// FindOne is similar to First except that you must
// specify an arbitrary filter condition in where clause.
func (o *orm) FindOne(v any, where Where, conditions ...Condition) error {
	if !IsPointer(v) {
		return ErrNotPointer
	}

	model := applyConditions(o.db.Where(where.Query, where.Args...), conditions...)
	return model.First(v).Error
}

// FindAll queries the database, populating slicePtr with the records
func (o *orm) FindAll(slicePtr any, conditions ...Condition) error {
	if !IsPointer(slicePtr) {
		return ErrNotPointer
	}

	model := applyConditions(o.db, conditions...)
	return model.Find(slicePtr).Error
}

// Represents a paginated query result based of limit/offset pagination
type PaginatedResult[T any] struct {
	Page       int  // Current page
	Limit      int  // Page size
	HasNext    bool // if there is a next page
	HasPrev    bool // is there is a previous page
	Count      int  // total number of records in the database.
	TotalPages int  // total number of pages (based on Count)
	Results    []T  // Slice of the query results
}

func Paginate[T any](table *T, page int, limit int, db *gorm.DB, conditions ...Condition) (PaginatedResult[T], error) {
	var count int64
	model := db.Model(table)
	err := model.Count(&count).Error

	if err != nil {
		return PaginatedResult[T]{}, err
	}

	if page == 0 {
		page = 1
	}

	results := PaginatedResult[T]{}
	model = applyConditions(model, conditions...)
	err = model.Offset(limit * (page - 1)).Limit(limit).Find(&results.Results).Error
	if err != nil {
		return PaginatedResult[T]{}, err
	}

	totalPages := int(count) / limit
	if int(count)%limit > 0 {
		totalPages++
	}

	results.Page = page
	results.Count = int(count)
	results.TotalPages = totalPages
	results.HasNext = page < totalPages && (len(results.Results) >= limit)
	results.HasPrev = page > 1
	return results, nil
}
