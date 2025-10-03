package system

import (
	"os"
	"runtime"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/sirupsen/logrus"
)

type PidHook struct {
	Pid int
}

func (h *PidHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *PidHook) Fire(entry *logrus.Entry) error {
	entry.Data["pid"] = h.Pid
	return nil
}

type MemStats struct {
	Alloc   uint64
	Percent float64
}

func GetMemStats() (MemStats, error) {

	// memoria usada por el heap del proceso
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// memAlloc := m.Alloc // bytes en uso por el heap

	// obtener memoria total del sistema
	vmStat, e := mem.VirtualMemory()
	if e != nil {
		return MemStats{}, e
	}

	// porcentaje de uso del proceso sobre la RAM total
	return MemStats{
		Alloc:   m.Alloc,
		Percent: (float64(m.Alloc) / float64(vmStat.Total)) * 100,
	}, nil

}

func GetCPUPercent() (float64, error) {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return 0, err
	}

	// devuelve porcentaje de CPU usado por el proceso
	return p.CPUPercent()
}
