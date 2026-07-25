#include "common.h"

char LICENSE[] SEC("license") = "Dual BSD/GPL";

SEC("uprobe/SSL_write")
int BPF_KPROBE(kubeinsights_ssl_write)
{
	struct kubeinsights_event *event = reserve_event(EVENT_TLS_HANDSHAKE);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

SEC("uprobe/SSL_read")
int BPF_KPROBE(kubeinsights_ssl_read)
{
	struct kubeinsights_event *event = reserve_event(EVENT_RESPONSE);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

