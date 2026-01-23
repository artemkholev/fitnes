package database

import (
	"fitness-bot/internal/models"
)

// CreateOrganization создаёт новую организацию
func (db *DB) CreateOrganization(org *models.Organization) error {
	return db.GORM.Create(org).Error
}

// GetOrganizationByCode получает организацию по коду
func (db *DB) GetOrganizationByCode(code string) (*models.Organization, error) {
	var org models.Organization
	err := db.GORM.Where("code = ?", code).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetAllOrganizations возвращает все организации
func (db *DB) GetAllOrganizations() ([]*models.Organization, error) {
	var orgs []*models.Organization
	err := db.GORM.Order("created_at DESC").Find(&orgs).Error
	return orgs, err
}

// GetOrganizationByID возвращает организацию по ID
func (db *DB) GetOrganizationByID(id int64) (*models.Organization, error) {
	var org models.Organization
	err := db.GORM.First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}
