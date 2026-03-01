package ws

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn}
}

func (c *Client) Send(ctx context.Context, payload []byte) error {
	deadline := time.Now().Add(5 * time.Second)
	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

func (c *Client) ConsumeLoop() {
	defer c.Close()

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *Client) Close() {
	_ = c.conn.Close()
}
