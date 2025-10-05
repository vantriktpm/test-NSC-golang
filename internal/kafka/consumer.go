package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"url-shortener/internal/models"

	"github.com/Shopify/sarama"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	client   sarama.Client
	consumer sarama.ConsumerGroup
	config   *ConsumerConfig
	handlers map[string]EventHandler
	mu       sync.RWMutex
}

// ConsumerConfig holds configuration for Kafka consumer
type ConsumerConfig struct {
	Brokers           []string
	GroupID           string
	TopicPrefix       string
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
	RetryAttempts     int
	RetryDelay        time.Duration
}

// EventHandler defines the interface for handling different event types
type EventHandler interface {
	HandleEvent(ctx context.Context, event *models.InventoryEvent) error
	GetEventType() models.InventoryEventType
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(config *ConsumerConfig) (*Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Session.Timeout = config.SessionTimeout
	saramaConfig.Consumer.Group.Heartbeat.Interval = config.HeartbeatInterval
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	client, err := sarama.NewClient(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	consumer, err := sarama.NewConsumerGroupFromClient(config.GroupID, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &Consumer{
		client:   client,
		consumer: consumer,
		config:   config,
		handlers: make(map[string]EventHandler),
	}, nil
}

// RegisterHandler registers an event handler
func (c *Consumer) RegisterHandler(handler EventHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[string(handler.GetEventType())] = handler
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context, topics []string) error {
	// Add topic prefix to topics
	prefixedTopics := make([]string, len(topics))
	for i, topic := range topics {
		prefixedTopics[i] = c.config.TopicPrefix + topic
	}

	// Start consuming in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Consumer context cancelled, stopping...")
				return
			default:
				err := c.consumer.Consume(ctx, prefixedTopics, c)
				if err != nil {
					log.Printf("Error consuming messages: %v", err)
					time.Sleep(c.config.RetryDelay)
				}
			}
		}
	}()

	// Handle errors
	go func() {
		for err := range c.consumer.Errors() {
			log.Printf("Consumer error: %v", err)
		}
	}()

	log.Printf("Consumer started for topics: %v", prefixedTopics)
	return nil
}

// Stop stops the consumer
func (c *Consumer) Stop() error {
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close client: %w", err)
	}
	return nil
}

// Sarama ConsumerGroupHandler implementation

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session setup")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Consumer group session cleanup")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Process message
			if err := c.processMessage(session.Context(), message); err != nil {
				log.Printf("Error processing message: %v", err)
				// Continue processing other messages
			}

			// Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// processMessage processes a single Kafka message
func (c *Consumer) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	// Parse event type from headers
	var eventType models.InventoryEventType
	for _, header := range message.Headers {
		if string(header.Key) == "eventType" {
			eventType = models.InventoryEventType(header.Value)
			break
		}
	}

	// Deserialize message
	var event models.InventoryEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal inventory event: %w", err)
	}

	// Get handler for event type
	c.mu.RLock()
	handler, exists := c.handlers[string(eventType)]
	c.mu.RUnlock()

	if !exists {
		log.Printf("No handler found for event type: %s", eventType)
		return nil
	}

	// Handle event
	if err := handler.HandleEvent(ctx, &event); err != nil {
		return fmt.Errorf("failed to handle event %s: %w", eventType, err)
	}

	log.Printf("Successfully processed event %s for product %s", eventType, event.ProductID)
	return nil
}

// InventoryEventHandler handles inventory events
type InventoryEventHandler struct {
	eventType models.InventoryEventType
	handler   func(ctx context.Context, event *models.InventoryEvent) error
}

// NewInventoryEventHandler creates a new inventory event handler
func NewInventoryEventHandler(eventType models.InventoryEventType, handler func(ctx context.Context, event *models.InventoryEvent) error) *InventoryEventHandler {
	return &InventoryEventHandler{
		eventType: eventType,
		handler:   handler,
	}
}

// HandleEvent implements EventHandler interface
func (h *InventoryEventHandler) HandleEvent(ctx context.Context, event *models.InventoryEvent) error {
	return h.handler(ctx, event)
}

// GetEventType implements EventHandler interface
func (h *InventoryEventHandler) GetEventType() models.InventoryEventType {
	return h.eventType
}
