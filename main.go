package main

import (
	"fmt"
	bloodlabnet "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net"
	bloodlabnetProtocol "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net/protocol"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"
)

func main() {

	var path string = ""
	var hostname string = ""
	var protocolType string = ""
	var protocolTypeImplementation bloodlabnetProtocol.Implementation
	var (
		startByte     string = ""
		endByte       string = ""
		lineBreakByte string = ""
	)

	app := &cli.App{
		Name:                 "BloodLab Net client helper",
		Usage:                "Transfer a file to a TCP Server over BloodlabNet library",
		EnableBashCompletion: true,
		Version:              "v0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "file",
				Usage:       "FilePath incl. path",
				Value:       "",
				Required:    true,
				Destination: &path,
			},
			&cli.StringFlag{
				Name:        "host",
				Usage:       "Hostname",
				Value:       "",
				Required:    true,
				Destination: &hostname,
			},
			&cli.StringFlag{
				Name:        "protocol",
				Usage:       "protocol",
				Required:    true,
				Value:       "raw",
				Destination: &protocolType,
			},
			&cli.StringFlag{
				Name:        "startbyte",
				Usage:       "startbyte for some protocols",
				Required:    true,
				Value:       "",
				Destination: &startByte,
			},
			&cli.StringFlag{
				Name:        "endbyte",
				Usage:       "endbyte for some protocols",
				Required:    true,
				Value:       "",
				Destination: &endByte,
			}, &cli.StringFlag{
				Name:        "linebreak",
				Usage:       "linebreak for some protocols",
				Required:    true,
				Value:       "",
				Destination: &lineBreakByte,
			},
		},
		Action: func(c *cli.Context) error {
			switch protocolType {
			case "raw":
				protocolTypeImplementation = bloodlabnetProtocol.Raw(bloodlabnetProtocol.DefaultRawProtocolSettings())
			case "stxetx":
				protocolTypeImplementation = bloodlabnetProtocol.STXETX(bloodlabnetProtocol.DefaultSTXETXProtocolSettings())
			case "lis1a1":
				protocolTypeImplementation = bloodlabnetProtocol.Lis1A1Protocol(bloodlabnetProtocol.DefaultLis1A1ProtocolSettings())
			case "mlp":
				config := bloodlabnetProtocol.DefaultMLLPProtocolSettings()
				if startByte != "" {
					startByteInt, err := strconv.Atoi(startByte)
					if err != nil {
						return fmt.Errorf("invalid startbyte: %s, err: %s", startByte, err.Error())
					}
					config = config.SetStartByte(byte(startByteInt))
				}

				if endByte != "" {
					endByteInt, err := strconv.Atoi(endByte)
					if err != nil {
						return fmt.Errorf("invalid endbyte: %s, err: %s", endByteInt, err.Error())
					}
					config = config.SetStartByte(byte(endByteInt))
				}
				protocolTypeImplementation = bloodlabnetProtocol.MLLP(config)
			default:
				return fmt.Errorf("invalid protocol type given: %s , supported:raw,stxetx,lis1a1,mlp", protocolType)

			}

			host, port, err := net.SplitHostPort(hostname)
			if err != nil {
				return fmt.Errorf("invalid host: %s err: %s", hostname, err.Error())
			}

			portInt, err := strconv.Atoi(port)
			if err != nil {
				return fmt.Errorf("invalid port in hostname: %s err: %s", port, err.Error())
			}

			bytes, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to open file with path: %s and error: %s", path, err.Error())
			}

			tcpClient := bloodlabnet.CreateNewTCPClient(host, portInt, protocolTypeImplementation, bloodlabnet.NoLoadBalancer, bloodlabnet.DefaultTCPServerSettings)
			err = tcpClient.Connect()
			if err != nil {
				return fmt.Errorf("cannot connect to host (%s): %s", hostname, err.Error())
			}

			_, err = tcpClient.Send([][]byte{bytes})
			if err != nil {
				return fmt.Errorf("failed to send file to host: %s", err.Error())
			}

			println("Successfully sent")

			tcpClient.Close()
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
