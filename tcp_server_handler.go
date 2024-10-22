package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	bloodlabNet "github.com/blutspende/go-bloodlab-net"
	bloodlabnet "github.com/blutspende/go-bloodlab-net"
	bloodlabnetProtocol "github.com/blutspende/go-bloodlab-net/protocol"
)

type TCPServerHandler interface {
	DataReceived(session bloodlabNet.Session, fileData []byte, receiveTimestamp time.Time) error
	Connected(con bloodlabNet.Session) error
	Disconnected(session bloodlabNet.Session)
	Error(session bloodlabNet.Session, typeOfError bloodlabNet.ErrorType, err error)
	DataReceivedAndisconnected() bool
}

type tcpServerHandler struct {
	showRawBytes           bool
	showLineBreaks         bool
	outPutFileName         string
	sendBackHost           string
	sendBackFile           string
	dataReceived           bool
	disconnected           bool
	protocolImplementation bloodlabnetProtocol.Implementation
}

func NewTCPServerHandler(showRawBytes, showLineBreaks bool, outPutFileName string, sendBackHost string, sendBackFile string, protocolImplementation bloodlabnetProtocol.Implementation) TCPServerHandler {
	return &tcpServerHandler{
		showRawBytes:           showRawBytes,
		showLineBreaks:         showLineBreaks,
		outPutFileName:         outPutFileName,
		sendBackHost:           sendBackHost,
		sendBackFile:           sendBackFile,
		protocolImplementation: protocolImplementation,
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

func (h *tcpServerHandler) DataReceived(session bloodlabNet.Session, fileData []byte, receiveTimestamp time.Time) error {
	_, err := session.RemoteAddress()
	if err != nil {
		println(fmt.Errorf("can not get remote address: %w", err))
		return err
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
	if len(h.sendBackHost) > 0 {
		h.sendBack(readAbleFile)
	}
	fmt.Fprintln(os.Stdout, string(readAbleFile))
	h.dataReceived = true

	return nil
}

func (h *tcpServerHandler) Disconnected(session bloodlabNet.Session) {
	// Is instrument whitelisted
	remoteAddress, err := session.RemoteAddress()
	if err != nil {
		println(fmt.Errorf("can not get remote address: %w", err))
	}

	println(fmt.Sprintf("Client with %s is disconnected", remoteAddress))

	h.disconnected = true
}

func (h *tcpServerHandler) DataReceivedAndisconnected() bool {
	if h.dataReceived && h.disconnected {
		return true
	}
	return false
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
	logFile.Close()

	println(fmt.Sprintf("Create log file %s", h.outPutFileName))
}

func (h *tcpServerHandler) sendBack(logData []byte) {

	var fileLines = make([][]byte, 0)

	sendHost, sendPort, err := net.SplitHostPort(h.sendBackHost)
	if err != nil {
		println(fmt.Errorf("invalid device: %s err: %s", h.sendBackHost, err.Error()))
		return
	}

	sendPortInt, err := strconv.Atoi(sendPort)
	if err != nil {
		println(fmt.Errorf("invalid port in hostname: %s err: %s", sendPort, err.Error()))
		return
	}

	if len(h.sendBackFile) > 0 {
		file, err := os.Open(h.sendBackFile)
		if err != nil {
			println(fmt.Errorf("failed to open file: %s error: %s", h.sendBackFile, err.Error()))
			return
		}
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			fileLines = append(fileLines, []byte(scanner.Text()))
		}
		file.Close()
	} else {
		fileLines = append(fileLines, logData)
	}

	tcpClient := bloodlabnet.CreateNewTCPClient(sendHost, sendPortInt, h.protocolImplementation, bloodlabnet.NoLoadBalancer, bloodlabnet.DefaultTCPClientSettings)
	err = tcpClient.Connect()
	if err != nil {
		println(fmt.Errorf("cannot connect to host (%s): %s", h.sendBackHost, err.Error()))
		return
	}

	n, err := tcpClient.Send(fileLines)
	if err != nil {
		println(fmt.Errorf("failed to send file to host: %s", err.Error()))
		return
	}
	time.Sleep(time.Second * 5)

	if n <= 0 {
		println("No data was sent by the client")
	} else {
		println("Successfully sent data")
	}

	tcpClient.Close()
}

func (h *tcpServerHandler) Error(session bloodlabNet.Session, typeOfError bloodlabNet.ErrorType, err error) {
	println(fmt.Errorf("error: %w", err))
}
