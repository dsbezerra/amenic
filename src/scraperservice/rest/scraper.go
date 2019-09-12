package rest

import (
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/dsbezerra/amenic/src/scraperservice/queue"
	"github.com/gin-gonic/gin"
)

// ScraperService ...
type ScraperService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeScrapers ...
func (rs *Service) ServeScrapers(r *gin.Engine) {
	s := &ScraperService{rs.data, rs.emitter}
	scrapers := r.Group("/scrapers", rest.AdminAuth(rs.data))
	scrapers.GET("", s.GetAll)
	scrapers.POST("/scraper/:id/run", s.RunScraper)
}

// GetAll ...
func (s *ScraperService) GetAll(c *gin.Context) {
	scrapers, err := s.data.GetScrapers(BuildScraperQuery(s.data, c))
	apiutil.SendSuccessOrError(c, scrapers, err)
}

// RunScraper ...
func (s *ScraperService) RunScraper(c *gin.Context) {
	err := queue.AddWork(queue.WorkRequest{
		ScraperID: c.Param("id"),
	})
	apiutil.SendSuccessOrError(c, "Scraper run emitted.", err)
}

// BuildScraperQuery builds movie query from request query string
func BuildScraperQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildScraperQuery(query)
}
