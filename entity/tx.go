package entity

type Transaction struct {
	ID      string `json:"id" binding:"required"`
	Payload string `json:"payload" binding:"required"`
}
