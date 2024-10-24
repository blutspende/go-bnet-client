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

func SendingCommand(c *cli.App) {

	var path string = ""
	var hostname string = ""
	var (
		startByte     string = ""
		endByte       string = ""
		lineBreakByte string = ""
	)

	command := cli.Command{
		Name:        "send",
		Description: "send a file to a tcp device",
		Usage:       "send <hostname> <protocol [raw|lis1a1|stxetx|mllp] Default:raw> <filename>",
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
		},
		Action: func(c *cli.Context) error {

			args := c.Args()
			if c.NArg() != 3 {
				return fmt.Errorf("required: <hostname> <protocol> <file>")
			}

			protocol, err := makeLowLevelProtocol(args.Get(1), startByte, endByte)
			if err != nil {
				return fmt.Errorf("invalid protocol '%w'", err)
			}

			host, port, err := net.SplitHostPort(args.Get(0))
			if err != nil {
				return fmt.Errorf("invalid host '%s' - %s", hostname, err.Error())
			}

			portInt, err := strconv.Atoi(port)
			if err != nil {
				return fmt.Errorf("invalid port in hostname: %s err: %s", port, err.Error())
			}

			file, err := os.Open(args.Get(2))
			if err != nil {
				return fmt.Errorf("failed to open file '%s' - %s", path, err.Error())
			}

			scanner := bufio.NewScanner(file)
			scanner.Split(bufio.ScanLines)
			var fileLines = make([][]byte, 0)
			for scanner.Scan() {
				fileLines = append(fileLines, []byte(scanner.Text()))
			}

			tcpClient := bloodlabnet.CreateNewTCPClient(host, portInt, protocol, bloodlabnet.NoLoadBalancer, bloodlabnet.DefaultTCPClientSettings)
			err = tcpClient.Connect()
			if err != nil {
				return fmt.Errorf("cannot connect to host (%s): %s", hostname, err.Error())
			}

			n, err := tcpClient.Send(fileLines)
			if err != nil {
				return fmt.Errorf("failed to send file to host: %s", err.Error())
			}

			if n <= 0 {
				println("No data was sent by the client")
			} else {
				println("Successfully sent data")
			}

			tcpClient.Close()
			return nil
		},
	}

	c.Commands = append(c.Commands, &command)
}
