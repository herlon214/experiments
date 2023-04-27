package downloader

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Result struct {
	Source string
	Body   []byte
}

type Downloader struct {
	input       <-chan string
	size        int
	enableCache bool
}

func NewDownloader(input <-chan string, size int, enableCache bool) *Downloader {
	return &Downloader{
		input:       input,
		size:        size,
		enableCache: enableCache,
	}
}

func (d *Downloader) Start(ctx context.Context) <-chan Result {
	output := make(chan Result, d.size)

	for i := 0; i < d.size; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case url := <-d.input:
					hash := md5.Sum([]byte(url))
					id := hex.EncodeToString(hash[:])

					// Read from cache
					if d.enableCache {
						file, err := os.ReadFile(fmt.Sprintf("data/%s", id))
						if err == nil {
							output <- Result{
								Source: url,
								Body:   file,
							}

							continue
						}
					}

					// Request data
					res, err := http.Get(url)
					if err != nil {
						fmt.Println("failed to request data", err.Error())
						continue
					}

					defer res.Body.Close()
					body, err := io.ReadAll(res.Body)
					if err != nil {
						fmt.Println("failed to read body", err.Error())
						continue
					}

					// Save to cache
					if d.enableCache {
						if err := os.WriteFile(fmt.Sprintf("data/%s", id), body, 0644); err != nil {
							fmt.Println("failed to save file", err.Error())
						}
					}

					output <- Result{
						Source: url,
						Body:   body,
					}
				}
			}
		}()
	}

	return output
}
