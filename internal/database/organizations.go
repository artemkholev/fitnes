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

// GetAllOrganizations возвращает все организации
func (db *DB) GetAllOrganizations(ctx context.Context) ([]*models.Organization, error) {
	query := `SELECT id, name, code, created_at FROM organizations ORDER BY created_at DESC`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*models.Organization
	for rows.Next() {
		org := &models.Organization{}
		if err := rows.Scan(&org.ID, &org.Name, &org.Code, &org.CreatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

// GetOrganizationByID возвращает организацию по ID
func (db *DB) GetOrganizationByID(ctx context.Context, id int64) (*models.Organization, error) {
	org := &models.Organization{}
	query := `SELECT id, name, code, created_at FROM organizations WHERE id = $1`
	err := db.Pool.QueryRow(ctx, query, id).Scan(&org.ID, &org.Name, &org.Code, &org.CreatedAt)
	if err != nil {
		return nil, err
	}
	return org, nil
}
