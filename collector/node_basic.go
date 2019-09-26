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
	"encoding/json"
	"github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	"strconv"
)

type linuxBasicCollector struct {
	hostName *prometheus.Desc
	cpu      *prometheus.Desc
	mem      *prometheus.Desc
	disk     *prometheus.Desc
	netDev   *prometheus.Desc
}

const (
	basicCollectorSubsystem = "basic"
)

func init() {
	registerCollector("basic", defaultEnabled, NewLinuxBasicCollector)
}

// NewCPUCollector returns a new Collector exposing kernel/system statistics.
func NewLinuxBasicCollector() (Collector, error) {
	return &linuxBasicCollector{
		hostName: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, basicCollectorSubsystem, "host_info"),
			"操作系统信息.",
			[]string{"hostname", "os", "platform", "platform_family", "platform_version",
				"host_id", "virtualization_system", "virtualization_role"}, nil,
		),
		cpu: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, basicCollectorSubsystem, "cpu"),
			"cpu信息.",
			[]string{"count", "core", "vendor_id", "model_name", "mhz"}, nil,
		),

		mem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, basicCollectorSubsystem, "mem"),
			"内存信息.",
			[]string{"total"}, nil,
		),
		disk: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, basicCollectorSubsystem, "disk"),
			"磁盘信息.",
			[]string{"total"}, nil,
		),
		// e.Name,string(s),e.HardwareAddr,strconv.Itoa(e.MTU)
		netDev: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, basicCollectorSubsystem, "net_dev"),
			"网卡信息.",
			[]string{"if_index", "if_name", "ip_address", "hw_address", "mtu"}, nil,
		),
	}, nil
}

// Update implements Collector and exposes cpu related metrics from /proc/stat and /sys/.../cpu/.
func (c *linuxBasicCollector) Update(ch chan<- prometheus.Metric) error {
	if err := c.updateCpuInfo(ch); err != nil {
		return err
	}
	if err := c.updateMemInfo(ch); err != nil {
		return err
	}

	if err := c.updateNetDevInfo(ch); err != nil {
		return err
	}

	if err := c.updateHostName(ch); err != nil {
		return err
	}

	if err := c.updateDisk(ch); err != nil {
		return err
	}
	return nil
}

func (c *linuxBasicCollector) updateDisk(ch chan<- prometheus.Metric) error {

	a, err := disk.Partitions(true)

	if err != nil {
		return err
	}
	var total uint64 = 0
	for _, e := range a {
		b, _ := disk.Usage(e.Mountpoint)
		total = total + b.Total
	}

	ch <- prometheus.MustNewConstMetric(c.disk, prometheus.CounterValue, 1, strconv.FormatUint(total, 10))
	return nil

}

func (c *linuxBasicCollector) updateHostName(ch chan<- prometheus.Metric) error {

	a, err := host.Info()
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(c.hostName, prometheus.CounterValue,
		1, a.Hostname, a.OS, a.Platform, a.PlatformFamily,
		a.PlatformVersion, a.HostID, a.VirtualizationSystem, a.VirtualizationRole)
	return nil

}

func (c *linuxBasicCollector) updateMemInfo(ch chan<- prometheus.Metric) error {
	a, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(c.mem, prometheus.CounterValue, 1, strconv.FormatUint(a.Total, 10))
	return nil
}

func (c *linuxBasicCollector) updateNetDevInfo(ch chan<- prometheus.Metric) error {
	a, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, e := range a {
		s, _ := json.Marshal(e.Addrs)
		ch <- prometheus.MustNewConstMetric(c.netDev, prometheus.CounterValue, 1, strconv.Itoa(e.Index), e.Name, string(s), e.HardwareAddr, strconv.Itoa(e.MTU))
	}
	return nil
}

func (c *linuxBasicCollector) updateCpuInfo(ch chan<- prometheus.Metric) error {
	a, err := cpu.Info()
	if err != nil {
		return err
	}
	if len(a) < 1 {
		return errors.New("no cpu info")
	}
	s := mapset.NewThreadUnsafeSet()
	for _, e := range a {
		s.Add(e.PhysicalID)
	}
	cores := mapset.NewThreadUnsafeSet()

	mHz := 0.0
	for _, e := range a {
		cores.Add(e.CoreID)
		mHz += float64(e.Mhz)
	}
	mHz = mHz / float64(len(a))
	coreNum := len(s.ToSlice()) * len(cores.ToSlice())
	ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, 1, strconv.Itoa(len(s.ToSlice())),
		strconv.Itoa(coreNum), a[0].VendorID, a[0].ModelName, strconv.FormatFloat(mHz, 'f', 0, 64))
	return nil
}
