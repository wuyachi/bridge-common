package neo3

import (
	"fmt"
	"github.com/joeqian10/neo-gogogo/helper"
	"github.com/joeqian10/neo3-gogogo/crypto"
	"github.com/joeqian10/neo3-gogogo/rpc"
	"github.com/polynetwork/bridge-common/chains"
	"github.com/polynetwork/bridge-common/util"
	"time"
)

type Rpc = rpc.RpcClient

type Client struct {
	*Rpc
	address string
}

func New(url string) *Client {
	client := rpc.NewClient(url)
	return &Client{
		Rpc:     client,
		address: url,
	}
}

func (c *Client) Address() string {
	return c.address
}

func (c *Client) GetLatestHeight() (uint64, error) {
	res := c.GetBlockCount()
	if res.ErrorResponse.Error.Message != "" {
		return 0, fmt.Errorf("%s", res.ErrorResponse.Error.Message)
	}
	return uint64(res.Result), nil
}

func (c *Client) GetPolyEpochHeight(ccm string) (height uint64, err error) {
	response := c.GetStorage(ccm, "AgE=")
	if response.HasError() {
		return 0, fmt.Errorf("[GetCurrentNeoChainSyncHeight] GetStorage error: %s", response.GetErrorInfo())
	}
	s := response.Result
	if s == "" {
		return 0, nil
	}
	b, err := crypto.Base64Decode(s)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		height = 0
	} else {
		height = helper.BytesToUInt64(b)
		height++ // means the next block header needs to be synced
	}
	return
}

type SDK struct {
	*chains.ChainSDK
	nodes   []*Client
	options *chains.Options
}

func (s *SDK) Node() *Client {
	return s.nodes[s.ChainSDK.Index()]
}

func (s *SDK) Select() *Client {
	return s.nodes[s.ChainSDK.Select()]
}

func NewSDK(chainID uint64, urls []string, interval time.Duration, maxGap uint64) (*SDK, error) {
	clients := make([]*Client, len(urls))
	nodes := make([]chains.SDK, len(urls))
	for i, url := range urls {
		client := New(url)
		nodes[i] = client
		clients[i] = client
	}
	sdk, err := chains.NewChainSDK(chainID, nodes, interval, maxGap)
	if err != nil {
		return nil, err
	}
	return &SDK{ChainSDK: sdk, nodes: clients}, nil
}

func WithOptions(chainID uint64, urls []string, interval time.Duration, maxGap uint64) (*SDK, error) {
	sdk, err := util.Single(&SDK{
		options: &chains.Options{
			ChainID:  chainID,
			Nodes:    urls,
			Interval: interval,
			MaxGap:   maxGap,
		},
	})
	if err != nil {
		return nil, err
	}
	return sdk.(*SDK), nil
}

func (s *SDK) Create() (interface{}, error) {
	return NewSDK(s.options.ChainID, s.options.Nodes, s.options.Interval, s.options.MaxGap)
}

func (s *SDK) Key() string {
	if s.ChainSDK != nil {
		return s.ChainSDK.Key()
	} else if s.options != nil {
		return s.options.Key()
	} else {
		panic("Unable to identify the sdk")
	}
}
