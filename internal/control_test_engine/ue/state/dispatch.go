package state

import (
	"my5G-RANTester/internal/control_test_engine/ue/context"
	data "my5G-RANTester/internal/control_test_engine/ue/data/service"
	"my5G-RANTester/internal/control_test_engine/ue/nas"
	"sync"
)

var m sync.Mutex

func DispatchState(ue *context.UEContext, message []byte, ueRegistrationSignal chan int, ueTerminationSignal chan int) {
	m.Lock()
	defer m.Unlock()
	// if state is PDU session inactive send to analyze NAS
	switch ue.GetStateSM() {

	case context.SM5G_PDU_SESSION_INACTIVE:
		nas.DispatchNas(ue, message)
	case context.SM5G_PDU_SESSION_ACTIVE_PENDING:
		nas.DispatchNas(ue, message)
	case context.SM5G_PDU_SESSION_ACTIVE:
		data.InitDataPlane(ue, message, ueRegistrationSignal, ueTerminationSignal)
	}
}
