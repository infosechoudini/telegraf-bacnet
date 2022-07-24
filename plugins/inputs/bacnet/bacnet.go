package bacnet

import (
	"net"
	"reflect"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	zbacnet "github.com/zmap/zgrab2/modules/bacnet"
)

type requestInfo struct {
	dstaddr     string //"ip"
	dstport     int
	objectId    uint32 //obTy<<22 | obInst
	ProperityId uint8
	parmName    string
	waitTime    time.Duration
}

func (b *requestInfo) SampleConfig() string {
	return `
	[[inputs.BACnet]]
	[[inputs.BACnet.Properity.Tags]]
	dstaddr = "192.168.5.32:47808"
	objectId = 8388608
	requestInfo.ProperityId = 85
	parmName = "name of parameter"
	waitTime = 500
	
	[[inputs.BACnet.Properity.Tags]]
	dstaddr = "192.168.5.32:56419"
	objectId = 0
	requestInfo.ProperityId = 85
	parmName = "name of parameter"
	waitTime = 500
	`
}

func (b *requestInfo) Description() string {
	return "BACnet Read Properity"
}

func init() {
	inputs.Add("bacnet", func() telegraf.Input { return &requestInfo{} })
}

func (b *requestInfo) Gather(acc telegraf.Accumulator) error {
	conn, err := net.Dial("udp", b.dstaddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	b.QueryDevice(conn, acc)

	return nil
}

func (b *requestInfo) QueryDevice(conn net.Conn, acc telegraf.Accumulator) {

	grouper := metric.NewSeriesGrouper()

	log := new(zbacnet.Log)

	t := reflect.TypeOf(log)

	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		reflect.ValueOf(log).MethodByName(method.Name).Call(nil)
	}

	rtags := map[string]string{
		"name":            log.ObjectName,
		"location":        log.Location,
		"instance_number": strconv.FormatUint(uint64(log.InstanceNumber), 10),
		"description":     log.Description,
	}

	rfields := map[string]string{
		"vendor_id":        strconv.FormatUint(uint64(log.VendorID), 10),
		"vendor_name":      log.VendorName,
		"firmware_version": log.FirmwareRevision,
		"app_version":      log.ApplicationSoftwareRevision,
	}

	grouper.Add("bacnet", rtags, time.Now(), "asset_id", rfields)

	// Add the metrics grouped by series to the accumulator
	for _, x := range grouper.Metrics() {
		acc.AddMetric(x)
	}
}
