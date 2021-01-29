package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Fetch latest MaxMind GeoIP Country database with license key
// into the OS's temporary directory and return path to it
func fetchGeolocationDatabase(license string) (string, error) {
	targetDir := os.TempDir()
	targetFile := filepath.Join(targetDir, "caddy-analytics-maxmind-geolite2-country.mmdb")
	url := "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=" + license + "&suffix=tar.gz"

	// Fetch the latest database
	fetch := func() error {

		// Open target file
		file, err := os.Create(targetFile)
		if err != nil {
			os.Remove(targetFile)
			return err
		}
		defer file.Close()

		// Download the database
		response, err := http.Get(url)
		if err != nil {
			os.Remove(targetFile)
			return err
		}
		defer response.Body.Close()
		if response.StatusCode == 401 {
			return errors.New("invalid MaxMind license key")
		} else if response.StatusCode != 200 {
			return errors.New("http request failed with status code " + strconv.Itoa(response.StatusCode))
		}

		// Decompress the file
		gzipReader, err := gzip.NewReader(response.Body)
		if err != nil {
			os.Remove(targetFile)
			return err
		}

		// Untar the file
		tarReader := tar.NewReader(gzipReader)
		if err != nil {
			os.Remove(targetFile)
			return err
		}

		// Search the archive for a .mmdb file and copy it to the target file
		for {
			header, err := tarReader.Next()
			if err != nil {
				os.Remove(targetFile)
				return err
			}
			if header.Typeflag == tar.TypeReg {
				if filepath.Ext(header.Name) == ".mmdb" {
					_, err = io.Copy(file, tarReader)
					if err != nil {
						os.Remove(targetFile)
						return err
					}
					return nil
				}
			}
		}
	}

	// Create temporary directory if it doesn't exist
	err := os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	// Get database file stats and fetch if it doesn't exist
	info, err := os.Stat(targetFile)
	if os.IsNotExist(err) {
		log.Print("No cached database found, fetching...")
		return targetFile, fetch()
	} else if err != nil {
		return "", err
	} else if info.Size() < 1024*8 {
		log.Print("Invalid database found, fetching...")
		return targetFile, fetch()
	}

	// Get only the headers of the external database
	response, err := http.Head(url)
	if err != nil {
		return "", err
	}
	if response.StatusCode == 401 {
		return "", errors.New("invalid MaxMind license key")
	} else if response.StatusCode != 200 {
		return "", errors.New("http request failed with status code " + strconv.Itoa(response.StatusCode))
	}

	// Extract the build time of the most recent database
	build, err := time.Parse(time.RFC1123, response.Header["Last-Modified"][0])
	if err != nil {
		return "", err
	}

	// Get new database if it is outdated
	if build.After(info.ModTime()) {
		log.Print("Cached geo database is outdated, fetching...")
		return targetFile, fetch()
	}

	// If already existed and not out-of-date, do nothing
	log.Print("Cached geo database is up-to-date...")
	return targetFile, nil
}
