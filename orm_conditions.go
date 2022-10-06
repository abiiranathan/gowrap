package gowrap

import "gorm.io/gorm"

// Interface that applies some operation to the GORM DB definition
// Multiple conditions are applied in the order specified.
// Select, Join, Where, Group, Having
type Condition interface {
	Apply(db *gorm.DB) *gorm.DB
}

func applyConditions(db *gorm.DB, conditions ...Condition) *gorm.DB {
	for _, condition := range conditions {
		db = condition.Apply(db)
	}
	return db
}

// Preload relationships.
//
// if your struct has no nested relationships, you can simply
// pass clause.Associations as the query.
type Preload struct {
	// Main reload string e.g "Orders.Products" or "Users"
	Query string

	// condition for preloading e.g  []any{"state NOT IN (?)", "cancelled"}
	Args []any
}

func (p *Preload) Apply(db *gorm.DB) *gorm.DB {
	return db.Preload(p.Query, p.Args...)
}

// Filter query results based on the where string
type Where struct {
	// where add condition e.g "username=? and password=?"
	Query string

	// arguments to the query e.g []any{"johndoe", "passwordhash"}
	Args []any
}

func (p Where) Apply(db *gorm.DB) *gorm.DB {
	return db.Where(p.Query, p.Args...)
}

// Add grouping. Group should apear after Join but before Where conditions
type Group struct {
	Name string // grouping condition e.g "category DESC"
}

func (g Group) Apply(db *gorm.DB) *gorm.DB {
	return db.Group(g.Name)
}

type Join struct {
	Query string
	Args  []any
}

func (j Join) Apply(db *gorm.DB) *gorm.DB {
	return db.Joins(j.Query, j.Args...)
}

type Select struct {
	Query string
	Args  []any
}

func (j Select) Apply(db *gorm.DB) *gorm.DB {
	return db.Select(j.Query, j.Args...)
}
