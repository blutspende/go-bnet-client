package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	bloodlabNet "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net"
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
		outPutFileName string
	)

	command := cli.Command{
		Name:    "listen",
		Aliases: nil,
		Usage: `listen for incomming message
		cli args -> listen <port> <protocol [raw|lis1a1|stxetx|mllp] Default:raw> <maxcon Default:10> <proxy: noproxy or haproxyv2> <logfile <optional>:  yes or no>`,
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
			if strings.ToLower(logfile) == "yes" {
				outPutFileName = fmt.Sprintf("log_%s.dat", time.Now().Format("20060102_150405"))
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

		readableByte, ok := ReplaceAbleBytes[fileByte]
		if ok {
			readAbleFile = append(readAbleFile, []byte(readableByte)...)
		} else {
			readAbleFile = append(readAbleFile, fileByte)
		}

	}

	if len(h.outPutFileName) > 0 {
		h.writeLogFile(string(readAbleFile))
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

func (h *tcpServerHandler) writeLogFile(logData string) {

	logFile, err := os.OpenFile(h.outPutFileName, os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		println(fmt.Errorf("can not open file: %w", err))
	}
	defer logFile.Close()
	logFileWriter := bufio.NewWriter(logFile)
	logDataLines := strings.Split(logData, "<CR>")
	for _, logDataLine := range logDataLines {
		_, err = logFileWriter.WriteString(logDataLine + "\n")
		if err != nil {
			println(fmt.Errorf("can not write to file: %w", err))
			return
		}
	}
	logFileWriter.Flush()
}

func (h *tcpServerHandler) Error(session bloodlabNet.Session, typeOfError bloodlabNet.ErrorType, err error) {
	println(fmt.Errorf("error: %w", err))
}
