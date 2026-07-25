#include "common.h"

char LICENSE[] SEC("license") = "Dual BSD/GPL";

SEC("kprobe/tcp_recvmsg")
int BPF_KPROBE(kubeinsights_tcp_recvmsg)
{
	struct kubeinsights_event *event = reserve_event(EVENT_HTTP);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

