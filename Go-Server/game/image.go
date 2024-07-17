package game

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net/http"
	"websocket/constants"

	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp" // Import the WebP format support
)

func CreateSpriteSheet(imageURLs []string) ([]byte, error) {

	// Create a new RGBA image for the sprite sheet
	spriteSheet := image.NewRGBA(image.Rect(0, 0, constants.SPRITE_SHEET_SIZE, constants.SPRITE_SHEET_SIZE))

	for i, url := range imageURLs {
		// Download the image
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error downloading image[%d]: %v\n", i, err)
			return nil, err
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			log.Printf("Error response for image[%d]: %v\n", i, resp.Status)
			return nil, fmt.Errorf("error response for image[%d]: %v", i, resp.Status)
		}

		// Read the image data
		imgData, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading image data[%d]: %v\n", i, err)
			return nil, fmt.Errorf("error reading image data[%d]: %v", i, err)
		}

		// Check if image data is read correctly
		//log.Printf("Image data[%d]: %v\n", i, imgData[:20]) // Print first 100 bytes for debugging

		// Decode the image
		img, _, err := image.Decode(bytes.NewReader(imgData))
		if err != nil {
			log.Printf("Error decoding image[%d]: %v\n", i, err)
			return nil, fmt.Errorf("error decoding image[%d]: %v", i, err)
		}

		// Resize the image to 512x512
		resizedImg := resize.Resize(constants.IMAGE_SIZE, constants.IMAGE_SIZE, img, resize.Lanczos3)

		// Calculate position
		x := (i % constants.IMAGE_COUNT_COLUMN) * constants.IMAGE_SIZE
		y := (i / constants.IMAGE_COUNT_COLUMN) * constants.IMAGE_SIZE

		// Draw the image onto the sprite sheet
		draw.Draw(spriteSheet, image.Rect(x, y, x+constants.IMAGE_SIZE, y+constants.IMAGE_SIZE), resizedImg, image.Point{0, 0}, draw.Src)
	}

	// Encode the sprite sheet to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, spriteSheet); err != nil {
		log.Printf("Error encoding sprite sheet: %v\n", err)
		return nil, err
	}

	return buf.Bytes(), nil
}
