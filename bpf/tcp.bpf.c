#include "common.h"

char LICENSE[] SEC("license") = "Dual BSD/GPL";

SEC("kprobe/tcp_retransmit_skb")
int BPF_KPROBE(kubeinsights_tcp_retransmit_skb)
{
	struct kubeinsights_event *event = reserve_event(EVENT_TCP_RETRANSMIT);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

SEC("kprobe/tcp_rcv_established")
int BPF_KPROBE(kubeinsights_tcp_rcv_established)
{
	struct kubeinsights_event *event = reserve_event(EVENT_TCP_RTT);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

