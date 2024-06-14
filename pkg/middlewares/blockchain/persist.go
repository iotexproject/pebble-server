package blockchain

import "github.com/ethereum/go-ethereum/core/types"

type persistence interface {
	Range(id string) (int64, int64, error)
	Upsert(id string, l types.Log)
	Query(id string, from, end int64) ([]types.Log, error)
}
