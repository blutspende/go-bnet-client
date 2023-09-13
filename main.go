package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	bloodlabnet "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net"
	bloodlabnetProtocol "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net/protocol"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "Bnet helper",
		Usage:                "Transfer a file to a TCP Server over BloodlabNet library or listen as a TCP Server",
		EnableBashCompletion: true,
	}

	SendingCommand(app)
	ListeningCommand(app)
	DeviceCommand(app)
	QueryCommand(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getLowLevelProtocol(protocol string, startByte, endByte string) (bloodlabnetProtocol.Implementation, error) {
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
		if startByte != "" {
			startByteInt, err := strconv.Atoi(startByte)
			if err != nil {
				return nil, fmt.Errorf("invalid startbyte: %s, err: %s", startByte, err.Error())
			}
			config = config.SetStartByte(byte(startByteInt))
		}

		if endByte != "" {
			endByteInt, err := strconv.Atoi(endByte)
			if err != nil {
				return nil, fmt.Errorf("invalid startbyte: %s, err: %s", endByte, err.Error())
			}
			config = config.SetEndByte(byte(endByteInt))
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
