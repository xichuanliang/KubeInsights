#include "common.h"

char LICENSE[] SEC("license") = "Dual BSD/GPL";

SEC("tc")
int kubeinsights_tc_ingress(struct __sk_buff *skb)
{
	struct kubeinsights_event *event = reserve_event(EVENT_NETWORK_HOP);
	if (!event)
		return 0;
	event->ifindex = skb->ifindex;
	event->trace_id = bpf_ktime_get_ns();
	submit_event(event);
	return TC_ACT_OK;
}

SEC("xdp")
int kubeinsights_xdp_pass(struct xdp_md *ctx)
{
	return XDP_PASS;
}

