package models

type BankRecord struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    string `json:"amount"`
	Timestamp int64  `json:"timestamp"`
	Type      int32  `json:"type"`
}
