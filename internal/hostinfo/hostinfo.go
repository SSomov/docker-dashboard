package hostinfo

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	psutilNet "github.com/shirou/gopsutil/v4/net"
)

type SystemMetrics struct {
	CPU       []float64                  `json:"cpu"`
	Memory    *mem.VirtualMemoryStat     `json:"memory"`
	Disk      []disk.PartitionStat       `json:"disk_partitions"`
	DiskUsage map[string]*disk.UsageStat `json:"disk_usage"`
	Load      *load.AvgStat              `json:"load"`
	Host      *host.InfoStat             `json:"host"`
	Net       []psutilNet.IOCountersStat `json:"net"`
}

func GetSystemMetrics() (*SystemMetrics, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	diskUsage := make(map[string]*disk.UsageStat)
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err == nil {
			diskUsage[p.Mountpoint] = usage
		}
	}
	loadStat, err := load.Avg()
	if err != nil {
		return nil, err
	}
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}
	netIO, err := psutilNet.IOCounters(false)
	if err != nil {
		return nil, err
	}

	return &SystemMetrics{
		CPU:       cpuPercent,
		Memory:    memStat,
		Disk:      partitions,
		DiskUsage: diskUsage,
		Load:      loadStat,
		Host:      hostInfo,
		Net:       netIO,
	}, nil
}
