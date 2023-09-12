package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	bloodlabNet "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net"
)

type TCPServerHandler interface {
	DataReceived(session bloodlabNet.Session, fileData []byte, receiveTimestamp time.Time)
	Connected(con bloodlabNet.Session) error
	Disconnected(session bloodlabNet.Session)
	Error(session bloodlabNet.Session, typeOfError bloodlabNet.ErrorType, err error)
	GetOutPutFileName() string
	DeleteOutPutFileName()
	DataReceivedAndisconnected() bool
}

type tcpServerHandler struct {
	showRawBytes     bool
	showLineBreaks   bool
	outPutFileName   string
	createOutPutFile bool
	outPutFileMask   string
	dataReceived     bool
	disconnected     bool
}

func NewTCPServerHandler(showRawBytes, showLineBreaks bool, createOutPutFile bool, outPutFileMask string) TCPServerHandler {
	return &tcpServerHandler{
		showRawBytes:     showRawBytes,
		showLineBreaks:   showLineBreaks,
		createOutPutFile: createOutPutFile,
		outPutFileMask:   outPutFileMask,
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

	if h.createOutPutFile && len(h.outPutFileName) == 0 {
		h.outPutFileName = fmt.Sprintf(h.outPutFileMask, time.Now().Format("20060102_150405"))
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
	h.dataReceived = true
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

func (h *tcpServerHandler) GetOutPutFileName() string {
	return h.outPutFileName
}

func (h *tcpServerHandler) DeleteOutPutFileName() {
	h.outPutFileName = ""
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

func (h *tcpServerHandler) Error(session bloodlabNet.Session, typeOfError bloodlabNet.ErrorType, err error) {
	println(fmt.Errorf("error: %w", err))
}
