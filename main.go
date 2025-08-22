package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Настройки Kafka
const (
	broker  = "localhost:9092"    // Адрес брокера Kafka
	topic   = "my-topic"        // Тема Kafka
	groupID = "my-consumer-group" // ID группы консьюмеров
)

// Producer отправляет сообщения в Kafka
func produceMessages(ctx context.Context) {
	// Конфигурация писателя
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topic,
		Balancer:               &kafka.Hash{}, // Хэш-балансировщик для предсказуемого партиционирования
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
		WriteTimeout:           10 * time.Second,
		BatchTimeout:           100 * time.Millisecond,
		MaxAttempts:            3,
	}

	defer writer.Close()

	// Отправка нескольких сообщений
	for i := 0; i < 5; i++ {
		err := writer.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(fmt.Sprintf("key-%d", i)),
				Value: []byte(fmt.Sprintf("Сообщение %d для Kafka", i)),
			},
		)
		if err != nil {
			log.Printf("Ошибка отправки сообщения %d: %v", i, err)
			continue
		}
		fmt.Printf("Сообщение %d отправлено\n", i)
	}
}

// Consumer читает сообщения из Kafka
func consumeMessages(ctx context.Context) {
	// Конфигурация читателя с группой консьюмеров
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{broker},
		GroupID:        groupID, // Указываем группу консьюмеров
		Topic:          topic,
		MinBytes:       1,               // Минимальный размер батча
		MaxBytes:       1e6,             // Максимальный размер батча (1MB)
		MaxWait:        5 * time.Second, // Максимальное время ожидания
		CommitInterval: 1 * time.Second, // Автоматический коммит каждую секунду
	})

	defer reader.Close()

	fmt.Println("Консьюмер запущен. Ожидание сообщений...")
	for {
		// Чтение сообщения
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Ошибка чтения сообщения: %v", err)
			if err == context.DeadlineExceeded {
				log.Println("Тайм-аут чтения, возможно, топик пуст. Продолжаем...")
				continue
			}
			return
		}
		fmt.Printf("Получено сообщение: key=%s, value=%s, offset=%d, partition=%d\n",
			string(msg.Key), string(msg.Value), msg.Offset, msg.Partition)

		// Явный коммит сообщения (дополнительно к автоматическому)
		err = reader.CommitMessages(ctx, msg)
		if err != nil {
			log.Printf("Ошибка коммита offset: %v", err)
		} else {
			fmt.Printf("Offset %d для сообщения успешно закоммичен\n", msg.Offset)
		}
	}
}

func main() {
	ctx := context.Background()

	// Запускаем продюсера
	go produceMessages(ctx)

	// Даем продюсеру время отправить сообщения
	time.Sleep(3 * time.Second)

	// Запускаем консьюмера
	consumeMessages(ctx)
}
