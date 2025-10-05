package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"url-shortener/internal/models"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
)

// Producer handles Kafka message production
type Producer struct {
	producer sarama.SyncProducer
	config   *ProducerConfig
}

// ProducerConfig holds configuration for Kafka producer
type ProducerConfig struct {
	Brokers       []string
	TopicPrefix   string
	RetryAttempts int
	RetryDelay    time.Duration
}

// NewProducer creates a new Kafka producer
func NewProducer(config *ProducerConfig) (*Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = config.RetryAttempts
	saramaConfig.Producer.Retry.Backoff = config.RetryDelay
	saramaConfig.Producer.Compression = sarama.SnappyCompression

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		config:   config,
	}, nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.producer.Close()
}

// PublishInventoryEvent publishes an inventory event to Kafka
func (p *Producer) PublishInventoryEvent(ctx context.Context, event *models.InventoryEvent) error {
	// Generate event ID if not provided
	if event.EventID == uuid.Nil {
		event.EventID = uuid.New()
	}

	// Generate correlation ID if not provided
	if event.CorrelationID == uuid.Nil {
		event.CorrelationID = uuid.New()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize event to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory event: %w", err)
	}

	// Create Kafka message
	message := &sarama.ProducerMessage{
		Topic: p.config.TopicPrefix + "inventory-events",
		Key:   sarama.StringEncoder(event.ProductID), // Partition by product ID
		Value: sarama.ByteEncoder(eventBytes),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("eventType"),
				Value: []byte(event.EventType),
			},
			{
				Key:   []byte("correlationId"),
				Value: []byte(event.CorrelationID.String()),
			},
		},
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send inventory event: %w", err)
	}

	log.Printf("Published inventory event %s to partition %d, offset %d",
		event.EventID, partition, offset)

	return nil
}

// PublishInventoryState publishes inventory state to Kafka
func (p *Producer) PublishInventoryState(ctx context.Context, state *models.InventoryState) error {
	// Serialize state to JSON
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory state: %w", err)
	}

	// Create Kafka message
	message := &sarama.ProducerMessage{
		Topic: p.config.TopicPrefix + "inventory-state",
		Key:   sarama.StringEncoder(state.ProductID), // Partition by product ID
		Value: sarama.ByteEncoder(stateBytes),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("version"),
				Value: []byte(fmt.Sprintf("%d", state.Version)),
			},
		},
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send inventory state: %w", err)
	}

	log.Printf("Published inventory state for product %s to partition %d, offset %d",
		state.ProductID, partition, offset)

	return nil
}

// PublishInventoryEventAsync publishes an inventory event asynchronously
func (p *Producer) PublishInventoryEventAsync(ctx context.Context, event *models.InventoryEvent) error {
	go func() {
		if err := p.PublishInventoryEvent(ctx, event); err != nil {
			log.Printf("Failed to publish inventory event asynchronously: %v", err)
		}
	}()
	return nil
}

// Helper methods for creating common events

// CreateInventoryCheckEvent creates an inventory check event
func (p *Producer) CreateInventoryCheckEvent(productID string, quantity int, userID string, metadata map[string]interface{}) *models.InventoryEvent {
	return &models.InventoryEvent{
		EventID:       uuid.New(),
		EventType:     models.InventoryEventTypeCheck,
		ProductID:     productID,
		UserID:        userID,
		Quantity:      quantity,
		Timestamp:     time.Now(),
		CorrelationID: uuid.New(),
		Metadata:      metadata,
	}
}

// CreateInventoryReserveEvent creates an inventory reserve event
func (p *Producer) CreateInventoryReserveEvent(productID string, quantity int, userID string, metadata map[string]interface{}) *models.InventoryEvent {
	return &models.InventoryEvent{
		EventID:       uuid.New(),
		EventType:     models.InventoryEventTypeReserve,
		ProductID:     productID,
		UserID:        userID,
		Quantity:      quantity,
		Timestamp:     time.Now(),
		CorrelationID: uuid.New(),
		Metadata:      metadata,
	}
}

// CreateInventoryConfirmEvent creates an inventory confirm event
func (p *Producer) CreateInventoryConfirmEvent(productID string, quantity int, userID string, correlationID uuid.UUID, metadata map[string]interface{}) *models.InventoryEvent {
	return &models.InventoryEvent{
		EventID:       uuid.New(),
		EventType:     models.InventoryEventTypeConfirm,
		ProductID:     productID,
		UserID:        userID,
		Quantity:      quantity,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Metadata:      metadata,
	}
}

// CreateInventoryReleaseEvent creates an inventory release event
func (p *Producer) CreateInventoryReleaseEvent(productID string, quantity int, userID string, correlationID uuid.UUID, metadata map[string]interface{}) *models.InventoryEvent {
	return &models.InventoryEvent{
		EventID:       uuid.New(),
		EventType:     models.InventoryEventTypeRelease,
		ProductID:     productID,
		UserID:        userID,
		Quantity:      quantity,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Metadata:      metadata,
	}
}
