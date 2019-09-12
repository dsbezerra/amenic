package persistence

import (
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
)

// DataAccessLayer is used to communicate with the database.
type DataAccessLayer interface {
	Setup()
	Close()

	DefaultQuery() Query

	BuildCityQuery(q map[string]string) Query
	BuildMovieQuery(q map[string]string) Query
	BuildSessionQuery(q map[string]string) Query
	BuildTheaterQuery(q map[string]string) Query
	BuildScraperQuery(q map[string]string) Query
	BuildImageQuery(q map[string]string) Query

	// ------ APIKey ------
	// InsertAPIKey inserts a single APIKey resource
	// @param apikey{models.APIKey} - An APIKey reource to be inserted
	InsertAPIKey(apikey models.APIKey) error

	// FindAPIKey retrieves a APIKey resource matching the given Query
	// @param	query{Query}  - Options used to retrieve data
	FindAPIKey(query Query) (*models.APIKey, error)

	// GetAPIKey retrieves a APIKey resource by ID
	// @param	id{string} 		- APIKey identifier
	// @param	query{Query}  - Options used to retrieve data
	GetAPIKey(id string, query Query) (*models.APIKey, error)

	// GetAPIKeys retrieves all APIKey resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetAPIKeys(query Query) ([]models.APIKey, error)

	// DeleteAPIKey removes a single APIKey matching the given id
	// @param	id{string} 		- APIKey identifier
	DeleteAPIKey(id string) error

	// ------ Theater ------
	CountTheaters(query Query) (int64, error)

	// InsertTheater inserts a single Theater resource
	// @param theater{models.Theater} - A Theater resource to insert
	InsertTheater(theater models.Theater) error

	// FindTheater retrieves a Theater resource matching the given Query
	// @param	query{Query}  - Options used to retrieve data
	FindTheater(query Query) (*models.Theater, error)

	// GetTheater retrieves a Theater resource by ID
	// @param	id{string} 		- Theater identifier
	// @param	query{Query}  - Options used to retrieve data
	GetTheater(id string, query Query) (*models.Theater, error)

	// GetTheaters retrieves all Theater resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetTheaters(query Query) ([]models.Theater, error)

	// DeleteTheater removes a single Theater matching the given id
	// @param	id{string} 		- Theater identifier
	DeleteTheater(id string) error

	// DeleteTheaters removes all Theaters matching the given Query
	// @param	query{Query} - Options used to retrieve data
	DeleteTheaters(query Query) (int64, error)

	// TODO:
	UpdateTheater(id string, m models.Theater) (int64, error)

	// ------ Image ------

	// InsertImage inserts a single Image resource
	// @param image{models.Image} - An Image resource to be inserted
	InsertImage(image models.Image) error

	// FindImage retrieves a Image resource matching the given Query
	// @param	query{Query}  - Options used to retrieve data
	FindImage(query Query) (*models.Image, error)

	// GetImage retrieves a Image resource by ID
	// @param	id{string} - Image identifier
	// @param	query{Query}  - Options used to retrieve data
	GetImage(id string, query Query) (*models.Image, error)

	// GetImages retrieves all Image resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetImages(query Query) ([]models.Image, error)

	// GetMovieImages retrieves all Image resources for the given movie ID
	// @param	query{Query} - Options used to retrieve data
	GetMovieImages(id string, query Query) ([]models.Image, error)

	// DeleteImage removes a single Image matching the given id
	// @param	id{string} 		- Image identifier
	DeleteImage(id string) error

	// DeleteImages removes all Images matching the given Query
	// @param	query{Query} - Options used to retrieve data
	DeleteImages(query Query) (int64, error)

	// DeleteImagesByIDs removes all Images matching the given IDs
	// @param	ids{[]string} - Slice of target IDs
	DeleteImagesByIDs(ids []string) (int64, error)

	// TODO:
	UpdateImage(id string, m models.Image) (int64, error)

	// ------ Movie ------
	CountMovies(query Query) (int64, error)

	// InsertMovie inserts a single Movie resource
	// @param movie{models.Movie} - A Movie resource to be inserted
	InsertMovie(movie models.Movie) error

	// FindMovie retrieves a Movie resource matching the given Query
	// @param	query{Query}  - Options used to retrieve data
	FindMovie(query Query) (*models.Movie, error)

	// FindMovieAndUpdate finds a single Movie matching the given query
	// and updates it, returning either the original or the updated.
	// @param	query{Query}  				- Options used to find movie
	// @param	update{interface{}}   - Update data
	FindMovieAndUpdate(query Query, update interface{}) (*models.Movie, error)

	// GetMovie retrieves a Movie resource by ID
	// @param	id{string} 								- Movie identifier
	// @param	query{Query}  - Options used to retrieve data
	GetMovie(id string, query Query) (*models.Movie, error)

	// GetMovies retrieves all Movie resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetMovies(query Query) ([]models.Movie, error)

	// Legacy
	OldGetNowPlayingMovies(query Query) ([]models.Movie, error)

	// GetNowPlayingMovies retrieves all now playing matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetNowPlayingMovies(query Query) ([]models.Movie, error)

	// GetUpcomingMovies retrieve all upcoming movies matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetUpcomingMovies(query Query) ([]models.Movie, error)

	// DeleteMovie removes a single Movie matching the given id
	// @param	id{string} 		- Movie identifier
	DeleteMovie(id string) error

	// DeleteMovies removes all Movies matching the given Query
	// @param	query{Query} - Options used to retrieve data
	DeleteMovies(query Query) (int64, error)

	// TODO:
	UpdateMovie(id string, m models.Movie) (int64, error)

	// ------ Notification ------

	// InsertNotification inserts a single Notification resource
	// @param notification{models.Notification} - A Notification resource to be inserted
	InsertNotification(notification models.Notification) error

	// FindNotification retrieves a Notification resource matching the given Query
	// @param	query{Query}  - Options used to retrieve data
	FindNotification(query Query) (*models.Notification, error)

	// GetNotification retrieves a Notification resource by ID
	// @param	id{string} 		- Notification identifier
	// @param	query{Query}  - Options used to retrieve data
	GetNotification(id string, query Query) (*models.Notification, error)

	// GetAllNotifications retrieves all Notification resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetNotifications(query Query) ([]models.Notification, error)

	// DeleteNotification removes a single Notification matching the given id
	// @param	id{string} 		- Notification identifier
	DeleteNotification(id string) error

	// DeleteNotifications removes all Notifications matching the given Query
	// @param	query{Query} - Options used to retrieve data
	DeleteNotifications(query Query) (int64, error)

	// ------ Price ------

	// InsertPrice inserts a single Price resource
	// @param price{models.Price} - A Price resource to be inserted
	InsertPrice(price models.Price) error

	// InsertPrices inserts a list of Price resources in a bulk operation
	// @param prices{[]models.Price} - A list of Price resources to be inserted
	InsertPrices(prices ...models.Price) error

	// GetPrice retrieves a Price resource by ID
	// @param	id{string} 		- Price identifier
	// @param	query{Query}  - Options used to retrieve data
	GetPrice(id string, query Query) (*models.Price, error)

	// GetPrices retrieves all Price resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetPrices(query Query) ([]models.Price, error)

	// DeletePrice removes a single Price matching the given id
	// @param	id{string} - Price identifier
	DeletePrice(id string) error

	// DeletePrices removes all Prices matching the given Query
	// @param	query{Query} - Options used to retrieve data
	DeletePrices(query Query) (int64, error)

	// ------ Score ------

	// InsertScore inserts a single Score resource
	// @param score{models.Score} - A Score resource to be inserted
	InsertScore(score models.Score) error

	// GetScore retrieves a Score resource by ID
	// @param	id{string} 		- Score identifier
	// @param	query{Query}  - Options used to retrieve data
	GetScore(id string, query Query) (*models.Score, error)

	// GetScores retrieves all Score resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetScores(query Query) ([]models.Score, error)

	// DeleteScore removes a single Score matching the given id
	// @param	id{string} - Score identifier
	DeleteScore(id string) error

	// DeleteScores removes all Scores matching the given Query
	// @param	query{Query} - Options used to retrieve data
	DeleteScores(query Query) (int64, error)

	// TODO:
	UpdateScore(id string, s models.Score) (int64, error)

	// ------ Scraper ------

	// InsertScraper inserts a single Scraper resource
	// @param scraper{models.Scraper} - A Scraper resource to be inserted
	InsertScraper(scraper models.Scraper) error

	// FindScraper retrieves a Scraper resource matching the given Query
	// @param	query{Query}  - Options used to retrieve data
	FindScraper(query Query) (*models.Scraper, error)

	// GetScraper retrieves a Scraper resource by ID
	// @param	id{string} 		- Scraper identifier
	// @param	query{Query}  - Options used to retrieve data
	GetScraper(id string, query Query) (*models.Scraper, error)

	// GetScrapers retrieves all Scraper resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetScrapers(query Query) ([]models.Scraper, error)

	// DeleteScraper removes a single Scraper matching the given id
	// @param	id{string} - Scraper identifier
	// DeleteScraper(id string) error

	// DeleteScrapers removes all Scrapers matching the given Query
	// @param	query{Query} - Options used to retrieve data
	// DeleteScrapers(query Query) error

	// TODO:
	UpdateScraper(id string, s models.Scraper) (int64, error)

	// ------ Scraper Run ------

	// InsertScraperRun inserts a single ScraperRun resource
	// @param scraperRun{models.ScraperRun} - A ScraperRun resource to be insert
	InsertScraperRun(scraperRun models.ScraperRun) error

	// GetScraperRun retrieves a ScraperRun resource by ID
	// @param	id{string} 		- ScraperRun identifier
	// @param	query{Query}  - Options used to retrieve data
	GetScraperRun(id string, query Query) (*models.ScraperRun, error)

	// ------ Session ------

	// InsertSession inserts a single Session resource
	// @param session{models.Session} - A Session resource to insert
	InsertSession(session models.Session) error

	// InsertSessions inserts a list of Session resources in a bulk operation
	// @param sessions{[]models.Session} - A list of Session resources to insert
	InsertSessions(sessions ...models.Session) error

	// GetSession retrieves a Session resource by ID
	// @param	id{string} 		- Session identifier
	// @param	query{Query}  - Options used to retrieve data
	GetSession(id string, query Query) (*models.Session, error)

	// GetSessions retrieves all Session resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetSessions(query Query) ([]models.Session, error)

	// DeleteSession removes a single Session matching the given id
	// @param	id{string} - Session identifier
	DeleteSession(id string) error

	// DeleteSessions removes all Sessions matching the given Query
	// @param	query{Query} - Options used to retrieve data
	DeleteSessions(query Query) (int64, error)

	// ------ Task ------

	// InsertTask inserts a single Task resource
	// @param task{models.Task} - A Task resource to be inserted
	InsertTask(task models.Task) error

	// GetTask retrieves a Task resource by ID
	// @param	id{string} 		- Task identifier
	// @param	query{Query}  - Options used to retrieve data
	GetTask(id string, query Query) (*models.Task, error)

	// GetTasks retrieves all Task resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetTasks(query Query) ([]models.Task, error)

	// DeleteTask removes a single Task matching the given id
	// @param	id{string} - Task identifier
	DeleteTask(id string) error

	// DeleteTasks removes all Tasks matching the given Query
	// @param	query{Query} - Options used to retrieve data
	DeleteTasks(query Query) (int64, error)

	// TODO:
	UpdateTask(id string, t models.Task) (int64, error)

	// ------ City ------

	// InsertCity inserts a single City resource
	// @param task{models.City} - A City resource to be inserted
	InsertCity(task models.City) error

	// GetCity retrieves a City resource by ID
	// @param	id{string} 		- City identifier
	// @param	query{Query}  - Options used to retrieve data
	GetCity(id string, query Query) (*models.City, error)

	// GetCities retrieves all City resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	GetCities(query Query) ([]models.City, error)

	// DeleteCity removes a single City matching the given id
	// @param	id{string} - City identifier
	DeleteCity(id string) error

	// TODO:
	UpdateCity(id string, c models.City) (int64, error)

	// DeleteCities removes all Cities matching the given Query
	// @param	query{Query} - Options used to retrieve data
	DeleteCities(query Query) (int64, error)

	// ------ State ------

	// InsertState inserts a single State resource
	// @param task{models.State} - A State resource to be inserted
	// InsertState(task models.State) error

	// GetState retrieves a State resource by ID
	// @param	id{string} 		- State identifier
	// @param	query{Query}  - Options used to retrieve data
	// GetState(id string, query Query) (*models.State, error)

	// GetStates retrieves all State resources matching the given Query
	// @param	query{Query} - Options used to retrieve data
	// GetStates(query Query) ([]models.State, error)

	// TODO:
	// UpdateState(id string, s models.State) (int64, error)

	// DeleteState removes a single State matching the given id
	// @param	id{string} - State identifier
	// DeleteState(id string) error

	// DeleteStates removes all Cities matching the given Query
	// @param	query{Query} - Options used to retrieve data
	// DeleteStates(query Query) (int64, error)
}

// Query ...
type Query interface {
	AddCondition(name string, value interface{}) Query
	GetConditions() interface{}
	GetCondition(name string) interface{}

	AddField(string) Query
	SetFields(interface{}) Query
	GetFields() interface{}

	SetSort(...string) Query
	GetSort() []string
	Sorting() bool

	SetLimit(int64) Query
	GetLimit() int64

	SetSkip(int64) Query
	GetSkip() int64

	AddInclude(...string) Query
	HasInclude() bool
	SetIncludes(interface{}) Query
	GetIncludes() interface{}
}
