package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type aries struct {
	Identifiers []string `json:"identifier,omitempty"`
	AdminURL    []string `json:"administrative_url,omitempty"`
}

type apTrustResult struct {
	Title         string `json:"title"`
	Identifier    string `json:"identifier"`
	AltIdentifier string `json:"alt_identifier"`
	BagName       string `json:"bag_name"`
	InDPN         bool   `json:"in_dpn"`
	FileCount     int    `json:"file_count"`
	Institution   string `json:"institution"`
}

type apTrustResp struct {
	Count    int
	Next     string
	Previous string
	Results  []apTrustResult
}

// Version of the service
const version = "1.0.0"

// Config info; APTrust host and API key
var apTrustURL string
var apTrustAPIUser string
var apTrustAPIKey string

// favHandler is a dummy handler to silence browser API requests that look for /favicon.ico
func favHandler(c *gin.Context) {
}

// versionHandler reports the version of the serivce
func versionHandler(c *gin.Context) {
	c.String(http.StatusOK, "Aries APTrust version %s", version)
}

// healthCheckHandler reports the health of the serivce
func healthCheckHandler(c *gin.Context) {
	hcMap := make(map[string]string)
	hcMap["AriesAPTrust"] = "true"
	// ping the api with a minimal request to see if it is alive
	url := fmt.Sprintf("%s/objects?perpage=1", apTrustURL)
	_, err := getAPIResponse(url)
	if err != nil {
		hcMap["APTrust"] = "false"
	} else {
		hcMap["APTrust"] = "true"
	}
	c.JSON(http.StatusOK, hcMap)
}

/// ariesPing handles requests to the aries endpoint with no params.
// Just returns and alive message
func ariesPing(c *gin.Context) {
	c.String(http.StatusOK, "APTrust Aries API")
}

// ariesLookup will query APTrust for information on the supplied identifer
func ariesLookup(c *gin.Context) {
	passedID := c.Param("id")
	out := aries{}
	hasMatches := false

	// Try several APTrust API endpoints to search for the ID.
	// Collect all and return results as JSON
	// endpoints: /objects?identifier=ID, /objects?alt_identifier=ID, /objects?bag_name=ID
	// eventually include: /files?identifier_like=ID where ID is filename. Use like because
	// identifer is <institutions_domain_name>/<intellectual_object_identifier_parent_object>/<individual_file_name>
	endpoints := []string{"objects?identifier", "objects?alt_identifier", "objects?bag_name"}
	for _, ep := range endpoints {
		url := fmt.Sprintf("%s/%s=%s", apTrustURL, ep, passedID)
		respStr, err := getAPIResponse(url)
		if err != nil {
			log.Printf("API returned error: %s", err.Error())
		} else {
			var resp apTrustResp
			marshallErr := json.Unmarshal([]byte(respStr), &resp)
			if marshallErr != nil {
				log.Printf("Unable to parse response: %s", marshallErr.Error())
				continue
			}

			if resp.Count > 0 {
				hasMatches = true
				for _, hit := range resp.Results {
					log.Printf("Match %s", hit.Title)
					out.Identifiers = append(out.Identifiers, hit.Identifier)
					out.Identifiers = append(out.Identifiers, hit.AltIdentifier)
					out.Identifiers = append(out.Identifiers, hit.BagName)
					idx := strings.Index(apTrustURL, "/member-api")
					url := fmt.Sprintf("%s/objects/%s", apTrustURL[0:idx], hit.Identifier)
					out.AdminURL = append(out.AdminURL, url)
				}
			}
		}
	}

	if hasMatches {
		c.JSON(http.StatusOK, out)
	} else {
		c.String(http.StatusNotFound, "%s not found", passedID)
	}
}

// getAPIResponse is a helper used to call a JSON endpoint and return the resoponse as a string
func getAPIResponse(url string) (string, error) {
	log.Printf("Get resonse for: %s", url)
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Unable to create request for %s", url)
		return "", err
	}

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Pharos-API-User", apTrustAPIUser)
	req.Header.Set("X-Pharos-API-Key", apTrustAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	respString := string(bodyBytes)
	if resp.StatusCode != 200 {
		return "", errors.New(respString)
	}
	return respString, nil
}

/**
 * MAIN
 */
func main() {
	log.Printf("===> Aries APTrust service staring up <===")

	// Get config params
	log.Printf("Read configuration...")
	var port int
	flag.IntVar(&port, "port", 8080, "Aries APTrust port (default 8080)")
	flag.StringVar(&apTrustURL, "aptrust", "", "APTrust base URL")
	flag.StringVar(&apTrustAPIUser, "apiuser", "", "APTrust API User")
	flag.StringVar(&apTrustAPIKey, "apikey", "", "APTrust API Key")
	flag.Parse()

	log.Printf("Setup routes...")
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.GET("/favicon.ico", favHandler)
	router.GET("/version", versionHandler)
	router.GET("/healthcheck", healthCheckHandler)
	api := router.Group("/api")
	{
		api.GET("/aries", ariesPing)
		api.GET("/aries/:id", ariesLookup)
	}

	portStr := fmt.Sprintf(":%d", port)
	log.Printf("Start Aries APTrust v%s on port %s", version, portStr)
	log.Fatal(router.Run(portStr))
}
