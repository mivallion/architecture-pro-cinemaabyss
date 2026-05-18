package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Proxy implements StrictServerInterface with reverse-proxy logic.
// It routes requests to either the monolith or the extracted microservices
// based on the Strangler Fig pattern configuration.
type Proxy struct {
	MonolithURL            string
	MoviesServiceURL       string
	EventsServiceURL       string
	GradualMigration       bool
	MoviesMigrationPercent int
}

var _ StrictServerInterface = (*Proxy)(nil)

func NewProxy(monolithURL, moviesServiceURL, eventsServiceURL string, gradualMigration bool, moviesMigrationPercent int) *Proxy {
	return &Proxy{
		MonolithURL:            monolithURL,
		MoviesServiceURL:       moviesServiceURL,
		EventsServiceURL:       eventsServiceURL,
		GradualMigration:       gradualMigration,
		MoviesMigrationPercent: moviesMigrationPercent,
	}
}

// shouldRouteToMoviesService determines if a movie request should go to the
// extracted microservice based on the migration percentage (Strangler Fig pattern).
func (p *Proxy) shouldRouteToMoviesService() bool {
	if !p.GradualMigration {
		return true
	}
	return rand.Intn(100) < p.MoviesMigrationPercent
}

// --- Health endpoints ---

func (p *Proxy) GetProxyHealth(ctx context.Context, request GetProxyHealthRequestObject) (GetProxyHealthResponseObject, error) {
	return GetProxyHealth200TextResponse("Strangler Fig Proxy is healthy"), nil
}

func (p *Proxy) GetMoviesServiceHealth(ctx context.Context, request GetMoviesServiceHealthRequestObject) (GetMoviesServiceHealthResponseObject, error) {
	resp, err := http.Get(p.MoviesServiceURL + "/api/movies/health")
	if err != nil {
		log.Printf("movies service health check failed: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadGateway, "movies service unreachable")
	}
	defer resp.Body.Close()

	status := resp.StatusCode == http.StatusOK
	return GetMoviesServiceHealth200JSONResponse{Status: &status}, nil
}

func (p *Proxy) GetEventsServiceHealth(ctx context.Context, request GetEventsServiceHealthRequestObject) (GetEventsServiceHealthResponseObject, error) {
	resp, err := http.Get(p.EventsServiceURL + "/api/events/health")
	if err != nil {
		log.Printf("events service health check failed: %v", err)
		return nil, echo.NewHTTPError(http.StatusBadGateway, "events service unreachable")
	}
	defer resp.Body.Close()

	status := resp.StatusCode == http.StatusOK
	return GetEventsServiceHealth200JSONResponse{Status: &status}, nil
}

// --- Movies endpoints ---

func (p *Proxy) GetAllMovies(ctx context.Context, request GetAllMoviesRequestObject) (GetAllMoviesResponseObject, error) {
	target := p.MonolithURL
	if p.shouldRouteToMoviesService() {
		target = p.MoviesServiceURL
	}
	result, err := fetchGET[[]Movie](target, "/api/movies")
	if err != nil {
		return nil, err
	}
	return GetAllMovies200JSONResponse(result), nil
}

func (p *Proxy) CreateMovie(ctx context.Context, request CreateMovieRequestObject) (CreateMovieResponseObject, error) {
	target := p.MonolithURL
	if p.shouldRouteToMoviesService() {
		target = p.MoviesServiceURL
	}
	result, err := fetchPOST[Movie](target, "/api/movies", request.Body)
	if err != nil {
		return nil, err
	}
	return CreateMovie201JSONResponse(result), nil
}

// --- Users endpoints (always route to monolith) ---

func (p *Proxy) GetAllUsers(ctx context.Context, request GetAllUsersRequestObject) (GetAllUsersResponseObject, error) {
	result, err := fetchGET[[]User](p.MonolithURL, "/api/users")
	if err != nil {
		return nil, err
	}
	return GetAllUsers200JSONResponse(result), nil
}

func (p *Proxy) CreateUser(ctx context.Context, request CreateUserRequestObject) (CreateUserResponseObject, error) {
	result, err := fetchPOST[User](p.MonolithURL, "/api/users", request.Body)
	if err != nil {
		return nil, err
	}
	return CreateUser201JSONResponse(result), nil
}

// --- Payments endpoints (always route to monolith) ---

func (p *Proxy) GetAllPayments(ctx context.Context, request GetAllPaymentsRequestObject) (GetAllPaymentsResponseObject, error) {
	path := "/api/payments"
	if request.Params.UserId != nil {
		path += "?user_id=" + strconv.Itoa(*request.Params.UserId)
	}
	result, err := fetchGET[[]Payment](p.MonolithURL, path)
	if err != nil {
		return nil, err
	}
	return GetAllPayments200JSONResponse(result), nil
}

func (p *Proxy) CreatePayment(ctx context.Context, request CreatePaymentRequestObject) (CreatePaymentResponseObject, error) {
	result, err := fetchPOST[Payment](p.MonolithURL, "/api/payments", request.Body)
	if err != nil {
		return nil, err
	}
	return CreatePayment201JSONResponse(result), nil
}

// --- Subscriptions endpoints (always route to monolith) ---

func (p *Proxy) GetAllSubscriptions(ctx context.Context, request GetAllSubscriptionsRequestObject) (GetAllSubscriptionsResponseObject, error) {
	path := "/api/subscriptions"
	if request.Params.UserId != nil {
		path += "?user_id=" + strconv.Itoa(*request.Params.UserId)
	}
	result, err := fetchGET[[]Subscription](p.MonolithURL, path)
	if err != nil {
		return nil, err
	}
	return GetAllSubscriptions200JSONResponse(result), nil
}

func (p *Proxy) CreateSubscription(ctx context.Context, request CreateSubscriptionRequestObject) (CreateSubscriptionResponseObject, error) {
	result, err := fetchPOST[Subscription](p.MonolithURL, "/api/subscriptions", request.Body)
	if err != nil {
		return nil, err
	}
	return CreateSubscription201JSONResponse(result), nil
}

// --- Events endpoints (always route to events service) ---

func (p *Proxy) CreateMovieEvent(ctx context.Context, request CreateMovieEventRequestObject) (CreateMovieEventResponseObject, error) {
	result, err := fetchPOST[EventResponse](p.EventsServiceURL, "/api/events/movie", request.Body)
	if err != nil {
		return nil, err
	}
	return CreateMovieEvent201JSONResponse(result), nil
}

func (p *Proxy) CreateUserEvent(ctx context.Context, request CreateUserEventRequestObject) (CreateUserEventResponseObject, error) {
	result, err := fetchPOST[EventResponse](p.EventsServiceURL, "/api/events/user", request.Body)
	if err != nil {
		return nil, err
	}
	return CreateUserEvent201JSONResponse(result), nil
}

func (p *Proxy) CreatePaymentEvent(ctx context.Context, request CreatePaymentEventRequestObject) (CreatePaymentEventResponseObject, error) {
	result, err := fetchPOST[EventResponse](p.EventsServiceURL, "/api/events/payment", request.Body)
	if err != nil {
		return nil, err
	}
	return CreatePaymentEvent201JSONResponse(result), nil
}

// --- Generic proxy helpers ---

func fetchGET[T any](target, path string) (T, error) {
	var zero T
	url := target + path
	log.Printf("[proxy] GET %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[proxy] GET %s failed: %v", url, err)
		return zero, echo.NewHTTPError(http.StatusBadGateway, "upstream service unreachable")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[proxy] GET %s returned %d: %s", url, resp.StatusCode, string(body))
		return zero, echo.NewHTTPError(resp.StatusCode, string(body))
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

func fetchPOST[T any](target, path string, body interface{}) (T, error) {
	var zero T
	url := target + path
	log.Printf("[proxy] POST %s", url)

	data, err := json.Marshal(body)
	if err != nil {
		return zero, fmt.Errorf("marshal request body: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("[proxy] POST %s failed: %v", url, err)
		return zero, echo.NewHTTPError(http.StatusBadGateway, "upstream service unreachable")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[proxy] POST %s returned %d: %s", url, resp.StatusCode, string(respBody))
		return zero, echo.NewHTTPError(resp.StatusCode, string(respBody))
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}
