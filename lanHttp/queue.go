package lanHttp

import (
	"github.com/shockerli/cvt"
	"sync"
)

type DownLoadQueue struct {
	Queue chan Task
	mutex sync.Mutex
}

func NewDownLoadQueue(maxQueueLen int) (dc *DownLoadQueue) {
	dc = &DownLoadQueue{}
	dc.Queue = make(chan Task, maxQueueLen)
	return dc
}

func (dc *DownLoadQueue) Push(data Task) bool {
	select {
	case dc.Queue <- data:
		return true
	default:
		return false
	}
}

func (dc *DownLoadQueue) Pop() (data Task, ret bool) {
	select {
	case data = <-dc.Queue:
		return data, true
	default:
		return data, false
	}
}

func (dc *DownLoadQueue) Clear() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	for {
		if len(dc.Queue) > 0 {
			select {
			case <-dc.Queue:
			}
		} else {
			break
		}
	}
}

func (dc *DownLoadQueue) Length() uint16 {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	return cvt.Uint16(len(dc.Queue))
}
