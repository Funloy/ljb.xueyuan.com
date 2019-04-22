package daemon

import (
	humanize "github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

//UsageState 数据使用状态
type UsageState struct {

	// 总容量
	Total string `json:"total"`
	// 可用容量
	Free string `json:"free"`
	// 已使用容量
	Used string `json:"used"`
	// 使用率
	UsedPercent string `json:"usedPercent"`
}

// QueryMemUsage 查询系统内存信息
func QueryMemUsage() *UsageState {
	vm, _ := mem.VirtualMemory()
	usage := &UsageState{
		Total:       humanize.Bytes(vm.Total),
		Free:        humanize.Bytes(vm.Available),
		Used:        humanize.Bytes(vm.Used),
		UsedPercent: humanize.FormatFloat("#.##", vm.UsedPercent) + "%",
	}
	return usage
}

// QueryDiskUsage 查询系统硬盘信息
func QueryDiskUsage() *UsageState {
	disk, _ := disk.Usage("/")
	usage := &UsageState{
		Total:       humanize.Bytes(disk.Total),
		Free:        humanize.Bytes(disk.Free),
		Used:        humanize.Bytes(disk.Used),
		UsedPercent: humanize.FormatFloat("#.##", disk.UsedPercent) + "%",
	}
	return usage
}
