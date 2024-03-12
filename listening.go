package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	bloodlabNet "github.com/blutspende/go-bloodlab-net"
	"github.com/urfave/cli/v2"
)

func ListeningCommand(app *cli.App) {
	var (
		listenPort     int
		maxConn        int
		protocol       string
		proxy          string
		err            error
		startByte      string
		endByte        string
		lineBreakByte  string
		rawBytes       bool
		showLinebreaks bool
	)

	command := cli.Command{
		Name:    "listen",
		Aliases: nil,
		Usage: `listen for incomming message
		cli args -> listen <listenport> <protocol [raw|lis1a1|stxetx|mllp]> <maxcon> <proxy: noproxy or haproxyv2> <logfile (optional):  yes or no>`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "startbyte",
				Usage:       "startbyte for some protocols",
				Required:    false,
				Value:       "",
				Destination: &startByte,
			},
			&cli.StringFlag{
				Name:        "endbyte",
				Usage:       "endbyte for some protocols",
				Required:    false,
				Value:       "",
				Destination: &endByte,
			}, &cli.StringFlag{
				Name:        "linebreak",
				Usage:       "linebreak for some protocols",
				Required:    false,
				Value:       "",
				Destination: &lineBreakByte,
			},
			&cli.StringFlag{
				Name:        "proxy",
				Usage:       "proxy (noproxy, haproxyv2)",
				Required:    false,
				Value:       "noproxy",
				Destination: &proxy,
			},
			&cli.BoolFlag{
				Name:        "raw",
				Usage:       "raw bytes in print",
				Required:    false,
				Value:       false,
				Destination: &rawBytes,
			},
			&cli.BoolFlag{
				Name:        "showLinebreaks",
				Usage:       "showLinebreaks bytes in print",
				Required:    false,
				Value:       false,
				Destination: &showLinebreaks,
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args()

			listenPort, err = strconv.Atoi(args.Get(0))
			if err != nil {
				return fmt.Errorf("invalid port number. Can not parse string to int: %w", err)
			}
			protocol = args.Get(1)
			maxConn, err = strconv.Atoi(args.Get(2))
			if err != nil {
				return fmt.Errorf("invalid max conn. Can not parse string to int: %w", err)
			}

			connectionType, err := getProxyCoonnectionType(proxy)
			if err != nil {
				return fmt.Errorf("invalid proxy type: %w", err)
			}

			protocolImplementation, err := getLowLevelProtocol(protocol, startByte, endByte)
			if err != nil {
				return fmt.Errorf("can not find protocol: %w", err)
			}

			logfile := args.Get(4)
			outPutFileName := ""
			if strings.ToLower(logfile) == "yes" {
				outPutFileName = fmt.Sprintf("listen_%s.log", time.Now().Format("20060102_150405"))
			}

			flags := c.Args().Slice()

			isRaw := sliceContains(flags, "--raw")
			if isRaw {
				rawBytes = !rawBytes
			}

			showLinebreaks = sliceContains(flags, "--showLinebreaks")

			tcpServerHandler := NewTCPServerHandler(rawBytes, showLinebreaks, outPutFileName, "", "", protocolImplementation)
			tcpServer := bloodlabNet.CreateNewTCPServerInstance(listenPort, protocolImplementation, connectionType, maxConn)
			tcpServer.Run(tcpServerHandler)
			return nil
		},
		Subcommands: nil,
	}

	app.Commands = append(app.Commands, &command)
}

func sliceContains(list []string, search string) bool {
	for _, item := range list {
		if item == search {
			return true
		}
	}
	return false
}
