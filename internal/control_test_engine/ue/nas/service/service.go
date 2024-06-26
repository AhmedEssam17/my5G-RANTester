package service

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/state"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/common/log"
)

const SM5G_PDU_SESSION_ACTIVE = 0x08

func CloseConn(ue *context.UEContext) {
	conn := ue.GetUnixConn()
	conn.Close()
}

func InitConn(ue *context.UEContext, ueRegistrationSignal chan int, ueTerminationSignal chan int) error {

	// initiated communication with GNB(unix sockets).
	gnbID, err := strconv.Atoi(string(ue.GetGnbId()))
	sockPath := fmt.Sprintf("/tmp/gnb%d.sock", gnbID)

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return fmt.Errorf("[UE] Error on Dial with server", err)
	}

	// store unix socket connection in the UE.
	ue.SetUnixConn(conn)

	// listen NAS.
	go UeListen(ue, ueRegistrationSignal, ueTerminationSignal)

	return nil
}

// ue listen unix sockets.
func UeListen(ue *context.UEContext, ueRegistrationSignal chan int, ueTerminationSignal chan int) {

	buf := make([]byte, 65535)
	conn := ue.GetUnixConn()

	
	// defer func() {
	// 	err := conn.Close()
	// 	log.Warn("*****Connection closed with UE-imsi = ", ue.GetMsin())
	// 	if err != nil {
	// 		fmt.Printf("Error in closing unix sockets for %s ue\n", ue.GetSupi())
	// 	}
	// }()
	
	
	for {

		// read message.
		if ue.GetStateSM() == SM5G_PDU_SESSION_ACTIVE{
			conn.SetReadDeadline(time.Time{})
		} else {
			timeoutDuration := 10 * time.Second
			conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		}
		n, err := conn.Read(buf[:])
		if err != nil {
			ueTerminationSignal <- 1
			log.Warn("****Read Error on ue = ", ue.GetMsin())
			break
		}

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NAS message.
		go state.DispatchState(ue, forwardData)
	}
}
