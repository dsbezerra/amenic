package mongolayer

import (
	"github.com/dsbezerra/amenic/src/lib/persistence"
)

func getTestingMongoDAL() (persistence.DataAccessLayer, error) {
	data, err := NewMongoDAL("mongodb://localhost/amenic-test")
	if err != nil {
		return nil, err
	}
	data.Setup()
	return data, err
}
