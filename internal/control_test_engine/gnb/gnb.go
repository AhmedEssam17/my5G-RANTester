package gnb

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/nas/message/sender"
	serviceNas "my5G-RANTester/internal/control_test_engine/gnb/nas/service"
	serviceNgap "my5G-RANTester/internal/control_test_engine/gnb/ngap/service"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"my5G-RANTester/internal/monitoring"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func InitGnbMonitored(conf config.Config, wg *sync.WaitGroup, monitorGnbs chan config.Config) {

	// instance new gnb.
	gnb := &context.GNBContext{}

	// new gnb context.
	gnb.NewRanGnbContext(
		conf.GNodeB.PlmnList.GnbId,
		conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc,
		conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.Sst,
		conf.GNodeB.SliceSupportList.Sd,
		conf.GNodeB.ControlIF.Ip,
		conf.GNodeB.DataIF.Ip,
		conf.GNodeB.ControlIF.Port,
		conf.GNodeB.DataIF.Port)

	// new AMF context.
	amf := gnb.NewGnBAmf(conf.AMF.Ip, conf.AMF.Port)

	triggerGnbs := make(chan int, 1)

	go func(triggerGnbs chan int, monitorGnbs chan config.Config) {
		for {
			select {
			case <-triggerGnbs:
				ranIdRange := gnb.IdUeGenerator
				var ranUeId int64 = 1
				for ranUeId = 1; ranUeId < ranIdRange; ranUeId++ {
					ue, err := gnb.GetGnbUe(ranUeId)
					if err != nil || ue == nil {
						log.Error("Failed to load UE")
					}
					sender.SendToUe(ue, []byte("TERMINATE"))
				}

				gnb.Terminate()
				monitorGnbs <- conf
				wg.Done()
				return
			}
		}
	}(triggerGnbs, monitorGnbs)

	// start communication with AMF(SCTP).
	if err := serviceNgap.InitConnMonitored(amf, gnb, triggerGnbs); err != nil {
		log.Fatal("Error in", err)
	} else {
		log.Info("[GNB] SCTP/NGAP service is running")
		// wg.Add(1)
	}

	// start communication with UE (server UNIX sockets).
	if err := serviceNas.InitServer(gnb, triggerGnbs); err != nil {
		log.Fatal("Error in ", err)
	} else {
		log.Info("[GNB] UNIX/NAS service is running")
	}

	trigger.SendNgSetupRequest(gnb, amf)

	// control the signals
	sigGnb := make(chan os.Signal, 1)
	signal.Notify(sigGnb, os.Interrupt)

	// Block until a signal is received.
	<-sigGnb
	gnb.Terminate()
	monitorGnbs <- conf
	wg.Done()
	// os.Exit(0)

}

func InitGnb(conf config.Config, wg *sync.WaitGroup) {

	// instance new gnb.
	gnb := &context.GNBContext{}

	// new gnb context.
	gnb.NewRanGnbContext(
		conf.GNodeB.PlmnList.GnbId,
		conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc,
		conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.Sst,
		conf.GNodeB.SliceSupportList.Sd,
		conf.GNodeB.ControlIF.Ip,
		conf.GNodeB.DataIF.Ip,
		conf.GNodeB.ControlIF.Port,
		conf.GNodeB.DataIF.Port)

	// start communication with AMF (server SCTP).

	// log.Info("conf.GNodeB.PlmnList.GnbId ", conf.GNodeB.PlmnList.GnbId)
	// log.Info("conf.GNodeB.PlmnList.Mcc ", conf.GNodeB.PlmnList.Mcc)
	// log.Info("conf.GNodeB.PlmnList.Mnc ", conf.GNodeB.PlmnList.Mnc)
	// log.Info("conf.GNodeB.PlmnList.Tac ", conf.GNodeB.PlmnList.Tac)
	// log.Info("conf.GNodeB.SliceSupportList.Sst ", conf.GNodeB.SliceSupportList.Sst)
	// log.Info("conf.GNodeB.SliceSupportList.Sd ", conf.GNodeB.SliceSupportList.Sd)
	// log.Info("conf.GNodeB.ControlIF.Ip ", conf.GNodeB.ControlIF.Ip)
	// log.Info("conf.GNodeB.DataIF.Ip ", conf.GNodeB.DataIF.Ip)
	// log.Info("conf.GNodeB.ControlIF.Port ", conf.GNodeB.ControlIF.Port)
	// log.Info("conf.GNodeB.DataIF.Port ", conf.GNodeB.DataIF.Port)

	// new AMF context.
	amf := gnb.NewGnBAmf(conf.AMF.Ip, conf.AMF.Port)

	// start communication with AMF(SCTP).
	if err := serviceNgap.InitConn(amf, gnb); err != nil {
		log.Fatal("Error in", err)
	} else {
		log.Info("[GNB] SCTP/NGAP service is running")
		// wg.Add(1)
	}

	triggerGnbs := make(chan int, 1)
	// start communication with UE (server UNIX sockets).
	if err := serviceNas.InitServer(gnb, triggerGnbs); err != nil {
		log.Fatal("Error in ", err)
	} else {
		log.Info("[GNB] UNIX/NAS service is running")
	}

	trigger.SendNgSetupRequest(gnb, amf)

	// control the signals
	sigGnb := make(chan os.Signal, 1)
	signal.Notify(sigGnb, os.Interrupt)

	// Block until a signal is received.
	<-sigGnb
	gnb.Terminate()
	wg.Done()
	// os.Exit(0)

}

func InitGnbForUeLatency(conf config.Config, sigGnb chan bool, synch chan bool) {

	// instance new gnb.
	gnb := &context.GNBContext{}

	// new gnb context.
	gnb.NewRanGnbContext(
		conf.GNodeB.PlmnList.GnbId,
		conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc,
		conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.Sst,
		conf.GNodeB.SliceSupportList.Sd,
		conf.GNodeB.ControlIF.Ip,
		conf.GNodeB.DataIF.Ip,
		conf.GNodeB.ControlIF.Port,
		conf.GNodeB.DataIF.Port)

	// start communication with AMF (server SCTP).

	// new AMF context.
	amf := gnb.NewGnBAmf(conf.AMF.Ip, conf.AMF.Port)

	// start communication with AMF(SCTP).
	if err := serviceNgap.InitConn(amf, gnb); err != nil {
		log.Info("Error in", err)

		synch <- false

		return
	} else {
		log.Info("[GNB] SCTP/NGAP service is running")
		// wg.Add(1)
	}

	triggerGnbs := make(chan int, 1)
	// start communication with UE (server UNIX sockets).
	if err := serviceNas.InitServer(gnb, triggerGnbs); err != nil {
		log.Info("Error in", err)

		synch <- false
	} else {
		log.Info("[GNB] UNIX/NAS service is running")

	}

	trigger.SendNgSetupRequest(gnb, amf)

	synch <- true

	// Block until a signal is received.
	<-sigGnb
	gnb.Terminate()
}

func InitGnbForLoadSeconds(conf config.Config, wg *sync.WaitGroup,
	monitor *monitoring.Monitor) {

	// instance new gnb.
	gnb := &context.GNBContext{}

	// new gnb context.
	gnb.NewRanGnbContext(
		conf.GNodeB.PlmnList.GnbId,
		conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc,
		conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.Sst,
		conf.GNodeB.SliceSupportList.Sd,
		conf.GNodeB.ControlIF.Ip,
		conf.GNodeB.DataIF.Ip,
		conf.GNodeB.ControlIF.Port,
		conf.GNodeB.DataIF.Port)

	// start communication with AMF (server SCTP).

	// new AMF context.
	amf := gnb.NewGnBAmf(conf.AMF.Ip, conf.AMF.Port)

	// start communication with AMF(SCTP).
	if err := serviceNgap.InitConn(amf, gnb); err != nil {
		log.Info("Error in ", err)

		time.Sleep(1000 * time.Millisecond)

		wg.Done()

		return
	} else {
		log.Info("[GNB] SCTP/NGAP service is running")
		// wg.Add(1)
	}

	trigger.SendNgSetupRequest(gnb, amf)

	// timeout is 1 second for receive NG Setup Response
	time.Sleep(1000 * time.Millisecond)

	// AMF responds message sends by Tester
	// means AMF is available
	if amf.GetState() == 0x01 {
		monitor.IncRqs()
	}

	gnb.Terminate()
	wg.Done()
	// os.Exit(0)
}

func InitGnbForAvaibility(conf config.Config,
	monitor *monitoring.Monitor) {

	// instance new gnb.
	gnb := &context.GNBContext{}

	// new gnb context.
	gnb.NewRanGnbContext(
		conf.GNodeB.PlmnList.GnbId,
		conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc,
		conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.Sst,
		conf.GNodeB.SliceSupportList.Sd,
		conf.GNodeB.ControlIF.Ip,
		conf.GNodeB.DataIF.Ip,
		conf.GNodeB.ControlIF.Port,
		conf.GNodeB.DataIF.Port)

	// start communication with AMF (server SCTP).

	// new AMF context.
	amf := gnb.NewGnBAmf(conf.AMF.Ip, conf.AMF.Port)

	// start communication with AMF(SCTP).
	if err := serviceNgap.InitConn(amf, gnb); err != nil {
		log.Info("Error in ", err)

		return

	} else {
		log.Info("[GNB] SCTP/NGAP service is running")

	}

	trigger.SendNgSetupRequest(gnb, amf)

	// timeout is 1 second for receive NG Setup Response
	time.Sleep(1000 * time.Millisecond)

	// AMF responds message sends by Tester
	// means AMF is available
	if amf.GetState() == 0x01 {
		monitor.IncAvaibility()

	}

	gnb.Terminate()
	// os.Exit(0)
}
