package db_models

type Status struct {
	ID          int64  `db:"id" json:"id"`
	Default     string `db:"default" json:"default"`
	DefaultName string `db:"default_name" json:"default_name"`
	Slug        string `db:"slug" json:"slug"`
	Name        string `db:"name" json:"name"`
	Color       string `db:"color" json:"color"`
	IsRoot      bool   `db:"is_root" json:"is_root"`
}
