package handlers

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"motrava/app/middleware"
	"motrava/core/models"
	portUsecase "motrava/core/port/usecase"
	"motrava/core/utils/response"
	wsInfra "motrava/infra/ws"
)

type WSHandler struct {
	tripUsecase portUsecase.TripUsecase
	hub         *wsInfra.Hub
	rdb         *redis.Client
	log         *slog.Logger
	jwtSecret   string
	jwtIssuer   string
}

func NewWSHandler(
	tripUsecase portUsecase.TripUsecase,
	hub *wsInfra.Hub,
	rdb *redis.Client,
	logger *slog.Logger,
	jwtSecret, jwtIssuer string,
) *WSHandler {
	return &WSHandler{
		tripUsecase: tripUsecase,
		hub:         hub,
		rdb:         rdb,
		log:         logger,
		jwtSecret:   jwtSecret,
		jwtIssuer:   jwtIssuer,
	}
}

func (h *WSHandler) Upgrade(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return response.Error(c, fiber.StatusUnauthorized, "token is required", nil)
	}

	userID, err := h.validateToken(token)
	if err != nil {
		h.log.Error("ws auth failed", "module", "ws_handler", "error", err)
		return response.Error(c, fiber.StatusUnauthorized, "invalid token", nil)
	}

	c.Locals("ws_user_id", userID.String())

	return websocket.New(func(conn *websocket.Conn) {
		h.handleConnection(conn, userID)
	})(c)
}

func (h *WSHandler) handleConnection(conn *websocket.Conn, userID uuid.UUID) {
	var client *wsInfra.Client
	tripID := ""
	defer func() {
		if client != nil {
			h.hub.Unregister(client)
		}
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.log.Error("ws read error", "module", "ws_handler", "error", err)
			break
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(msg, &raw); err != nil {
			continue
		}

		msgType, _ := raw["type"].(string)
		h.log.Info("ws message", "module", "ws_handler", "type", msgType, "body", string(msg))

		switch msgType {
		case "auth":
			tripID, _ = raw["trip_id"].(string)
			if tripID == "" {
				continue
			}

			client = wsInfra.NewClient(h.hub, tripID, userID.String())
			h.hub.Register(client)

			go func() {
				for data := range client.Send {
					if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
						h.log.Error("ws write error", "module", "ws_handler", "error", err)
						return
					}
				}
			}()

		case "end_trip":
			if tripID == "" {
				continue
			}
			if _, err := h.tripUsecase.EndTrip(userID.String(), tripID); err != nil {
				h.log.Error("ws end trip failed", "module", "ws_handler", "error", err)
			}
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"trip_ended","trip_id":"`+tripID+`"}`))
			return

		default:
			if tripID == "" || client == nil {
				continue
			}

			_, hasLat := raw["latitude"].(float64)
			_, hasLon := raw["longitude"].(float64)
			if !hasLat || !hasLon {
				continue
			}

			tid, ok := raw["trip_id"].(string)
			if ok && tid != "" {
				tripID = tid
			}

			payload, ok := h.parseLocationPayload(raw)
			if !ok {
				continue
			}
			payload.TripID = tripID

			recordedAt, err := time.Parse(time.RFC3339, payload.Timestamp)
			if err != nil {
				recordedAt = time.Now().UTC()
			}

			tripPoint := models.TripPoint{
				ID:         uuid.New(),
				TripID:     uuid.MustParse(tripID),
				Latitude:   payload.Latitude,
				Longitude:  payload.Longitude,
				Speed:      payload.Speed,
				Heading:    payload.Heading,
				Accuracy:   payload.Accuracy,
				Altitude:   payload.Altitude,
				Battery:    payload.Battery,
				RecordedAt: recordedAt,
			}

			go func() {
				if err := h.tripUsecase.ProcessLocation(userID.String(), tripID, tripPoint); err != nil {
					h.log.Error("ws process location failed", "module", "ws_handler", "error", err)
				}
			}()

			h.hub.UpdateLastPoint(tripID, tripPoint)

			h.hub.HandlePosition(payload, payload.Speed)
		}
	}
}

func (h *WSHandler) parseLocationPayload(raw map[string]interface{}) (wsInfra.PositionUpdate, bool) {
	lat, _ := raw["latitude"].(float64)
	lon, _ := raw["longitude"].(float64)
	heading, _ := raw["heading"].(float64)
	accuracy, _ := raw["accuracy"].(float64)
	altitude, _ := raw["altitude"].(float64)
	speed, _ := raw["speed"].(float64)
	battery, _ := raw["battery"].(float64)
	timestamp, _ := raw["timestamp"].(string)

	return wsInfra.PositionUpdate{
		Latitude:  lat,
		Longitude: lon,
		Speed:     speed,
		Heading:   heading,
		Accuracy:  accuracy,
		Altitude:  altitude,
		Battery:   int(battery),
		Timestamp: timestamp,
	}, true
}

func (h *WSHandler) validateToken(tokenString string) (uuid.UUID, error) {
	claims := &middleware.AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, fiber.ErrUnauthorized
	}

	if h.jwtIssuer != "" && claims.Issuer != h.jwtIssuer {
		return uuid.Nil, fiber.ErrUnauthorized
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fiber.ErrUnauthorized
	}

	return userID, nil
}


