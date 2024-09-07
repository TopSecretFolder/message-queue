package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	redisimpl "github.com/TopSecretFolder/message-queue/internal/redis_impl"
	"github.com/TopSecretFolder/message-queue/shared/queue"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/ws/:source_id", handleWebsocket)
	e.GET("/", about)
	e.Logger.Fatal(e.Start(":8080"))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var queueProvider = redisimpl.NewClient("redis:6379")

func about(c echo.Context) error {
	c.String(200, "message-queue server")
	return nil
}

func handleWebsocket(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	sourceID := c.Param("source_id")

	// when we get a message, queue it up in the queue provider
	key := queue.Key(fmt.Sprintf("%s:%s", "queue", sourceID))
	q := queue.Instance[Message](key, queueProvider)

	go queueReadLoop(c.Request().Context(), conn, q)

	socketReadLoop(c.Request().Context(), conn, OnSocketMessage)

	return nil
}

type Message struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Data     string `json:"Data"`
}

func socketReadLoop(
	ctx context.Context,
	conn *websocket.Conn,
	onMessage func(m Message),
) {
	count := 0
	for {
		count = throttleLog(count, "reading socket")
		select {
		case <-ctx.Done():
			return
		default:

			var message Message
			err := conn.ReadJSON(&message)
			if err != nil {
				log.Println("websocket read error:", err)
				break
			}

			onMessage(message)

		}
	}
}

func OnSocketMessage(m Message) {

	q := queue.Instance[Message](queue.GetKey(m.TargetID), queueProvider)

	err := q.Enqueue(m)
	if err != nil {
		log.Println(err)
	}
}

func queueReadLoop(ctx context.Context, conn *websocket.Conn, q queue.Queue[Message]) {
	count := 0
	for {
		count = throttleLog(count, "reading queue")
		select {
		case <-ctx.Done():
			return
		default:
			m, err := q.Dequeue()
			if q.IsEmptyErr(err) {
				continue
			}
			if err != nil {
				log.Println("could not dequeue from the queue provider:", err)
				break
			}

			err = conn.WriteJSON(m)
			if err != nil {
				log.Println("could not write to socket:", err)
				return
			}
		}
	}
}

func throttleLog(count int, loggables ...interface{}) int {
	count += 1

	if count%100000 == 0 {
		fmt.Println(loggables...)
		count = 0
	}

	return count
}
