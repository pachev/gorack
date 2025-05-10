// @title           Gorack API
// @version         1.0
// @description     A simple API for calculating barbell weight plates.
// @termsOfService  http://example.com/terms/

// @contact.name   Your Name
// @contact.url    http://www.github.com/pachev
// @contact.email  your.email@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /v1/api

// @tag.name Rack
// @tag.description Operations for calculating barbell weight plates

package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/mitchellh/mapstructure"

   httpSwagger "github.com/swaggo/http-swagger/v2"
   _ "github.com/pachev/gorack/docs"
)

// WeightCache provides in-memory caching for weight calculations
type WeightCache struct {
	cache map[string]*ReturnedValueStandard
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewWeightCache creates a new cache with the specified TTL
func NewWeightCache(ttl time.Duration) *WeightCache {
	return &WeightCache{
		cache: make(map[string]*ReturnedValueStandard),
		ttl:   ttl,
	}
}

// Get retrieves a cached result if it exists
func (wc *WeightCache) Get(key string) (*ReturnedValueStandard, bool) {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	result, found := wc.cache[key]
	return result, found
}

// Set stores a calculation result in the cache
func (wc *WeightCache) Set(key string, value *ReturnedValueStandard) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.cache[key] = value
	
	// Set up automatic expiration if TTL is positive
	if wc.ttl > 0 {
		go func(k string) {
			time.Sleep(wc.ttl)
			wc.mu.Lock()
			delete(wc.cache, k)
			wc.mu.Unlock()
		}(key)
	}
}

// Clear empties the cache
func (wc *WeightCache) Clear() {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.cache = make(map[string]*ReturnedValueStandard)
}

// Global cache instance
var weightCache *WeightCache

// Routes function that sets up the initial Chi Router
func Routes() *chi.Mux {
	router := chi.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		MaxAge:         300, // Maximum value not ignored by any of major browsers
	})

	router.Use(corsMiddleware.Handler)
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.RedirectSlashes,
		middleware.Recoverer,
		middleware.RequestID,
	)

	// Health check endpoints
	router.Get("/health", HealthCheck)
	router.Get("/status", HealthCheck)
	router.Get("/", HealthCheck)
	
	router.Get("/docs/*", httpSwagger.Handler())
	return router
}

// WeightAmounts is a translation for keynames in amounts (weight per single plate)
var WeightAmounts = map[string]float32{
	"Hundos":         100,
	"FortyFives":     45,
	"ThirtyFives":    35,
	"TwentyFives":    25,
	"Tens":           10,
	"Fives":          5,
	"TwoDotFives":    2.5,
	"OneDotTwoFives": 1.25,
}

func main() {
	// Initialize cache with TTL from environment or 1 hour if not set
	cacheTTL := getEnvDuration("CACHE_TTL", 1*time.Hour)
	weightCache = NewWeightCache(cacheTTL)
	
	router := Routes()

	router.Route("/v1/api", func(r chi.Router) {
		r.Post("/rack", RackEmPost)
		r.Get("/rack", RackEmGet)
	})

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route)
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error())
	}

	port := getEnv("API_PORT", "8080")
	log.Printf("Starting server on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}



// RackEmPost godoc
// @Summary      Calculate plates with custom plate availability
// @Description  Returns an optimal plate configuration for a given target weight with custom available plates
// @Tags         Rack
// @Accept       json
// @Produce      json
// @Param        request    body     RackInputStandard  true  "Desired weight and available plates"
// @Success      200  {object}  ReturnedValueStandard
// @Failure      400  {object}  ErrResponse
// @Failure      500  {object}  ErrResponse
// @Router       /rack [post]
func RackEmPost(w http.ResponseWriter, r *http.Request) {
	input := &RackInputStandard{}

	if err := render.Bind(r, input); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Generate cache key for this specific input
	cacheKey := generateCacheKey(input)
	
	// Check cache first
	if cachedResult, found := weightCache.Get(cacheKey); found {
		render.JSON(w, r, cachedResult)
		return
	}

	// Cache miss, calculate and store
	results, err := CalculateWeight(input)
	if err != nil {
		log.Printf("Error calculating weight for POST: %v\nInput: %+v\n", err, input)
		render.Render(w, r, ErrInternal())
		return
	}
	
	// Store in cache
	weightCache.Set(cacheKey, results)
	
	render.JSON(w, r, results)
}

// RackEmGet godoc
// @Summary      Calculate plates using default plate availability
// @Description  Returns an optimal plate configuration for a given target weight
// @Tags         Rack
// @Accept       json
// @Produce      json
// @Param        weight    query     int  true  "Desired weight in pounds"
// @Success      200  {object}  ReturnedValueStandard
// @Failure      400  {object}  ErrResponse
// @Failure      500  {object}  ErrResponse
// @Router       /rack [get]
func RackEmGet(w http.ResponseWriter, r *http.Request) {
	inputWithDefaults := AssumeDefaults()
	weightStr := r.URL.Query().Get("weight")
	if weightStr == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("query parameter 'weight' is required")))
		return
	}
	weight, err := strconv.ParseInt(weightStr, 10, 64)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid 'weight' parameter: must be an integer")))
		return
	}
	if int(weight) <= 0 {
		render.Render(w, r, ErrInvalidRequest(errors.New("'weight' must be a positive integer")))
		return
	}
	inputWithDefaults.DesiredWeight = int(weight)

	// Validate DesiredWeight against BarWeight for GET requests
	if inputWithDefaults.DesiredWeight <= inputWithDefaults.BarWeight {
		render.Render(w, r, ErrInvalidRequest(errors.New("desired weight must be greater than bar weight")))
		return
	}

	// Check cache for GET request with standard plates
	cacheKey := fmt.Sprintf("get:%d", inputWithDefaults.DesiredWeight)
	if cachedResult, found := weightCache.Get(cacheKey); found {
		render.JSON(w, r, cachedResult)
		return
	}

	results, calcErr := CalculateWeight(&inputWithDefaults)
	if calcErr != nil {
		log.Printf("Error calculating weight for GET: %v\nInput: %+v\n", calcErr, inputWithDefaults)
		render.Render(w, r, ErrInternal())
		return
	}
	
	// Store in cache
	weightCache.Set(cacheKey, results)
	
	render.JSON(w, r, results)
}

// generateCacheKey creates a unique key for caching based on input parameters
func generateCacheKey(input *RackInputStandard) string {
	return fmt.Sprintf("post:bar=%d:desired=%d:h=%d:45=%d:35=%d:25=%d:10=%d:5=%d:2.5=%d:1.25=%d",
		input.BarWeight,
		input.DesiredWeight,
		input.Hundos,
		input.FortyFives,
		input.ThirtyFives,
		input.TwentyFives,
		input.Tens,
		input.Fives,
		input.TwoDotFives,
		input.OneDotTwoFives,
	)
}

// CalculateWeight is the core logic for calculating plates needed.
// Input represents available plates. Output represents plates to use.
func CalculateWeight(inputAvailablePlates *RackInputStandard) (*ReturnedValueStandard, error) {
	platesToUse := map[string]int{} // Stores count of each plate type (pair) to load
	currentBarWeight := inputAvailablePlates.BarWeight
	if currentBarWeight < 0 { // Ensure bar weight is not negative
	    currentBarWeight = 0
    }
	
	leftOver := inputAvailablePlates.DesiredWeight - currentBarWeight
	achievedWeight := currentBarWeight

	// Create a mutable copy of the input available plates for deduction during calculation
	currentAvailablePlates := *inputAvailablePlates

	// Define order of plates to try, from heaviest to lightest.
	orderedPlateNames := []string{
		"Hundos", "FortyFives", "ThirtyFives", "TwentyFives",
		"Tens", "Fives", "TwoDotFives", "OneDotTwoFives",
	}

	for leftOver > 0 {
		foundPlateInIteration := false
		for _, plateName := range orderedPlateNames {
			
			var plateAvailableCount int
			val := reflect.ValueOf(&currentAvailablePlates).Elem()
			fieldVal := val.FieldByName(plateName)
			if fieldVal.IsValid() {
				plateAvailableCount = int(fieldVal.Int())
			} else {
				// This should not happen if orderedPlateNames matches RackInputStandard fields
				return nil, errors.New("internal error: plate name mismatch: " + plateName)
			}


			if plateAvailableCount == 0 {
				continue
			}

			plateWeightPerSingle, ok := WeightAmounts[plateName]
			if !ok {
				return nil, errors.New("internal error: weight definition missing for " + plateName)
			}
			weightOfPair := int(plateWeightPerSingle * 2)

			if weightOfPair > 0 && weightOfPair <= leftOver {
				leftOver -= weightOfPair
				platesToUse[plateName]++
				achievedWeight += weightOfPair
				currentAvailablePlates.DecreaseWeight(plateName)
				foundPlateInIteration = true
				break // Greedily take the heaviest possible, then restart outer loop for next heaviest
			}
		}
		if !foundPlateInIteration {
			break // No suitable plate could be added in this pass
		}
	}

	// Prepare the output struct containing plates to use
	outputPlates := RackInputStandard{
		BarWeight:     currentBarWeight,
		DesiredWeight: inputAvailablePlates.DesiredWeight,
	}
	// Populate outputPlates with the counts from platesToUse map
	decoderConfig := &mapstructure.DecoderConfig{
		Result:   &outputPlates,
		TagName:  "json",
		Squash:   true,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
	    log.Printf("Error creating mapstructure decoder: %v", err)
		return nil, errors.New("internal error creating decoder")
	}
	if err := decoder.Decode(platesToUse); err != nil {
	    log.Printf("Error decoding plates map to struct: %v", err)
		return nil, errors.New("internal error decoding result")
	}


	return &ReturnedValueStandard{
		RackInputStandard: &outputPlates,
		AchievedWeight:    achievedWeight,
		Message:           "You got this!",
	}, nil
}

/* Models */

// RackInputStandard defines the structure for API input (available plates)
// and also for the plates to be used in the output.
// Plate counts are number of PAIRS.
type RackInputStandard struct {
	BarWeight      int `json:"barWeight,omitempty"`
	Hundos         int `json:"hundreds,omitempty"` // JSON tag "hundreds" for API compatibility
	FortyFives     int `json:"fortyFives,omitempty"`
	ThirtyFives    int `json:"thirtyFives,omitempty"`
	TwentyFives    int `json:"twentyFives,omitempty"`
	Tens           int `json:"tens,omitempty"`
	Fives          int `json:"fives,omitempty"`
	TwoDotFives    int `json:"twoDotFives,omitempty"`
	OneDotTwoFives int `json:"oneDotTwoFives,omitempty"`
	DesiredWeight  int `json:"desiredWeight"` // Required in input
}

// Bind is a method on RackInputStandard to process and validate the request payload.
func (ris *RackInputStandard) Bind(r *http.Request) error {
	if ris.BarWeight == 0 { // If not provided, default to standard Olympic bar
		ris.BarWeight = AssumeDefaults().BarWeight
	}
    if ris.BarWeight < 0 {
        return errors.New("bar weight cannot be negative")
    }
	if ris.DesiredWeight == 0 {
		return errors.New("a valid desired weight must be provided")
	}
	if ris.DesiredWeight <= ris.BarWeight {
		return errors.New("desired weight must be greater than bar weight")
	}
	// Plate counts (Hundos, FortyFives, etc.) default to 0 if not in payload,
	// meaning "0 pairs available" for POST requests.
	return nil
}

// DecreaseWeight reduces the count of a specific plate type.
// This is called on the *copy* of available plates during calculation.
func (ris *RackInputStandard) DecreaseWeight(plateName string) {
	switch plateName {
	case "Hundos":
		ris.Hundos--
	case "FortyFives":
		ris.FortyFives--
	case "ThirtyFives":
		ris.ThirtyFives--
	case "TwentyFives":
		ris.TwentyFives--
	case "Tens":
		ris.Tens--
	case "Fives":
		ris.Fives--
	case "TwoDotFives":
		ris.TwoDotFives--
	case "OneDotTwoFives":
		ris.OneDotTwoFives--
	}
}

// ReturnedValueStandard is the structure of the JSON response.
type ReturnedValueStandard struct {
	*RackInputStandard        // Embeds the plates *to use* for the lift
	AchievedWeight     int    `json:"achievedWeight"`
	Message            string `json:"message,omitempty"`
}

// HealthCheck godoc
// @Summary      Health check endpoint
// @Description  Returns status of the API server
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
// @Router       /status [get]
// @Router       /v1/api/health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := map[string]string{
		"status":  "ok",
		"service": "gorack-api",
		"version": "1.0",
	}
	render.JSON(w, r, status)
}

/* Util Functions */

// getEnv retrieves an environment variable or returns a fallback value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvDuration retrieves a time.Duration environment variable or returns a fallback value.
func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

// ErrResponse is a generic renderer for API error responses.
type ErrResponse struct {
	Err            error  `json:"-"` // Low-level runtime error (not exposed to client)
	HTTPStatusCode int    `json:"-"` // HTTP response status code
	StatusText     string `json:"status"`          // User-level status message
	AppCode        int64  `json:"code,omitempty"`  // Application-specific error code
	ErrorText      string `json:"error,omitempty"` // Application-level error message for debugging
}

// Render sets the HTTP status code for the error response.
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest creates a standardized "400 Bad Request" response.
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

// ErrInternal creates a standardized "500 Internal Server Error" response.
func ErrInternal() render.Renderer {
	return &ErrResponse{
		Err:            nil, // Specific error is logged server-side
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "Internal Server Error.",
		ErrorText:      "An unexpected error occurred. Please try again later.",
	}
}

// AssumeDefaults provides a default set of available plates and bar weight.
// Used for GET requests where the user doesn't specify their available equipment.
func AssumeDefaults() RackInputStandard {
	return RackInputStandard{
		BarWeight:      45,  // Standard Olympic bar weight in lbs
		Hundos:         10, 
		FortyFives:     10,
		ThirtyFives:    10,
		TwentyFives:    10,
		Tens:           10,
		Fives:          10,
		TwoDotFives:    10,
		OneDotTwoFives: 10,
	}
}
