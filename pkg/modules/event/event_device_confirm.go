package event

import (
	"context"
	"crypto/sha256"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
	"google.golang.org/protobuf/proto"

	pebble "github.com/machinefi/sprout-pebble-sequencer/contract"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/pebblepb"
)

func init() {
	e := &DeviceConfirm{}
	registry(e.Topic(), func() Event { return &DeviceConfirm{} })
}

type DeviceConfirm struct {
	IMEI
	SignatureValidator
	pkg *pebblepb.ConfirmPackage
}

func (e *DeviceConfirm) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__MQTT
}

func (e *DeviceConfirm) Topic() string { return "device/+/confirm" }

func (e *DeviceConfirm) UnmarshalTopic(topic []byte) error {
	return (&TopicParser{e, topic, "device", "confirm"}).Unmarshal()
}

func (e *DeviceConfirm) Unmarshal(v any) (err error) {
	data, ok := v.([]byte)
	must.BeTrueWrap(ok, "assertion unmarshal with bytes")

	defer func() { err = WrapUnmarshalError(err, e) }()

	pkg := &pebblepb.ConfirmPackage{}
	if err = proto.Unmarshal(data, pkg); err != nil {
		return errors.Wrap(err, "failed to unmarshal proto")
	}
	e.pkg = pkg

	var (
		sig   = pkg.GetSignature()
		owner = pkg.GetOwner()
		ts    = pkg.GetTimestamp()
	)

	if len(sig) != 64 {
		return errors.Errorf("unexpected sig, expect 64 bytes but got %d", len(sig))
	}
	e.sig = append(sig, 0)

	buf := make([]byte, len(owner)+4)
	copy(buf, owner)
	gByteOrder.PutUint32(buf[len(owner):], ts)

	sum := sha256.Sum256(buf)
	e.hash = sum[:]

	return nil
}

type message struct {
	IMEI        string `json:"imei"`
	Owner       string `json:"owner"`
	Timestamp   uint32 `json:"timestamp"`
	Signature   string `json:"signature"`
	GasLimit    string `json:"gasLimit"`
	DataChannel uint32 `json:"dataChannel"`
}

func (e *DeviceConfirm) Handle(ctx context.Context) (err error) {
	defer func() {
		err = WrapHandleError(err, e)
	}()

	// if !contexts.IMEIFilter().MustFrom(ctx).NeedHandle(e.Imei) {
	// 	return errors.Errorf("imei %s not in whitelist", e.Imei)
	// }

	dev := &models.Device{ID: e.Imei}
	err = FetchByPrimary(ctx, dev)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch dev: %s", dev.ID)
	}

	if dev.Status != models.PROPOSAL {
		return errors.Errorf("device `%s` is %d, donnot need confirm", dev.ID, dev.Status)
	}

	e.addr = common.HexToAddress(dev.Address)
	if !e.Validate() {
		return WrapValidateError(e)
	}

	operatorPrivateKey := contexts.OperatorPrivateKey().MustFrom(ctx)
	client, err := ethclient.Dial("https://babel-api.mainnet.iotex.io")
	if err != nil {
		return errors.Wrap(err, "failed to dial chain endpoint")
	}
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to get chain id")
	}
	instance, err := pebble.NewPebble(common.HexToAddress("0xC9D7D9f25b98119DF5b2303ac0Df6b15C982BbF5"), client)
	if err != nil {
		return errors.Wrap(err, "failed to new erc20 contract instance")
	}
	prv, err := crypto.HexToECDSA(operatorPrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to parse private key")
	}

	_, err = instance.Confirm(
		&bind.TransactOpts{
			From: crypto.PubkeyToAddress(prv.PublicKey),
			Signer: func(a common.Address, t *types.Transaction) (*types.Transaction, error) {
				return types.SignTx(t, types.NewLondonSigner(chainID), prv)
			},
		},
		e.Imei, common.BytesToAddress(e.pkg.GetOwner()), e.pkg.GetTimestamp(), e.sig,
		big.NewInt(200000), uint32(dev.DataChannel),
	)
	if err != nil {
		return errors.Wrap(err, "failed to confirm device")
	}

	return nil
}
