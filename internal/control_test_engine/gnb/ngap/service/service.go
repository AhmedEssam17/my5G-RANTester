package service

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap"
	"syscall"
	"unsafe"

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

	// buf := rem.ToRawSockAddrBuf()
	// param := sctp.GetAddrsOld{
	//  AddrNum: int32(len(buf)),
	//  Addrs:   uintptr(uintptr(unsafe.Pointer(&buf[0]))),
	// }
	// optlen := unsafe.Sizeof(param)
	// _, _, err = conn.Getsockopt(sctp.SCTP_SOCKOPT_CONNECTX3, uintptr(unsafe.Pointer(&param)), uintptr(unsafe.Pointer(&optlen)))

	type PAddrParams struct {
		AssocID    int32                  // todo: how can we get associd for a peer ?
		Address    syscall.RawSockaddrAny // todo: correct the type
		HBInterval uint32
		PathMaxRxt uint16
		PathMTU    uint32
		SACKDelay  uint32
		Flags      uint32
	}
	peerAddrParamsOptions := PAddrParams{}
	peerAddrParamsOptionslen := unsafe.Sizeof(peerAddrParamsOptions)
	_, _, err = conn.Getsockopt(sctp.SCTP_PEER_ADDR_PARAMS, uintptr(unsafe.Pointer(&peerAddrParamsOptions)), uintptr(peerAddrParamsOptionslen))
	if err != nil {
		log.Info("SCTP Getsockopt SCTP_PEER_ADDR_PARAMS failed with error: ", err)
		return err
	}
	log.Info("peerAddrParamsOptions Before = ", peerAddrParamsOptions)
	peerAddrParamsOptions.HBInterval = 5
	_, _, err = conn.Setsockopt(sctp.SCTP_PEER_ADDR_PARAMS, uintptr(unsafe.Pointer(&peerAddrParamsOptions)), uintptr(peerAddrParamsOptionslen))
	if err != nil {
		log.Info("SCTP Setsockopt SCTP_PEER_ADDR_PARAMS failed with error: ", err)
		return err
	}
	log.Info("peerAddrParamsOptions After = ", peerAddrParamsOptions)

	type RTOInfo struct {
		AssocID    int32 // todo: how can we get associd for a peer ?
		RTOInitial uint32
		RTOMax     uint32
		RTOMin     uint32
	}
	rtoInfoOptions := RTOInfo{}
	rtoInfoOptionslen := unsafe.Sizeof(rtoInfoOptions)
	if _, _, err := conn.Getsockopt(sctp.SCTP_RTOINFO, uintptr(unsafe.Pointer(&rtoInfoOptions)), uintptr(unsafe.Pointer(&rtoInfoOptionslen))); err != nil {
		log.Info("SCTP Getsockopt SCTP_RTOINFO failed with error: ", err)
		return err
	}
	log.Info("rtoInfoOptions Before = ", rtoInfoOptions)
	rtoInfoOptions.RTOMax = 5000
	if _, _, err := conn.Setsockopt(sctp.SCTP_RTOINFO, uintptr(unsafe.Pointer(&rtoInfoOptions)), uintptr(rtoInfoOptionslen)); err != nil {
		log.Info("SCTP Setsockopt SCTP_RTOINFO failed with error: ", err)
		return err
	}
	log.Info("rtoInfoOptions After = ", rtoInfoOptions)
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
		n, info, err := conn.SCTPRead(buf[:])
		if err != nil {
			triggerGnbs <- 1
			log.Info("SCTPRead Error = ", err)
			break
		}

		if info != nil {
			log.Info("[GNB][SCTP] Receive message in ", info.Stream, " stream\n")
		}

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NGAP message.
		go ngap.Dispatch(amf, gnb, forwardData)
	}

}
