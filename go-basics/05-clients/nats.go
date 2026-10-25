package clients

import (
	"errors"
	"time"

	"github.com/nats-io/nats.go"
)

// ═════════════════════════════════════════════════════════════
// PART 4: MESSAGING WITH NATS
// Services talk to each other through messages instead of direct calls:
// the sender doesn't need to know who listens. NATS is simple and fast; the
// same ideas (publish/subscribe, request/reply, queue groups) apply to
// Kafka, RabbitMQ and Google Pub/Sub.
// Docs: https://docs.nats.io/using-nats/developer
// Monitoring UI: http://localhost:58222
// ═════════════════════════════════════════════════════════════

var ErrUnknownProduct = errors.New("unknown product")

type Order struct {
	ID        string `json:"id"`
	AmountYen int    `json:"amount_yen"`
}

// ─── 13. Connecting ───────────────────────────────────────────
// NewNATS connects to url ("nats://localhost:54222") with a client name
// (nats.Name("go-basics")), so it's identifiable in the monitoring UI.
func NewNATS(url string) (*nats.Conn, error) {
	return nil, ErrTODO
}

// ─── 14. Publish / subscribe ──────────────────────────────────
// PublishOrder publishes the order as JSON on subject "orders.created".
func PublishOrder(nc *nats.Conn, o Order) error {
	return ErrTODO
}

// SubscribeOrders calls handle for every order published on
// "orders.created" (decode the JSON; skip messages that aren't valid).
// The caller unsubscribes with the returned subscription.
func SubscribeOrders(nc *nats.Conn, handle func(Order)) (*nats.Subscription, error) {
	return nil, ErrTODO
}

// ─── 15. Request / reply ──────────────────────────────────────
// Like an HTTP call, but over NATS: any service subscribed to the subject
// can answer.

// ServePrices answers requests on "prices.get". The request data is a
// product name; reply with its price as text ("1200"), or "unknown".
// Hint: msg.Respond([]byte(...))
func ServePrices(nc *nats.Conn, prices map[string]int) (*nats.Subscription, error) {
	return nil, ErrTODO
}

// GetPrice asks "prices.get" for a product's price, waiting up to timeout.
// "unknown" → ErrUnknownProduct. No responder → return the NATS error.
func GetPrice(nc *nats.Conn, product string, timeout time.Duration) (int, error) {
	return 0, ErrTODO
}

// ─── 16. Queue groups (load balancing) ────────────────────────
// StartWorkers starts n subscribers on subject "jobs" in the queue group
// "workers". NATS delivers each message to only ONE member of a queue group,
// so the work is shared, like a worker pool across machines.
// handle receives the worker number (0..n-1) and the message text.
func StartWorkers(nc *nats.Conn, n int, handle func(worker int, job string)) ([]*nats.Subscription, error) {
	return nil, ErrTODO
}
