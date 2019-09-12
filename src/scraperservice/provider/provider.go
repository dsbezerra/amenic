package provider

import (
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
)

const (
	// ProviderCinemais ...
	ProviderCinemais = "cinemais"
	// ProviderIbicinemas ...
	ProviderIbicinemas = "ibicinemas"
)

// Provider ...
type Provider interface {
	Init(data persistence.DataAccessLayer) error

	GetNowPlaying() ([]models.Movie, error)
	GetUpcoming() ([]models.Movie, error)
	GetSchedule() ([]models.Session, error)
	GetPrices() ([]models.Price, error)
}

// NewProvider ...
func NewProvider(data persistence.DataAccessLayer, name string, id string) Provider {
	var p Provider
	var err error

	switch name {
	case ProviderCinemais:
		p = NewCinemais(ComplexCode(id))
	case ProviderIbicinemas:
		p = NewIbicinemas()
	}
	if p != nil {
		err = p.Init(data)
	}
	if err != nil {
		return nil
	}
	return p
}
