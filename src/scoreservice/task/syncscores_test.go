package task

import (
	"log"
	"testing"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
)

const (
	testConnection = "mongodb://localhost/amenic-test"
)

func TestSyncScores(t *testing.T) {
	data := newMockDataAccessLayer()
	err := SyncScores(data)
	exitIfErr(err)
}

func newMockDataAccessLayer() persistence.DataAccessLayer {
	data, err := mongolayer.NewMongoDAL(testConnection)
	exitIfErr(err)
	return data
}

func exitIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
