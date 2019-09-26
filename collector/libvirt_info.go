// +build libvirt

// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"encoding/xml"
	"github.com/libvirt/libvirt-go"
	"github.com/prometheus/client_golang/prometheus"
	"log"
)

// LibvirtExporter implements a Prometheus exporter for libvirt state.
type LibvirtExporter struct {
	uri                string
	exportNovaMetadata bool

	libvirtUpDesc *prometheus.Desc

	libvirtDomainActive *prometheus.Desc
	libvirtDomainTotal  *prometheus.Desc

	// domain info
	libvirtDomainInfoMaxMemDesc    *prometheus.Desc
	libvirtDomainInfoMemoryDesc    *prometheus.Desc
	libvirtDomainInfoNrVirtCpuDesc *prometheus.Desc
	libvirtDomainInfoCpuTimeDesc   *prometheus.Desc
	libvirtDomainInfoDomainState   *prometheus.Desc

	//domain cpu info
	libvirtDomainCpuCpuTime    *prometheus.Desc
	libvirtDomainCpuUserTime   *prometheus.Desc
	libvirtDomainCpuSystemTime *prometheus.Desc
	libvirtDomainCpuVcpuTime   *prometheus.Desc

	//domain mem info
	libvirtDomainMemUnused     *prometheus.Desc
	libvirtDomainMemAvailable  *prometheus.Desc
	libvirtDomainMemUsable     *prometheus.Desc
	libvirtDomainMemRss        *prometheus.Desc
	libvirtDomainMemLastUpdate *prometheus.Desc

	//domain block info
	libvirtDomainBlockCapacity            *prometheus.Desc
	libvirtDomainBlockAllocation          *prometheus.Desc
	libvirtDomainBlockPhysical            *prometheus.Desc
	libvirtDomainBlockRdBytesDesc         *prometheus.Desc
	libvirtDomainBlockRdReqDesc           *prometheus.Desc
	libvirtDomainBlockRdTotalTimesDesc    *prometheus.Desc
	libvirtDomainBlockWrBytesDesc         *prometheus.Desc
	libvirtDomainBlockWrReqDesc           *prometheus.Desc
	libvirtDomainBlockWrTotalTimesDesc    *prometheus.Desc
	libvirtDomainBlockFlushReqDesc        *prometheus.Desc
	libvirtDomainBlockFlushTotalTimesDesc *prometheus.Desc

	// domain interface info
	libvirtDomainInterfaceAddresses     *prometheus.Desc
	libvirtDomainInterfaceRxBytesDesc   *prometheus.Desc
	libvirtDomainInterfaceRxPacketsDesc *prometheus.Desc
	libvirtDomainInterfaceRxErrsDesc    *prometheus.Desc
	libvirtDomainInterfaceRxDropDesc    *prometheus.Desc
	libvirtDomainInterfaceTxBytesDesc   *prometheus.Desc
	libvirtDomainInterfaceTxPacketsDesc *prometheus.Desc
	libvirtDomainInterfaceTxErrsDesc    *prometheus.Desc
	libvirtDomainInterfaceTxDropDesc    *prometheus.Desc

	// domain interface params
}

func init() {
	registerCollector("libvirt", defaultEnabled, NewLibvirtExporter)
}

// NewLibvirtExporter creates a new Prometheus exporter for libvirt.使用uri和是否导出nova信息2个参数启动exporter
func NewLibvirtExporter() (Collector, error) {
	var domainLabels = []string{"domain", "uuid", "name", "flavor", "project_name"}
	return &LibvirtExporter{
		uri:                "qemu:///system",
		exportNovaMetadata: true,
		libvirtUpDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "", "up"),
			"Whether scraping libvirt's metrics was successful.",
			nil,
			nil),
		libvirtDomainActive: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "", "active"),
			"the number of active domains.",
			nil,
			nil),
		libvirtDomainTotal: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "", "total"),
			"the number of active and inactive domains (total).",
			nil,
			nil),
		// domain info
		libvirtDomainInfoDomainState: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "domain_state"),
			"the state of the domain.",
			domainLabels,
			nil),
		libvirtDomainInfoMaxMemDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "maximum_memory_bytes"),
			"Maximum allowed memory of the domain, in bytes.",
			domainLabels,
			nil),
		libvirtDomainInfoMemoryDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "memory_usage_bytes"),
			"Memory usage of the domain, in bytes.",
			domainLabels,
			nil),
		libvirtDomainInfoNrVirtCpuDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "virtual_cpus"),
			"Number of virtual CPUs for the domain.",
			domainLabels,
			nil),
		libvirtDomainInfoCpuTimeDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "cpu_time_seconds_total"),
			"Amount of CPU time used by the domain, in seconds.",
			domainLabels,
			nil),
		libvirtDomainCpuCpuTime: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_cpu_state", "cpu_cpu_time_ns"),
			"Cpu time used in ns.",
			domainLabels,
			nil),
		libvirtDomainCpuUserTime: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_cpu_state", "cpu_user_time_ns"),
			"Cpu time used by user in ns.",
			domainLabels,
			nil),
		libvirtDomainCpuSystemTime: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_cpu_state", "cpu_system_time_ns"),
			"Cpu time used by system in ns.",
			domainLabels,
			nil),
		libvirtDomainCpuVcpuTime: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_cpu_state", "cpu_vcpu_time_ns"),
			"vcpu time used in ns.",
			domainLabels,
			nil),
		// domain memory info
		libvirtDomainMemUnused: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_mem_state", "mem_unused"),
			"The amount of memory left completely unused by the system. This value is expressed in kB.",
			domainLabels,
			nil),
		libvirtDomainMemAvailable: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_mem_state", "mem_available"),
			"The total amount of usable memory as seen by the domain. This value is expressed in kB.",
			domainLabels,
			nil),
		libvirtDomainMemUsable: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_mem_state", "mem_usable"),
			"How much the balloon can be inflated without pushing the guest system to swap, corresponds to 'Available' in /proc/meminfo",
			domainLabels,
			nil),
		libvirtDomainMemRss: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_mem_state", "mem_rss"),
			"Resident Set Size of the process running the domain. This value is in kB",
			domainLabels,
			nil),
		libvirtDomainMemLastUpdate: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_mem_state", "mem_last_update"),
			"Timestamp of the last update of statistics, in seconds.",
			domainLabels,
			nil),
		// domain block info
		libvirtDomainBlockCapacity: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "block_capacity"),
			"logical size in bytes of the image (how much storage the guest will see).",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockAllocation: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "block_allocation"),
			"host storage in bytes occupied by the image (such as highest allocated extent if there are no holes, similar to 'du').",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockPhysical: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "block_physical"),
			"host physical size in bytes of the image container (last offset, similar to 'ls'.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockRdBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "read_bytes_total"),
			"Number of bytes read from a block device, in bytes.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockRdReqDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "read_requests_total"),
			"Number of read requests from a block device.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockRdTotalTimesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "read_seconds_total"),
			"Amount of time spent reading from a block device, in seconds.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockWrBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "write_bytes_total"),
			"Number of bytes written from a block device, in bytes.",
			append(domainLabels, "source_file", "target_device"),
			nil),

		libvirtDomainBlockWrReqDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "write_requests_total"),
			"Number of write requests from a block device.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockWrTotalTimesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "write_seconds_total"),
			"Amount of time spent writing from a block device, in seconds.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockFlushReqDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "flush_requests_total"),
			"Number of flush requests from a block device.",
			append(domainLabels, "source_file", "target_device"),
			nil),
		libvirtDomainBlockFlushTotalTimesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_block_stats", "flush_seconds_total"),
			"Amount of time spent flushing of a block device, in seconds.",
			append(domainLabels, "source_file", "target_device"),
			nil),

		libvirtDomainInterfaceAddresses: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_info", "interface_info_addresses"),
			"Network interface info .",
			append(domainLabels, "source_bridge", "target_device", "domain_interface"),
			nil),
		libvirtDomainInterfaceRxBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_bytes_total"),
			"Number of bytes received on a network interface, in bytes.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceRxPacketsDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_packets_total"),
			"Number of packets received on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceRxErrsDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_errors_total"),
			"Number of packet receive errors on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceRxDropDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_drops_total"),
			"Number of packet receive drops on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceTxBytesDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_bytes_total"),
			"Number of bytes transmitted on a network interface, in bytes.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceTxPacketsDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_packets_total"),
			"Number of packets transmitted on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceTxErrsDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_errors_total"),
			"Number of packet transmit errors on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
		libvirtDomainInterfaceTxDropDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_drops_total"),
			"Number of packet transmit drops on a network interface.",
			append(domainLabels, "source_bridge", "target_device"),
			nil),
	}, nil
}

// Describe returns metadata for all Prometheus metrics that may be exported.
func (e *LibvirtExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.libvirtUpDesc
	ch <- e.libvirtDomainActive
	ch <- e.libvirtDomainTotal

	ch <- e.libvirtDomainCpuCpuTime
	ch <- e.libvirtDomainCpuUserTime
	ch <- e.libvirtDomainCpuSystemTime
	ch <- e.libvirtDomainCpuVcpuTime

	ch <- e.libvirtDomainInfoDomainState
	ch <- e.libvirtDomainInfoMaxMemDesc
	ch <- e.libvirtDomainInfoMemoryDesc
	ch <- e.libvirtDomainInfoNrVirtCpuDesc
	ch <- e.libvirtDomainInfoCpuTimeDesc

	// domain memory info
	ch <- e.libvirtDomainMemUnused
	ch <- e.libvirtDomainMemAvailable
	ch <- e.libvirtDomainMemUsable
	ch <- e.libvirtDomainMemRss
	ch <- e.libvirtDomainMemLastUpdate

	ch <- e.libvirtDomainBlockCapacity
	ch <- e.libvirtDomainBlockAllocation
	ch <- e.libvirtDomainBlockPhysical
	ch <- e.libvirtDomainBlockRdBytesDesc
	ch <- e.libvirtDomainBlockRdReqDesc
	ch <- e.libvirtDomainBlockRdTotalTimesDesc
	ch <- e.libvirtDomainBlockWrBytesDesc
	ch <- e.libvirtDomainBlockWrReqDesc
	ch <- e.libvirtDomainBlockWrTotalTimesDesc
	ch <- e.libvirtDomainBlockFlushReqDesc
	ch <- e.libvirtDomainBlockFlushTotalTimesDesc
}

// Collect scrapes Prometheus metrics from libvirt.
func (e *LibvirtExporter) Update(ch chan<- prometheus.Metric) error {
	err := e.CollectFromLibvirt(ch)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(
			e.libvirtUpDesc,
			prometheus.GaugeValue,
			1.0)
	} else {
		log.Printf("Failed to scrape metrics: %s", err)
		ch <- prometheus.MustNewConstMetric(
			e.libvirtUpDesc,
			prometheus.GaugeValue,
			0.0)
		return err
	}
	return nil
}

// CollectFromLibvirt obtains Prometheus metrics from all domains in a
// libvirt setup.
func (e *LibvirtExporter) CollectFromLibvirt(ch chan<- prometheus.Metric) error {
	conn, err := libvirt.NewConnect(e.uri)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Use ListDomains() as opposed to using ListAllDomains(), as
	// the latter is unsupported when talking to a system using
	// libvirt 0.9.12 or older.
	domainIds, err := conn.ListDomains()
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainActive,
		prometheus.GaugeValue,
		float64(len(domainIds)))

	//allDomain, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	allDomain, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainTotal,
		prometheus.GaugeValue,
		float64(len(allDomain)+len(domainIds)))
	for _, ad := range allDomain {
		ad.Free()
	}

	for _, id := range domainIds {
		domain, err := conn.LookupDomainById(id)
		if err == nil {
			err = e.CollectDomain(ch, domain)
			domain.Free()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CollectDomain extracts Prometheus metrics from a libvirt domain.
func (e *LibvirtExporter) CollectDomain(ch chan<- prometheus.Metric, domain *libvirt.Domain) error {
	// Decode XML description of domain to get block device names, etc.
	xmlDesc, err := domain.GetXMLDesc(0)
	if err != nil {
		return err
	}
	var desc Domain
	err = xml.Unmarshal([]byte(xmlDesc), &desc)
	if err != nil {
		return err
	}
	domainName, err := domain.GetName()
	if err != nil {
		return err
	}
	var domainUUID = desc.UUID

	// Extract domain label valuies
	var domainLabelValues []string
	if e.exportNovaMetadata {
		var (
			novaName        = desc.Metadata.NovaInstance.Name
			novaFlavor      = desc.Metadata.NovaInstance.Flavor.Name
			novaProjectName = desc.Metadata.NovaInstance.Owner.ProjectName
		)
		domainLabelValues = []string{domainName, domainUUID, novaName, novaFlavor, novaProjectName}
	} else {
		domainLabelValues = []string{domainName, domainUUID}
	}

	// Report domain info.
	info, err := domain.GetInfo()
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainInfoDomainState,
		prometheus.GaugeValue,
		float64(info.State),
		domainLabelValues...)
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainInfoMaxMemDesc,
		prometheus.GaugeValue,
		float64(info.MaxMem)*1024,
		domainLabelValues...)
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainInfoMemoryDesc,
		prometheus.GaugeValue,
		float64(info.Memory)*1024,
		domainLabelValues...)
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainInfoNrVirtCpuDesc,
		prometheus.GaugeValue,
		float64(info.NrVirtCpu),
		domainLabelValues...)
	ch <- prometheus.MustNewConstMetric(
		e.libvirtDomainInfoCpuTimeDesc,
		prometheus.CounterValue,
		float64(info.CpuTime)/1e9,
		domainLabelValues...)

	// Report cpu statistics
	// -1返回的是整个虚机的cpu统计信息,因此虽然是个数组，但是只有1个
	cpuStates, err := domain.GetCPUStats(-1, 0, 0)
	if err != nil {
		return err
	}
	cpuState := cpuStates[0]
	if cpuState.CpuTimeSet {
		ch <- prometheus.MustNewConstMetric(
			e.libvirtDomainCpuCpuTime,
			prometheus.CounterValue,
			float64(cpuState.CpuTime),
			append(domainLabelValues)...)
	}
	if cpuState.SystemTimeSet {
		ch <- prometheus.MustNewConstMetric(
			e.libvirtDomainCpuSystemTime,
			prometheus.CounterValue,
			float64(cpuState.SystemTime),
			append(domainLabelValues)...)
	}
	if cpuState.UserTimeSet {
		ch <- prometheus.MustNewConstMetric(
			e.libvirtDomainCpuUserTime,
			prometheus.CounterValue,
			float64(cpuState.UserTime),
			append(domainLabelValues)...)
	}
	if cpuState.VcpuTimeSet {
		ch <- prometheus.MustNewConstMetric(
			e.libvirtDomainCpuVcpuTime,
			prometheus.CounterValue,
			float64(cpuState.VcpuTime),
			append(domainLabelValues)...)
	}

	// Report memory statistics
	memStats, err := domain.MemoryStats(15, 0) // 15 DOMAIN_MEMORY_STAT totally,flags not used.
	if err != nil {
		return err
	}
	// see https://libvirt.org/html/libvirt-libvirt-domain.html#VIR_DOMAIN_MEMORY_STAT_UNUSED
	for _, memStat := range memStats {
		switch memStat.Tag {
		case 4:
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainMemUnused,
				prometheus.GaugeValue,
				float64(memStat.Val),
				append(domainLabelValues)...)
		case 5:
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainMemAvailable,
				prometheus.GaugeValue,
				float64(memStat.Val),
				append(domainLabelValues)...)
		case 7:
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainMemRss,
				prometheus.GaugeValue,
				float64(memStat.Val),
				append(domainLabelValues)...)
		case 8:
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainMemUsable,
				prometheus.GaugeValue,
				float64(memStat.Val),
				append(domainLabelValues)...)
		case 9:
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainMemLastUpdate,
				prometheus.GaugeValue,
				float64(memStat.Val),
				append(domainLabelValues)...)
		default:
			// no need to do
		}
	}

	// Report block device statistics.
	for _, disk := range desc.Devices.Disks {
		if disk.Device == "cdrom" || disk.Device == "fd" {
			continue
		}
		blockStats, err := domain.BlockStats(disk.Target.Device)
		if err != nil {
			return err
		}

		//block total info
		blockInfo, err := domain.GetBlockInfo(disk.Target.Device, 0)
		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			e.libvirtDomainBlockCapacity,
			prometheus.GaugeValue,
			float64(blockInfo.Capacity),
			append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		ch <- prometheus.MustNewConstMetric(
			e.libvirtDomainBlockAllocation,
			prometheus.GaugeValue,
			float64(blockInfo.Allocation),
			append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		ch <- prometheus.MustNewConstMetric(
			e.libvirtDomainBlockPhysical,
			prometheus.GaugeValue,
			float64(blockInfo.Physical),
			append(domainLabelValues, disk.Source.File, disk.Target.Device)...)

		//block rx info.
		if blockStats.RdBytesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockRdBytesDesc,
				prometheus.CounterValue,
				float64(blockStats.RdBytes),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.RdReqSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockRdReqDesc,
				prometheus.CounterValue,
				float64(blockStats.RdReq),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.RdTotalTimesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockRdTotalTimesDesc,
				prometheus.CounterValue,
				float64(blockStats.RdTotalTimes)/1e9,
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.WrBytesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockWrBytesDesc,
				prometheus.CounterValue,
				float64(blockStats.WrBytes),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.WrReqSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockWrReqDesc,
				prometheus.CounterValue,
				float64(blockStats.WrReq),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.WrTotalTimesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockWrTotalTimesDesc,
				prometheus.CounterValue,
				float64(blockStats.WrTotalTimes)/1e9,
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.FlushReqSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockFlushReqDesc,
				prometheus.CounterValue,
				float64(blockStats.FlushReq),
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		if blockStats.FlushTotalTimesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainBlockFlushTotalTimesDesc,
				prometheus.CounterValue,
				float64(blockStats.FlushTotalTimes)/1e9,
				append(domainLabelValues, disk.Source.File, disk.Target.Device)...)
		}
		// Skip "Errs", as the documentation does not clearly
		// explain what this means.
	}

	// Report network interface statistics.
	for _, iface := range desc.Devices.Interfaces {
		if iface.Target.Device == "" {
			continue
		}
		interfaceStats, err := domain.InterfaceStats(iface.Target.Device)
		if err != nil {
			return err
		}

		// network rx state
		if interfaceStats.RxBytesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceRxBytesDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxBytes),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.RxPacketsSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceRxPacketsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxPackets),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.RxErrsSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceRxErrsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxErrs),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.RxDropSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceRxDropDesc,
				prometheus.CounterValue,
				float64(interfaceStats.RxDrop),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.TxBytesSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceTxBytesDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxBytes),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.TxPacketsSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceTxPacketsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxPackets),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.TxErrsSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceTxErrsDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxErrs),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
		if interfaceStats.TxDropSet {
			ch <- prometheus.MustNewConstMetric(
				e.libvirtDomainInterfaceTxDropDesc,
				prometheus.CounterValue,
				float64(interfaceStats.TxDrop),
				append(domainLabelValues, iface.Source.Bridge, iface.Target.Device)...)
		}
	}

	return nil
}
