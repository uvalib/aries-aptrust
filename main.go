package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Version of the service
const version = "1.0.0"

// favHandler is a dummy handler to silence browser API requests that look for /favicon.ico
func favHandler(c *gin.Context) {
}

// versionHandler reports the version of the serivce
func versionHandler(c *gin.Context) {
	c.String(http.StatusOK, "Aries-APTrust version %s", version)
}

// healthCheckHandler reports the health of the serivce
func healthCheckHandler(c *gin.Context) {
	// hcMap := make(map[string]string)
	// hcMap["Aries"] = "true"
	// for _, svc := range services {
	// 	if pingService(svc, false) {
	// 		hcMap[svc.Name] = "true"
	// 	} else {
	// 		hcMap[svc.Name] = "false"
	// 	}
	// }
	// c.JSON(http.StatusOK, hcMap)
	c.String(http.StatusOK, "alive")
}

/// ariesPing handles requests to the aries endpoint with no params.
// Just returns and alive message
func ariesPing(c *gin.Context) {
	c.String(http.StatusOK, "APTrust Aries API")
}

// ariesLookup rwill query APTrust for information on the supplied identifer
func ariesLookup(c *gin.Context) {
	passedID := c.Param("id")
	c.String(http.StatusNotFound, "%s not found", passedID)
}

/**
 * MAIN
 */
func main() {
	log.Printf("===> Aries APTrust proxy staring up <===")

	// Get config params
	log.Printf("Read configuration...")
	var port int
	flag.IntVar(&port, "port", 8080, "Aries-APTrust port (default 8080)")
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
	log.Printf("Start Aries APTrust Proxy v%s on port %s", version, portStr)
	log.Fatal(router.Run(portStr))
}
