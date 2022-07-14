package main

import (
	"fmt"
	"strconv"
	"time"

	bloodlabNet "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net"
	"github.com/urfave/cli/v2"
)

func ListeningCommand(app *cli.App) {
	var (
		listenPort    int
		maxConn       int
		protocol      string
		proxy         string
		err           error
		startByte     string
		endByte       string
		lineBreakByte string
	)

	command := cli.Command{
		Name:        "listen",
		Aliases:     nil,
		Usage:       "listen <port> <protocol> <maxConn>",
		UsageText:   "",
		Description: "",
		ArgsUsage:   "",
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
			givenFlags := c.NArg()
			if givenFlags != 4 {
				return fmt.Errorf("invalid amount of arguments. Required: <listenPort> <protocol> <maxconnections> <proxy>")
			}

			listenPort, err = strconv.Atoi(args.Get(0))
			if err != nil {
				return fmt.Errorf("invalid port number. Can not parse string to int: %w", err)
			}
			protocol = args.Get(1)
			maxConn, err = strconv.Atoi(args.Get(2))
			if err != nil {
				return fmt.Errorf("invalid max conn. Can not parse string to int: %w", err)
			}

			proxy = args.Get(3)
			connectionType, err := getProxyCoonnectionType(proxy)
			if err != nil {
				return fmt.Errorf("invalid proxy type: %w", err)
			}

			protocolImplementation, err := getLowLevelProtocol(protocol, startByte, endByte)
			if err != nil {
				return fmt.Errorf("can not find protocol: %w", err)
			}

			tcpHandler := NewTCPServerHandler()
			tcpServer := bloodlabNet.CreateNewTCPServerInstance(listenPort, protocolImplementation, connectionType, maxConn)
			tcpServer.Run(tcpHandler)
			return nil
		},
		Subcommands: nil,
	}

	app.Commands = append(app.Commands, &command)
}

type TCPServerHandler interface {
	DataReceived(session bloodlabNet.Session, fileData []byte, receiveTimestamp time.Time)
	Connected(con bloodlabNet.Session)
	Disconnected(session bloodlabNet.Session)
	Error(session bloodlabNet.Session, typeOfError bloodlabNet.ErrorType, err error)
}

type tcpServerHandler struct {
}

func NewTCPServerHandler() TCPServerHandler {
	return &tcpServerHandler{}
}
func (h *tcpServerHandler) Connected(session bloodlabNet.Session) {
	// Is instrument whitelisted
	remoteAddress, err := session.RemoteAddress()
	if err != nil {
		println(fmt.Errorf("can not get remote address: %w", err))
		session.Close()
	}
	println(fmt.Sprintf("Client successfully conntected with IP: %s", remoteAddress))
}

func (h *tcpServerHandler) DataReceived(session bloodlabNet.Session, fileData []byte, receiveTimestamp time.Time) {
	remoteAddress, err := session.RemoteAddress()
	if err != nil {
		println(fmt.Errorf("can not get remote address: %w", err))
	}

	println(fmt.Sprintf("Client(%s) successfully transferred Data(%d bytes): %s ", remoteAddress, len(fileData), string(fileData)))
}

func (h *tcpServerHandler) Disconnected(session bloodlabNet.Session) {
	// Is instrument whitelisted
	remoteAddress, err := session.RemoteAddress()
	if err != nil {
		println(fmt.Errorf("can not get remote address: %w", err))
	}

	println(fmt.Sprintf("Client with %s is disconnected", remoteAddress))

}

func (h *tcpServerHandler) Error(session bloodlabNet.Session, typeOfError bloodlabNet.ErrorType, err error) {
	println(fmt.Errorf("error: %w", err))
}
