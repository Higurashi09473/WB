package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)


func test() {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	// to produce messages
	go func(){
		defer  wg.Done()

		topic := "my-topic"
		partition := 0
	
		conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
		if err != nil {
			log.Fatal("failed to dial leader:", err)
		}
	
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte("one!")},
			kafka.Message{Value: []byte("two!")},
			kafka.Message{Value: []byte("three!")},
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
	
		if err := conn.Close(); err != nil {
			log.Fatal("failed to close writer:", err)
		}
		
	}()

	go func(){
		defer  wg.Done()

		// to consume messages
		topic := "my-topic"
		partition := 0
	
		conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
		if err != nil {
			log.Fatal("failed to dial leader:", err)
		}
	
		conn.SetReadDeadline(time.Now().Add(10*time.Second))
		batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max
	
		b := make([]byte, 10e3) // 10KB max per message
		for {
			n, err := batch.Read(b)
			if err != nil {
				break
			}
			fmt.Println(string(b[:n]))
		}
	
		if err := batch.Close(); err != nil {
			log.Fatal("failed to close batch:", err)
		}
	
		if err := conn.Close(); err != nil {
			log.Fatal("failed to close connection:", err)
		}
	}()
	wg.Wait()
}