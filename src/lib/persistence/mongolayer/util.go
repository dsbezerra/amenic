package mongolayer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dsbezerra/amenic/src/lib/util/scheduleutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// LookupInfo ...
type LookupInfo struct {
	srcCollection  *Collection
	fromCollection *Collection

	// Lookup fields
	from         string
	localField   string
	foreignField string
	as           string

	// Used in pipeline
	cond []bson.D
	sort bson.D
}

func buildPipeline(collectionName string, opts *QueryOptions) mongo.Pipeline {
	p := make(mongo.Pipeline, 0)

	// Build $match
	if len(opts.Conditions) > 0 {
		and := make([]bson.D, 0)
		for f, v := range opts.Conditions {

			switch v.(type) {
			case bson.D:
				and = append(and, bson.D{
					{Key: f, Value: v.(bson.D)},
				})
			default:
				and = append(and, bson.D{
					{Key: f, Value: bson.D{
						{Key: "$eq", Value: v},
					}},
				})
			}
		}
		p = append(p, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "$and", Value: and},
			}},
		})
	}

	// Build $lookup(s)
	for _, included := range opts.Includes {
		lookup := buildLookup(collectionName, included)
		if lookup != nil {
			p = append(p, lookup...)
		}
	}

	if len(opts.Sort) > 0 {
		p = append(p, bson.D{
			{Key: "$sort", Value: SortToBSON("", opts.Sort...)},
		})
	}

	// Some endpoints set -1 to disable skip/limit
	if opts.Skip > -1 {
		p = append(p, bson.D{
			{Key: "$skip", Value: opts.Skip},
		})
	}

	if opts.Limit > -1 {
		p = append(p, bson.D{
			{Key: "$limit", Value: opts.Limit},
		})
	}

	// Build projection
	project := bson.M{}
	if len(opts.Fields) == 0 {
		opts.Fields = DefaultOptions(collectionName).Fields
	}
	for f := range opts.Fields {
		project[f] = 1
	}
	for _, included := range opts.Includes {
		project[included.Field] = 1
	}
	if len(project) > 0 {
		p = append(p, bson.D{
			{Key: "$project", Value: project},
		})
	}
	return p
}

func getCollection(s string) (*Collection, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("invalid string")
	}

	var err error
	var name string
	var collt CollectionType

	switch s {
	case "apikeys", "apikey", "api_keys":
		name = CollectionAPIKeys
		collt = APIKeys

	case "cities", "city":
		name = CollectionCities
		collt = Cities

	case "images", "image":
		name = CollectionImages
		collt = Images

	case "movies", "movie":
		name = CollectionMovies
		collt = Movies

	case "notifications", "notification":
		name = CollectionNotifications
		collt = Notifications

	case "prices", "price":
		name = CollectionPrices
		collt = Prices

	case "scrapers", "scraper":
		name = CollectionScrapers
		collt = Scrapers

	case "scraper_runs", "scraper_run", "runs", "run":
		name = CollectionScores
		collt = ScraperRuns

	case "scores", "score":
		name = CollectionScores
		collt = Scores

	case "sessions", "session":
		name = CollectionSessions
		collt = Sessions

	case "theaters", "theater":
		name = CollectionTheaters
		collt = Theaters

	default:
		err = fmt.Errorf("couldn't map string %s to any of the database collections", s)
	}

	if err != nil {
		return nil, err
	}

	return &Collection{name, collt}, err
}

func buildLookup(src string, include QueryInclude) []bson.D {
	lookup, err := calculateLookupInfo(src, include)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	// Build lookup pipeline
	and := []bson.D{
		bson.D{
			{Key: "$eq", Value: []string{fmt.Sprintf("$%s", lookup.foreignField), "$$localField"}},
		},
	}
	and = append(and, lookup.cond...)
	lookupPipeline := []bson.D{
		bson.D{
			{
				Key: "$match",
				Value: bson.D{
					{
						Key: "$expr",
						Value: bson.D{
							{
								Key:   "$and",
								Value: and,
							},
						},
					},
				},
			},
		},
	}
	if len(lookup.sort) > 0 {
		lookupPipeline = append(lookupPipeline, bson.D{{Key: "$sort", Value: lookup.sort}})
	}

	if len(include.Fields) > 0 {
		project := bson.M{}
		for _, field := range include.Fields {
			project[field] = 1
		}
		lookupPipeline = append(lookupPipeline, bson.D{{Key: "$project", Value: project}})
	}

	return []bson.D{
		bson.D{{
			Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: lookup.from},
				{Key: "as", Value: lookup.as},
				{Key: "let", Value: bson.D{
					{Key: "localField", Value: fmt.Sprintf("$%s", lookup.localField)},
				}},
				{Key: "pipeline", Value: lookupPipeline},
			},
		}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: fmt.Sprintf("$%s", lookup.as)},
		}}},
	}
}

func calculateLookupInfo(collection string, include QueryInclude) (*LookupInfo, error) {
	// Map collection and include field to respective collection types
	src, err := getCollection(collection)
	if err != nil {
		return nil, err
	}

	from, err := getCollection(include.Field)
	if err != nil {
		return nil, err
	}

	// Build lookup information for each case
	result := LookupInfo{
		srcCollection:  src,
		fromCollection: from,
		from:           from.Name,
		cond:           make([]bson.D, 0),
		sort:           bson.D{},
	}

	fType := from.Type

	// NOTE: Temporary.
	//
	// We will be including this default options for showtimes include for while.
	if fType == Sessions {
		period := scheduleutil.GetWeekPeriod(nil)
		result.cond = append(result.cond, bson.D{
			{
				Key: "$gte", Value: []bson.D{
					bson.D{
						{
							Key: "$startTime", Value: period.Start,
						},
					},
				},
			},
		})
		result.sort = bson.D{
			{
				Key:   "movieId",
				Value: -1,
			},
			{
				Key:   "version",
				Value: 1,
			},
			{
				Key:   "format",
				Value: 1,
			},
			{
				Key:   "startTime",
				Value: 1,
			},
		}
	}

	switch src.Type {
	case Theaters:
		if fType == Prices {
			result.foreignField = "theaterId"
			result.localField = "_id"
			result.as = "prices"
		} else if fType == Sessions {
			result.foreignField = "theaterId"
			result.localField = "_id"
			result.as = "sessions"
		} else if fType == Cities {
			result.foreignField = "_id"
			result.localField = "cityId"
			result.as = "city"
		}

	case Movies:
		if fType == Scores {
			result.foreignField = "movieId"
			result.localField = "_id"
			result.as = "scores"
		} else if fType == Sessions {
			result.foreignField = "movieId"
			result.localField = "_id"
			result.as = "sessions"
		}

	case Images:
		if fType == Movies {
			result.localField = "movieId"
			result.foreignField = "_id"
			result.as = "movie"
		}

	case Sessions:
		if fType == Theaters {
			result.localField = "theaterId"
			result.foreignField = "_id"
			result.as = "theater"
		} else if fType == Movies {
			result.localField = "movieId"
			result.foreignField = "_id"
			result.as = "movie"
		}

	case Prices:
		if fType == Theaters {
			result.localField = "theaterId"
			result.foreignField = "_id"
			result.as = "theater"
		}

	case Cities:
	}

	if result.fromCollection.Type == None {
		return nil, fmt.Errorf("relation not defined between models src: %s and target: %s", collection, include.Field)
	}

	return &result, nil
}

// SortToBSON converts a sort array to a bson object
func SortToBSON(prefix string, sort ...string) bson.D {
	result := bson.D{}

	for _, f := range sort {
		if f == "" {
			continue
		}

		v := 0
		if f[0] == '-' {
			v = -1
		} else if f[0] == '+' {
			v = 1
		}

		// If our field doesn't have an - or +, defaults to ascending + == 1
		if v == 0 {
			v = 1
		} else {
			f = f[1:] // Delete -/+ prefix
		}

		field := f
		if prefix != "" {
			field = fmt.Sprintf("%s.%s", prefix, field)
		}

		result = append(result, bson.E{Key: field, Value: v})
	}

	return result
}

// DefaultOptions ...
func DefaultOptions(collectionName string) *QueryOptions {
	result := QueryOptions{
		Fields:     bson.M{},
		Conditions: bson.M{},
		Limit:      10,
		Skip:       0,
	}

	if collectionName == CollectionCities {
		result.Fields = bson.M{
			"_id":       1,
			"stateId":   1,
			"name":      1,
			"timeZone":  1,
			"createdAt": 1,
			"updatedAt": 1,
		}
	} else if collectionName == CollectionTheaters {
		result.Fields = bson.M{
			"_id":        1,
			"cityId":     1,
			"hidden":     1,
			"internalId": 1,
			"name":       1,
			"shortName":  1,
			"images":     1,
			"createdAt":  1,
			"updatedAt":  1,
		}
	} else if collectionName == CollectionSessions {
		result.Fields = bson.M{
			"_id":       1,
			"movieId":   1,
			"theaterId": 1,
			"movieSlug": 1,
			"room":      1,
			"startTime": 1,
			"format":    1,
			"version":   1,
			"timeZone":  1,
			"hidden":    1,
		}
	}

	return &result
}
