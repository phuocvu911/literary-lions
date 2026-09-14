package database

import (
	"database/sql"
	"lions/internal/models"
)

func ListCategories(db *sql.DB) ([]models.Category, error) {
	categories := []models.Category{}

	query := "SELECT id, name FROM categories ORDER by name;"
	rows, err := db.Query(query)
	if err != nil {
		return categories, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category
		err = rows.Scan(
			&category.ID,
			&category.Name,
		)
		if err != nil {
			return categories, err
		}
		categories = append(categories, category)
	}
	if err = rows.Err(); err != nil {
		return categories, err
	}

	return categories, nil
}


// func GetCateByID(db *sql.DB, id int64) ([]string,error) {

// }
