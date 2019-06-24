package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/dbus"
	lowdbus "github.com/godbus/dbus"
)

func main() {
	cgroot := "/sys/fs/cgroup"
	ensureCgroupV2(cgroot)

	conn, err := dbus.NewSystemConnection()
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	slice := "xdoj4.slice"
	_, err = conn.StartTransientUnit(slice, "fail", []dbus.Property{
		dbus.PropDescription("XDOJ v4 Prototype"),
	}, nil)
	if err != nil {
		log.Panic(err)
	}
	defer conn.StopUnit(slice, "fail", nil)

	cg1, err := conn.GetUnitTypeProperty(slice, "Slice", "ControlGroup")
	if err != nil {
		log.Panic(err)
	}

	cg := cgroot + cg1.Value.Value().(string)
	log.Printf("control group = %s", cg)

	name := "hello"
	unit := fmt.Sprintf("xdoj4-%s.service", name)

	lowConn, err := lowdbus.SystemBusPrivate()
	if err != nil {
		log.Panic(err)
	}
	defer lowConn.Close()

	methods := []lowdbus.Auth{lowdbus.AuthExternal(strconv.Itoa(os.Getuid()))}
	err = lowConn.Auth(methods)
	if err != nil {
		log.Panic(err)
	}

	if err = lowConn.Hello(); err != nil {
		log.Panic(err)
	}

	err = lowConn.BusObject().AddMatchSignal(
		"org.freedesktop.DBus.Properties",
		"PropertiesChanged",
		lowdbus.WithMatchObjectPath(dbusPath(unit)),
		lowdbus.WithMatchOption("arg0", "org.freedesktop.systemd1.Unit"),
	).Store()
	if err != nil {
		log.Panic(err)
	}

	chsgn := make(chan *lowdbus.Signal)
	lowConn.Signal(chsgn)

	desc := fmt.Sprintf("XDOJ v4 Prototype - Solution %s", name)
	prop := []dbus.Property{
		dbus.PropDescription(desc),
		dbus.PropExecStart([]string{"/home/xry111/loop"}, false),
		dbus.PropRemainAfterExit(true),
		newProperty("DynamicUser", true),
		newProperty("RuntimeMaxUSec", uint64(1000000)),
		newProperty("MemoryMax", uint64(128<<20)),
		newProperty("MemorySwapMax", uint64(0)),
		newProperty("Slice", slice),
		newProperty("CPUQuotaPerSecUSec", uint64(1000000)),
	}

	_, err = conn.StartTransientUnit(unit, "fail", prop, nil)
	if err != nil {
		log.Panic(err)
	}
	defer conn.StopUnit(unit, "fail", nil)

	lastActiveState := ""
	lastSubState := ""
	for v := range chsgn {
		mp := v.Body[1].(map[string]lowdbus.Variant)
		activeState := mp["ActiveState"].Value().(string)
		subState := mp["SubState"].Value().(string)
		if lastActiveState != activeState || lastSubState != subState {
			log.Printf(
				"service %s is now %s (%s)",
				unit, activeState, subState,
			)
			lastActiveState = activeState
			lastSubState = subState
		}
		if subState == "failed" || subState == "exited" {
			break
		}
	}
	lowConn.RemoveSignal(chsgn)

	if lastSubState == "failed" {
		defer conn.ResetFailedUnit(unit)
	}

	cpuUsage1, err := conn.GetServiceProperty(unit, "CPUUsageNSec")
	if err != nil {
		log.Panic(err)
	}
	cpuUsage := time.Duration(cpuUsage1.Value.Value().(uint64))
	log.Printf("consumed %v CPU time", cpuUsage)

	timeProp, err := conn.GetServiceProperty(unit, "ExecMainStartTimestampMonotonic")
	if err != nil {
		log.Panic(err)
	}
	t0 := 1000 * timeProp.Value.Value().(uint64)

	timeProp, err = conn.GetServiceProperty(unit, "ExecMainExitTimestampMonotonic")
	if err != nil {
		log.Panic(err)
	}
	t1 := 1000 * timeProp.Value.Value().(uint64)

	log.Printf("t0 = %v, t1 = %v", t0, t1)
	log.Printf("service runtime: %v", t1 - t0)

	oom, err := getOOMCount(cg)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("OOM count = %d", oom)

	resultProp, err := conn.GetServiceProperty(unit, "Result")
	result := resultProp.Value.Value().(string)
	log.Printf("result = %s", result)
}

func dbusPath(u string) lowdbus.ObjectPath {
	path := lowdbus.ObjectPath(dbus.PathBusEscape(u))
	return "/org/freedesktop/systemd1/unit/" + path
}
