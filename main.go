package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/mitchellh/mapstructure"
)

func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON), // Set content-Type headers as application/json
		middleware.Logger,          // Log API request calls
		middleware.DefaultCompress, // Compress results, mostly gzipping assets and json
		middleware.RedirectSlashes, // Redirect slashes to no slash URL versions
		middleware.Recoverer,       // Recover from panics without crashing server
		middleware.RequestID,
		middleware.URLFormat,
	)
	return router
}

// DefaultWeights to use for calculations when none are given
var DefaultWeights = AssumeDefaults()

// WeightAmounts is a translation for keynames in amounts
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
	router := Routes()

	router.Route("/v1", func(r chi.Router) {
		r.Post("/rack", RackEmPost)
		r.Get("/rack", RackEmGet) // assumes all default values and returns
	})

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route) // Walk and print out all routes
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	port := getEnv("API_PORT", "8080")

	log.Fatal(http.ListenAndServe(":"+port, router)) // Note, the port is usually gotten from the environment.
}

/* Main Logic */

//RackEmPost the main function that calculates a desired weight based on some inputs
func RackEmPost(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, ErrInternal())
}

//RackEmGet the main function that calculates a desired weight based on default inputs
func RackEmGet(w http.ResponseWriter, r *http.Request) {
	weight, err := strconv.ParseInt(r.URL.Query().Get("weight"), 10, 64)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	results, hasError := CalculateWeight(&DefaultWeights, weight)

	if hasError {
		render.Render(w, r, ErrInternal())
		return
	}

	fmt.Println(results)
	render.JSON(w, r, results)
}

//CalculateWeight Main logic for calculating weight based on a set of inputs
func CalculateWeight(input *RackInputStandard, weight int64) (RackInputStandard, bool) {
	//TODO: add logic for returning errors
	rawResult := map[string]int{}
	leftOver := int(weight) - input.BarWeight
	reflection := reflect.ValueOf(input).Elem()

	for leftOver > 0 {
		found := false
		for i := 0; i < reflection.NumField(); i++ {
			field := reflection.Field(i)
			fieldName := reflection.Type().Field(i).Name
			fieldAmount := field.Interface().(int)

			if fieldName == "BarWeight" || fieldAmount == 0 {
				continue
			}
			amount := int(WeightAmounts[fieldName] * 2)

			if amount <= leftOver {
				leftOver -= amount
				rawResult[fieldName]++
				found = true
				break
			}
		}
		if !found {
			break
		}
		fmt.Println("leftover", leftOver)
	}

	var result RackInputStandard
	mapstructure.Decode(rawResult, &result)
	return result, false
}

/* Models */

// RackInputStandard is an Input to be used in calculations
type RackInputStandard struct {
	BarWeight      int `json:"barWeight,omitempty"`
	Hundos         int `json:"hundreds,omitempty"`
	FortyFives     int `json:"fortyFives,omitempty"`
	ThirtyFives    int `json:"thirtyFives,omitempty"`
	TwentyFives    int `json:"twentyFives,omitempty"`
	Tens           int `json:"tens,omitempty"`
	Fives          int `json:"fives,omitempty"`
	TwoDotFives    int `json:"twoDotFives,omitempty"`
	OneDotTwoFives int `json:"oneDotTwoFives,omitempty"`
}

func (a *RackInputStandard) Bind(r *http.Request) error {
	return nil
}

/* Util Functions */

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// ErrResponse renderer type for handling all sorts of errors.
//
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest is used to return an invalid response to client
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

// ErrInternal is used to return an internal error response
func ErrInternal() render.Renderer {
	return &ErrResponse{
		Err:            nil,
		HTTPStatusCode: 500,
		StatusText:     "Internal Error",
		ErrorText:      "Internal Error, we apologize",
	}
}

// AssumeDefaults sets default amounts for input if none are provided
func AssumeDefaults() RackInputStandard {
	return RackInputStandard{
		BarWeight:      45,
		Hundos:         0,
		FortyFives:     6,
		ThirtyFives:    6,
		TwentyFives:    6,
		Tens:           6,
		Fives:          6,
		TwoDotFives:    6,
		OneDotTwoFives: 0,
	}
}
