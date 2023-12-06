package database

import (
	"WB0/internal/nats_streaming/model"
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

type tempItems struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func GetAllDataFromDB(conn *pgxpool.Pool) ([]model.Order, error) {
	query := `
SELECT
    d.name, d.phone, d.zip, d.city, d.address, d.region, d.email, 
    o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
    o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,  
    p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
    p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
    i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size, 
    i.total_price, i.nm_id, i.brand, i.status
FROM
    delivery d 
		JOIN
    order_info o ON d.id = o.delivery_id
		JOIN 
    payment p ON o.id = p.order_id
		JOIN 
    items i ON o.id = i.order_id
	`

	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order

	for rows.Next() {
		var temp tempItems
		var orderInfo model.Order
		err := rows.Scan(
			&orderInfo.Delivery.Name, &orderInfo.Delivery.Phone, &orderInfo.Delivery.Zip,
			&orderInfo.Delivery.City, &orderInfo.Delivery.Address, &orderInfo.Delivery.Region,
			&orderInfo.Delivery.Email,
			&orderInfo.OrderUID, &orderInfo.TrackNumber, &orderInfo.Entry, &orderInfo.Locale,
			&orderInfo.InternalSignature, &orderInfo.CustomerID, &orderInfo.DeliveryService,
			&orderInfo.Shardkey, &orderInfo.SmID, &orderInfo.DateCreated, &orderInfo.OofShard,
			&orderInfo.Payment.Transaction, &orderInfo.Payment.RequestID, &orderInfo.Payment.Currency,
			&orderInfo.Payment.Provider, &orderInfo.Payment.Amount, &orderInfo.Payment.PaymentDt,
			&orderInfo.Payment.Bank, &orderInfo.Payment.DeliveryCost, &orderInfo.Payment.GoodsTotal,
			&orderInfo.Payment.CustomFee,
			&temp.ChrtID, &temp.TrackNumber, &temp.Price,
			&temp.Rid, &temp.Name, &temp.Sale, &temp.Size, &temp.TotalPrice, &temp.NmID,
			&temp.Brand, &temp.Status,
		)
		orderInfo.Items = make(model.Items, 0)
		orderInfo.Items = model.Items{
			{ChrtID: temp.ChrtID, TrackNumber: temp.TrackNumber, Price: temp.Price, Rid: temp.Rid, Name: temp.Name, Sale: temp.Sale, Size: temp.Size, TotalPrice: temp.TotalPrice, NmID: temp.NmID, Brand: temp.Brand, Status: temp.Status},
		}
		if err != nil {
			return nil, err
		}
		orders = append(orders, orderInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
