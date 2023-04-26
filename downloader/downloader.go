package downloader

import (
	"context"
	"io"
	"net/http"
)

type Result struct {
	Source string
	Body   []byte
}

type Downloader struct {
	input <-chan string
	size  int
}

func NewDownloader(input <-chan string, size int) *Downloader {
	return &Downloader{
		input: input,
		size:  size,
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
					res, err := http.Get(url)
					if err != nil {
						panic(err)
					}

					defer res.Body.Close()
					body, err := io.ReadAll(res.Body)
					if err != nil {
						panic(err)
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
