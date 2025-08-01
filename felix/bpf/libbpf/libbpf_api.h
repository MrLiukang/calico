// Copyright (c) 2021-2022 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include "libbpf.h"
#include <linux/limits.h>
#include <net/if.h>
#include <bpf.h>
#include <stdlib.h>
#include <errno.h>
#include "globals.h"
#include "ip_addr.h"

static void set_errno(int ret) {
	errno = ret >= 0 ? ret : -ret;
}

struct bpf_object* bpf_obj_open(char *filename) {
	struct bpf_object *obj;
	obj = bpf_object__open(filename);
	int err = libbpf_get_error(obj);
	if (err) {
		obj = NULL;
	}
	set_errno(err);
	return obj;
}

void bpf_obj_load(struct bpf_object *obj) {
	set_errno(bpf_object__load(obj));
}

int bpf_program_fd(struct bpf_object *obj, char *secname)
{
	int fd = bpf_program__fd(bpf_object__find_program_by_name(obj, secname));
	if (fd < 0) {
		errno = -fd;
	}

	return fd;
}

void bpf_set_attach_type(struct bpf_object *obj, char *progName, uint attach_type)
{
	struct bpf_program *prog = bpf_object__find_program_by_name(obj, progName);
        if (prog == NULL) {
                errno = ENOENT;
                return;
        }
	int ret =  bpf_program__set_expected_attach_type(prog, attach_type);
	if (ret) {
		set_errno(ret);
	}
	return;
}

void bpf_get_prog_name(uint prog_id, char *prog_name) {
	struct bpf_prog_info info = {};
        int prog_fd = bpf_prog_get_fd_by_id(prog_id);
        if (prog_fd < 0) {
		set_errno(-prog_fd);
		return;
        }
	int len = sizeof(info);
	int err = bpf_prog_get_info_by_fd(prog_fd, &info, &len);
	if (err) {
		set_errno(err);
		return;
	}
	memcpy(prog_name, info.name, strlen(info.name));
}

struct bpf_link *bpf_link_open(char *path) {
	struct bpf_link *link = bpf_link__open(path);
	int err = libbpf_get_error(link);
	if (err) {
		set_errno(err);
		return NULL;
	}
	return link;
}

int bpf_update_link(struct bpf_link *link, struct bpf_object *obj, char *progName)
{
	struct bpf_program *prog = bpf_object__find_program_by_name(obj, progName);
        if (prog == NULL) {
                errno = ENOENT;
                return -1;
        }
	int err = bpf_link__update_program(link, prog);
	set_errno(err);
	return err;
}

int bpf_ctlb_get_prog_fd(int target_fd, int attach_type) {
       int err;
        __u32 attach_flags, prog_cnt, prog_id;

        err = bpf_prog_query(target_fd, attach_type, 0, &attach_flags, &prog_id, &prog_cnt);
        if (err) {
                goto out;
        }
        int prog_fd = bpf_prog_get_fd_by_id(prog_id);
        if (prog_fd < 0) {
                err = -prog_fd;
                goto out;
        }
out:
        set_errno(err);
        return prog_fd;
}


void bpf_ctlb_detach_legacy(int prog_fd, int target_fd, int attach_type) {
        set_errno(bpf_prog_detach2(prog_fd, target_fd, attach_type));
}

int bpf_program_query(int ifindex, int attach_type, int flags, uint *attach_flags, uint *prog_ids, uint *prog_cnt) {
	return bpf_prog_query(ifindex, attach_type, 0, attach_flags, prog_ids, prog_cnt);
}

void bpf_tc_program_attach(struct bpf_object *obj, char *secName, int ifIndex, bool ingress, int prio, uint handle)
{
	DECLARE_LIBBPF_OPTS(bpf_tc_hook, hook,
			.attach_point = ingress ? BPF_TC_INGRESS : BPF_TC_EGRESS,
			);
	DECLARE_LIBBPF_OPTS(bpf_tc_opts, attach, .priority=prio,);
	if (prio) {
		attach.handle = handle;
		attach.flags = BPF_TC_F_REPLACE;
	}

	attach.prog_fd = bpf_program__fd(bpf_object__find_program_by_name(obj, secName));
	if (attach.prog_fd < 0) {
		errno = -attach.prog_fd;
		return;
	}
	hook.ifindex = ifIndex;
	set_errno(bpf_tc_attach(&hook, &attach));
}

struct bpf_link* bpf_tcx_program_attach(struct bpf_object *obj, char *secName, int ifIndex)
{
	DECLARE_LIBBPF_OPTS(bpf_tcx_opts, attach); 
	struct bpf_program *prog = bpf_object__find_program_by_name(obj, secName);
	if (!prog) {
		errno = ENOENT;
		return NULL;
	}
	struct bpf_link *link =  bpf_program__attach_tcx(prog, ifIndex, &attach);
	int err = libbpf_get_error(link);
        if (err) {
                link = NULL;
        }
        set_errno(err);
        return link;
}

void bpf_tc_program_detach(int ifindex, int handle, int pref, bool ingress)
{
	DECLARE_LIBBPF_OPTS(bpf_tc_hook, hook,
			.ifindex = ifindex,
			.attach_point = ingress ? BPF_TC_INGRESS : BPF_TC_EGRESS,
			);
	DECLARE_LIBBPF_OPTS(bpf_tc_opts, opts,
			.handle = handle,
			.priority = pref,
			);

	set_errno(bpf_tc_detach(&hook, &opts));
}

struct bpf_tc_opts bpf_tc_program_query(int ifindex, int handle, int pref, bool ingress)
{
	DECLARE_LIBBPF_OPTS(bpf_tc_hook, hook,
			.ifindex = ifindex,
			.attach_point = ingress ? BPF_TC_INGRESS : BPF_TC_EGRESS,
			);
	DECLARE_LIBBPF_OPTS(bpf_tc_opts, opts,
			.handle = handle,
			.priority = pref,
			);

	set_errno(bpf_tc_query(&hook, &opts));

	return opts;
}

void bpf_tc_create_qdisc(int ifIndex)
{
	DECLARE_LIBBPF_OPTS(bpf_tc_hook, hook, .attach_point = BPF_TC_INGRESS);
	hook.ifindex = ifIndex;
	set_errno(bpf_tc_hook_create(&hook));
}

void bpf_tc_remove_qdisc(int ifindex)
{
        DECLARE_LIBBPF_OPTS(bpf_tc_hook, hook,
			.attach_point = BPF_TC_EGRESS | BPF_TC_INGRESS,
			.ifindex = ifindex,
			);

        set_errno(bpf_tc_hook_destroy(&hook));
}

int bpf_update_jump_map(struct bpf_object *obj, char* mapName, char *progName, int progIndex) {
	struct bpf_program *prog_name = bpf_object__find_program_by_name(obj, progName);
	if (prog_name == NULL) {
		errno = ENOENT;
		return -1;
	}
	int prog_fd = bpf_program__fd(prog_name);
	if (prog_fd < 0) {
		errno = -prog_fd;
		return prog_fd;
	}
	int map_fd = bpf_object__find_map_fd_by_name(obj, mapName);
	if (map_fd < 0) {
		errno = -map_fd;
		return map_fd;
	}
	return bpf_map_update_elem(map_fd, &progIndex, &prog_fd, 0);
}

int bpf_link_destroy(struct bpf_link *link) {
	return bpf_link__destroy(link);
}

void bpf_tc_set_globals(struct bpf_map *map,
			char *iface_name,
			char* host_ip,
			char* intf_ip,
			char* host_ip6,
			char* intf_ip6,
			uint ext_to_svc_mark,
			ushort tmtu,
			ushort vxlanPort,
			ushort psnat_start,
			ushort psnat_len,
			char* host_tunnel_ip,
			char* host_tunnel_ip6,
			uint flags,
			ushort wg_port,
			ushort wg6_port,
			ushort profiling,
			uint natin,
			uint natout,
			uint overlay_tunnel_id,
			uint log_filter_jmp,
			uint *jumps,
			uint *jumps6,
			ushort ingress_packet_rate,
			ushort ingress_packet_burst,
			ushort egress_packet_rate,
			ushort egress_packet_burst)
{
	struct cali_tc_global_data v4 = {
		.tunnel_mtu = tmtu,
		.vxlan_port = vxlanPort,
		.ext_to_svc_mark = ext_to_svc_mark,
		.psnat_start = psnat_start,
		.psnat_len = psnat_len,
		.flags = flags,
		.wg_port = wg_port,
		.profiling = profiling,
		.natin_idx = natin,
		.natout_idx = natout,
		.overlay_tunnel_id = overlay_tunnel_id,
		.log_filter_jmp = log_filter_jmp,
		.ingress_packet_rate = ingress_packet_rate,
		.ingress_packet_burst = ingress_packet_burst,
		.egress_packet_rate = egress_packet_rate,
		.egress_packet_burst = egress_packet_burst,
	};

	strncpy(v4.iface_name, iface_name, sizeof(v4.iface_name));
	v4.iface_name[sizeof(v4.iface_name)-1] = '\0';

	struct cali_tc_global_data v6 = v4;
	struct cali_tc_preamble_globals data;

	memcpy(&v4.host_ip, host_ip, 16);
	memcpy(&v4.intf_ip, intf_ip, 16);
	memcpy(&v4.host_tunnel_ip, host_tunnel_ip, 16);

	memcpy(&v6.host_ip, host_ip6, 16);
	memcpy(&v6.intf_ip, intf_ip6, 16);
	memcpy(&v6.host_tunnel_ip, host_tunnel_ip6, 16);

	int i;

	for (i = 0; i < sizeof(v4.jumps)/sizeof(uint); i++) {
		v4.jumps[i] = jumps[i];
	}

	for (i = 0; i < sizeof(v6.jumps)/sizeof(uint); i++) {
		v6.jumps[i] = jumps6[i];
	}

	v6.wg_port = wg6_port;

	data.v4 = v4;
	data.v6 = v6;
	set_errno(bpf_map__set_initial_value(map, (void*)(&data), sizeof(data)));
}


void bpf_ct_cleanup_set_globals(
    struct bpf_map *map,
    uint64_t creation_grace,

    uint64_t tcp_syn_sent,
    uint64_t tcp_established,
    uint64_t tcp_fins_seen,
    uint64_t tcp_reset_seen,

    uint64_t udp_timeout,
    uint64_t generic_timeout,
    uint64_t icmp_timeout
) {
	struct cali_ct_cleanup_globals data = {
		.creation_grace = creation_grace,
		.tcp_syn_sent = tcp_syn_sent,
		.tcp_established = tcp_established,
		.tcp_fins_seen = tcp_fins_seen,
		.tcp_reset_seen = tcp_reset_seen,
		.udp_timeout = udp_timeout,
		.generic_timeout = generic_timeout,
		.icmp_timeout = icmp_timeout,
	};

	set_errno(bpf_map__set_initial_value(map, (void*)(&data), sizeof(data)));
}

int bpf_xdp_program_id(int ifIndex) {
	__u32 prog_id = 0, flags = 0;
	int err;

	err = bpf_xdp_query_id(ifIndex, flags, &prog_id);
	set_errno(err);
	return prog_id;
}

int bpf_program_attach_xdp(struct bpf_object *obj, char *name, int ifIndex, int old_id, __u32 flags)
{
	int err = 0;
	struct bpf_link *link = NULL;
	struct bpf_program *prog, *first_prog = NULL;
	DECLARE_LIBBPF_OPTS(bpf_xdp_attach_opts, opts,
		.old_prog_fd = bpf_prog_get_fd_by_id(old_id));

	if (!(prog = bpf_object__find_program_by_name(obj, name))) {
		err = ENOENT;
		goto out;
	}

	int prog_fd = bpf_program__fd(prog);
	if (prog_fd < 0) {
		errno = -prog_fd;
		return prog_fd;
	}

	err = bpf_xdp_attach(ifIndex, prog_fd, flags, &opts);
	set_errno(err);
	return err;

out:
	set_errno(err);
	return err;
}

struct bpf_link *bpf_program_attach_cgroup(struct bpf_object *obj, int cgroup_fd, char *name)
{
	int err = 0;
	struct bpf_link *link = NULL;
	struct bpf_program *prog;

	if (!(prog = bpf_object__find_program_by_name(obj, name))) {
		err = ENOENT;
		goto out;
	}

	link = bpf_program__attach_cgroup(prog, cgroup_fd);
	err = libbpf_get_error(link);
	if (err) {
		link = NULL;
		goto out;
	}

out:
	set_errno(err);
	return link;
}

void bpf_program_attach_cgroup_legacy(struct bpf_object *obj, int cgroup_fd, char *name)
{
	int err = 0, prog_fd;
	struct bpf_program *prog;
	enum bpf_attach_type attach_type;

	if (!(prog = bpf_object__find_program_by_name(obj, name))) {
		err = ENOENT;
		goto out;
	}

	prog_fd = bpf_program__fd(prog);
	if (prog_fd < 0) {
		err = EINVAL;
		goto out;
	}

	attach_type = bpf_program__get_expected_attach_type(prog);
	err = bpf_prog_attach(prog_fd, cgroup_fd, attach_type, 0);

out:
	set_errno(err);
}

void bpf_ctlb_set_globals(struct bpf_map *map, uint udp_not_seen_timeo, bool exclude_udp)
{
	struct cali_ctlb_globals data = {
		.udp_not_seen_timeo = udp_not_seen_timeo,
		.exclude_udp = exclude_udp,
	};

	set_errno(bpf_map__set_initial_value(map, (void*)(&data), sizeof(data)));
}

void bpf_xdp_set_globals(struct bpf_map *map, char *iface_name, uint *jumps, uint *jumpsV6)
{
	struct cali_xdp_preamble_globals data = {
	};

	strncpy(data.v4.iface_name, iface_name, sizeof(data.v4.iface_name));
	data.v4.iface_name[sizeof(data.v4.iface_name)-1] = '\0';
	data.v6 = data.v4;

	int i;

	for (i = 0; i < sizeof(data.v4.jumps)/sizeof(__u32); i++) {
		data.v4.jumps[i] = jumps[i];
	}

	for (i = 0; i < sizeof(data.v6.jumps)/sizeof(__u32); i++) {
		data.v6.jumps[i] = jumpsV6[i];
	}

	set_errno(bpf_map__set_initial_value(map, (void*)(&data), sizeof(data)));
}

void bpf_map_set_max_entries(struct bpf_map *map, uint max_entries) {
	set_errno(bpf_map__set_max_entries(map, max_entries));
}

int num_possible_cpu()
{
    return libbpf_num_possible_cpus();
}
