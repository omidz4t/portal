package bot

// dcWorkers / dcQueue bound OnNewMsg work so a flood cannot spawn a goroutine
// per message. Enqueue is non-blocking so Bot.Run never waits on handlers.
const (
	dcWorkers = 8
	dcQueue   = 64
)

type dcJob struct {
	accID uint32
	msgID uint32
}

type dcWorkPool struct {
	jobs chan dcJob
}

func newDCWorkPool(workers, queue int, handle func(accID, msgID uint32)) *dcWorkPool {
	if workers < 1 {
		workers = 1
	}
	if queue < 1 {
		queue = 1
	}
	p := &dcWorkPool{jobs: make(chan dcJob, queue)}
	for i := 0; i < workers; i++ {
		go func() {
			for job := range p.jobs {
				func() {
					defer func() { _ = recover() }()
					handle(job.accID, job.msgID)
				}()
			}
		}()
	}
	return p
}

// tryEnqueue reports whether the job was accepted. False means the queue is full.
func (p *dcWorkPool) tryEnqueue(accID, msgID uint32) bool {
	if p == nil {
		return false
	}
	select {
	case p.jobs <- dcJob{accID: accID, msgID: msgID}:
		return true
	default:
		return false
	}
}
