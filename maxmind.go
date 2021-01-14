package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Fetch latest MaxMind GeoIP Country database with license key
// into the OS's temporary directory and return path to it
func fetchGeolocationDatabase(license string) (string, error) {
	targetFile := filepath.Join(os.TempDir(), "caddy-analytics-maxmind-geolite2-country.mmdb")
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

	// Get database file stats and fetch if it doesn't exist
	info, err := os.Stat(targetFile)
	if os.IsNotExist(err) {
		log.Print("No cached database found, downloading now...")
		return targetFile, fetch()
	} else if err != nil {
		return "", err
	}

	// Get only the headers of the external database
	response, err := http.Head(url)
	if err != nil {
		return "", err
	}

	// Extract the build time of the most recent database
	build, err := time.Parse(time.RFC1123, response.Header["Last-Modified"][0])
	if err != nil {
		return "", err
	}

	// Get new database if it is outdated
	if build.After(info.ModTime()) {
		log.Print("Cached geolocation database is outdated, fetching new one...")
		return targetFile, fetch()
	}

	// If already existed and not out-of-date, do nothing
	log.Print("Cached geolocation database is up-to-date, skipping downloading new one...")
	return targetFile, nil
}
