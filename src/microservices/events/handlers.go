package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type EventHandler struct {
	producer *KafkaProducer
}

func NewEventHandler(producer *KafkaProducer) *EventHandler {
	return &EventHandler{producer: producer}
}

func (h *EventHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]bool{"status": true})
}

func (h *EventHandler) CreateMovieEvent(c echo.Context) error {
	var evt MovieEvent
	if err := c.Bind(&evt); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	envelope := buildEvent("movie", evt)
	payload, err := json.Marshal(envelope)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	partition, offset, err := h.producer.ProduceEvent(c.Request().Context(), "movie-events", "movie", payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, EventResponse{
		Status:    "success",
		Partition: partition,
		Offset:    offset,
		Event:     envelope,
	})
}

func (h *EventHandler) CreateUserEvent(c echo.Context) error {
	var evt UserEvent
	if err := c.Bind(&evt); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	envelope := buildEvent("user", evt)
	payload, err := json.Marshal(envelope)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	partition, offset, err := h.producer.ProduceEvent(c.Request().Context(), "user-events", "user", payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, EventResponse{
		Status:    "success",
		Partition: partition,
		Offset:    offset,
		Event:     envelope,
	})
}

func (h *EventHandler) CreatePaymentEvent(c echo.Context) error {
	var evt PaymentEvent
	if err := c.Bind(&evt); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	envelope := buildEvent("payment", evt)
	payload, err := json.Marshal(envelope)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	partition, offset, err := h.producer.ProduceEvent(c.Request().Context(), "payment-events", "payment", payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, EventResponse{
		Status:    "success",
		Partition: partition,
		Offset:    offset,
		Event:     envelope,
	})
}

func buildEvent(eventType string, payload interface{}) Event {
	id := fmt.Sprintf("evt_%d", time.Now().UnixNano())

	b, _ := json.Marshal(payload)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)

	return Event{
		ID:        id,
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Payload:   m,
	}
}
