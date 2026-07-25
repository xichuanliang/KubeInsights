package topology

import "net"

type Device struct {
	Name         string   `json:"name"`
	MTU          int      `json:"mtu"`
	HardwareAddr string   `json:"hardwareAddr,omitempty"`
	Flags        []string `json:"flags,omitempty"`
	Addresses    []string `json:"addresses,omitempty"`
}

type Discoverer struct{}

func NewDiscoverer() *Discoverer {
	return &Discoverer{}
}

func (d *Discoverer) Discover() ([]Device, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	devices := make([]Device, 0, len(ifaces))
	for _, iface := range ifaces {
		device := Device{
			Name:         iface.Name,
			MTU:          iface.MTU,
			HardwareAddr: iface.HardwareAddr.String(),
			Flags:        flags(iface.Flags),
		}
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				device.Addresses = append(device.Addresses, addr.String())
			}
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func flags(f net.Flags) []string {
	names := []struct {
		flag net.Flags
		name string
	}{
		{net.FlagUp, "up"},
		{net.FlagBroadcast, "broadcast"},
		{net.FlagLoopback, "loopback"},
		{net.FlagPointToPoint, "point_to_point"},
		{net.FlagMulticast, "multicast"},
		{net.FlagRunning, "running"},
	}
	out := []string{}
	for _, item := range names {
		if f&item.flag != 0 {
			out = append(out, item.name)
		}
	}
	return out
}
