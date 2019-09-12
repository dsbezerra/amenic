package cloudinary

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // Add support for jpegs
	_ "image/png"  // Add support for pngs
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
)

const (
	CloudinaryUri = ""
	BackdropPath  = "amenic/backdrop"
	PosterPath    = "amenic/poster"
)

var (
	ErrServiceNotInitialied = errors.New("cloudinary service is not initiliazed")
	ErrIvalidImageType      = errors.New("invalid image type")
	service                 *Service
)

// InitService ...
func InitService(uri string) {
	s, err := Dial(uri)
	if err == nil {
		service = s
	}
}

// DeleteImage ...
func DeleteImage(pathOrPublicId string) error {
	substring := fmt.Sprintf("%s/image/upload/", service.CloudName())
	if index := strings.Index(pathOrPublicId, substring); index > -1 {
		pathOrPublicId = pathOrPublicId[index+len(substring):]
	}
	return service.Delete(pathOrPublicId, "", ImageType)
}

// UploadWebImage ...
func UploadWebImage(urlstring string, stype string) (*models.Image, error) {
	if service == nil {
		return nil, ErrServiceNotInitialied
	}

	itype := models.ImageType(stype)
	valid := models.CheckForValidImageType(&itype)
	if !valid {
		return nil, ErrIvalidImageType
	}

	URL, err := url.Parse(urlstring)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(urlstring)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 && resp.StatusCode > 304 {
		return nil, errors.New("couldn't retrieve image from given url. Status: " + resp.Status)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image/") {
		return nil, errors.New("this url does not contain an image")
	}

	_, name := filepath.Split(URL.Path)
	dest, err := copyToTempFile(name, resp.Body)
	if err != nil {
		return nil, err
	}
	defer os.Remove(dest)
	return uploadImage(dest, itype)
}

// UploadMultipartImage ...
func UploadMultipartImage(file *multipart.FileHeader, stype string) (*models.Image, error) {
	if service == nil {
		return nil, ErrServiceNotInitialied
	}

	itype := models.ImageType(stype)
	valid := models.CheckForValidImageType(&itype)
	if !valid {
		return nil, ErrIvalidImageType
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// TODO(diego): Check for image file.
	// Supported ones may be PNG, JPEG and WebP.
	dest, err := copyToTempFile(file.Filename, src)
	if err != nil {
		return nil, err
	}

	// NOTE(diego):
	// The file will be stored in the temp dir of the executing operating system
	// and will be deleted someday.
	// But to ensure we always have space we will remove it right after
	// the upload.
	defer os.Remove(dest)
	return uploadImage(dest, itype)
}

func uploadImage(src string, itype models.ImageType) (*models.Image, error) {
	if service == nil {
		return nil, ErrServiceNotInitialied
	}
	//
	// NOTE(diego): First we open and try to decode image.
	// The decode step is used to retrieve width and height and also make sure
	// that our file is an image (supported image formats are PNG and JPG)
	//
	// After this we generate our filename, append the extension from the original file
	// and upload to Cloudinary.
	//
	// If we succeed, we build the image model which will be returned to caller so it
	// can handler as intended.
	//
	// If we fail, we return the error.
	//
	file, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode
	im, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}
	// Reset offset because we reuse it below
	file.Seek(0, 0)

	// Builds path for the given type
	path := "any"
	switch itype {
	case "backdrop":
		path = BackdropPath
	case "poster":
		path = PosterPath
	}

	filename := generateFilename()
	publicId := fmt.Sprintf("%s/%s", path, filename)
	_, err = service.UploadImage(publicId, file, "")
	if err != nil {
		return nil, err
	}

	// Create image model
	secureURL := service.Url(publicId, ImageType)
	mimage := &models.Image{
		Name:      filename,
		Path:      publicId,
		Type:      string(itype),
		Width:     im.Width,
		Height:    im.Height,
		URL:       strings.Replace(secureURL, "https", "http", 1),
		SecureURL: secureURL,
	}

	// Try to get checksum
	file.Seek(0, 0)
	buf, err := ioutil.ReadAll(file)
	if err == nil {
		sha1, err := FileChecksumFromData(buf)
		if err == nil {
			mimage.Checksum = sha1
		}
	}
	createdAt := time.Now()
	mimage.CreatedAt = &createdAt
	return mimage, nil
}

func copyToTempFile(name string, r io.Reader) (string, error) {
	dest := filepath.Join(os.TempDir(), name)
	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, r)
	if err != nil {
		return "", err
	}
	return dest, nil
}
