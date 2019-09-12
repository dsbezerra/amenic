package mongolayer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/util/mathutil"
	"github.com/dsbezerra/amenic/src/lib/util/stringutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CollectionAPIKeys       = "api_keys"
	CollectionCities        = "cities"
	CollectionImages        = "images"
	CollectionMovies        = "movies"
	CollectionNotifications = "notifications"
	CollectionPrices        = "prices"
	CollectionScores        = "scores"
	CollectionScrapers      = "scrapers"
	CollectionScraperRuns   = "scraper_runs"
	CollectionSessions      = "sessions"
	CollectionTasks         = "tasks"
	CollectionTheaters      = "theaters"
)

type (
	CollectionType uint

	// Collection used to represent model as integers
	Collection struct {
		Name string
		Type CollectionType
	}

	// QueryInclude TODO
	QueryInclude struct {
		Field  string // Name of the model to include, if plural it must include an array of the model
		Fields []string
	}

	// QueryOptions TODO
	QueryOptions struct {
		Conditions bson.M
		Fields     bson.M
		Sort       []string
		Limit      int64
		Skip       int64
		Includes   []QueryInclude

		sorting bool
	}
)

const (
	None CollectionType = iota
	APIKeys
	Cities
	Theaters
	Images
	Movies
	Notifications
	Prices
	Scores
	Scrapers
	ScraperRuns
	Sessions
)

type (
	// MongoDAL represents a mgo.Database
	MongoDAL struct {
		client *mongo.Client
		db     *mongo.Database
		name   string
	}
)

// NewMongoDAL ...
func NewMongoDAL(connection string) (persistence.DataAccessLayer, error) {
	clientOptions := options.Client().ApplyURI(connection)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	name := stringutil.StringAfterLast(connection, '/')
	return &MongoDAL{
		client: client,
		db:     client.Database(name),
		name:   name,
	}, err
}

// DefaultQuery ...
func (m *MongoDAL) DefaultQuery() persistence.Query {
	return DefaultOptions("")
}

// Close ...
func (m *MongoDAL) Close() {
	if m.client != nil {
		m.client.Disconnect(context.Background())
	}
}

// C ...
func (m *MongoDAL) C(collectionName string) *mongo.Collection {
	return m.db.Collection(collectionName)
}

// AggregateOne ...
func (m *MongoDAL) AggregateOne(collectionName string, id interface{}, pipeline interface{}, result interface{}) error {
	ctx := context.Background()
	cursor, err := m.db.Collection(collectionName).Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.Decode(&result)
}

// AggregateOneWithQuery ...
func (m *MongoDAL) AggregateOneWithQuery(collectionName string, id interface{}, query persistence.Query, result interface{}) error {
	opts := query.(*QueryOptions)
	opts.Conditions["_id"] = id
	return m.AggregateOne(collectionName, id, buildPipeline(collectionName, opts), result)
}

// AggregateAll ...
func (m *MongoDAL) AggregateAll(collectionName string, pipeline interface{}, result interface{}) error {
	ctx := context.Background()
	cursor, err := m.db.Collection(collectionName).Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, &result)
}

// AggregateAllWithQuery ...
func (m *MongoDAL) AggregateAllWithQuery(collectionName string, query persistence.Query, result interface{}) error {
	return m.AggregateAll(collectionName, buildPipeline(collectionName, query.(*QueryOptions)), result)
}

// Count returns the total number of documents in the collection.
func (m *MongoDAL) Count(collectionName string, query persistence.Query) (int64, error) {
	return m.db.Collection(collectionName).CountDocuments(context.Background(), query.GetConditions())
}

// InsertOne ...
func (m *MongoDAL) InsertOne(collectionName string, doc interface{}) error {
	_, err := m.db.Collection(collectionName).InsertOne(context.Background(), doc)
	return err
}

// InsertMany ...
func (m *MongoDAL) InsertMany(collectionName string, docs []interface{}) error {
	_, err := m.db.Collection(collectionName).InsertMany(context.Background(), docs)
	return err
}

// FindOne ...
func (m *MongoDAL) FindOne(collectionName string, query persistence.Query, result interface{}) (interface{}, error) {
	options := options.FindOneOptions{
		Projection: query.GetFields(),
	}
	doc := m.db.Collection(collectionName).FindOne(context.Background(), query.GetConditions(), &options)
	err := doc.Decode(&result)
	return result, err
}

// Get ...
func (m *MongoDAL) Get(collectionName string, id interface{}, query persistence.Query, result interface{}) (interface{}, error) {
	var err error
	if query.HasInclude() {
		err = m.AggregateOneWithQuery(collectionName, id, query, result)
	} else {
		_, err = m.FindOne(collectionName, query.AddCondition("_id", id), result)
	}
	return result, err
}

// DeleteOne ...
func (m *MongoDAL) DeleteOne(collectionName string, id interface{}) error {
	_, err := m.db.Collection(collectionName).DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

// UpdateId ...
func (m *MongoDAL) UpdateId(collectionName string, id interface{}, data interface{}) (int64, error) {
	result, err := m.db.Collection(collectionName).UpdateOne(context.Background(), collectionName, bson.M{"_id": id})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

func getFindOneOptions(query persistence.Query) *options.FindOneOptions {
	return &options.FindOneOptions{Projection: query.GetFields()}
}

func getFindOptions(query persistence.Query) *options.FindOptions {
	sort := bson.M{}
	for _, v := range query.GetSort() {
		value := 1
		if v[0] == '-' {
			value = -1
			v = v[1:]
		} else if v[0] == '+' {
			v = v[1:]
		}
		sort[v] = value
	}
	opts := options.FindOptions{
		Projection: query.GetFields(),
		Sort:       sort,
	}

	limit := query.GetLimit()
	if limit > -1 {
		opts.Limit = &limit
	}

	skip := query.GetSkip()
	if skip > -1 {
		opts.Skip = &skip
	}

	return &opts
}

func getFindOneAndUpdateOptions(query persistence.Query) *options.FindOneAndUpdateOptions {
	return &options.FindOneAndUpdateOptions{Projection: query.GetFields()}
}

// Setup ...
func (m *MongoDAL) Setup() {
	// notificationsCollection := m.C(CollectionNotifications)

	// APIKeys
	apiKeysCollection := m.C(CollectionAPIKeys)
	EnsureUniqueIndex(apiKeysCollection, "key")
	EnsureIndexes(apiKeysCollection, []string{
		"owner",
		"user_type",
	})

	// Movies
	moviesCollection := m.C(CollectionMovies)
	// TODO: See if we need to add this:
	// https://docs.mongodb.com/manual/tutorial/specify-language-for-text-index/
	EnsureTextIndexes(moviesCollection, []string{
		"title",
		"originalTitle",
	})
	EnsureIndexes(moviesCollection, []string{
		"tmdbId",
		"imdbId",
		"claqueteId",
		"slug",
		"title",
		"originalTitle",
		"hidden",
		"releaseDate",
	})

	// Cities
	citiesCollection := m.C(CollectionCities)
	EnsureIndexes(citiesCollection, []string{
		"name",
		"state",
		"timeZone",
	})

	// Theaters
	theatersCollection := m.C(CollectionTheaters)
	EnsureTextIndexes(theatersCollection, []string{
		"name",
		"shortName",
	})
	EnsureIndexes(theatersCollection, []string{
		"cityId",
		"internalId",
		"hidden",
	})

	// Scrapers
	scrapersCollection := m.C(CollectionScrapers)
	EnsureIndexes(scrapersCollection, []string{
		"theaterId",
		"type",
		"provider",
	})

	// scraperRunsCollection := m.C(CollectionScraperRuns)

	scoresCollection := m.C(CollectionScores)
	EnsureIndex(scoresCollection, "movieId")

	sessionsCollection := m.C(CollectionSessions)
	EnsureIndexes(sessionsCollection, []string{
		"movieSlug",
		"theaterId",
		"movieId",
		"startTime",
		"hidden",
		"room",
		"version",
		"format",
	})

	pricesCollection := m.C(CollectionPrices)
	EnsureIndex(pricesCollection, "theaterId")

	// EnsureUniqueIndex(notificationsCollection, "nowPlaying")
}

// EnsureTextIndexes ...
func EnsureTextIndexes(c *mongo.Collection, keys []string) {
	indexKeys := bson.M{}
	for _, k := range keys {
		indexKeys[k] = "text"
	}
	model := mongo.IndexModel{
		Keys: indexKeys,
	}
	_, err := c.Indexes().CreateOne(context.Background(), model)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

// EnsureIndex ensures that a given key in a given collection is an index.
func EnsureIndex(c *mongo.Collection, key string) {
	ensureIndex(c, []string{key}, false, true, true)
}

// EnsureIndexes ...
func EnsureIndexes(c *mongo.Collection, keys []string) {
	ensureIndexes(c, keys, false, true, true)
}

// EnsureUniqueIndex ensures that a given key in a given collection is an unique index.
func EnsureUniqueIndex(c *mongo.Collection, key string) {
	ensureIndex(c, []string{key}, true, true, true)
}

// EnsureUniqueIndexes ...
func EnsureUniqueIndexes(c *mongo.Collection, keys []string) {
	ensureIndexes(c, keys, true, true, true)
}

func ensureIndexes(c *mongo.Collection, keys []string, unique, bg, sparse bool) {
	indexModels := make([]mongo.IndexModel, 0)
	for _, k := range keys {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.M{
				k: 1,
			},
			Options: (&options.IndexOptions{}).
				SetBackground(bg).
				SetUnique(unique).
				SetSparse(sparse),
		})
	}
	_, err := c.Indexes().CreateMany(context.Background(), indexModels)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func ensureIndex(c *mongo.Collection, keys []string, unique, bg, sparse bool) {
	indexKeys := bson.M{}
	for _, k := range keys {
		indexKeys[k] = 1
	}
	model := mongo.IndexModel{
		Keys: indexKeys,
		Options: (&options.IndexOptions{}).
			SetBackground(bg).
			SetUnique(unique).
			SetSparse(sparse),
	}
	_, err := c.Indexes().CreateOne(context.Background(), model)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func (q *QueryOptions) AddCondition(name string, value interface{}) persistence.Query {
	q.Conditions[name] = value
	return q
}

func (q *QueryOptions) GetConditions() interface{} {
	return q.Conditions
}

func (q *QueryOptions) GetCondition(name string) interface{} {
	return q.Conditions[name]
}

func (q *QueryOptions) AddField(field string) persistence.Query {
	q.Fields[field] = 1
	return q
}

func (q *QueryOptions) SetFields(fields interface{}) persistence.Query {
	q.Fields = fields.(bson.M)
	return q
}

func (q *QueryOptions) GetFields() interface{} {
	return q.Fields
}

func (q *QueryOptions) SetSort(sort ...string) persistence.Query {
	q.Sort = sort
	q.sorting = len(sort) > 0
	return q
}

func (q *QueryOptions) GetSort() []string {
	return q.Sort
}

func (q *QueryOptions) Sorting() bool {
	return q.sorting || len(q.Sort) > 0
}

func (q *QueryOptions) SetSkip(skip int64) persistence.Query {
	q.Skip = skip
	return q
}

func (q *QueryOptions) GetSkip() int64 {
	return q.Skip
}

func (q *QueryOptions) SetLimit(limit int64) persistence.Query {
	q.Limit = limit
	return q
}

func (q *QueryOptions) GetLimit() int64 {
	return q.Limit
}

func (q *QueryOptions) HasInclude() bool {
	return len(q.Includes) > 0
}

func (q *QueryOptions) AddInclude(include ...string) persistence.Query {
	// TODO: Improve if necessary.
	for _, i := range include {
		q.Includes = append(q.Includes, QueryInclude{
			Field: i,
		})
	}
	return q
}

func (q *QueryOptions) SetIncludes(includes interface{}) persistence.Query {
	q.Includes = includes.([]QueryInclude)
	return q
}

func (q *QueryOptions) GetIncludes() interface{} {
	return q.Includes
}

// BuildQuery ...
func BuildQuery(collectionName string, q map[string]string) persistence.Query {
	if q == nil {
		q = make(map[string]string)
	}
	query := DefaultOptions(collectionName)
	if len(q) > 0 {
		fields, ok := q["fields"]
		if ok {
			query.SetFields(parseFieldsQuery(fields))
		}
		sort, ok := q["sort"]
		if ok {
			query.SetSort(parseSortQuery(sort)...)
		}
		limit, ok := q["limit"]
		if ok {
			value, err := parseLimitQuery(limit)
			if err == nil {
				query.SetLimit(value)
			}
		}
		skip, ok := q["skip"]
		if ok {
			value, err := parseSkipQuery(skip)
			if err == nil {
				query.SetSkip(value)
			}
		}
		include, ok := q["include"]
		if ok {
			len := len(include)
			if len > 0 && strings.HasPrefix(include, "{") && strings.HasSuffix(include, "}") {
				query.SetIncludes(parseIncludeQuery(include[1 : len-1]))
			}
		}
	}
	return query
}

func parseFieldsQuery(fields string) bson.M {
	result := bson.M{}
	fields = strings.Replace(fields, " ", "", -1)
	for _, f := range strings.Split(fields, ",") {
		if len(f) > 0 {
			sign := -1
			if f[0] == '-' {
				sign = 0
			} else if f[0] == '+' {
				sign = 1
			}

			if sign != -1 {
				result[f[1:]] = sign
			} else {
				result[f] = 1
			}
		} else {
			result[f] = 1
		}
	}
	return result
}

func parseSortQuery(sort string) []string {
	sort = strings.Replace(sort, " ", "", -1)
	return strings.Split(sort, ",")
}

func parseLimitQuery(limit string) (int64, error) {
	value, err := strconv.ParseInt(limit, 10, 32)
	if err != nil {
		return -1, err
	}
	return value, nil
}

func parseSkipQuery(skip string) (int64, error) {
	value, err := strconv.ParseInt(skip, 10, 32)
	if err != nil {
		return 0, err
	}
	return mathutil.MaxInt64(0, value), nil
}

func parseIncludeQuery(include string) []QueryInclude {
	models := strings.Split(include, ",")
	result := make([]QueryInclude, len(models))

	for index, v := range models {
		name, remainder := stringutil.BreakByToken(v, '{')
		if name == "" {
			continue
		}

		i := QueryInclude{Field: name}
		if remainder != "" {
			fields, _ := stringutil.BreakByToken(remainder, '}')
			if fields != "" {
				// NOTE: Maybe add default fields of the collection here
				for _, f := range strings.Split(fields, "\\n") {
					i.Fields = append(i.Fields, strings.TrimSpace(f))
				}
			}
		}

		result[index] = i
	}

	return result
}

func getCurrentTime() *time.Time {
	now := time.Now()
	return &now
}
