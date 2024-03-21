package templates

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/ue"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestAttachUeWithConfiguration() {

	wg := sync.WaitGroup{}

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	go gnb.InitGnb(cfg, &wg)

	wg.Add(1)

	time.Sleep(1 * time.Second)
	ueRegistrationSignal := make(chan int, 1)
	cfg.Ue.UeSessionId = 1
	go ue.RegistrationUe(cfg, &wg, ueRegistrationSignal)

	wg.Add(1)

	wg.Wait()
}
