package route


import (
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

var (
	nGetBestRoute uintptr
	defaultGateway string
)

type (
	DWORD               uint32
	ULONG               uint32
	NET_IFINDEX         ULONG
	IF_INDEX            NET_IFINDEX
	NL_ROUTE_PROTOCOL   int32
	MIB_IPFORWARD_PROTO NL_ROUTE_PROTOCOL
	MIB_IPFORWARD_TYPE  int32
)

func init() {
	iphlp, err := syscall.LoadLibrary("iphlpapi.dll")
	if err != nil {
		logger.Fatalf("looadlibrary iphlpapi.dll error: %+v", err)
	}
	defer syscall.FreeLibrary(iphlp)
	nGetBestRoute = getProcAddr(iphlp, "GetBestRoute")

	if defaultGateway, err = getDefaultGateway(); err != nil {
		logger.Fatalf("get default gateway error: %+v", err)
	}
	logger.Printf("system gateway is %s", defaultGateway)
}

func getProcAddr(lib syscall.Handle, name string) uintptr {
	addr, err := syscall.GetProcAddress(lib, name)
	if err != nil {
		panic(name + " " + err.Error())
	}
	return addr
}

type MIB_IPFORWARDROW struct {
	DwForwardDest      DWORD
	DwForwardMask      DWORD
	DwForwardPolicy    DWORD
	DwForwardNextHop   DWORD
	DwForwardIfIndex   IF_INDEX
	ForwardType        MIB_IPFORWARD_TYPE
	ForwardProto       MIB_IPFORWARD_PROTO
	DwForwardAge       DWORD
	DwForwardNextHopAS DWORD
	DwForwardMetric1   DWORD
	DwForwardMetric2   DWORD
	DwForwardMetric3   DWORD
	DwForwardMetric4   DWORD
	DwForwardMetric5   DWORD
}

func dwordIP(d DWORD) (ip net.IP) {
	ip = make(net.IP, net.IPv4len)
	ip[0] = byte(d & 0xff)
	ip[1] = byte((d >> 8) & 0xff)
	ip[2] = byte((d >> 16) & 0xff)
	ip[3] = byte((d >> 24) & 0xff)
	return
}

func ipDword(ip net.IP) (d DWORD) {
	ip = ip.To4()
	d |= DWORD(ip[0]) << 0
	d |= DWORD(ip[1]) << 8
	d |= DWORD(ip[2]) << 16
	d |= DWORD(ip[3]) << 24
	return
}

// find system default gateway
func getDefaultGateway() (string, error) {
	var row MIB_IPFORWARDROW
	_, _, err := syscall.Syscall(nGetBestRoute, 3,
		uintptr(ipDword(net.ParseIP("8.8.8.8"))),
		uintptr(ipDword(net.ParseIP("0.0.0.0"))),
		uintptr(unsafe.Pointer(&row)))
	if err != syscall.Errno(0) {
		return "", err
	}
	return dwordIP(row.DwForwardNextHop).String(), nil
}

// set route for server
func setServerRoute(serverIp string) error {
	if out, err := RunCmd("route", "add", fmt.Sprintf("%s/32", serverIp), defaultGateway); err != nil {
		return errors.Wrap(err, string(out))
	}
	return nil
}

// set traffic route
func setTrafficRoute(tunnelGateway string) error {
	// set 0.0.0.0 to use virtual gateway
	if out, err := RunCmd("route", "add", "0.0.0.0/0", tunnelGateway); err != nil {
		return errors.Wrap(err, string(out))
	}
	// clear old traffic route
	if out, err := RunCmd("route", "delete", "0.0.0.0/0", defaultGateway); err != nil {
		return errors.Wrap(err, string(out))
	}
	return nil
}

// restore network
func RestoreRoute(tunnelGateway string, serverIp string) error {
	// set 0.0.0.0 to use virtual gateway
	if out, err := RunCmd("route", "add", "0.0.0.0/0", defaultGateway); err != nil {
		return errors.Wrap(err, string(out))
	}
	// clear old traffic route
	if out, err := RunCmd("route", "delete", "0.0.0.0/0", tunnelGateway); err != nil {
		return errors.Wrap(err, string(out))
	}
	// remove server route
	if out, err := RunCmd("route", "delete", fmt.Sprintf("%s/32", serverIp), defaultGateway); err != nil {
		return errors.Wrap(err, string(out))
	}
	return nil
}

// setup route
func SetupRoute(tunnelGateway string, serverIp string) error {
	if err := setServerRoute(serverIp); err != nil {
		return err
	}
	if err := setTrafficRoute(tunnelGateway); err != nil {
		return err
	}
	return nil
}
