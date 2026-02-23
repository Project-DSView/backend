package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Project-DSView/backend/go/pkg/logger"
	"github.com/streadway/amqp"
)

type RabbitMQConfig struct {
	URL      string
	Exchange string
}

type RabbitMQService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *RabbitMQConfig
}

type QueueMessage struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// Queue types
const (
	QueueTypeCodeExecution  = "code_execution"
	QueueTypeReview         = "review"
	QueueTypeFileProcessing = "file_processing"
)

// connectWithRetry attempts to connect to RabbitMQ with retry logic
func connectWithRetry(url string, maxRetries int, retryDelay time.Duration) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < maxRetries; i++ {
		logger.Infof("Attempting to connect to RabbitMQ (attempt %d/%d)...", i+1, maxRetries)

		conn, err = amqp.Dial(url)
		if err == nil {
			logger.Infof("Successfully connected to RabbitMQ on attempt %d", i+1)
			return conn, nil
		}

		logger.Warnf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			logger.Infof("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

func NewRabbitMQService(config *RabbitMQConfig) (*RabbitMQService, error) {
	conn, err := connectWithRetry(config.URL, 30, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		config.Exchange, // name
		"direct",        // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Note: Queues are now created dynamically per course
	// No global queues are created at startup

	return &RabbitMQService{
		conn:    conn,
		channel: channel,
		config:  config,
	}, nil
}

// GetQueueName generates a course-specific queue name
func GetQueueName(queueType, courseID string) string {
	return fmt.Sprintf("%s_course_%s", queueType, courseID)
}

// EnsureQueueExists ensures that a queue exists, creating it if necessary
func (r *RabbitMQService) EnsureQueueExists(queueName string) error {
	// Try to declare queue (idempotent operation)
	_, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	// Bind queue to exchange with routing key = queue name
	err = r.channel.QueueBind(
		queueName,         // queue name
		queueName,         // routing key
		r.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue %s: %w", queueName, err)
	}

	return nil
}

// PublishMessage publishes a message to the specified course-specific queue
func (r *RabbitMQService) PublishMessage(ctx context.Context, queueType, courseID string, message *QueueMessage) error {
	// Generate course-specific queue name
	queueName := GetQueueName(queueType, courseID)

	// Ensure queue exists
	if err := r.EnsureQueueExists(queueName); err != nil {
		return fmt.Errorf("failed to ensure queue exists: %w", err)
	}

	message.CreatedAt = time.Now()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.Publish(
		r.config.Exchange, // exchange
		queueName,         // routing key (use queue name as routing key)
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   message.CreatedAt,
			MessageId:   message.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Infof("Message published to queue %s: %s", queueName, message.ID)
	return nil
}

// ConsumeMessages starts consuming messages from the specified course-specific queue
func (r *RabbitMQService) ConsumeMessages(ctx context.Context, queueType, courseID string, handler func(*QueueMessage) error) error {
	// Generate course-specific queue name
	queueName := GetQueueName(queueType, courseID)

	// Ensure queue exists before consuming
	if err := r.EnsureQueueExists(queueName); err != nil {
		return fmt.Errorf("failed to ensure queue exists: %w", err)
	}

	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				var queueMessage QueueMessage
				if err := json.Unmarshal(msg.Body, &queueMessage); err != nil {
					logger.Errorf("Failed to unmarshal message: %v", err)
					msg.Nack(false, false)
					continue
				}

				if err := handler(&queueMessage); err != nil {
					logger.Errorf("Failed to process message %s: %v", queueMessage.ID, err)
					msg.Nack(false, true) // Requeue on error
				} else {
					msg.Ack(false)
					logger.Infof("Message processed successfully: %s", queueMessage.ID)
				}
			}
		}
	}()

	return nil
}

// GetQueueInfo returns information about the specified course-specific queue
func (r *RabbitMQService) GetQueueInfo(queueType, courseID string) (*amqp.Queue, error) {
	queueName := GetQueueName(queueType, courseID)
	queue, err := r.channel.QueueInspect(queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect queue %s: %w", queueName, err)
	}
	return &queue, nil
}

// PurgeQueue removes all messages from the specified course-specific queue
func (r *RabbitMQService) PurgeQueue(queueType, courseID string) (int, error) {
	queueName := GetQueueName(queueType, courseID)
	count, err := r.channel.QueuePurge(queueName, false)
	if err != nil {
		return 0, fmt.Errorf("failed to purge queue %s: %w", queueName, err)
	}
	return count, nil
}

// DeleteQueue deletes a course-specific queue
func (r *RabbitMQService) DeleteQueue(queueType, courseID string) error {
	queueName := GetQueueName(queueType, courseID)
	_, err := r.channel.QueueDelete(queueName, false, false, false)
	if err != nil {
		return fmt.Errorf("failed to delete queue %s: %w", queueName, err)
	}
	logger.Infof("Queue deleted: %s", queueName)
	return nil
}

// DeleteCourseQueues deletes all queues for a specific course
func (r *RabbitMQService) DeleteCourseQueues(courseID string) error {
	queueTypes := []string{QueueTypeCodeExecution, QueueTypeReview, QueueTypeFileProcessing}
	for _, queueType := range queueTypes {
		if err := r.DeleteQueue(queueType, courseID); err != nil {
			// Log error but continue with other queues
			logger.Warnf("Failed to delete queue %s for course %s: %v", queueType, courseID, err)
		}
	}
	return nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQService) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			logger.Warnf("Failed to close channel: %v", err)
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			logger.Warnf("Failed to close connection: %v", err)
		}
	}
	return nil
}

// Health check
func (r *RabbitMQService) HealthCheck(ctx context.Context) error {
	if r.conn == nil || r.channel == nil {
		return fmt.Errorf("RabbitMQ connection is not initialized")
	}

	// Try to declare a test queue to check connection
	testQueue := "health_check_test"
	_, err := r.channel.QueueDeclare(
		testQueue,
		false, // not durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Clean up test queue
	_, err = r.channel.QueueDelete(testQueue, false, false, false)
	if err != nil {
		logger.Warnf("Failed to delete test queue: %v", err)
	}

	return nil
}
