package api

type Register struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type Auth struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type Item struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        uint   `json:"cost"`
	Weight      uint   `json:"weight"`
	Type        string `json:"type"`
}

type CreateOrderItem struct {
	Id    uint `json:"id"`
	Count uint `json:"count"`
}

type CreateOrderReq struct {
	Address string            `json:"address"`
	Items   []CreateOrderItem `json:"items"`
}

type GetOrderItem struct {
	Id    uint `json:"id"`
	Count uint `json:"count"`
	Cost  uint `json:"cost"`
}

type GetOrderReq struct {
	Address string         `json:"address"`
	Items   []GetOrderItem `json:"items"`
	Cost    uint           `json:"cost"`
	Status  string         `json:"status"`
}
