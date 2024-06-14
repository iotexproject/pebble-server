package blockchain

import (
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Blockchain struct {
	Clients EthClients
	Monitor Monitor

	insts   sync.Map
	persist persistence
}

func (bc *Blockchain) Init() error {
	if bc.Clients == nil {
		bc.Clients = make(EthClients)
	}

	if err := bc.Clients.Init(); err != nil {
		return err
	}

	if err := bc.Monitor.Init(); err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) ClientByName(name string) *EthClient {
	return bc.Clients[name]
}

func (bc *Blockchain) ClientByEndpoint(ep string) *EthClient {
	for _, c := range bc.Clients {
		if c.Endpoint == ep {
			return c
		}
	}
	return nil
}

func (bc *Blockchain) NewMonitor(id, metaID string, from, end int64) (*MonitorInstance, error) {
	meta := bc.Monitor.metas[metaID]
	if meta == nil {
		return nil, errors.Errorf("monitor meta not found: %s", metaID)
	}

	client := bc.ClientByEndpoint(meta.Endpoint)

	if id == "" {
		id = uuid.NewString()
	}

	inst := &MonitorInstance{
		ID:          id,
		MonitorMeta: *meta,
		From:        from,
		End:         end,
		client:      client,
		stop:        make(chan struct{}),
		persis:      bc.persist,
	}
	bc.insts.Store(inst.ID, inst)
	return inst, nil
}

func (bc *Blockchain) NewMonitorDefault(id, metaID string) (*MonitorInstance, error) {
	_, latest, err := bc.persist.Range(metaID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query latest from persistence")
	}
	return bc.NewMonitor(id, metaID, latest, -1)
}
