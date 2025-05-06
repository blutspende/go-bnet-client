package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	bloodlabnet "github.com/blutspende/go-bnet"
	bloodlabnetProtocol "github.com/blutspende/go-bnet/protocol"
	"github.com/urfave/cli/v2"
)

var Version = "0.5.3"

func main() {
	app := &cli.App{
		Name:                 "Bnet-tool",
		Usage:                "Communicate with lab equipment over low-level protocols mllp, lisA1, StxEtx",
		EnableBashCompletion: true,
		Version:              Version,
	}

	SendingCommand(app)
	ListeningCommand(app)
	DeviceCommand(app)
	QueryCommand(app)
	FTPServerCommand(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func makeLowLevelProtocol(protocol string, startBytesStr, endBytesStr string) (bloodlabnetProtocol.Implementation, error) {
	var protocolTypeImplementation bloodlabnetProtocol.Implementation
	switch protocol {
	case "raw":
		protocolTypeImplementation = bloodlabnetProtocol.Raw(bloodlabnetProtocol.DefaultRawProtocolSettings())
	case "stxetx":
		protocolTypeImplementation = bloodlabnetProtocol.STXETX(bloodlabnetProtocol.DefaultSTXETXProtocolSettings())
	case "lis1a1":
		protocolTypeImplementation = bloodlabnetProtocol.Lis1A1Protocol(bloodlabnetProtocol.DefaultLis1A1ProtocolSettings())
	case "mllp":
		config := bloodlabnetProtocol.DefaultMLLPProtocolSettings()
		if startBytesStr != "" {
			startBytes, err := hex.DecodeString(startBytesStr)
			if err != nil {
				return nil, fmt.Errorf("invalid startbytes: %s, %s", startBytesStr, err.Error())
			}
			config = config.SetStartBytes(startBytes)
		}

		if endBytesStr != "" {
			endBytes, err := hex.DecodeString(endBytesStr)
			if err != nil {
				return nil, fmt.Errorf("invalid endbytes: %s, %s", endBytesStr, err.Error())
			}
			config = config.SetEndBytes(endBytes)
		}
		protocolTypeImplementation = bloodlabnetProtocol.MLLP(config)
	default:
		return nil, fmt.Errorf("invalid protocol type given: %s , supported:raw,stxetx,lis1a1,mllp", protocol)

	}

	return protocolTypeImplementation, nil
}

func getProxyCoonnectionType(proxy string) (bloodlabnet.ConnectionType, error) {
	switch strings.ToLower(proxy) {
	case "noproxy":
		return bloodlabnet.NoLoadBalancer, nil
	case "haproxyv2":
		return bloodlabnet.HAProxySendProxyV2, nil
	case "":
		return bloodlabnet.NoLoadBalancer, nil
	default:
		return bloodlabnet.NoLoadBalancer, fmt.Errorf("invalid proxy connection type. valid values: 'noproxy', 'haproxyv2' given: %s", strings.ToLower(proxy))
	}
}
