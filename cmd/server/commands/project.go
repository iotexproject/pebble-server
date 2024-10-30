package commands

import (
	"bytes"
	"compress/zlib"
	"context"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/confcmd"
)

//go:embed confirm.json
var abi string

func NewDefaultSproutConfigGenerator() *SproutConfigGenerator {
	return &SproutConfigGenerator{
		Network:     "mainnet",
		OutputPath:  ".",
		WasmVersion: "v0.0.1",
	}
}

type SproutConfigGenerator struct {
	Network     string `               help:"iotx network testnet or mainnet"`
	OutputPath  string `               help:"config file output path"`
	WasmVersion string `               help:"wasm version"`
	WasmPath    string `cmd:",require" help:"wasm code path"`
	DataSource  string `cmd:",require" help:"project datasource"`
}

var _ confcmd.Executor = (*SproutConfigGenerator)(nil)

func (g *SproutConfigGenerator) Use() string {
	return "project"
}

func (g *SproutConfigGenerator) Short() string {
	return "generate sprout project configuration"
}

type Project struct {
	DataSourceURI  string    `json:"datasourceURI,omitempty"`
	DefaultVersion string    `json:"defaultVersion"`
	Versions       []*Config `json:"versions"`
}

type Config struct {
	Version string  `json:"version"`
	VMType  string  `json:"vmType"`
	Output  *Output `json:"output"`
	Code    string  `json:"code"`
}

type Output struct {
	Type     string `json:"type"`
	Ethereum struct {
		ChainEndpoint   string `json:"chainEndpoint"`
		ContractAddress string `json:"contractAddress"`
		ContractMethod  string `json:"contractMethod"`
		ContractAbiJSON string `json:"contractAbiJSON"`
	} `json:"ethereum"`
}

func (g *SproutConfigGenerator) Exec(cmd *cobra.Command, args ...string) error {
	content, err := os.ReadFile(g.WasmPath)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(nil)
	w := zlib.NewWriter(b)
	defer w.Close()

	_, err = w.Write(content)
	if err != nil {
		return nil
	}
	w.Close()
	code := hex.EncodeToString(b.Bytes())

	output := &Output{
		Type: "ethereumContract",
		Ethereum: struct {
			ChainEndpoint   string `json:"chainEndpoint"`
			ContractAddress string `json:"contractAddress"`
			ContractMethod  string `json:"contractMethod"`
			ContractAbiJSON string `json:"contractAbiJSON"`
		}{
			ContractMethod:  "confirm",
			ContractAbiJSON: abi,
		},
	}
	project := &Project{
		DataSourceURI:  g.DataSource,
		DefaultVersion: g.WasmVersion,
		Versions: []*Config{{
			Version: g.WasmVersion,
			VMType:  "wasm",
			Output:  output,
			Code:    code,
		}},
	}

	switch g.Network {
	default:
		output.Ethereum.ChainEndpoint = "https://babel-api.mainnet.iotex.io"
		output.Ethereum.ContractAddress = "0xC9D7D9f25b98119DF5b2303ac0Df6b15C982BbF5"
	case "testnet":
		output.Ethereum.ChainEndpoint = "https://babel-api.testnet.iotex.io"
		output.Ethereum.ContractAddress = "0x1AA325E5144f763a520867c56FC77cC1411430d0"
	}

	f, err := os.OpenFile(
		filepath.Join(g.OutputPath, "project.json"),
		os.O_RDWR|os.O_CREATE, 0666,
	)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	return enc.Encode(project)
}

func GenerateSproutConfig(_ context.Context) *cobra.Command {
	return confcmd.NewCommand(NewDefaultSproutConfigGenerator())
}
