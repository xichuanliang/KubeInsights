#include "common.h"

char LICENSE[] SEC("license") = "Dual BSD/GPL";

SEC("tracepoint/syscalls/sys_enter_connect")
int kubeinsights_sys_enter_connect(void *ctx)
{
	struct kubeinsights_event *event = reserve_event(EVENT_SOCKET_CONNECT);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

