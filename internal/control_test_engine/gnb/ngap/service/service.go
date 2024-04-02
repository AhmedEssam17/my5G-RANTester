package service

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap"

	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
)

func InitConn(amf *context.GNBAmf, gnb *context.GNBContext) error {

	// check AMF IP and AMF port.
	remote := fmt.Sprintf("%s:%d", amf.GetAmfIp(), amf.GetAmfPort())
	local := fmt.Sprintf("%s:%d", gnb.GetGnbIp(), gnb.GetGnbPort())

	// log.Info("Remote address: ", remote)
	// log.Info("Local address: ", local)

	rem, err := sctp.ResolveSCTPAddr("sctp", remote)
	if err != nil {
		// log.Info("sctp remote error: ", err)
		return err
	}
	loc, err := sctp.ResolveSCTPAddr("sctp", local)
	if err != nil {
		// log.Info("sctp local error: ", err)
		return err
	}

	// streams := amf.GetTNLAStreams()

	// log.Info("before conn")

	conn, err := sctp.DialSCTPExt(
		"sctp",
		loc,
		rem,
		sctp.InitMsg{NumOstreams: 2, MaxInstreams: 2})
	if err != nil {
		// log.Info("conn error", err)
		amf.SetSCTPConn(nil)
		return err
	}

	// set streams and other information about TNLA

	// successful established SCTP (TNLA - N2)
	amf.SetSCTPConn(conn)
	gnb.SetN2(conn)

	conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO | sctp.SCTP_EVENT_SHUTDOWN)

	go GnbListen(amf, gnb)

	return nil
}

func GnbListen(amf *context.GNBAmf, gnb *context.GNBContext) {

	buf := make([]byte, 65535)
	conn := amf.GetSCTPConn()

	/*
		defer func() {
			err := conn.Close()
			if err != nil {
				log.Info("[GNB][SCTP] Error in closing SCTP association for %d AMF\n", amf.GetAmfId())
			}
		}()
	*/

	for {
		n, info, err := conn.SCTPRead(buf[:])
		if err != nil {
			// log.Info("SCTPRead Error ", err)
			break
		}

		log.Info("[GNB][SCTP] Receive message in ", info.Stream, " stream\n")

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NGAP message.
		go ngap.Dispatch(amf, gnb, forwardData)
	}
}

func InitConnMonitored(amf *context.GNBAmf, gnb *context.GNBContext, triggerGnbs chan int) error {

	// check AMF IP and AMF port.
	remote := fmt.Sprintf("%s:%d", amf.GetAmfIp(), amf.GetAmfPort())
	local := fmt.Sprintf("%s:%d", gnb.GetGnbIp(), gnb.GetGnbPort())

	rem, err := sctp.ResolveSCTPAddr("sctp", remote)
	if err != nil {
		// log.Info("sctp remote error: ", err)
		return err
	}
	loc, err := sctp.ResolveSCTPAddr("sctp", local)
	if err != nil {
		// log.Info("sctp local error: ", err)
		return err
	}

	conn, err := sctp.DialSCTPExt(
		"sctp",
		loc,
		rem,
		sctp.InitMsg{NumOstreams: 2, MaxInstreams: 2})
	if err != nil {
		// log.Info("conn error", err)
		amf.SetSCTPConn(nil)
		return err
	}

	// set streams and other information about TNLA

	// successful established SCTP (TNLA - N2)
	amf.SetSCTPConn(conn)
	gnb.SetN2(conn)

	conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO | sctp.SCTP_EVENT_SHUTDOWN | sctp.SCTP_EVENT_PEER_ERROR | sctp.SCTP_EVENT_SEND_FAILURE)

	go GnbListenMonitored(amf, gnb, triggerGnbs)

	return nil
}

func GnbListenMonitored(amf *context.GNBAmf, gnb *context.GNBContext, triggerGnbs chan int) {

	buf := make([]byte, 65535)
	conn := amf.GetSCTPConn()

	/*
		defer func() {
			err := conn.Close()
			if err != nil {
				log.Info("[GNB][SCTP] Error in closing SCTP association for %d AMF\n", amf.GetAmfId())
			}
		}()
	*/

	for {
		n, _, err := conn.SCTPRead(buf[:])
		if err != nil {
			triggerGnbs <- 1
			log.Info("SCTPRead Error = ", err)
			break
		}

		// log.Info("[GNB][SCTP] Receive message in ", info.Stream, " stream\n")

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NGAP message.
		go ngap.Dispatch(amf, gnb, forwardData)
	}

}
