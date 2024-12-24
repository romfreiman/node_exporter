// Copyright 2024 The Prometheus Authors
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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type testAcceleratorCollector struct {
	xc Collector
}

func (c testAcceleratorCollector) Collect(ch chan<- prometheus.Metric) {
	c.xc.Update(ch)
}

func (c testAcceleratorCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func TestAccelerator(t *testing.T) {
	testcase := `# HELP node_accelerator_card_info Accelerator card info including vendor, model and pci id (address)
	# TYPE node_accelerator_card_info counter
	node_accelerator_card_info{id="0000:00:02.0",model="A100",vendor="NVIDIA"} 1
	node_accelerator_card_info{id="0000:00:09.0",model="A100",vendor="NVIDIA"} 1
	node_accelerator_card_info{id="0000:00:1f.5",model="RTX_4090",vendor="NVIDIA"} 1
	`
	vendorToDeviceMap, err := prepareVendorModelData("testdata/accelerators_test_data.yaml")
	if err != nil {
		t.Fatal(err)
	}

	*sysPath = "fixtures/sys"
	logger := log.NewLogfmtLogger(os.Stderr)
	c := &acceleratorsCollector{
		pciDevicesPath:    filepath.Join(*sysPath, "bus/pci/devices"),
		logger:            logger,
		vendorToDeviceMap: vendorToDeviceMap,
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(&testAcceleratorCollector{xc: c})

	sink := make(chan prometheus.Metric)
	go func() {
		err = c.Update(sink)
		if err != nil {
			panic(fmt.Errorf("failed to update collector: %s", err))
		}
		close(sink)
	}()

	err = testutil.GatherAndCompare(reg, strings.NewReader(testcase))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_prepareVendorModelData_badMapping(t *testing.T) {
	_, err := prepareVendorModelData("testdata/accelerators_test_data_duplicated_vendors.bad.yaml")
	if err == nil {
		t.Fatal(err)
	}

	_, err = prepareVendorModelData("testdata/accelerators_test_data_duplicated_device_ids.bad.yaml")
	if err == nil {
		t.Fatal(err)
	}
}
