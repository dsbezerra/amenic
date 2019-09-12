package extractutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/amenic/src/lib/util/stringutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// UserAgents is a list of user agents that the scraper can use to trick the webmasters
// and hopefully don't get block
var UserAgents = [...]string{
	// Linus Firefox
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:43.0) Gecko/20100101 Firefox/43.0",
	// Mac Firefox
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.11; rv:43.0) Gecko/20100101 Firefox/43.0",
	// Mac Safari 4
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_2; de-at) AppleWebKit/531.21.8 (KHTML, like Gecko) Version/4.0.4 Safari/531.21.10",
	// Mac Safari
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2) AppleWebKit/601.3.9 (KHTML, like Gecko) Version/9.0.2 Safari/601.3.9",
	// Windows Chrome
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.125 Safari/537.36",
	// Windows IE 10
	"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; WOW64; Trident/6.0)",
	// Windows IE 11
	"Mozilla/5.0 (Windows NT 6.3; WOW64; Trident/7.0; rv:11.0) like Gecko",
	// Windows Edge
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2486.0 Safari/537.36 Edge/13.10586",
	// Windows Firefox
	"Mozilla/5.0 (Windows NT 6.3; WOW64; rv:43.0) Gecko/20100101 Firefox/43.0",
	// iPhone
	"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B5110e Safari/601.1",
	// iPad
	"Mozilla/5.0 (iPad; CPU OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
	// Android
	"Mozilla/5.0 (Linux; Android 5.1.1; Nexus 7 Build/LMY47V) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.76 Safari/537.36",
}

// GetRandomUserAgent retrieves a random user agent
func GetRandomUserAgent() string {
	result := ""

	// Using current time nanosecond as seed
	seed := time.Now().Nanosecond()

	// Seed the random
	rand.Seed(int64(seed))

	// Get random user-agent
	size := len(UserAgents)
	result = UserAgents[rand.Int31n(int32(size))]

	return result
}

// NewDocument gets a new goquery.Document from a given website
func NewDocument(url, charset string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", GetRandomUserAgent())

	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New("debug")
	}
	defer response.Body.Close()

	needsToDecode := false

	// Check if we need to decode body
	ct := response.Header.Get("Content-Type")
	if ct != "" {
		values := strings.Split(ct, ";")
		for _, value := range values {
			if value == "" {
				continue
			}

			lw := strings.ToLower(value)
			if strings.Contains(lw, "charset=") {
				_, remainder := stringutil.BreakByToken(lw, '=')
				if remainder != "utf-8" && remainder != "" {
					charset = remainder
					needsToDecode = true
				}
			}
		}
	}

	var doc *goquery.Document
	if needsToDecode {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		// Convert charset to utf-8
		output, err := convertToUTF8(body, charset)
		if err != nil {
			return nil, err
		}

		doc, err = goquery.NewDocumentFromReader(strings.NewReader(output))
	} else {
		doc, err = goquery.NewDocumentFromResponse(response)
	}

	if err != nil {
		return nil, err
	}

	// NOTE: For some unknown reason doc.Url can be nil, so we make sure here that we have it by
	// assigning the URL from Request.
	if doc != nil && doc.Url == nil && response.Request.URL != nil {
		URL := *response.Request.URL
		doc.Url = &URL
	}

	return doc, err
}

func convertToUTF8(body []byte, from string) (string, error) {
	var dec *encoding.Decoder
	var err error

	from = strings.ToUpper(from)

	switch from {
	case "ISO-8859-1":
		dec = charmap.ISO8859_1.NewDecoder()
	default:
		err = fmt.Errorf("enconding %s is not supported", from)
	}

	if err != nil {
		return string(body), err
	}

	out, err := dec.Bytes(body)
	if err != nil {
		return string(body), err
	}

	return string(out), err
}
