package rest

import (
	"github.com/dsbezerra/amenic/src/imageservice/cloudinary"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ImageService ...
type ImageService struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

type imageUploadBody struct {
	Type    string `form:"type"`    // Type of image: backdrop, poster, etc...
	MovieID string `form:"movieId"` // ID of movie which this image belongs. (optional)
	URL     string `form:"url"`     // URL of image
	Main    bool   `form:"main"`    // Whether this image should be marked as main for the given movie ID (optional, there's only effect if a movie id is passed too)
}

// ServeImages ...
func (rs *Service) ServeImages(r *gin.Engine) {
	s := &ImageService{rs.data, rs.emitter}

	images := r.Group("/images")
	images.GET("/:id", s.Get)
	images.GET("/", s.GetAll)

	images.POST("/upload", s.Upload)
	images.DELETE("/:id", s.Delete)
}

// Get TODO
func (s *ImageService) Get(c *gin.Context) {
	image, err := s.data.GetImage(c.Param("id"), BuildImageQuery(s.data, c))
	apiutil.SendSuccessOrError(c, image, err)
}

// GetAll TODO
func (s *ImageService) GetAll(c *gin.Context) {
	images, err := s.data.GetImages(BuildImageQuery(s.data, c))
	apiutil.SendSuccessOrError(c, images, err)
}

// Upload ...
func (s *ImageService) Upload(c *gin.Context) {
	var body imageUploadBody
	c.ShouldBind(&body)

	var result *models.Image
	var err error

	// NOTE(diego): We should replace this with an interface to make it
	// work in case we need to support other databases.
	var movieID primitive.ObjectID
	if body.MovieID != "" {
		movieID, err = primitive.ObjectIDFromHex(body.MovieID)
		if err != nil {
			apiutil.SendBadRequest(c)
			return
		}
	}

	imageType := body.Type
	if body.URL != "" {
		result, err = cloudinary.UploadWebImage(body.URL, imageType)
	} else {
		file, err := c.FormFile("file")
		if err != nil {
			apiutil.SendBadRequest(c)
			return
		}
		result, err = cloudinary.UploadMultipartImage(file, imageType)
	}

	if err == nil {
		if result.ID.IsZero() {
			result.ID = primitive.NewObjectID()
		}
		result.Main = body.Main

		// Check if we will need to update Movie doc.
		var updateMovie bool
		var itype = models.ImageType(result.Type)
		if !movieID.IsZero() {
			updateMovie = body.Main && (itype == models.ImageTypeBackdrop || itype == models.ImageTypePoster)
			result.MovieID = movieID
		}
		err = s.data.InsertImage(*result)
		if err != nil {
			// TODO: Diagnostic
		} else {
			if updateMovie {
				updateMovieDoc := func(ID primitive.ObjectID, data persistence.DataAccessLayer) {
					if itype == "" {
						return
					}

					k := string(itype)

					// NOTE(diego): Commented for now since itype matches our collection field name
					// switch itype {
					// case models.ImageTypeBackdrop:
					// 	k = "backdrop"
					// case models.ImageTypePoster:
					// 	k = "poster"
					// default:
					// 	return
					// }

					filter := data.DefaultQuery().AddCondition("_id", ID)
					_, err = data.FindMovieAndUpdate(filter, bson.M{"$set": bson.M{k: result.SecureURL}})
					if err != nil {
						// TODO: Diagnostic or ignore.
					} else {

						// Get the previous images that could be marked as main
						// so we can remove it later in case we change
						previous, _ := data.GetImages(data.DefaultQuery().
							AddCondition("movieId", ID).
							AddCondition("type", k).
							AddCondition("main", true).
							SetLimit(-1)) // Ensure we don't limit our results

						if previous != nil {
							// Update all previous marked as main
							for _, im := range previous {
								if im.Name == result.Name {
									continue
								}
								im.Main = false
								_, err := data.UpdateImage(im.ID.Hex(), im)
								if err != nil {
									// TODO: Diagnostic or ignore.
								}
							}
						}
					}
				}
				go updateMovieDoc(movieID, s.data)
			}
		}
	}

	apiutil.SendSuccessOrError(c, result, err)
}

// Delete route deletes an image, returning not_found if the image does not exist
// in the database, protected_resource if the image is being used by other resource,
// or other error in case we fail to delete it either from Cloudinary or our database.
func (s *ImageService) Delete(c *gin.Context) {
	image, err := s.data.GetImage(c.Param("id"), s.data.DefaultQuery())
	if err != nil {
		apiutil.SendNotFound(c)
		return
	}

	if image.Main {
		apiutil.SendProtectedResource(c)
		return
	}

	err = cloudinary.DeleteImage(image.SecureURL)
	if err != nil {
		apiutil.SendSuccessOrError(c, nil, err)
		return
	}

	err = s.data.DeleteImage(c.Param("id"))
	apiutil.SendSuccessOrError(c, 1, err)
}

// BuildImageQuery builds image query from request query string
func BuildImageQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildImageQuery(query)
}
