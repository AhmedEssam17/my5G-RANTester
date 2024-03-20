package templates

import (
	"math/rand"

	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"strconv"

	"my5G-RANTester/internal/control_test_engine/ue"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestMultiUesMultiGNBsSeq(numUes int, numGNBs int, i int) {

	log.Info("Num of UEs = ", numUes)
	log.Info("Num of GNBs = ", numGNBs)

	wg := sync.WaitGroup{}

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	if i != 0 {
		gnbIDInt, _ := strconv.Atoi(cfg.GNodeB.PlmnList.GnbId)
		gnbIDInt += numGNBs * i
		cfg.GNodeB.PlmnList.GnbId = fmt.Sprintf("%06d", gnbIDInt)

		msinInt, _ := strconv.Atoi(cfg.Ue.Msin)
		msinInt += numUes * i
		cfg.Ue.Msin = strconv.Itoa(msinInt)

		cfg.GNodeB.ControlIF.Port += 10 * i
	}

	gnbControlPort := cfg.GNodeB.ControlIF.Port
	gnbID, err := strconv.Atoi(string(cfg.GNodeB.PlmnList.GnbId))
	if err != nil {
		log.Error("Failed to extract gnbID")
	}

	baseGnbID := gnbID

	monitorGnbs := make(chan config.Config, numGNBs)

	go func(monitorGnbs chan config.Config) {
		for {
			select {
			case gnbCfg := <-monitorGnbs:
				go gnb.InitGnbMonitored(gnbCfg, &wg, monitorGnbs)
				wg.Add(1)
				time.Sleep(1 * time.Second)
			}
		}
	}(monitorGnbs)

	for i := 0; i < numGNBs; i++ {
		//newGnbID := fmt.Sprintf("%d", gnbID)
		newGnbID := constructGnbID(gnbID)

		cfg.GNodeB.PlmnList.GnbId = newGnbID
		cfg.GNodeB.ControlIF.Port = gnbControlPort + i
		log.Info("Initializing gnb with GnbId = ", cfg.GNodeB.PlmnList.GnbId)
		log.Info("Initializing gnb with gnbControlPort = ", cfg.GNodeB.ControlIF.Port)

		go gnb.InitGnbMonitored(cfg, &wg, monitorGnbs)
		wg.Add(1)
		time.Sleep(1 * time.Second)

		gnbID++
	}

	// time.Sleep(1 * time.Second)

	ueRegistrationSignal := make(chan int, 1)

	msin := cfg.Ue.Msin
	var ueSessionIdOffset uint8 = 0
	var ueSessionId uint8
	// startTime := time.Now()
	for i := 1; i <= numUes; i++ {

		offset := rand.Intn(numGNBs)
		cfg.GNodeB.ControlIF.Port = gnbControlPort + offset
		cfg.GNodeB.PlmnList.GnbId = constructGnbID(baseGnbID + offset)

		log.Info("Registering ue with gnbControlPort = ", cfg.GNodeB.ControlIF.Port)

		imsi := imsiGenerator(i, msin)
		log.Info("[TESTER] TESTING REGISTRATION USING IMSI ", imsi, " UE")
		cfg.Ue.Msin = imsi

		if (i+int(ueSessionIdOffset))%256 == 0 {
			ueSessionIdOffset++
		}
		ueSessionId = uint8(i) + ueSessionIdOffset

		go ue.RegistrationUe(cfg, ueSessionId, &wg, ueRegistrationSignal)
		wg.Add(1)

		select {
		case <-ueRegistrationSignal:
			log.Info("[TESTER] IMSI ", imsi, " UE REGISTERED OK")
		case <-time.After(60 * time.Second):
			log.Info("[TESTER] IMSI ", imsi, " UE REGISTER TIMEOUT")
		}

		sleepTime := 200 * time.Millisecond
		time.Sleep(sleepTime)
	}

	wg.Wait()
}

func constructGnbID(gnbID int) string {
	var newGnbID string

	if gnbID <= 9 {
		newGnbID = fmt.Sprintf("00000%d", gnbID)
	} else if gnbID > 9 && gnbID <= 99 {
		newGnbID = fmt.Sprintf("0000%d", gnbID)
	} else if gnbID > 99 && gnbID <= 999 {
		newGnbID = fmt.Sprintf("000%d", gnbID)
	}

	return newGnbID
}
