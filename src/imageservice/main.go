package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/dsbezerra/amenic/src/imageservice/cloudinary"
	"github.com/dsbezerra/amenic/src/imageservice/listener"
	"github.com/dsbezerra/amenic/src/imageservice/rest"
	"github.com/dsbezerra/amenic/src/imageservice/task"
	"github.com/dsbezerra/amenic/src/lib/config"
	"github.com/dsbezerra/amenic/src/lib/messagequeue"
	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type ImageFile struct {
	ID                    string     `json:"id,omitempty"`                      // MD5 hash of URL without scheme http(s)://
	Name                  string     `json:"name,omitempty"`                    // Only the name of file without extension
	Filename              string     `json:"filename,omitempty"`                // Name of file including extension
	AbsoluteFilepath      string     `json:"absolute_path,omitempty"`           // Absolute path of file after persisted in disk
	Extension             string     `json:"extension,omitempty"`               // Extension of image
	Filepath              string     `json:"filepath,omitempty"`                // Relative path of file to the cwd
	Checksum              string     `json:"checksum,omitempty"`                // Checksum of file data
	URL                   string     `json:"url,omitempty"`                     // Source URL used to retrieve file
	LastModifiedFormatted string     `json:"last_modified_formatted,omitempty"` // Last time the file was modified by the source in the RFC1123 format
	LastModified          *time.Time `json:"last_modified,omitempty"`           // Last time the file was modified by the source server
	Length                int        `json:"length,omitempty"`                  // Length in bytes of the file
}

var (
	ErrNoImage          = errors.New("no image file")
	ErrImageNotModified = errors.New("image not modified")
)

const (
	ServiceName = "Image"
)

var (
	ctx = &Context{
		Service: ServiceName,
		Log:     logrus.WithFields(logrus.Fields{"App": ServiceName}),
	}
)

// Context ...
type Context struct {
	Service  string
	Log      *logrus.Entry
	Config   *config.ServiceConfig
	Data     persistence.DataAccessLayer
	Emitter  messagequeue.EventEmitter
	Listener messagequeue.EventListener
}

func main() {
	settings, err := config.LoadConfiguration()
	if err != nil {
		ctx.Log.Fatal(err)
	}
	ctx.Config = settings

	conn, err := amqp.Dial(settings.AMQPMessageBroker)
	if err != nil {
		ctx.Log.Fatal(err)
	}

	eventEmitter, err := messagequeue.NewAMQPEventEmitter(conn, "events")
	if err != nil {
		ctx.Log.Fatal(err)
	}
	ctx.Emitter = eventEmitter

	eventListener, err := messagequeue.NewAMQPEventListener(conn, "events", ServiceName)
	if err != nil {
		ctx.Log.Fatal(err)
	}
	ctx.Listener = eventListener

	data, err := mongolayer.NewMongoDAL(settings.DBConnection)
	if err != nil {
		ctx.Log.Fatal(err)
	}
	data.Setup()
	defer data.Close()

	ctx.Data = data
	ctx.Log.Info("Database setup completed!")

	cloudinary.InitService(settings.ImageServiceConnection)

	// Start event processor.
	p := listener.EventProcessor{
		Data:          data,
		Log:           ctx.Log,
		EventListener: eventListener,
	}
	go p.ProcessEvents()
	go task.RunAll(data)

	// Catch signal so we can shutdown gracefully
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		router := ctx.buildRouter()

		// Serve API
		rest.ServeAPI(router, data, eventEmitter)
		router.Run(settings.RESTEndpoint)
	}()

	// Wait for a signal
	sig := <-sigCh
	ctx.Log.WithField("signal", sig).Info("Signal received. Shutting down.")
}

// BuildRouter ...
func (ctx *Context) buildRouter() *gin.Engine {
	if ctx.Config.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.SecureJsonPrefix(")]}',\n")
	r.Use(helmet.Default())
	r.Use(
		cors.Default(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	return r
}

// func main() {
// 	URLS := []string{
// 		"http://www.claquete.com/fotos/filmes/poster/9266_grande.jpg",
// 		"http://www.claquete.com/fotos/filmes/poster/12077_grande.jpg",
// 	}

// 	for _, u := range URLS {
// 		url, err := stripScheme(u)
// 		if err != nil {
// 			// Ignore ...
// 			continue
// 		}

// 		// Check if already exists
// 		lastModified := ""

// 		ID := MD5Hash(url)
// 		fn := fmt.Sprintf("./data/%s.json", ID)
// 		_, err = os.Stat(fn)
// 		if err == nil {
// 			// NOTE: treat as found and fill last modified
// 			var v *ImageFile
// 			ReadJSONFile(fn, &v)

// 			if v != nil && v.LastModifiedFormatted != "" {
// 				lastModified = v.LastModifiedFormatted
// 			}
// 		}

// 		image, err := downloadImage(u, lastModified, "./data")
// 		if err != nil {
// 			if err == ErrImageNotModified {
// 				fmt.Printf("this image has not modified since %s\n", lastModified)
// 			} else {
// 				log.Fatal(err)
// 			}
// 			continue
// 		}

// 		fmt.Println("Brand new image downloaded.")
// 		fmt.Println("Saving in database...")
// 		err = WriteJSONFile(fmt.Sprintf("./data/%s.json", ID), image, 0)
// 		if err != nil {
// 			fmt.Println("something happened while writing file...")
// 		}
// 	}
// }

// func downloadImage(url, lastModified, outDir string) (*ImageFile, error) {
// 	req, err := http.NewRequest("GET", url, nil)
// 	if lastModified != "" {
// 		req.Header.Set("if-modified-since", lastModified)
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if lastModified != "" && resp.StatusCode == 304 {
// 		return nil, ErrImageNotModified
// 	}

// 	if !isImageFile(resp.Header.Get("content-type")) {
// 		return nil, ErrNoImage
// 	}

// 	data, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	filename := stringutil.StringAfterLast(url, '/')
// 	if filename == "" {
// 		// TODO: Generate some random filename
// 	}

// 	fp := path.Join(outDir, filename)
// 	err = ioutil.WriteFile(fp, data, 0)
// 	if err != nil {
// 		return nil, err
// 	}

// 	absFp, err := filepath.Abs(fp)
// 	if err != nil {
// 		return nil, err
// 	}

// 	result := &ImageFile{
// 		Filename:         filename,
// 		Filepath:         fp,
// 		AbsoluteFilepath: absFp,
// 		Extension:        filepath.Ext(fp),
// 		Length:           len(data),
// 	}

// 	dot := strings.LastIndex(filename, ".")
// 	if dot > -1 {
// 		result.Name = filename[0:dot]
// 	}

// 	lt := resp.Header.Get("last-modified")
// 	if lt != "" {
// 		t, err := time.Parse(time.RFC1123, lt)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		result.LastModified = &t
// 		result.LastModifiedFormatted = lt
// 	}

// 	// Strip http:// or https:// to avoid different md5 hashes
// 	r, err := stripScheme(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	result.ID = MD5Hash(r)
// 	result.Checksum = MD5Checksum(data)
// 	result.URL = r
// 	return result, nil
// }

// func isImageFile(contentType string) bool {
// 	return strings.HasPrefix(contentType, "image")
// }

// // MD5Hash ...
// func MD5Hash(s string) string {
// 	h := md5.New()
// 	io.WriteString(h, s)
// 	return fmt.Sprintf("%x", h.Sum(nil))
// }

// // MD5Checksum ...
// func MD5Checksum(data []byte) string {
// 	return fmt.Sprintf("%x", md5.Sum(data))
// }

// // WriteJSONFile ...
// func WriteJSONFile(filename string, data interface{}, perm os.FileMode) error {
// 	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
// 	if err != nil {
// 		return err
// 	}
// 	enc := json.NewEncoder(f)
// 	enc.SetIndent("", "  ")
// 	enc.Encode(data)
// 	if err1 := f.Close(); err == nil {
// 		err = err1
// 	}
// 	return err
// }

// // ReadJSONFile reads a JSON and fills the given interface
// func ReadJSONFile(filename string, v interface{}) error {
// 	data, err := ioutil.ReadFile(filename)
// 	if err != nil {
// 		return err
// 	}
// 	return json.Unmarshal(data, v)
// }

// // stripScheme returns a copy of string s without http:// scheme
// func stripScheme(s string) (string, error) {
// 	_, err := url.Parse(s)
// 	if err != nil {
// 		return "", err
// 	}
// 	return strings.NewReplacer([]string{
// 		"http://", "",
// 		"https://", "",
// 	}...).Replace(s), nil
// }
