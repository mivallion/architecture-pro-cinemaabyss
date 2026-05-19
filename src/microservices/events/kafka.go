package main

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/segmentio/kafka-go"
)

// ---------- Producer ----------

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers string) *KafkaProducer {
	brokerList := strings.Split(brokers, ",")
	for i := range brokerList {
		brokerList[i] = strings.TrimSpace(brokerList[i])
	}

	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerList...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaProducer) ProduceEvent(ctx context.Context, topic string, key string, value []byte) (partition int, offset int64, err error) {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return 0, 0, err
	}

	// kafka-go does not return partition/offset per-message from WriteMessages in the
	// high-level writer.  We return 0,0 to satisfy the signature; the response will
	// show zeroes.  This is a known limitation of the batch writer API.
	return 0, 0, nil
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// ---------- Consumer ----------

type KafkaConsumer struct {
	brokers []string
	readers []*kafka.Reader
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewKafkaConsumer(brokers string) *KafkaConsumer {
	brokerList := strings.Split(brokers, ",")
	for i := range brokerList {
		brokerList[i] = strings.TrimSpace(brokerList[i])
	}
	return &KafkaConsumer{
		brokers: brokerList,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	topics := []string{"movie-events", "user-events", "payment-events"}
	for _, topic := range topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers: c.brokers,
			Topic:   topic,
			GroupID: "events-service-group",
		})
		c.readers = append(c.readers, reader)

		c.wg.Add(1)
		go func(r *kafka.Reader, t string) {
			defer c.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				m, err := r.ReadMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					log.Printf("[consumer] error reading from topic=%s: %v", t, err)
					continue
				}

				log.Printf("[consumer] topic=%s partition=%d offset=%d key=%s value=%s",
					m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
			}
		}(reader, topic)
	}

	c.wg.Wait()
}

func (c *KafkaConsumer) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	for _, r := range c.readers {
		_ = r.Close()
	}
	c.wg.Wait()
}
