package main

import (
	"fmt"
	bloodlabnet "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net"
	bloodlabnetProtocol "github.com/DRK-Blutspende-BaWueHe/go-bloodlab-net/protocol"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	app := &cli.App{
		Name:                 "Bnet helper",
		Usage:                "Transfer a file to a TCP Server over BloodlabNet library or listen as a TCP Server",
		EnableBashCompletion: true,
	}

	SendingCommand(app)
	ListeningCommand(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getLowLevelProtocol(protocol string, startByte, endByte string) (bloodlabnetProtocol.Implementation, error) {
	var protocolTypeImplementation bloodlabnetProtocol.Implementation
	switch protocol {
	case "raw":
		protocolTypeImplementation = bloodlabnetProtocol.Raw(bloodlabnetProtocol.DefaultRawProtocolSettings())
	case "stxetx":
		protocolTypeImplementation = bloodlabnetProtocol.STXETX(bloodlabnetProtocol.DefaultSTXETXProtocolSettings())
	case "lis1a1":
		protocolTypeImplementation = bloodlabnetProtocol.Lis1A1Protocol(bloodlabnetProtocol.DefaultLis1A1ProtocolSettings())
	case "mllp":
		config := bloodlabnetProtocol.DefaultMLLPProtocolSettings()
		if startByte != "" {
			startByteInt, err := strconv.Atoi(startByte)
			if err != nil {
				return nil, fmt.Errorf("invalid startbyte: %s, err: %s", startByte, err.Error())
			}
			config = config.SetStartByte(byte(startByteInt))
		}

		if endByte != "" {
			endByteInt, err := strconv.Atoi(endByte)
			if err != nil {
				return nil, fmt.Errorf("invalid endbyte: %s, err: %s", endByteInt, err.Error())
			}
			config = config.SetEndByte(byte(endByteInt))
		}
		protocolTypeImplementation = bloodlabnetProtocol.MLLP(config)
	default:
		return nil, fmt.Errorf("invalid protocol type given: %s , supported:raw,stxetx,lis1a1,mllp", protocol)

	}

	return protocolTypeImplementation, nil
}

func getProxyCoonnectionType(proxy string) (bloodlabnet.ConnectionType, error) {
	switch strings.ToLower(proxy) {
	case "noproxy":
		return bloodlabnet.NoLoadBalancer, nil
	case "haproxyv2":
		return bloodlabnet.HAProxySendProxyV2, nil
	case "":
		return bloodlabnet.NoLoadBalancer, nil
	default:
		return bloodlabnet.NoLoadBalancer, fmt.Errorf("invalid proxy connection type. valid values: 'noproxy', 'haproxyv2' given: %s", strings.ToLower(proxy))
	}
}

func logOutput(fileName string) func() {
	outFile := fileName
	// open file read/write | create if not exist | clear file at open if exists
	f, _ := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	// save existing stdout | MultiWriter writes to saved stdout and file
	out := os.Stdout
	mw := io.MultiWriter(out, f)

	// get pipe reader and writer | writes to pipe writer come out pipe reader
	r, w, _ := os.Pipe()

	// replace stdout,stderr with pipe writer | all writes to stdout, stderr will go through pipe instead (fmt.print, log)
	os.Stdout = w
	os.Stderr = w

	// writes with log.Print should also write to mw
	log.SetOutput(mw)

	//create channel to control exit | will block until all copies are finished
	exit := make(chan bool)

	go func() {
		// copy all reads from pipe to multiwriter, which writes to stdout and file
		_, _ = io.Copy(mw, r)
		// when r or w is closed copy will finish and true will be sent to channel
		exit <- true
	}()

	// function to be deferred in main until program exits
	return func() {
		// close writer then block on exit channel | this will let mw finish writing before the program exits
		_ = w.Close()
		<-exit
		// close file after all writes have finished
		_ = f.Close()
	}

}
