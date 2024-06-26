package service

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/nas"
	"net"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func InitServer(gnb *context.GNBContext) error {

	// initiated GNB server with unix sockets.
	
	gnbID, err := strconv.Atoi(string(gnb.GetGnbId()))
	sockPath := fmt.Sprintf("/tmp/gnb%d.sock", gnbID)

	os.Remove(sockPath)
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		fmt.Errorf("Listen error: ", err)
	}

	gnb.SetListener(ln)

	/*
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
		go func(ln net.Listener, c chan os.Signal) {
			sig := <-c
			log.Printf("Caught signal %s: shutting down.", sig)
			ln.Close()
			os.Exit(0)
		}(ln, sigc)
	*/

	go gnbListen(gnb)

	return nil
}

func gnbListen(gnb *context.GNBContext) {

	// log.Info("Before ln := gnb.GetListener()")
	ln := gnb.GetListener()
	if ln == nil {
		log.Error("Listener is nil. Not initialized Correctly.")
		return
	}
	// log.Info("After ln := gnb.GetListener() ln: ", ln)

	for {

		// log.Info("Inside nas/service for loop")

		// log.Info("Before fd, err := ln.Accept()")
		fd, err := ln.Accept()
		// log.Info("After fd, err := ln.Accept() fd: ", fd)

		if err != nil {
			log.Info("[GNB][UE] Accept error: ", err)
			break
		}

		// TODO this region of the code may induces race condition.

		// new instance GNB UE context
		// store UE in UE Pool
		// store UE connection
		// select AMF and get sctp association
		// make a tun interface
		// log.Info("Before ue := gnb.NewGnBUe(fd)")
		ue := gnb.NewGnBUe(fd)
		// log.Info("After ue := gnb.NewGnBUe(fd)")
		if ue == nil {
			break
		}

		// accept and handle connection.
		// log.Info("Before go processingConn(ue, gnb)")
		go processingConn(ue, gnb)
	}

}

func processingConn(ue *context.GNBUe, gnb *context.GNBContext) {

	buf := make([]byte, 65535)

	// log.Info("Before conn := ue.GetUnixSocket()")
	conn := ue.GetUnixSocket()
	// log.Info("After conn := ue.GetUnixSocket() Conn: ", conn)

	for {

		n, err := conn.Read(buf[:])
		if err != nil {
			return
		}

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// send to dispatch.
		// log.Info("Before go nas.Dispatch(ue, forwardData, gnb)")
		go nas.Dispatch(ue, forwardData, gnb)
	}
}
