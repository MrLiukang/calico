// Project Calico BPF dataplane programs.
// Copyright (c) 2022 Tigera, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later

#ifndef __CALI_COUNTERS_H__
#define __CALI_COUNTERS_H__

#include "bpf.h"

#define MAX_COUNTERS_SIZE 23

typedef __u64 counters_t[MAX_COUNTERS_SIZE];

struct counters_key {
	__u32 ifindex;
	__u32 hook;
};

#define COUNTERS_TC_INGRESS	0
#define COUNTERS_TC_EGRESS	1
#define COUNTERS_XDP		2

CALI_MAP(cali_counters, 4,
		BPF_MAP_TYPE_PERCPU_HASH,
		struct counters_key, counters_t, 20000,
		0)

CALI_MAP(cali_counters_scratch, 3,
		BPF_MAP_TYPE_PERCPU_ARRAY,
		__u32, counters_t, 1, 0)

static CALI_BPF_INLINE counters_t *counters_get(int ifindex)
{
	struct counters_key key = {
		.ifindex = ifindex,
	};

	if (CALI_F_XDP) {
		key.hook = COUNTERS_XDP;
	} else if (CALI_F_TO_HEP) {
		key.hook = COUNTERS_TC_EGRESS;
	} else if (CALI_F_FROM_HEP) {
		key.hook = COUNTERS_TC_INGRESS;
	} else if (CALI_F_TO_WEP) {
		key.hook = COUNTERS_TC_EGRESS;
	} else if (CALI_F_FROM_WEP) {
		key.hook = COUNTERS_TC_INGRESS;
	}

	return cali_counters_lookup_elem(&key);
}

static CALI_BPF_INLINE void counter_inc(struct cali_tc_ctx *ctx, int type)
{
	if (ctx->counters) {
		((__u64 *)((ctx)->counters))[type]++;
	}
}

static CALI_BPF_INLINE void deny_reason(struct cali_tc_ctx *ctx, int reason)
{
	ctx->fwd.reason = reason;
	counter_inc(ctx, reason);
}

#endif /* __CALI_COUNTERS_H__ */
