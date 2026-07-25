#include "common.h"

char LICENSE[] SEC("license") = "Dual BSD/GPL";

SEC("tracepoint/sched/sched_switch")
int kubeinsights_sched_switch(void *ctx)
{
	struct kubeinsights_event *event = reserve_event(EVENT_APPLICATION);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

