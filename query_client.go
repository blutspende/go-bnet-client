package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"

	bloodlabnet "github.com/blutspende/go-bloodlab-net"
	"github.com/urfave/cli/v2"
)

func QueryCommand(app *cli.App) {

	var (
		deviceHost     string
		devicePort     string
		devicePortInt  int
		protocol       string
		proxy          string
		startByte      string
		endByte        string
		lineBreakByte  string
		rawBytes       bool
		showLinebreaks bool
	)

	command := cli.Command{
		Name:    "query",
		Aliases: nil,
		Usage:   `query <filename> <raw|lis1a1|stxetx|mllp> <devicehost> <listenport> [--proxy haproxyv2]`,
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
				Usage:       "proxy noproxy (default), haproxyv2",
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

			deviceHost, devicePort, err = net.SplitHostPort(args.Get(2))
			if err != nil {
				return fmt.Errorf("invalid device: %s err: %s", args.Get(2), err.Error())
			}

			devicePortInt, err = strconv.Atoi(devicePort)
			if err != nil {
				return fmt.Errorf("invalid port in hostname: %s err: %s", devicePort, err.Error())
			}

			protocolImplementation, err := makeLowLevelProtocol(protocol, startByte, endByte)
			if err != nil {
				return fmt.Errorf("invalid protocol %s - %w", protocol, err)
			}

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

			tcpClient := bloodlabnet.CreateNewTCPClient(
				deviceHost,
				devicePortInt,
				protocolImplementation,
				bloodlabnet.NoLoadBalancer,
				bloodlabnet.DefaultTCPClientSettings)

			err = tcpClient.Connect()
			if err != nil {
				return fmt.Errorf("cannot connect to host (%s): %s", args.Get(2), err.Error())
			}

			n, err := tcpClient.Send(fileLines)
			if err != nil {
				return fmt.Errorf("failed to send file to host: %s", err.Error())
			}
			if n <= 0 {
				return fmt.Errorf("no data was sent by the client")
			}
			println("Successfully sent data.")

			bytes, err := tcpClient.Receive()
			if err != nil {
				return fmt.Errorf("failed while waiting for a response - %s", err.Error())
			}

			fmt.Printf("Device-Host's response (%d bytes) (Non-printable characters in brackets <xx> decimal)\n", len(bytes))
			fmt.Printf("--- start\n")
			for _, x := range bytes {
				if x < 32 {
					fmt.Printf("<%02d>", x)
				} else {
					fmt.Printf("%c", x)
				}
			}
			fmt.Printf("\n--- end\n")

			tcpClient.Close()

			return nil

		},
		Subcommands: nil,
	}

	app.Commands = append(app.Commands, &command)
}
