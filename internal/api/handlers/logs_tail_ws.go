package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/platformbuilds/mirador-core/internal/models"
)

// GET /api/v1/logs/tail (upgrades to WS)
// query params: query, since, tenantId, sampling
func (h *LogsHandler) TailWS(c *gin.Context) {
	var (
		query    = c.Query("query")
		tenantID = c.Query("tenantId")
		since    = parseInt64Default(c.Query("since"), time.Now().Add(-5*time.Minute).UnixMilli())
		sampling = parseIntDefault(c.Query("sampling"), 1)
	)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  8 << 10,
		WriteBufferSize: 64 << 10,
		CheckOrigin:     func(*http.Request) bool { return true }, // TODO: tighten CORS in prod
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("ws upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	type msg struct {
		Type string      `json:"type"` // row|stats|heartbeat|error
		Data interface{} `json:"data"`
	}

	// bounded channel for backpressure
	const bufSize = 2048
	rowsCh := make(chan map[string]any, bufSize)
	var dropped int64
	var wg sync.WaitGroup

	// writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Second) // heartbeat
		defer ticker.Stop()

		for {
			select {
			case row, ok := <-rowsCh:
				if !ok {
					return
				}
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteJSON(msg{Type: "row", Data: row}); err != nil {
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				_ = conn.WriteJSON(msg{Type: "heartbeat", Data: map[string]any{"ts": time.Now().UnixMilli()}})
			}
		}
	}()

	// reader (no-op: just to detect close)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				close(rowsCh)
				return
			}
		}
	}()

	// stream query
	sampleN := sampling
	if sampleN <= 1 {
		sampleN = 1
	}
	rowIdx := int64(0)

	_, qerr := h.logs.ExecuteQueryStream(c, &models.LogsQLQueryRequest{
		Query:    query,
		Start:    since,
		End:      time.Now().UnixMilli(),
		TenantID: tenantID,
	}, func(row map[string]any) error {
		n := atomic.AddInt64(&rowIdx, 1)
		if sampleN > 1 && (n%int64(sampleN)) != 0 {
			return nil
		}
		select {
		case rowsCh <- row:
		default:
			atomic.AddInt64(&dropped, 1) // backpressure: drop latest
		}
		return nil
	})
	close(rowsCh)
	wg.Wait()

	// send final stats
	_ = conn.WriteJSON(map[string]any{
		"type": "stats",
		"data": map[string]any{
			"dropped": dropped,
			"sent":    rowIdx - dropped,
			"sampleN": sampleN,
		},
	})

	_ = qerr // we could also send an "error" frame if qerr != nil
}

func parseInt64Default(s string, def int64) int64 {
	if s == "" {
		return def
	}
	var n int64
	_, err := fmt.Sscan(s, &n)
	if err != nil {
		return def
	}
	return n
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	var n int
	_, err := fmt.Sscan(s, &n)
	if err != nil {
		return def
	}
	return n
}
