package dropdown

import (
	"context"

	"github.com/go-pg/pg/v9"
)

type Repository interface {
	GetAirlineLogos(ctx context.Context) ([]AirlineLogoModel, error)
}

type AirlineLogoModel struct {
	UUID     string `pg:"uuid" json:"uuid"`
	Code     string `pg:"code" json:"code"`
	Name     string `pg:"name" json:"name"`
	LogoURL  string `pg:"logo_url" json:"logo_url"`
	IsActive bool   `pg:"is_active" json:"is_active"`
}

type repository struct {
	// db connection would go here if needed
}

func NewRepository() Repository {
	return &repository{}
}

func (r *repository) GetAirlineLogos(ctx context.Context) ([]AirlineLogoModel, error) {
	dbValue := ctx.Value("postgreSQLConn")
	if dbValue == nil {
		return []AirlineLogoModel{}, nil
	}

	db, ok := dbValue.(*pg.DB)
	if !ok || db == nil {
		return []AirlineLogoModel{}, nil
	}

	var airlines []AirlineLogoModel
	_, err := db.QueryContext(ctx, &airlines, `
		SELECT uuid, code, name, logo_url, is_active 
		FROM airline_logos 
		WHERE is_active = true 
		ORDER BY name
	`)

	if err != nil {
		// If no rows found, return empty slice instead of error
		if err == pg.ErrNoRows {
			return []AirlineLogoModel{}, nil
		}
		return nil, err
	}

	return airlines, nil
}
