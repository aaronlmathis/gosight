package runner

/*
type BackgroundRunner struct {
	sys *SystemContext
	wg  sync.WaitGroup
}

func NewBackgroundRunner(sys *SystemContext) *BackgroundRunner {
	return &BackgroundRunner{sys: sys}
}

func (b *BackgroundRunner) Start() {
	// each worker runs independently
	b.wg.Add(1)
	go b.runAlertSweep()

	b.wg.Add(1)
	go b.runRuleReloader()
}

func (b *BackgroundRunner) runAlertSweep() {
	defer b.wg.Done()
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		b.sys.Tele.Evaluator.Sweep()
	}
}

func (b *BackgroundRunner) runRuleReloader() {
	defer b.wg.Done()
	for {
		if updated := checkForRuleFileChange(); updated {
			newRules := LoadRulesFromYAML(...)
			b.sys.Tele.Evaluator.Reload(newRules)
		}
		time.Sleep(30 * time.Second)
	}
}
*/
