package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/cenkalti/dominantcolor"
	"github.com/herlon214/experiments/downloader"
)

type DatasetImage struct {
	Src    string `json:"src"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}
type DatasetRow struct {
	Image DatasetImage `json:"image"`
}
type DatasetRows struct {
	Row DatasetRow `json:"row"`
}
type Dataset struct {
	Rows []DatasetRows `json:"rows"`
}

type AlgoliaRecord struct {
	ObjectID           string           `json:"objectID"`
	ImageURL           string           `json:"imageURL"`
	ColorsHex          []string         `json:"colorsHex"`
	QunatizedColorsHex map[int][]string `json:"quantizedColorsHex"`
}

func main() {
	// Connect and authenticate with your Algolia app
	client := search.NewClient(os.Getenv("ALGOLIA_APP_ID"), os.Getenv("ALGOLIA_WRITE_KEY"))

	// Create a new index and add a record
	index := client.InitIndex("exp_image_search")
	printBlocks(index)
	// generateData()
}

func printBlocks(index *search.Index) {
	for i := 1; i < 9; i++ {
		values, err := index.SearchForFacetValues(fmt.Sprintf("quantizedColorsHex.%d", i), "")
		if err != nil {
			panic(err)
		}

		colors := make([]string, 0)
		for _, item := range values.FacetHits {
			colors = append(colors, item.Value)
		}

		err = generateColorGrid(colors, fmt.Sprintf("blocks_%dbits.png", i))
		if err != nil {
			fmt.Println("Error generating color grid:", err)
		}
	}

}

func generateData(index *search.Index) {
	size := 20
	input := make(chan string, size)
	d := downloader.NewDownloader(input, size)

	urls := loadItems()

	output := d.Start(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for _, url := range urls {
			input <- url
		}
	}()

	go func() {
		defer wg.Done()

		var wg2 sync.WaitGroup
		wg2.Add(len(urls))
		for i := 0; i < size; i++ {
			go func() {
				for result := range output {
					wg2.Done()

					fmt.Println("received body", len(result.Body), result.Source)
					img, _, err := image.Decode(bytes.NewReader(result.Body))
					if err != nil {
						panic(err)
					}

					colors := dominantcolor.FindN(img, 5)
					colorsHex := make([]string, 0)
					quantizedHexColorsBits := make(map[int][]string, 0)
					hash := md5.Sum([]byte(result.Source))
					id := hex.EncodeToString(hash[:])
					for _, color := range colors {
						colorsHex = append(colorsHex, dominantcolor.Hex(color))
					}

					// Quantize colors in different bit sizes
					for i := 1; i < 9; i++ {
						quantized := make([]string, 0)
						for _, color := range colors {
							quantized = append(quantized, dominantcolor.Hex(quantizeColor(color, i)))
						}

						quantizedHexColorsBits[i] = quantized
					}

					resSave, err := index.PartialUpdateObject(AlgoliaRecord{
						ObjectID:           id,
						ImageURL:           result.Source,
						ColorsHex:          colorsHex,
						QunatizedColorsHex: quantizedHexColorsBits,
					})
					if err != nil {
						panic(err)
					}
					resSave.Wait()
				}
			}()
		}

		wg2.Wait()
	}()

	wg.Wait()
}

func loadItems() []string {
	res, err := http.Get("https://datasets-server.huggingface.co/first-rows?dataset=sasha%2Fdog-food&config=sasha--dog-food&split=train")
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	values := make([]string, 0)

	var output Dataset
	err = json.Unmarshal(body, &output)
	if err != nil {
		panic(err)
	}

	for _, item := range output.Rows {
		values = append(values, item.Row.Image.Src)
	}

	return values
}

func quantizeColor(col color.RGBA, bits int) color.RGBA {
	maxValue := int(math.Pow(2, float64(bits)))
	colorStep := 256 / maxValue

	rQuantized := (int(col.R) / colorStep) * colorStep
	gQuantized := (int(col.G) / colorStep) * colorStep
	bQuantized := (int(col.B) / colorStep) * colorStep

	return color.RGBA{uint8(rQuantized), uint8(gQuantized), uint8(bQuantized), col.A}
}

func generateColorGrid(colors []string, outputFile string) error {
	const blockSize = 100
	rows := 1
	cols := len(colors)

	imgWidth := cols * blockSize
	imgHeight := rows * blockSize

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	for i, hexColor := range colors {
		col, err := parseHexColor(hexColor)
		if err != nil {
			return err
		}

		block := image.Rect(i*blockSize, 0, (i+1)*blockSize, blockSize)
		draw.Draw(img, block, &image.Uniform{col}, image.ZP, draw.Src)
	}

	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	return png.Encode(output, img)
}

func parseHexColor(s string) (color.RGBA, error) {
	c := color.RGBA{}

	if s[0] != '#' || (len(s) != 7 && len(s) != 9) {
		return c, fmt.Errorf("invalid hex color format")
	}

	r, err := strconv.ParseUint(s[1:3], 16, 8)
	if err != nil {
		return c, err
	}
	g, err := strconv.ParseUint(s[3:5], 16, 8)
	if err != nil {
		return c, err
	}
	b, err := strconv.ParseUint(s[5:7], 16, 8)
	if err != nil {
		return c, err
	}

	c.R = uint8(r)
	c.G = uint8(g)
	c.B = uint8(b)
	c.A = 0xFF

	if len(s) == 9 {
		a, err := strconv.ParseUint(s[7:9], 16, 8)
		if err != nil {
			return c, err
		}
		c.A = uint8(a)
	}

	return c, nil
}
