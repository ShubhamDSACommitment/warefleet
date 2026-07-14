// Package mqtt bridges the fleet manager to the robots over MQTT.
//
// Topics:
//
//	warefleet/orders                 (in)  new tasks from the order feeder
//	warefleet/robots/+/state         (in)  robot heartbeats (RobotState)
//	warefleet/robots/<id>/assignment (out) TaskAssignment
//	warefleet/robots/<id>/plan       (out) PathPlan
//
// QoS: orders/assignments/plans are QoS 1 (must arrive), heartbeats QoS 0
// (next one is a second away). JSON field names are defined by model/types.go
// and mirrored by the ROS bridge (warefleet_agent/mqtt_bridge.py).
package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/yourname/warefleet/fleet_manager/internal/model"
)

const (
	topicOrders     = "warefleet/orders"
	topicStates     = "warefleet/robots/+/state"
	topicAssignment = "warefleet/robots/%s/assignment"
	topicPlan       = "warefleet/robots/%s/plan"

	opTimeout = 10 * time.Second
)

// Handlers are the callbacks the coordinator registers for inbound messages.
type Handlers struct {
	OnOrder      func(model.Task)
	OnRobotState func(model.Robot)
}

// Client wraps the paho MQTT client with WareFleet-specific pub/sub helpers.
type Client struct {
	broker string
	inner  paho.Client
}

func New(broker string) *Client {
	return &Client{broker: broker}
}

// Connect dials the broker and subscribes to the inbound topics.
func (c *Client) Connect(h Handlers) error {
	opts := paho.NewClientOptions().
		AddBroker(c.broker).
		SetClientID("warefleet-fm").
		SetAutoReconnect(true).
		SetOrderMatters(false)
	opts.OnConnectionLost = func(_ paho.Client, err error) {
		log.Printf("mqtt: connection lost: %v (auto-reconnecting)", err)
	}
	// Subscriptions live in OnConnect so they survive broker restarts.
	opts.OnConnect = func(cl paho.Client) {
		cl.Subscribe(topicOrders, 1, func(_ paho.Client, m paho.Message) {
			var t model.Task
			if err := json.Unmarshal(m.Payload(), &t); err != nil {
				log.Printf("mqtt: bad order JSON: %v", err)
				return
			}
			if t.CreatedAt.IsZero() {
				t.CreatedAt = time.Now() // feeder may omit it; greedy sorts by age
			}
			h.OnOrder(t)
		})
		cl.Subscribe(topicStates, 0, func(_ paho.Client, m paho.Message) {
			var r model.Robot
			if err := json.Unmarshal(m.Payload(), &r); err != nil {
				log.Printf("mqtt: bad robot state JSON on %s: %v", m.Topic(), err)
				return
			}
			h.OnRobotState(r)
		})
		log.Printf("mqtt: connected to %s, subscribed to orders + robot states", c.broker)
	}

	c.inner = paho.NewClient(opts)
	tok := c.inner.Connect()
	if !tok.WaitTimeout(opTimeout) {
		return fmt.Errorf("mqtt: connect to %s: timeout after %s", c.broker, opTimeout)
	}
	return tok.Error()
}

// PublishAssignment implements coordination.Publisher.
func (c *Client) PublishAssignment(a model.Assignment) error {
	return c.publishJSON(fmt.Sprintf(topicAssignment, a.RobotID), a)
}

// PublishPathPlan implements coordination.Publisher.
func (c *Client) PublishPathPlan(p model.PathPlan) error {
	return c.publishJSON(fmt.Sprintf(topicPlan, p.RobotID), p)
}

func (c *Client) publishJSON(topic string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("mqtt: marshal for %s: %w", topic, err)
	}
	tok := c.inner.Publish(topic, 1, false, b)
	if !tok.WaitTimeout(opTimeout) {
		return fmt.Errorf("mqtt: publish %s: timeout after %s", topic, opTimeout)
	}
	return tok.Error()
}
