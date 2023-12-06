package subscriber

import (
	"WB0/config"
	"WB0/internal/cache"
	"WB0/internal/interfaces/database"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/stan.go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Subscriber() {
	sc, err := stan.Connect("test-cluster", "subscriber-client")
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()
	// Тема, на которую будем подписываться
	subject := "model"

	// Имя очереди, используемое для создания долговечной подписки
	queue := "queue"
	cacheMem := cache.New(0, 0)
	dbParams := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", config.User, config.Password, config.DBname)
	pool, err := pgxpool.Connect(context.Background(), dbParams)
	if err != nil {
		log.Fatal(err)
	}
	orders, err := database.GetAllDataFromDB(pool)
	cacheMem.RestoreFromBD(orders)
	cacheMem.PrintCache()
	pool.Close()
	go func() {
		http.HandleFunc("/", cache.IndexHandler)
		http.HandleFunc("/order", cacheMem.HandleGetByID)
		fmt.Println("Server is running on :8080")
		http.ListenAndServe(":8080", nil)
	}()
	// Функция обработки сообщений
	messageHandler := func(msg *stan.Msg) {
		log.Printf("Data is received")
		database.ConnectDBandInsert(msg.Data, cacheMem)
	}

	// Создание долговечной подписки
	_, err = sc.QueueSubscribe(subject, queue, messageHandler, stan.DurableName(queue))
	if err != nil {
		log.Fatal(err)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	<-signalChannel
}
