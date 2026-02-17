package html

import (
	"encoding/base64"
	"fmt"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/labstack/echo/v4"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RegisterImageRoutes registers image-related routes such as /image/webp
func RegisterImageRoutes(e *echo.Echo) {
	e.GET("/image/webp", func(c echo.Context) error {
		src := c.QueryParam("src")
		wStr := c.QueryParam("w")
		hStr := c.QueryParam("h")
		typeStr := c.QueryParam("type")
		qStr := c.QueryParam("q")

		// Set cache control headers
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")

		if src == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "src parameter is required"})
		}

		// Parse width and height
		var width, height int
		if wStr != "" {
			width, _ = strconv.Atoi(wStr)
		}
		if hStr != "" {
			height, _ = strconv.Atoi(hStr)
		}

		imgType := "webp"
		if typeStr != "" {
			imgType = typeStr
		}
		quality := 95
		if qStr != "" {
			if q, err := strconv.Atoi(qStr); err == nil && q >= 1 && q <= 100 {
				quality = q
			}
		}

		// Generate cache key using base64 (URL encoding, no padding)
		cacheKeyRaw := fmt.Sprintf("%s_%d_%d_%s_%d", src, width, height, imgType, quality)
		cacheKey := base64.RawURLEncoding.EncodeToString([]byte(cacheKeyRaw))
		/*if len(cacheKey) > 64 {
			cacheKey = cacheKey[:64] // truncate for filesystem safety
		}*/
		var ext string
		switch imgType {
		case "jpeg", "jpg":
			ext = ".jpg"
		case "png":
			ext = ".png"
		case "webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
		cacheDir := "var/cache/image_cache"
		cachePath := filepath.Join(cacheDir, cacheKey+ext)

		// Serve from cache if exists
		if f, err := os.Open(cachePath); err == nil {
			defer f.Close()
			switch imgType {
			case "jpeg", "jpg":
				c.Response().Header().Set("Content-Type", "image/jpeg")
			case "png":
				c.Response().Header().Set("Content-Type", "image/png")
			case "webp":
				c.Response().Header().Set("Content-Type", "image/webp")
			default:
				c.Response().Header().Set("Content-Type", "image/jpeg")
			}
			io.Copy(c.Response(), f)
			return nil
		}

		// Load and process image
		var img image.Image
		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
			resp, err := http.Get(src)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "failed to fetch remote image"})
			}
			defer resp.Body.Close()
			img, _, err = image.Decode(resp.Body)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to decode remote image"})
			}
		} else {
			file, err := os.Open(src)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "image not found"})
			}
			defer file.Close()
			img, _, err = image.Decode(file)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to decode image"})
			}
		}

		// Resize if width or height is specified, keeping aspect ratio by default
		if (width > 0 || height > 0) && img != nil {
			origBounds := img.Bounds()
			origW := origBounds.Dx()
			origH := origBounds.Dy()

			// If both dimensions specified, fit and pad
			if width > 0 && height > 0 {
				// Calculate target dimensions while maintaining aspect ratio
				ratioW := float64(width) / float64(origW)
				ratioH := float64(height) / float64(origH)
				var resizeW, resizeH int

				if ratioW < ratioH {
					// Width is the constraining factor
					resizeW = width
					resizeH = int(float64(origH) * ratioW)
				} else {
					// Height is the constraining factor
					resizeH = height
					resizeW = int(float64(origW) * ratioH)
				}

				// Resize the image with better quality settings
				resized := imaging.Resize(img, resizeW, resizeH, imaging.CatmullRom)

				// Create a new white background image of the target size
				background := imaging.New(width, height, image.White)

				// Calculate position to center the resized image
				posX := (width - resizeW) / 2
				posY := (height - resizeH) / 2

				// Paste the resized image onto the white background
				img = imaging.Paste(background, resized, image.Point{posX, posY})
			} else if width > 0 {
				// Only width specified - calculate height
				height = int(float64(width) * float64(origH) / float64(origW))
				img = imaging.Resize(img, width, height, imaging.CatmullRom)
			} else if height > 0 {
				// Only height specified - calculate width
				width = int(float64(height) * float64(origW) / float64(origH))
				img = imaging.Resize(img, width, height, imaging.CatmullRom)
			}
		}
		// Ensure cache directory exists
		os.MkdirAll(cacheDir, 0755)
		f, err := os.Create(cachePath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create cache file"})
		}
		defer f.Close()

		// Encode and save to cache, and also serve
		switch imgType {
		case "jpeg", "jpg":
			c.Response().Header().Set("Content-Type", "image/jpeg")
			imaging.Encode(io.MultiWriter(c.Response(), f), img, imaging.JPEG, imaging.JPEGQuality(quality))
		case "png":
			c.Response().Header().Set("Content-Type", "image/png")
			imaging.Encode(io.MultiWriter(c.Response(), f), img, imaging.PNG)
		case "webp":
			c.Response().Header().Set("Content-Type", "image/webp")
			opts := &webp.Options{
				Quality:  float32(quality),
				Exact:    true, // Preserve color accuracy
				Lossless: true, // Use lossless compression for best quality
			}
			webp.Encode(io.MultiWriter(c.Response(), f), img, opts)
		default:
			c.Response().Header().Set("Content-Type", "image/jpeg")
			imaging.Encode(io.MultiWriter(c.Response(), f), img, imaging.JPEG, imaging.JPEGQuality(quality))
		}
		return nil
	})
}
