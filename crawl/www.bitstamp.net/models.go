package main

type BtcData struct {
	Event   string `json:"event"`
	Data    *Adata `json:"data"`
	Channel string `json:"channel"`
}

type Adata struct {
	Amount        float64 `json:"amount"`
	Buy_order_id  int64   `json:"buy_order_id"`
	Sell_order_id int64   `json:"sell_order_id"`
	Amount_str    string  `json:"amount_str"`
	Price_str     string  `json:"price_str"`
	Timestamp     string  `json:"timestamp"`
	Price         float64 `json:"price"`
	Type          int     `json:"type"`
	Id            int64   `json:"id"`
}
