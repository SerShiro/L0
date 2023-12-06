package database

import (
	"WB0/config"
	"WB0/internal/cache"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
)

func ConnectDBandInsert(dataForDB []byte, cacheMem *cache.Cache) {

	dbParams := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", config.User, config.Password, config.DBname)
	pool, err := pgx.Connect(context.Background(), dbParams)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close(context.Background())
	insertOrderData(pool, dataForDB, cacheMem)
	if err != nil {
		log.Fatal(err)
	}

}
