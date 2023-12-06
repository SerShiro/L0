package publisher

import (
	"WB0/internal/nats_streaming/model"
	"encoding/json"
	"github.com/nats-io/stan.go"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func Publisher() {
	sc, err := stan.Connect("test-cluster", "publisher-client")
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()
	// Тема, в которую будем публиковать сообщения
	subject := "model"
	order, err := readJSON("data.json")
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(order)
	if err != nil {
		log.Fatal(err)
	}
	err = sc.Publish(subject, jsonData)
	if err != nil {
		log.Printf("Error publishing message: %v\n", err)
	} else {
		log.Printf("Data is published by publisher 1")
	}
}

// PublishNewData отправляем еще одно сообщение для теста
func PublishNewData() {
	sc, err := stan.Connect("test-cluster", "publisher-client-new")
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	subject := "model"
	order, err := readJSON("data2.json")
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(order)
	if err != nil {
		log.Fatal(err)
	}

	err = sc.Publish(subject, jsonData)
	if err != nil {
		log.Printf("Error publishing message: %v\n", err)
	} else {
		log.Printf("Data is published by publisher 2")

	}
}

func readJSON(filename string) (*model.Order, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(currentDir, filename)

	fileData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	var order model.Order

	err = json.Unmarshal(fileData, &order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}
