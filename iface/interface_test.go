// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package iface

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type ifaceInfoSuite struct {
	suite.Suite
	MockIfaceInfo string
}

func (iis *ifaceInfoSuite) SetupSuite() {
	ifaceInfo = iis.MockIfaceInfo
	if err := createMockIfaceInfo(); err != nil {
		iis.T().Skip("Could not find network interface test file!", err)
	}
}

func (iis *ifaceInfoSuite) TearDownSuite() {
	removeIfaceLoadInfo()
}

func (iis *ifaceInfoSuite) TestGetStats() {
	Convey("Given interface info map", iis.T(), func() {
		ifacePlg := New()

		Convey("and mock memory info file created", func() {
			assert.Equal(iis.T(), "mockIfaceInfo", ifaceInfo)
		})

		Convey("When reading interface statistics from file", func() {
			err := ifacePlg.getStats()

			Convey("No error should be reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Proper statistics values are returned", func() {
				So(len(ifacePlg.stats), ShouldEqual, 2)

				So(ifacePlg.stats["p3p1"], ShouldHaveSameTypeAs, map[string]interface{}{})
				p3p1 := ifacePlg.stats["p3p1"].(map[string]interface{})
				So(len(p3p1), ShouldEqual, 16)

				val, ok := p3p1["bytes_recv"].(int64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, 1412848320)

				val, ok = p3p1["packets_recv"].(int64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, 12238775)

				val, ok = p3p1["packets_sent"].(int64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, 17015516)

				So(ifacePlg.stats["lo"], ShouldHaveSameTypeAs, map[string]interface{}{})
				lo := ifacePlg.stats["lo"].(map[string]interface{})
				So(len(lo), ShouldEqual, 16)

				val, ok = lo["fifo_sent"].(int64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, 0)

				val, ok = lo["errs_recv"].(int64)
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, 0)
			})
		})
	})
}

func (iis *ifaceInfoSuite) TestGetMetricTypes() {
	Convey("Given interface info plugin initialized", iis.T(), func() {
		ifacePlg := New()

		Convey("When one wants to get list of available meterics", func() {
			mts, err := ifacePlg.GetMetricTypes(plugin.Config{})

			Convey("Then error should not be reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then list of metrics is returned", func() {
				So(len(mts), ShouldEqual, 16)

				namespaces := []string{}
				for _, m := range mts {
					namespaces = append(namespaces, m.Namespace.String())
				}

				So(namespaces, ShouldContain, "/intel/procfs/iface/*/errs_recv")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/frame_recv")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/packets_recv")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/drop_recv")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/fifo_recv")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/compressed_recv")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/multicast_recv")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/bytes_recv")

				So(namespaces, ShouldContain, "/intel/procfs/iface/*/errs_sent")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/frame_sent")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/packets_sent")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/drop_sent")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/fifo_sent")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/compressed_sent")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/multicast_sent")
				So(namespaces, ShouldContain, "/intel/procfs/iface/*/bytes_sent")
			})
		})
	})
}

func (iis *ifaceInfoSuite) TestCollectMetrics() {
	Convey("Given interface info plugin initialized", iis.T(), func() {
		ifacePlg := New()
		cfg := plugin.Config{}
		Convey("When one wants to get values for given metric types", func() {
			mTypes := []plugin.Metric{
				{
					Namespace: plugin.NewNamespace("intel", "procfs", "iface", "p3p1", "bytes_sent"),
					Config:    cfg,
				},
				{
					Namespace: plugin.NewNamespace("intel", "procfs", "iface", "lo", "packets_recv"),
					Config:    cfg,
				},
			}

			metrics, err := ifacePlg.CollectMetrics(mTypes)

			Convey("Then no errors should be reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then proper metrics values are returned", func() {
				So(len(metrics), ShouldEqual, 2)

				stats := map[string]interface{}{}
				for _, m := range metrics {
					n := m.Namespace.String()
					stats[n] = m.Data
				}

				So(len(metrics), ShouldEqual, len(stats))

				So(stats["/intel/procfs/iface/p3p1/bytes_sent"], ShouldNotBeNil)
				So(stats["/intel/procfs/iface/lo/packets_recv"], ShouldNotBeNil)
			})
		})
	})
}

func TestGetStatsSuite(t *testing.T) {
	suite.Run(t, &ifaceInfoSuite{MockIfaceInfo: "mockIfaceInfo"})
}

func createMockIfaceInfo() error {
	ifaceInfoContent, err := ioutil.ReadFile("../examples/test/proc.net.dev")
	if err != nil {
		return err
	}

	f, err := os.Create(ifaceInfo)
	if err != nil {
		return err
	}

	f.Write(ifaceInfoContent)
	return nil
}

func removeIfaceLoadInfo() {
	os.Remove(ifaceInfo)
}
