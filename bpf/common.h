#ifndef KUBEINSIGHTS_COMMON_H
#define KUBEINSIGHTS_COMMON_H

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

#define TASK_COMM_LEN 16
#define IFACE_LEN 16
#define META_LEN 64

enum kubeinsights_event_type {
	EVENT_SOCKET_CONNECT = 1,
	EVENT_TCP_RTT = 2,
	EVENT_TCP_RETRANSMIT = 3,
	EVENT_NETWORK_HOP = 4,
	EVENT_HTTP = 5,
	EVENT_TLS_HANDSHAKE = 6,
	EVENT_APPLICATION = 7,
	EVENT_DEPENDENCY = 8,
	EVENT_RESPONSE = 9,
	EVENT_SOCKET_CLOSE = 10,
};

struct kubeinsights_event {
	__u64 timestamp_ns;
	__u64 trace_id;
	__u64 socket_cookie;
	__u32 pid;
	__u32 tid;
	__u32 ifindex;
	__u32 event_type;
	__u32 src_ip;
	__u32 dst_ip;
	__u16 src_port;
	__u16 dst_port;
	__u64 duration_ns;
	char comm[TASK_COMM_LEN];
	char interface[IFACE_LEN];
	char metadata[META_LEN];
};

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 24);
} events SEC(".maps");

static __always_inline struct kubeinsights_event *reserve_event(__u32 event_type)
{
	struct kubeinsights_event *event;
	__u64 pid_tgid = bpf_get_current_pid_tgid();

	event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
	if (!event)
		return 0;

	__builtin_memset(event, 0, sizeof(*event));
	event->timestamp_ns = bpf_ktime_get_ns();
	event->pid = pid_tgid >> 32;
	event->tid = (__u32)pid_tgid;
	event->event_type = event_type;
	bpf_get_current_comm(&event->comm, sizeof(event->comm));
	return event;
}

static __always_inline void submit_event(struct kubeinsights_event *event)
{
	if (event)
		bpf_ringbuf_submit(event, 0);
}

#endif

