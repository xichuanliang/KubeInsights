#include "common.h"

char LICENSE[] SEC("license") = "Dual BSD/GPL";

SEC("tracepoint/block/block_rq_issue")
int kubeinsights_block_issue(void *ctx)
{
	struct kubeinsights_event *event = reserve_event(EVENT_DEPENDENCY);
	if (!event)
		return 0;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return 0;
}

