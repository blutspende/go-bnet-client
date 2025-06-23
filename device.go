package main

import (
	"fmt"
	"strconv"
	"time"

	bloodlabnet "github.com/blutspende/go-bnet"
	"github.com/urfave/cli/v2"
)

func DeviceCommand(app *cli.App) {
	var (
		listenPort     int
		maxConn        int
		protocol       string
		proxy          string
		startBytes     string
		endBytes       string
		lineBreakByte  string
		rawBytes       bool
		showLinebreaks bool
		err            error
	)

	command := cli.Command{
		Name:    "device",
		Aliases: nil,
		Usage: `simulate device recieve query, send query back
		cli args -> device <protocol [raw|lis1a1|stxetx|mllp]> <listenport> <maxcon> <proxy: noproxy or haproxyv2> <queryhost> <queryanswerfile>`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "startbytes",
				Usage:       "startbytes for MLLP protocol provided in the format '0A0D1C'",
				DefaultText: "0B",
				Required:    false,
				Value:       "0B",
				Destination: &startBytes,
			},
			&cli.StringFlag{
				Name:        "endbytes",
				Usage:       "endbytes for MLLP protocol provided in the format '0A0D1C'",
				DefaultText: "1C0D",
				Required:    false,
				Value:       "1C0D",
				Destination: &endBytes,
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

			protocol = args.Get(0)
			listenPort, err = strconv.Atoi(args.Get(1))
			if err != nil {
				return fmt.Errorf("invalid port number. Can not parse string to int: %w", err)
			}

			maxConn, err = strconv.Atoi(args.Get(2))
			if err != nil {
				return fmt.Errorf("invalid max conn. Can not parse string to int: %w", err)
			}

			connectionType, err := getProxyCoonnectionType(proxy)
			if err != nil {
				return fmt.Errorf("invalid proxy type: %w", err)
			}

			queryhost := args.Get(4)

			queryAnswerFile := args.Get(5)

			protocolImplementation, err := makeLowLevelProtocol(protocol, startBytes, endBytes)
			if err != nil {
				return fmt.Errorf("can not find protocol: %w", err)
			}

			flags := c.Args().Slice()

			isRaw := sliceContains(flags, "--raw")
			if isRaw {
				rawBytes = !rawBytes
			}

			showLinebreaks = sliceContains(flags, "--showLinebreaks")
			outPutFileName := fmt.Sprintf("device_%s.log", time.Now().Format("20060102_150405"))
			tcpServerHandler := NewTCPServerHandler(rawBytes, showLinebreaks, outPutFileName, queryhost, queryAnswerFile, protocolImplementation)
			tcpServer := bloodlabnet.CreateNewTCPServerInstance(listenPort, protocolImplementation, connectionType, maxConn)
			tcpServer.Run(tcpServerHandler)
			return nil
		},
		Subcommands: nil,
	}

	app.Commands = append(app.Commands, &command)
}
