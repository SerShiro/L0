package database

import (
	"WB0/internal/cache"
	"WB0/internal/nats_streaming/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
)

func insertOrderData(conn *pgx.Conn, orderJSON []byte, cacheMem *cache.Cache) error {
	var order model.Order

	if err := json.Unmarshal(orderJSON, &order); err != nil {
		return err
	}

	tx, err := conn.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	var deliveryID int
	// Вставка данных в таблицу delivery
	err = tx.QueryRow(context.Background(),
		"INSERT INTO delivery (name, phone, zip, city, address, region, email) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email).Scan(&deliveryID)
	if err != nil {
		log.Println("Error inserting data into delivery_info:", err)
		return err
	}
	var orderID int
	// Вставка данных в таблицу order_info
	err = tx.QueryRow(context.Background(),
		"INSERT INTO order_info (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, delivery_id) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id",
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard, deliveryID).Scan(&orderID)
	if err != nil {
		log.Println("Error inserting data into order_info:", err)
		return err
	}

	// Вставка данных в таблицу payment
	_, err = tx.Exec(context.Background(),
		"INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee, order_id) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee, orderID)
	if err != nil {
		log.Println("Error inserting data into payment_info:", err)
		return err
	}
	// Вставка данных в таблицу items
	for _, item := range order.Items {
		_, err = tx.Exec(context.Background(),
			"INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, order_id) "+
				"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
			item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size,
			item.TotalPrice, item.NmID, item.Brand, item.Status, orderID)
		if err != nil {
			log.Println("Error inserting data into items_info:", err)
			return err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	fmt.Println("Data inserted successfully into tables")
	return nil
}
