package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ToJsonString(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func PrettyJSON(value interface{}) string {
	statsJson, err := ToJsonString(value)

	if err != nil {
		return "Failed to convert to JSON"
	}

	var prettyJSON bytes.Buffer
	prettyJSONError := json.Indent(&prettyJSON, []byte(statsJson), "", "  ")

	if prettyJSONError != nil {
		return "Failed to format JSON"
	}

	return prettyJSON.String()
}

func CalculateDirSize(dirPath string) float32 {
	size := float32(0)
	files, _ := os.ReadDir(dirPath)

	for _, file := range files {
		filePath := dirPath + "/" + file.Name()
		stat, _ := file.Info()

		if stat.IsDir() {
			size += CalculateDirSize(filePath)
		} else {
			size += float32(stat.Size())
		}
	}

	return size
}

func FormatFileSize(bytes int64) string {
	const unit = 1024

	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	// Show with 2 decimal precision
	kb := float64(bytes) / unit

	if kb > 1024 {
		mb := kb / 1024
		return fmt.Sprintf("%.2fKB (%.2fMB)", kb, mb)
	}

	return fmt.Sprintf("%.2fKB", kb)
}

// Common asset file extensions
var assetExtensions = map[string]bool{
	".jpg":   true,
	".jpeg":  true,
	".png":   true,
	".gif":   true,
	".svg":   true,
	".ico":   true,
	".webp":  true,
	".avif":  true,
	".woff":  true,
	".woff2": true,
	".ttf":   true,
	".eot":   true,
}

// FindAndCalculateAssetsSize recursively finds all asset files in the build directory
// and returns their total size in bytes
func FindAndCalculateAssetsSize(buildPath string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(buildPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if the file extension matches any asset extension
		ext := strings.ToLower(filepath.Ext(path))
		if assetExtensions[ext] {
			totalSize += info.Size()
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return totalSize, nil
}

// IsAssetFile checks if a file is an asset based on its extension
func IsAssetFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return assetExtensions[ext]
}
