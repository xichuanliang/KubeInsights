#include "common.h"

char LICENSE[] SEC("license") = "Dual BSD/GPL";

SEC("kprobe/tcp_v4_connect")
int BPF_KPROBE(kubeinsights_tcp_v4_connect)
{
	struct kubeinsights_event *event = reserve_event(EVENT_SOCKET_CONNECT);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

SEC("kprobe/tcp_close")
int BPF_KPROBE(kubeinsights_tcp_close)
{
	struct kubeinsights_event *event = reserve_event(EVENT_SOCKET_CLOSE);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

