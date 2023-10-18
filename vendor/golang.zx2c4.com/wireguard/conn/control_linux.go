//go:build linux && !android

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2023 WireGuard LLC. All Rights Reserved.
 */

package conn

import (
	"fmt"
	"net/netip"
	"unsafe"

	"golang.org/x/sys/unix"
)

func (e *StdNetEndpoint) SrcIP() netip.Addr {
	switch len(e.src) {
	case unix.CmsgSpace(unix.SizeofInet4Pktinfo):
		info := (*unix.Inet4Pktinfo)(unsafe.Pointer(&e.src[unix.CmsgLen(0)]))
		return netip.AddrFrom4(info.Spec_dst)
	case unix.CmsgSpace(unix.SizeofInet6Pktinfo):
		info := (*unix.Inet6Pktinfo)(unsafe.Pointer(&e.src[unix.CmsgLen(0)]))
		// TODO: set zone. in order to do so we need to check if the address is
		// link local, and if it is perform a syscall to turn the ifindex into a
		// zone string because netip uses string zones.
		return netip.AddrFrom16(info.Addr)
	}
	return netip.Addr{}
}

func (e *StdNetEndpoint) SrcIfidx() int32 {
	switch len(e.src) {
	case unix.CmsgSpace(unix.SizeofInet4Pktinfo):
		info := (*unix.Inet4Pktinfo)(unsafe.Pointer(&e.src[unix.CmsgLen(0)]))
		return info.Ifindex
	case unix.CmsgSpace(unix.SizeofInet6Pktinfo):
		info := (*unix.Inet6Pktinfo)(unsafe.Pointer(&e.src[unix.CmsgLen(0)]))
		return int32(info.Ifindex)
	}
	return 0
}

func (e *StdNetEndpoint) SrcToString() string {
	return e.SrcIP().String()
}

// getSrcFromControl parses the control for PKTINFO and if found updates ep with
// the source information found.
func getSrcFromControl(control []byte, ep *StdNetEndpoint) {
	ep.ClearSrc()

	var (
		hdr  unix.Cmsghdr
		data []byte
		rem  []byte = control
		err  error
	)

	for len(rem) > unix.SizeofCmsghdr {
		hdr, data, rem, err = unix.ParseOneSocketControlMessage(rem)
		if err != nil {
			return
		}

		if hdr.Level == unix.IPPROTO_IP &&
			hdr.Type == unix.IP_PKTINFO {

			if ep.src == nil || cap(ep.src) < unix.CmsgSpace(unix.SizeofInet4Pktinfo) {
				ep.src = make([]byte, 0, unix.CmsgSpace(unix.SizeofInet4Pktinfo))
			}
			ep.src = ep.src[:unix.CmsgSpace(unix.SizeofInet4Pktinfo)]

			hdrBuf := unsafe.Slice((*byte)(unsafe.Pointer(&hdr)), unix.SizeofCmsghdr)
			copy(ep.src, hdrBuf)
			copy(ep.src[unix.CmsgLen(0):], data)
			return
		}

		if hdr.Level == unix.IPPROTO_IPV6 &&
			hdr.Type == unix.IPV6_PKTINFO {

			if ep.src == nil || cap(ep.src) < unix.CmsgSpace(unix.SizeofInet6Pktinfo) {
				ep.src = make([]byte, 0, unix.CmsgSpace(unix.SizeofInet6Pktinfo))
			}

			ep.src = ep.src[:unix.CmsgSpace(unix.SizeofInet6Pktinfo)]

			hdrBuf := unsafe.Slice((*byte)(unsafe.Pointer(&hdr)), unix.SizeofCmsghdr)
			copy(ep.src, hdrBuf)
			copy(ep.src[unix.CmsgLen(0):], data)
			return
		}
	}
}

// setSrcControl sets an IP{V6}_PKTINFO in control based on the source address
// and source ifindex found in ep. control's len will be set to 0 in the event
// that ep is a default value.
func setSrcControl(control *[]byte, ep *StdNetEndpoint) {
	if cap(*control) < len(ep.src) {
		return
	}
	*control = (*control)[:0]
	*control = append(*control, ep.src...)
}

const (
	sizeOfGSOData = 2
)

// getGSOSize parses control for UDP_GRO and if found returns its GSO size data.
func getGSOSize(control []byte) (int, error) {
	var (
		hdr  unix.Cmsghdr
		data []byte
		rem  = control
		err  error
	)

	for len(rem) > unix.SizeofCmsghdr {
		hdr, data, rem, err = unix.ParseOneSocketControlMessage(rem)
		if err != nil {
			return 0, fmt.Errorf("error parsing socket control message: %w", err)
		}
		if hdr.Level == unix.SOL_UDP && hdr.Type == unix.UDP_GRO && len(data) >= sizeOfGSOData {
			var gso uint16
			copy(unsafe.Slice((*byte)(unsafe.Pointer(&gso)), sizeOfGSOData), data[:sizeOfGSOData])
			return int(gso), nil
		}
	}
	return 0, nil
}

// setGSOSize sets a UDP_SEGMENT in control based on gsoSize. It leaves existing
// data in control untouched.
func setGSOSize(control *[]byte, gsoSize uint16) {
	existingLen := len(*control)
	avail := cap(*control) - existingLen
	space := unix.CmsgSpace(sizeOfGSOData)
	if avail < space {
		return
	}
	*control = (*control)[:cap(*control)]
	gsoControl := (*control)[existingLen:]
	hdr := (*unix.Cmsghdr)(unsafe.Pointer(&(gsoControl)[0]))
	hdr.Level = unix.SOL_UDP
	hdr.Type = unix.UDP_SEGMENT
	hdr.SetLen(unix.CmsgLen(sizeOfGSOData))
	copy((gsoControl)[unix.SizeofCmsghdr:], unsafe.Slice((*byte)(unsafe.Pointer(&gsoSize)), sizeOfGSOData))
	*control = (*control)[:existingLen+space]
}

// controlSize returns the recommended buffer size for pooling sticky and UDP
// offloading control data.
var controlSize = unix.CmsgSpace(unix.SizeofInet6Pktinfo) + unix.CmsgSpace(sizeOfGSOData)

const StdNetSupportsStickySockets = true
