package task

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/dsbezerra/amenic/src/lib/persistence"
	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/persistence/mongolayer"
	"go.mongodb.org/mongo-driver/bson"
)

type checksumJob struct {
	index int    // Index of the image resource to calculate checksum
	u     string // URL of the image resource
}

type checksumResult struct {
	index    int    // Index of the image resource to calculate checksum
	checksum string // Calculated SHA1 Checksum
}

// EnsureImagesChecksum ensure our own uploaded images have checksums persisted
func EnsureImagesChecksum(data persistence.DataAccessLayer) error {
	startTime := time.Now()

	var q persistence.Query = &mongolayer.QueryOptions{Conditions: bson.M{}}
	q.AddCondition("checksum", bson.M{
		"$in": []interface{}{nil, ""},
	})
	images, err := data.GetImages(q)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		fmt.Println("There are no images without checksum.")
		return nil
	}

	jobs := make(chan checksumJob, 10)
	results := make(chan checksumResult, 10)

	go allocateChecksumJobs(jobs, images)
	done := make(chan bool)
	go resultChecksumJobs(done, results)
	checksumPool(data, images, jobs, results, 10)
	<-done

	fmt.Println("EnsureImagesChecksum took ", time.Now().Sub(startTime).Seconds(), "seconds")
	return nil
}

func allocateChecksumJobs(jobs chan checksumJob, images []models.Image) {
	for index, image := range images {
		job := checksumJob{index, image.URL}
		jobs <- job
	}
	close(jobs)
}

func resultChecksumJobs(done chan bool, results chan checksumResult) {
	for result := range results {
		fmt.Printf("Job id %d, finished with Checksum %s\n", result.index, result.checksum)
	}
	done <- true
}

func checksumWorker(data persistence.DataAccessLayer, images []models.Image, wg *sync.WaitGroup, jobs chan checksumJob, results chan checksumResult) {
	size := len(images)
	for job := range jobs {
		fmt.Printf("Downloading image %s...\n", job.u)
		time.Sleep(2 * time.Second)
		buf, err := downloadImage(job.u)
		if err != nil {
			fmt.Printf("Oops! An error ocurred while downloading image. Error: %s\n", err.Error())
			// Ignore error for now.
			results <- checksumResult{}
		} else {
			fmt.Printf("Image downloaded.\nGenerating SHA1 checksum...\n")
			chk, err := dataChecksum(buf)
			if err != nil {
				fmt.Printf("Oops! An error ocurred while generating checksum. Error: %s\n", err.Error())
				results <- checksumResult{}
			} else {
				fmt.Printf("Checksum %s generated.\n", chk)

				// Update database now
				if job.index >= 0 && job.index < size {
					image := images[job.index]
					image.Checksum = chk
					now := time.Now()
					image.UpdatedAt = &now
					_, err = data.UpdateImage(image.ID.Hex(), image)
					if err != nil {
						fmt.Printf("Oops! An error occured while updating image %s. Error: %s\n", image.ID.Hex(), err.Error())
					}
				}

				results <- checksumResult{job.index, chk}
			}
		}
	}
	wg.Done()
}

func checksumPool(data persistence.DataAccessLayer, images []models.Image, jobs chan checksumJob, results chan checksumResult, n int) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go checksumWorker(data, images, &wg, jobs, results)
	}
	wg.Wait()
	close(results)
}

func downloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !isImageFile(resp.Header.Get("content-type")) {
		return nil, ErrNoImage
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, err
}
