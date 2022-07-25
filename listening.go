package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	bloodlabNet "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net"
	"github.com/urfave/cli/v2"
)

const (
	CR byte = 0x0D
)

var replaceAbleBytes = map[byte]string{
	0x00: "<NUL>",
	0x01: "<SOH>",
	0x02: "<STX>",
	0x03: "<ETX>",
	0x04: "<EOT>",
	0x05: "<ENQ>",
	0x06: "<ACK>",
	0x07: "<BEL>",
	0x08: "<BS>",
	0x09: "<HT>",
	0x0A: "<LF>",
	0x0B: "<VT>",
	0x0C: "<FF>",
	0x0D: "<CR>",
	0x0E: "<SO>",
	0x0F: "<SI>",
	0x10: "<DLE>",
	0x11: "<DC1>",
	0x12: "<DC2>",
	0x13: "<DC3>",
	0x14: "<DC4>",
	0x15: "<NAK>",
	0x16: "<SYN>",
	0x17: "<ETB>",
	0x18: "<CAN>",
	0x19: "<EM>",
	0x1A: "<SUB>",
	0x1B: "<ESC>",
	0x1C: "<FS>",
	0x1D: "<GS>",
	0x1E: "<RS>",
	0x1F: "<US>",
}

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
		outPutFileName string
	)

	command := cli.Command{
		Name:        "listen",
		Aliases:     nil,
		Usage:       "listen <port> <protocol [raw|lis1a1|stxetx|mllp] Default:raw> <maxcon Default:10> <proxy(optional): noproxy or haproxyv2>",
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

			flags := c.Args().Slice()

			isRaw := sliceContains(flags, "--raw")
			if isRaw {
				rawBytes = !rawBytes
			}

			showLinebreaks = sliceContains(flags, "--showLinebreaks")
			tcpHandler := NewTCPServerHandler(rawBytes, showLinebreaks, outPutFileName)
			tcpServer := bloodlabNet.CreateNewTCPServerInstance(listenPort, protocolImplementation, connectionType, maxConn)
			tcpServer.Run(tcpHandler)
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

type TCPServerHandler interface {
	DataReceived(session bloodlabNet.Session, fileData []byte, receiveTimestamp time.Time)
	Connected(con bloodlabNet.Session) error
	Disconnected(session bloodlabNet.Session)
	Error(session bloodlabNet.Session, typeOfError bloodlabNet.ErrorType, err error)
}

type tcpServerHandler struct {
	showRawBytes   bool
	showLineBreaks bool
	outPutFileName string
}

func NewTCPServerHandler(showRawBytes, showLineBreaks bool, outPutFileName string) TCPServerHandler {
	return &tcpServerHandler{
		showRawBytes:   showRawBytes,
		showLineBreaks: showLineBreaks,
		outPutFileName: outPutFileName,
	}
}
func (h *tcpServerHandler) Connected(session bloodlabNet.Session) error {
	// Is instrument whitelisted
	remoteAddress, err := session.RemoteAddress()
	if err != nil {
		println(fmt.Errorf("can not get remote address: %w", err))
		session.Close()
		return err
	}
	println(fmt.Sprintf("Client successfully conntected with IP: %s", remoteAddress))
	return nil
}

func (h *tcpServerHandler) DataReceived(session bloodlabNet.Session, fileData []byte, receiveTimestamp time.Time) {
	_, err := session.RemoteAddress()
	if err != nil {
		println(fmt.Errorf("can not get remote address: %w", err))
	}

	readAbleFile := make([]byte, 0)
	for _, fileByte := range fileData {
		if fileByte == CR && !h.showLineBreaks && h.showRawBytes {
			readAbleFile = append(readAbleFile, []byte("\n")...)
			continue
		}

		if h.showRawBytes {
			readAbleFile = append(readAbleFile, fileByte)
			continue
		}

		readableByte, ok := replaceAbleBytes[fileByte]
		if ok {
			readAbleFile = append(readAbleFile, []byte(readableByte)...)
		} else {
			readAbleFile = append(readAbleFile, fileByte)
		}

	}
	fmt.Fprintln(os.Stdout, string(readAbleFile))
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
