package cli

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"log"
	"os"
)

type Client struct {
	// Fabric network information
	ConfigPath string
	OrgName    string
	OrgAdmin   string
	OrgUser    string

	// sdk clients
	SDK *fabsdk.FabricSDK
	rc  *resmgmt.Client
	cc  *channel.Client

	// Same for each peer
	ChannelID string
	CCID      string // chaincode ID, eq name
	CCPath    string // chaincode source path, 是GOPATH下的某个目录
	CCGoPath  string // GOPATH used for chaincode
}

func New(cfg, org, admin, user string) *Client {
	c := &Client{
		ConfigPath: cfg,
		OrgName:    org,
		OrgAdmin:   admin,
		OrgUser:    user,

		CCID:      "example2",
		CCPath:    "github.com/hyperledger/fabric-samples/chaincode/chaincode_example02/go/", // 相对路径是从GOPAHT/src开始的
		CCGoPath:  os.Getenv("GOPATH"),
		ChannelID: "mychannel",
	}

	// create sdk
	sdk, err := fabsdk.New(config.FromFile(c.ConfigPath))
	if err != nil {
		log.Panicf("failed to create fabric sdk: %s", err)
	}
	c.SDK = sdk
	log.Println("Initialized fabric sdk")

	c.rc, c.cc = NewSdkClient(sdk, c.ChannelID, c.OrgName, c.OrgAdmin, c.OrgUser)

	return c
}

// NewSdkClient create resource client and channel client
func NewSdkClient(sdk *fabsdk.FabricSDK, channelID, orgName, orgAdmin, OrgUser string) (rc *resmgmt.Client, cc *channel.Client) {
	var err error

	// create rc
	rcp := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
	rc, err = resmgmt.New(rcp)
	if err != nil {
		log.Panicf("failed to create resource client: %s", err)
	}
	log.Println("Initialized resource client")

	// create cc
	ccp := sdk.ChannelContext(channelID, fabsdk.WithUser(OrgUser))
	cc, err = channel.New(ccp)
	if err != nil {
		log.Panicf("failed to create channel client: %s", err)
	}
	log.Println("Initialized channel client")

	return rc, cc
}

// RegisterChaincodeEvent more easy than event client to registering chaincode event.
func (c *Client) RegisterChaincodeEvent(ccid, eventName string) (fab.Registration, <-chan *fab.CCEvent, error) {
	return c.cc.RegisterChaincodeEvent(ccid, eventName)
}
func (c *Client)GetChannelConfig(v string, peer string) {
	org1ChannelContext := c.SDK.ChannelContext(c.ChannelID, fabsdk.WithUser(c.OrgAdmin), fabsdk.WithOrg(c.OrgName))
	channelCtx, err := org1ChannelContext()
	if err != nil {
		log.Fatalf("Failed to get channel client context: %s", err)
	}
	//orderers := channelCtx.EndpointConfig().ChannelOrderers(c.ChannelID)
	//orderer, err := channelCtx.InfraProvider().CreateOrdererFromConfig(&orderers[0])
	
	cs := channelCtx.ChannelService()
	cfg, err := cs.Config()
	
	if err != nil {
		log.Fatalf("Failed to create new channel config: %s", err)
	}
	
	queryChannelCfg(channelCtx, cfg)
}
func queryChannelCfg(channelCtx contextAPI.Channel, cfg fab.ChannelConfig) {
	reqCtx, cancel := context.NewRequest(channelCtx, context.WithTimeoutType(fab.OrdererResponse))
	defer cancel()
	response, err := cfg.Query(reqCtx)
	if err != nil {
		log.Fatal(err)
	}
	expected := "orderer.example.com:7050"
	found := false
	for _, o := range response.Orderers() {
		if o == expected {
			found = true
			break
		}
	}
	if !found {
		log.Fatalf("Expected orderer %s, got %s", expected, response.Orderers())
	}
}
