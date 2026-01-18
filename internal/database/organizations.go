package database

import (
	"context"
	"fitness-bot/internal/models"
)

func (db *DB) CreateOrganization(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO organizations (name, code)
		VALUES ($1, $2)
		RETURNING id, created_at
	`
	return db.Pool.QueryRow(ctx, query, org.Name, org.Code).Scan(&org.ID, &org.CreatedAt)
}

func (db *DB) GetOrganizationByCode(ctx context.Context, code string) (*models.Organization, error) {
	org := &models.Organization{}
	query := `SELECT id, name, code, created_at FROM organizations WHERE code = $1`
	err := db.Pool.QueryRow(ctx, query, code).Scan(&org.ID, &org.Name, &org.Code, &org.CreatedAt)
	if err != nil {
		return nil, err
	}
	return org, nil
}
