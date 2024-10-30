package clients

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/machinefi/ioconnect-go/pkg/ioconnect"
)

type Client struct {
	owner common.Address
	jwk   *ioconnect.JWK
}

func (c *Client) KeyAgreementKID() string {
	return c.jwk.KeyAgreementKID()
}

func (c *Client) DID() string {
	return c.jwk.DID()
}

func (c *Client) Doc() *ioconnect.Doc {
	return c.jwk.Doc()
}

func (c *Client) Owner() common.Address {
	return c.owner
}
