package middlewares

import (
	"strconv"

	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// MaxLimit for GetAll queries
	MaxLimit = 50
)

// BaseParseQuery ...
func BaseParseQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := map[string]string{}
		for name, value := range c.Request.URL.Query() {
			query[name] = value[0]
		}
		limit, ok := query["limit"]
		if !ok {
			query["limit"] = "25"
		} else {
			value, err := strconv.Atoi(limit)
			if err == nil && value > MaxLimit {
				query["limit"] = "50"
			}
		}
		_, ok = query["skip"]
		if !ok {
			query["skip"] = "0"
		}
		c.Set("query_options", query)
	}
}

// ValidObjectIDHex middleware to check if we have a valid object hex id
func ValidObjectIDHex() gin.HandlerFunc {
	return func(c *gin.Context) {
		ID := c.Param("id")
		if ID != "" {
			_, err := primitive.ObjectIDFromHex(ID)
			if err != nil {
				apiutil.SendBadRequest(c)
				c.Abort()
				return
			}
		}
	}
}
