package monitor

import (
	"context"
	"log/slog"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	"github.com/iotexproject/pebble-server/contract/project"
)

type (
	ScannedBlockNumber       func() (uint64, error)
	UpsertScannedBlockNumber func(uint64) error
	UpsertProjectMetadata    func(projectID uint64, key [32]byte, value []byte) error
)

type Handler struct {
	ScannedBlockNumber
	UpsertScannedBlockNumber
	UpsertProjectMetadata
}

type ContractAddr struct {
	IOID         common.Address
	IOIDRegistry common.Address
	Project      common.Address
}

type contract struct {
	h                    *Handler
	projectAddr          common.Address
	beginningBlockNumber uint64
	listStepSize         uint64
	watchInterval        time.Duration
	client               *ethclient.Client
	projectInstance      *project.Project
}

var (
	projectAddMetadataTopic = crypto.Keccak256Hash([]byte("AddMetadata(uint256,string,bytes32,bytes)"))
)

var allTopic = []common.Hash{
	projectAddMetadataTopic,
}

func (c *contract) processLogs(logs []types.Log) error {
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].BlockNumber != logs[j].BlockNumber {
			return logs[i].BlockNumber < logs[j].BlockNumber
		}
		return logs[i].TxIndex < logs[j].TxIndex
	})

	for _, l := range logs {
		switch l.Topics[0] {
		case projectAddMetadataTopic:
			e, err := c.projectInstance.ParseAddMetadata(l)
			if err != nil {
				return errors.Wrap(err, "failed to parse project add metadata event")
			}
			if err := c.h.UpsertProjectMetadata(e.ProjectId.Uint64(), e.Key, e.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *contract) list() (uint64, error) {
	head := c.beginningBlockNumber
	h, err := c.h.ScannedBlockNumber()
	if err != nil {
		return 0, err
	}
	head = max(head, h)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{c.projectAddr},
		Topics:    [][]common.Hash{allTopic},
	}
	ctx := context.Background()
	from := head + 1
	to := from
	for {
		header, err := c.client.HeaderByNumber(ctx, nil)
		if err != nil {
			return 0, errors.Wrap(err, "failed to retrieve latest block header")
		}
		currentHead := header.Number.Uint64()
		to = from + c.listStepSize
		if to > currentHead {
			to = currentHead
		}
		if from > to {
			break
		}
		slog.Debug("listing chain", "from", from, "to", to)
		query.FromBlock = new(big.Int).SetUint64(from)
		query.ToBlock = new(big.Int).SetUint64(to)
		logs, err := c.client.FilterLogs(ctx, query)
		if err != nil {
			return 0, errors.Wrap(err, "failed to filter contract logs")
		}
		if err := c.processLogs(logs); err != nil {
			return 0, err
		}
		if err := c.h.UpsertScannedBlockNumber(to); err != nil {
			return 0, err
		}
		from = to + 1
	}
	slog.Info("contract data synchronization completed", "current_height", to)
	return to, nil
}

func (c *contract) watch(listedBlockNumber uint64) {
	scannedBlockNumber := listedBlockNumber
	query := ethereum.FilterQuery{
		Addresses: []common.Address{c.projectAddr},
		Topics:    [][]common.Hash{allTopic},
	}
	ticker := time.NewTicker(c.watchInterval)

	go func() {
		for range ticker.C {
			target := scannedBlockNumber + 1

			query.FromBlock = new(big.Int).SetUint64(target)
			query.ToBlock = new(big.Int).SetUint64(target)
			logs, err := c.client.FilterLogs(context.Background(), query)
			if err != nil {
				if !strings.Contains(err.Error(), "start block > tip height") {
					slog.Error("failed to filter contract logs", "error", err)
				}
				continue
			}
			slog.Debug("listing chain", "from", target, "to", target)
			if err := c.processLogs(logs); err != nil {
				slog.Error("failed to process logs", "error", err)
				continue
			}
			if err := c.h.UpsertScannedBlockNumber(target); err != nil {
				slog.Error("failed to upsert scanned block number", "error", err)
				continue
			}
			scannedBlockNumber = target
		}
	}()
}

func Run(h *Handler, projectAddr common.Address, beginningBlockNumber uint64, client *ethclient.Client) error {
	projectInstance, err := project.NewProject(projectAddr, client)
	if err != nil {
		return errors.Wrap(err, "failed to new project contract instance")
	}

	c := &contract{
		h:                    h,
		projectAddr:          projectAddr,
		beginningBlockNumber: beginningBlockNumber,
		listStepSize:         500,
		watchInterval:        1 * time.Second,
		client:               client,
		projectInstance:      projectInstance,
	}

	listedBlockNumber, err := c.list()
	if err != nil {
		return err
	}
	go c.watch(listedBlockNumber)

	return nil
}
