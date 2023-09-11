package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	bloodlabnet "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net"
	"github.com/urfave/cli/v2"
)

func QueryCommand(app *cli.App) {
	var (
		listenPort     int
		maxConn        int
		protocol       string
		proxy          string
		startByte      string
		endByte        string
		lineBreakByte  string
		rawBytes       bool
		showLinebreaks bool
		outPutFileName string
	)

	command := cli.Command{
		Name:    "query",
		Aliases: nil,
		Usage: `send filecontent to device, waiting for answer, log answer to file
		cli args -> query <filename> <protocol [raw|lis1a1|stxetx|mllp] Default:raw> <device> <listenport> <maxcon Default:10> <proxy: noproxy or haproxyv2>`,
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

			filename := args.Get(0)
			file, err := os.Open(filename)
			if err != nil {
				return fmt.Errorf("failed to open file: %s error: %s", filename, err.Error())
			}

			protocol = args.Get(1)

			deviceHost, devicePort, err := net.SplitHostPort(args.Get(2))
			if err != nil {
				return fmt.Errorf("invalid device: %s err: %s", args.Get(2), err.Error())
			}

			devicePortInt, err := strconv.Atoi(devicePort)
			if err != nil {
				return fmt.Errorf("invalid port in hostname: %s err: %s", devicePort, err.Error())
			}

			listenPort, err = strconv.Atoi(args.Get(3))
			if err != nil {
				return fmt.Errorf("invalid port number. Can not parse string to int: %w", err)
			}

			maxConn, err = strconv.Atoi(args.Get(4))
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

			outPutFileName = fmt.Sprintf("log_%s.txt", time.Now().Format("20060102_150405"))

			flags := c.Args().Slice()

			isRaw := sliceContains(flags, "--raw")
			if isRaw {
				rawBytes = !rawBytes
			}

			scanner := bufio.NewScanner(file)
			scanner.Split(bufio.ScanLines)
			var fileLines = make([][]byte, 0)
			for scanner.Scan() {
				fileLines = append(fileLines, []byte(scanner.Text()))
			}

			tcpClient := bloodlabnet.CreateNewTCPClient(deviceHost, devicePortInt, protocolImplementation, bloodlabnet.NoLoadBalancer, bloodlabnet.DefaultTCPClientSettings)
			err = tcpClient.Connect()
			if err != nil {
				return fmt.Errorf("cannot connect to host (%s): %s", args.Get(2), err.Error())
			}

			n, err := tcpClient.Send(fileLines)
			if err != nil {
				return fmt.Errorf("failed to send file to host: %s", err.Error())
			}
			time.Sleep(time.Second * 5)

			if n <= 0 {
				println("No data was sent by the client")
			} else {
				println("Successfully sent data")
			}

			tcpClient.Close()

			showLinebreaks = sliceContains(flags, "--showLinebreaks")
			tcpServerHandler := NewTCPServerHandler(rawBytes, showLinebreaks, outPutFileName)
			tcpServer := bloodlabnet.CreateNewTCPServerInstance(listenPort, protocolImplementation, connectionType, maxConn)
			tcpServer.Run(tcpServerHandler)
			return nil
		},
		Subcommands: nil,
	}

	app.Commands = append(app.Commands, &command)
}
