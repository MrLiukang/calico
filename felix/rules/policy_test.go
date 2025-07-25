// Copyright (c) 2016-2025 Tigera, Inc. All rights reserved.
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

package rules_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/projectcalico/calico/felix/environment"
	"github.com/projectcalico/calico/felix/generictables"
	"github.com/projectcalico/calico/felix/ipsets"
	"github.com/projectcalico/calico/felix/iptables"
	"github.com/projectcalico/calico/felix/proto"
	. "github.com/projectcalico/calico/felix/rules"
	"github.com/projectcalico/calico/felix/types"
)

var ruleTestData = []TableEntry{
	Entry("Empty rule", 4, &proto.Rule{}, ""),

	// Non-negated matches...

	Entry("Protocol name", 4,
		&proto.Rule{Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}}},
		"-p tcp"),
	Entry("Protocol num", 4,
		&proto.Rule{Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Number{Number: 8}}},
		"-p 8"),

	Entry("Source net", 4,
		&proto.Rule{SrcNet: []string{"10.0.0.0/16"}},
		"--source 10.0.0.0/16"),
	Entry("Source IP set", 4,
		&proto.Rule{SrcIpSetIds: []string{"ipsetid1"}},
		"-m set --match-set cali40ipsetid1 src"),
	Entry("Source IP sets", 4,
		&proto.Rule{SrcIpSetIds: []string{"ipsetid1", "ipsetid2"}},
		"-m set --match-set cali40ipsetid1 src -m set --match-set cali40ipsetid2 src"),
	Entry("Source ports", 4,
		&proto.Rule{SrcPorts: []*proto.PortRange{{First: 10, Last: 12}}},
		"-m multiport --source-ports 10:12"),
	Entry("Source ports (multiple)", 4,
		&proto.Rule{SrcPorts: []*proto.PortRange{
			{First: 10, Last: 12},
			{First: 20, Last: 30},
			{First: 8080, Last: 8080},
		}},
		"-m multiport --source-ports 10:12,20:30,8080"),
	Entry("ICMP", 4,
		&proto.Rule{Icmp: &proto.Rule_IcmpType{IcmpType: 10}},
		"-m icmp --icmp-type 10"),
	Entry("ICMP with code", 4,
		&proto.Rule{Icmp: &proto.Rule_IcmpTypeCode{IcmpTypeCode: &proto.IcmpTypeAndCode{Type: 10, Code: 12}}},
		"-m icmp --icmp-type 10/12"),
	Entry("ICMP", 6,
		&proto.Rule{Icmp: &proto.Rule_IcmpType{IcmpType: 10}},
		"-m icmp6 --icmpv6-type 10"),
	Entry("ICMP with code", 6,
		&proto.Rule{Icmp: &proto.Rule_IcmpTypeCode{IcmpTypeCode: &proto.IcmpTypeAndCode{Type: 10, Code: 12}}},
		"-m icmp6 --icmpv6-type 10/12"),

	Entry("Dest net", 4,
		&proto.Rule{DstNet: []string{"10.0.0.0/16"}},
		"--destination 10.0.0.0/16"),
	Entry("Dest IP set", 4,
		&proto.Rule{DstIpSetIds: []string{"ipsetid1"}},
		"-m set --match-set cali40ipsetid1 dst"),
	Entry("Dest IP sets", 4,
		&proto.Rule{DstIpSetIds: []string{"ipsetid1", "ipsetid2"}},
		"-m set --match-set cali40ipsetid1 dst -m set --match-set cali40ipsetid2 dst"),
	Entry("Dest ports", 4,
		&proto.Rule{DstPorts: []*proto.PortRange{{First: 10, Last: 12}}},
		"-m multiport --destination-ports 10:12"),
	Entry("Dest ports (multiple)", 4,
		&proto.Rule{DstPorts: []*proto.PortRange{
			{First: 10, Last: 12},
			{First: 20, Last: 30},
			{First: 8080, Last: 8080},
		}},
		"-m multiport --destination-ports 10:12,20:30,8080"),

	// Negated matches...

	Entry("Protocol name", 4,
		&proto.Rule{NotProtocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}}},
		"! -p tcp"),
	Entry("Protocol num", 4,
		&proto.Rule{NotProtocol: &proto.Protocol{NumberOrName: &proto.Protocol_Number{Number: 8}}},
		"! -p 8"),

	Entry("Negated source net", 4,
		&proto.Rule{NotSrcNet: []string{"10.0.0.0/16"}},
		"! --source 10.0.0.0/16"),
	Entry("Negated source IP set", 4,
		&proto.Rule{NotSrcIpSetIds: []string{"ipsetid1"}},
		"-m set ! --match-set cali40ipsetid1 src"),
	Entry("Negated source IP set v6", 6,
		&proto.Rule{NotSrcIpSetIds: []string{"ipsetid1"}},
		"-m set ! --match-set cali60ipsetid1 src"),
	Entry("Negated source IP sets", 4,
		&proto.Rule{NotSrcIpSetIds: []string{"ipsetid1", "ipsetid2"}},
		"-m set ! --match-set cali40ipsetid1 src -m set ! --match-set cali40ipsetid2 src"),
	Entry("Negated source ports", 4,
		&proto.Rule{NotSrcPorts: []*proto.PortRange{{First: 10, Last: 12}}},
		"-m multiport ! --source-ports 10:12"),
	Entry("Negated source ports (multiple)", 4,
		&proto.Rule{NotSrcPorts: []*proto.PortRange{
			{First: 10, Last: 12},
			{First: 20, Last: 30},
			{First: 8080, Last: 8080},
		}},
		"-m multiport ! --source-ports 10:12,20:30,8080"),
	Entry("Negated source ports (>15) should be broken into blocks", 4,
		&proto.Rule{NotSrcPorts: []*proto.PortRange{
			{First: 1, Last: 2},
			{First: 3, Last: 4},
			{First: 5, Last: 6},
			{First: 7, Last: 8},
			{First: 9, Last: 10},
			{First: 11, Last: 12},
			{First: 13, Last: 14},
			{First: 15, Last: 16},
		}},
		"-m multiport ! --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 -m multiport ! --source-ports 15:16"),
	Entry("ICMP", 4,
		&proto.Rule{NotIcmp: &proto.Rule_NotIcmpType{NotIcmpType: 10}},
		"-m icmp ! --icmp-type 10"),
	Entry("ICMP with code", 4,
		&proto.Rule{NotIcmp: &proto.Rule_NotIcmpTypeCode{NotIcmpTypeCode: &proto.IcmpTypeAndCode{Type: 10, Code: 12}}},
		"-m icmp ! --icmp-type 10/12"),
	Entry("ICMP", 6,
		&proto.Rule{NotIcmp: &proto.Rule_NotIcmpType{NotIcmpType: 10}},
		"-m icmp6 ! --icmpv6-type 10"),
	Entry("ICMP with code", 6,
		&proto.Rule{NotIcmp: &proto.Rule_NotIcmpTypeCode{NotIcmpTypeCode: &proto.IcmpTypeAndCode{Type: 10, Code: 12}}},
		"-m icmp6 ! --icmpv6-type 10/12"),

	Entry("Dest net", 4,
		&proto.Rule{NotDstNet: []string{"10.0.0.0/16"}},
		"! --destination 10.0.0.0/16"),
	Entry("Dest IP set", 4,
		&proto.Rule{NotDstIpSetIds: []string{"ipsetid1"}},
		"-m set ! --match-set cali40ipsetid1 dst"),
	Entry("Dest IP set", 6,
		&proto.Rule{NotDstIpSetIds: []string{"ipsetid1"}},
		"-m set ! --match-set cali60ipsetid1 dst"),
	Entry("Dest IP sets", 4,
		&proto.Rule{NotDstIpSetIds: []string{"ipsetid1", "ipsetid2"}},
		"-m set ! --match-set cali40ipsetid1 dst -m set ! --match-set cali40ipsetid2 dst"),
	Entry("Dest ports", 4,
		&proto.Rule{NotDstPorts: []*proto.PortRange{{First: 10, Last: 12}}},
		"-m multiport ! --destination-ports 10:12"),
	Entry("Dest ports (>15) should be broken into blocks", 4,
		&proto.Rule{NotDstPorts: []*proto.PortRange{
			{First: 1, Last: 2},
			{First: 3, Last: 4},
			{First: 5, Last: 6},
			{First: 7, Last: 8},
			{First: 9, Last: 10},
			{First: 11, Last: 12},
			{First: 13, Last: 14},
			{First: 15, Last: 16},
		}},
		"-m multiport ! --destination-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 -m multiport ! --destination-ports 15:16"),
	Entry("Dest ports (multiple)", 4,
		&proto.Rule{NotDstPorts: []*proto.PortRange{
			{First: 10, Last: 12},
			{First: 20, Last: 30},
			{First: 8080, Last: 8080},
		}},
		"-m multiport ! --destination-ports 10:12,20:30,8080"),
}

// Test data for rules that should be filtered out (return no iptables rules)
var filteredRuleTestData = []TableEntry{
	Entry("Negated catch-all IPv4 source net (should be filtered out)", 4,
		&proto.Rule{NotSrcNet: []string{"0.0.0.0/0"}}),
	Entry("Negated catch-all IPv6 source net (should be filtered out)", 6,
		&proto.Rule{NotSrcNet: []string{"::/0"}}),
	Entry("Negated catch-all IPv4 destination net (should be filtered out)", 4,
		&proto.Rule{NotDstNet: []string{"0.0.0.0/0"}}),
	Entry("Negated catch-all IPv6 destination net (should be filtered out)", 6,
		&proto.Rule{NotDstNet: []string{"::/0"}}),
}

var _ = Describe("Protobuf rule to iptables rule conversion", func() {
	rrConfigNormal := Config{
		IPIPEnabled:       true,
		IPIPTunnelAddress: nil,
		IPSetConfigV4:     ipsets.NewIPVersionConfig(ipsets.IPFamilyV4, "cali", nil, nil),
		IPSetConfigV6:     ipsets.NewIPVersionConfig(ipsets.IPFamilyV6, "cali", nil, nil),
		MarkAccept:        0x80,
		MarkPass:          0x100,
		MarkScratch0:      0x200,
		MarkScratch1:      0x400,
		MarkDrop:          0x800,
		MarkEndpoint:      0xff000,
		LogPrefix:         "calico-packet",
	}

	DescribeTable(
		"Allow rules should be correctly rendered",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			rules := renderer.ProtoRuleToIptablesRules(in, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			// For allow, should be one match rule that sets the mark, then one that reads the
			// mark and returns.
			Expect(len(rules)).To(Equal(2))
			Expect(rules[0].Match.Render()).To(Equal(expMatch))
			Expect(rules[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x80}))
			Expect(rules[1]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x80),
				Action: iptables.ReturnAction{},
			}))

			// Explicit allow should be treated the same as empty.
			in.Action = "allow"
			rules2 := renderer.ProtoRuleToIptablesRules(in, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			Expect(rules2).To(Equal(rules))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Allow rules should be correctly rendered with flowlogs enabled",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = true
			renderer := NewRenderer(rrConfigNormal)
			rules := renderer.ProtoRuleToIptablesRules(in, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			// For allow, should be one match rule that sets the mark, then one that reads the
			// mark and returns.
			Expect(len(rules)).To(Equal(3))
			Expect(rules[0].Match.Render()).To(Equal(expMatch))
			Expect(rules[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x80}))
			Expect(rules[1]).To(Equal(generictables.Rule{
				Match: iptables.Match().MarkSingleBitSet(0x80),
				Action: iptables.NflogAction{
					Group:  1,
					Prefix: "API0|default.foo",
				},
			}))
			Expect(rules[2]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x80),
				Action: iptables.ReturnAction{},
			}))

			// Explicit allow should be treated the same as empty.
			in.Action = "allow"
			rules2 := renderer.ProtoRuleToIptablesRules(in, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			Expect(rules2).To(Equal(rules))
		},
		ruleTestData...,
	)

	DescribeTable(
		"pass rules should be correctly rendered",
		func(ipVer int, in *proto.Rule, expMatch string) {
			for _, action := range []string{"next-tier", "pass"} {
				By("Rendering for action " + action)
				rrConfigNormal.FlowLogsEnabled = false
				renderer := NewRenderer(rrConfigNormal)
				in.Action = action
				rules := renderer.ProtoRuleToIptablesRules(in, uint8(ipVer),
					RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
				// For pass, should be one match rule that sets the mark, then one
				// that reads the mark and returns.
				Expect(len(rules)).To(Equal(2))
				Expect(rules[0].Match.Render()).To(Equal(expMatch))
				Expect(rules[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x100}))
				Expect(rules[1]).To(Equal(generictables.Rule{
					Match:  iptables.Match().MarkSingleBitSet(0x100),
					Action: iptables.ReturnAction{},
				}))
			}
		},
		ruleTestData...,
	)

	DescribeTable(
		"pass rules should be correctly rendered with flowlogs enabled",
		func(ipVer int, in *proto.Rule, expMatch string) {
			for _, action := range []string{"next-tier", "pass"} {
				By("Rendering for action " + action)
				rrConfigNormal.FlowLogsEnabled = true
				renderer := NewRenderer(rrConfigNormal)
				in.Action = action
				rules := renderer.ProtoRuleToIptablesRules(in, uint8(ipVer),
					RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
				// For pass, should be one match rule that sets the mark, then one
				// that reads the mark and returns.
				Expect(len(rules)).To(Equal(3))
				Expect(rules[0].Match.Render()).To(Equal(expMatch))
				Expect(rules[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x100}))
				Expect(rules[1]).To(Equal(generictables.Rule{
					Match: iptables.Match().MarkSingleBitSet(0x100),
					Action: iptables.NflogAction{
						Group:  1,
						Prefix: "PPI0|default.foo",
					},
				}))
				Expect(rules[2]).To(Equal(generictables.Rule{
					Match:  iptables.Match().MarkSingleBitSet(0x100),
					Action: iptables.ReturnAction{},
				}))
			}
		},
		ruleTestData...,
	)

	DescribeTable(
		"Log rules should be correctly rendered",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			logRule := in
			logRule.Action = "log"
			rules := renderer.ProtoRuleToIptablesRules(logRule, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			// For deny, should be one match rule that just does the DROP.
			Expect(len(rules)).To(Equal(1))
			Expect(rules[0].Match.Render()).To(Equal(expMatch))
			Expect(rules[0].Action).To(Equal(iptables.LogAction{Prefix: "calico-packet"}))

			// Enabling flow log must not have any effect
			rrConfigNormal.FlowLogsEnabled = true
			renderer = NewRenderer(rrConfigNormal)
			rules2 := renderer.ProtoRuleToIptablesRules(logRule, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			Expect(rules2).To(Equal(rules))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Log rules should be correctly rendered with non-default prefix",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			rrConfigPrefix := rrConfigNormal
			rrConfigPrefix.LogPrefix = "foobar"
			renderer := NewRenderer(rrConfigPrefix)
			logRule := in
			logRule.Action = "log"
			rules := renderer.ProtoRuleToIptablesRules(logRule, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			// For deny, should be one match rule that just does the DROP.
			Expect(len(rules)).To(Equal(1))
			Expect(rules[0].Match.Render()).To(Equal(expMatch))
			Expect(rules[0].Action).To(Equal(iptables.LogAction{Prefix: "foobar"}))

			// Enabling flow log must not have any effect
			rrConfigPrefix.FlowLogsEnabled = true
			renderer = NewRenderer(rrConfigPrefix)
			rules2 := renderer.ProtoRuleToIptablesRules(logRule, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			Expect(rules2).To(Equal(rules))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Deny (DROP) rules should be correctly rendered",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			denyRule.Action = "deny"
			rules := renderer.ProtoRuleToIptablesRules(denyRule, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			// For deny, should be one match rule that just does the DROP.
			Expect(len(rules)).To(Equal(2))
			Expect(rules[0].Match.Render()).To(Equal(expMatch))
			Expect(rules[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x800}))
			Expect(rules[1]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.DropAction{},
			}))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Deny (DROP) rules should be correctly rendered with flowlogs enabled",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = true
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			denyRule.Action = "deny"
			rules := renderer.ProtoRuleToIptablesRules(denyRule, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			// For deny, should be one match rule that just does the DROP.
			Expect(len(rules)).To(Equal(3))
			Expect(rules[0].Match.Render()).To(Equal(expMatch))
			Expect(rules[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x800}))
			Expect(rules[1]).To(Equal(generictables.Rule{
				Match: iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.NflogAction{
					Group:  1,
					Prefix: "DPI0|default.foo",
				},
			}))
			Expect(rules[2]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.DropAction{},
			}))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Inbound deny rules should be correctly rendered within a policy",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			denyRule.Action = "deny"
			policyID := &types.PolicyID{
				Tier: "default",
				Name: "default.foo",
			}
			policy := &proto.Policy{
				Namespace:     "",
				InboundRules:  []*proto.Rule{denyRule},
				OutboundRules: []*proto.Rule{},
				Untracked:     false,
				PreDnat:       false,
			}

			chains := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			Expect(chains[0].Name).To(Equal("cali-pi-default/default.foo"))
			Expect(chains[1].Name).To(Equal("cali-po-default/default.foo"))

			inbound := chains[0].Rules
			outbound := chains[1].Rules
			Expect(inbound).To(HaveLen(2))
			Expect(outbound).To(ConsistOf(
				generictables.Rule{
					Comment: []string{"Policy default.foo egress"},
				}))
			Expect(inbound[0].Match.Render()).To(Equal(expMatch))
			Expect(inbound[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x800}))
			Expect(inbound[1]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.DropAction{},
			}))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Inbound deny rules should be correctly rendered within a policy with flowlogs enabled",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = true
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			denyRule.Action = "deny"
			policyID := &types.PolicyID{
				Tier: "default",
				Name: "default.foo",
			}
			policy := &proto.Policy{
				Namespace:     "",
				InboundRules:  []*proto.Rule{denyRule},
				OutboundRules: []*proto.Rule{},
				Untracked:     false,
				PreDnat:       false,
			}

			chains := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			Expect(chains[0].Name).To(Equal("cali-pi-default/default.foo"))
			Expect(chains[1].Name).To(Equal("cali-po-default/default.foo"))

			inbound := chains[0].Rules
			outbound := chains[1].Rules
			Expect(inbound).To(HaveLen(3))
			Expect(outbound).To(ConsistOf(
				generictables.Rule{
					Comment: []string{"Policy default.foo egress"},
				}))
			Expect(inbound[0].Match.Render()).To(Equal(expMatch))
			Expect(inbound[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x800}))
			Expect(inbound[1]).To(Equal(generictables.Rule{
				Match: iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.NflogAction{
					Group:  1,
					Prefix: "DPI0|default.foo",
				},
			}))
			Expect(inbound[2]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.DropAction{},
			}))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Outbound deny rules should be correctly rendered within a policy",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			denyRule.Action = "deny"
			policyID := &types.PolicyID{
				Tier: "default",
				Name: "default.foo",
			}
			policy := &proto.Policy{
				Namespace:     "",
				InboundRules:  []*proto.Rule{},
				OutboundRules: []*proto.Rule{denyRule},
				Untracked:     false,
				PreDnat:       false,
			}

			chains := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			Expect(chains[0].Name).To(Equal("cali-pi-default/default.foo"))
			Expect(chains[1].Name).To(Equal("cali-po-default/default.foo"))

			inbound := chains[0].Rules
			outbound := chains[1].Rules

			Expect(inbound).To(ConsistOf(
				generictables.Rule{
					Comment: []string{"Policy default.foo ingress"},
				}))
			Expect(outbound).To(HaveLen(2))
			Expect(outbound[0].Match.Render()).To(Equal(expMatch))
			Expect(outbound[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x800}))
			Expect(outbound[1]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.DropAction{},
			}))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Outbound deny rules should be correctly rendered within a policy with flowlogs enabled",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = true
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			denyRule.Action = "deny"
			policyID := &types.PolicyID{
				Tier: "default",
				Name: "default.foo",
			}
			policy := &proto.Policy{
				Namespace:     "",
				InboundRules:  []*proto.Rule{},
				OutboundRules: []*proto.Rule{denyRule},
				Untracked:     false,
				PreDnat:       false,
			}

			chains := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			Expect(chains[0].Name).To(Equal("cali-pi-default/default.foo"))
			Expect(chains[1].Name).To(Equal("cali-po-default/default.foo"))

			inbound := chains[0].Rules
			outbound := chains[1].Rules

			Expect(inbound).To(ConsistOf(
				generictables.Rule{
					Comment: []string{"Policy default.foo ingress"},
				}))
			Expect(outbound).To(HaveLen(3))
			Expect(outbound[0].Match.Render()).To(Equal(expMatch))
			Expect(outbound[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x800}))
			Expect(outbound[1]).To(Equal(generictables.Rule{
				Match: iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.NflogAction{
					Group:  2,
					Prefix: "DPE0|default.foo",
				},
			}))
			Expect(outbound[2]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.DropAction{},
			}))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Inbound deny rules should be correctly rendered within a staged policy",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			denyRule.Action = "deny"
			policyID := &types.PolicyID{
				Tier: "default",
				Name: "staged:default.foo",
			}
			policy := &proto.Policy{
				Namespace:     "",
				InboundRules:  []*proto.Rule{denyRule},
				OutboundRules: []*proto.Rule{},
				Untracked:     false,
				PreDnat:       false,
			}

			chains := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			// Staged policies should be skipped.
			Expect(chains).To(HaveLen(0))

			// Should be same with flowlogs enabled.
			rrConfigNormal.FlowLogsEnabled = true
			renderer = NewRenderer(rrConfigNormal)
			chainsWithFlowLogs := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			// Staged policies should be skipped.
			Expect(chainsWithFlowLogs).To(HaveLen(0))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Outbound deny rules should be correctly rendered within a staged policy",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			denyRule.Action = "deny"
			policyID := &types.PolicyID{
				Tier: "default",
				Name: "staged:default.foo",
			}
			policy := &proto.Policy{
				Namespace:     "",
				InboundRules:  []*proto.Rule{},
				OutboundRules: []*proto.Rule{denyRule},
				Untracked:     false,
				PreDnat:       false,
			}

			chains := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			// Staged policies should be skipped.
			Expect(chains).To(HaveLen(0))

			// Should be same with flowlogs enabled.
			rrConfigNormal.FlowLogsEnabled = true
			renderer = NewRenderer(rrConfigNormal)
			chainsWithFlowLogs := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			// Staged policies should be skipped.
			Expect(chainsWithFlowLogs).To(HaveLen(0))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Inbound allow rules should be correctly rendered within a staged policy",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			policyID := &types.PolicyID{
				Tier: "default",
				Name: "staged:default.foo",
			}
			policy := &proto.Policy{
				Namespace:     "",
				InboundRules:  []*proto.Rule{denyRule},
				OutboundRules: []*proto.Rule{},
				Untracked:     false,
				PreDnat:       false,
			}

			chains := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			// Staged policies should be skipped.
			Expect(chains).To(HaveLen(0))

			// Should be same with flowlogs enabled.
			rrConfigNormal.FlowLogsEnabled = true
			renderer = NewRenderer(rrConfigNormal)
			chainsWithFlowLogs := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			// Staged policies should be skipped.
			Expect(chainsWithFlowLogs).To(HaveLen(0))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Outbound allow rules should be correctly rendered within a staged policy",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			denyRule := in
			policyID := &types.PolicyID{
				Tier: "default",
				Name: "staged:default.foo",
			}
			policy := &proto.Policy{
				Namespace:     "",
				InboundRules:  []*proto.Rule{},
				OutboundRules: []*proto.Rule{denyRule},
				Untracked:     false,
				PreDnat:       false,
			}

			chains := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			// Staged policies should be skipped.
			Expect(chains).To(HaveLen(0))

			// Should be same with flowlogs enabled.
			rrConfigNormal.FlowLogsEnabled = true
			renderer = NewRenderer(rrConfigNormal)
			chainsWithFlowLogs := renderer.PolicyToIptablesChains(policyID, policy, uint8(ipVer))
			// Staged policies should be skipped.
			Expect(chainsWithFlowLogs).To(HaveLen(0))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Deny (REJECT) rules should be correctly rendered",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigReject := rrConfigNormal
			rrConfigReject.FilterDenyAction = "REJECT"
			rrConfigReject.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigReject)
			denyRule := in
			denyRule.Action = "deny"
			rules := renderer.ProtoRuleToIptablesRules(denyRule, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			// For deny, should be one match rule that just does the REJECT.
			Expect(len(rules)).To(Equal(2))
			Expect(rules[0].Match.Render()).To(Equal(expMatch))
			Expect(rules[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x800}))
			Expect(rules[1]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.RejectAction{},
			}))
		},
		ruleTestData...,
	)

	DescribeTable(
		"Deny (REJECT) rules should be correctly rendered with flowlogs enabled",
		func(ipVer int, in *proto.Rule, expMatch string) {
			rrConfigReject := rrConfigNormal
			rrConfigReject.FilterDenyAction = "REJECT"
			rrConfigReject.FlowLogsEnabled = true
			renderer := NewRenderer(rrConfigReject)
			denyRule := in
			denyRule.Action = "deny"
			rules := renderer.ProtoRuleToIptablesRules(denyRule, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			// For deny, should be one match rule that just does the REJECT.
			Expect(len(rules)).To(Equal(3))
			Expect(rules[0].Match.Render()).To(Equal(expMatch))
			Expect(rules[0].Action).To(Equal(iptables.SetMarkAction{Mark: 0x800}))
			Expect(rules[1]).To(Equal(generictables.Rule{
				Match: iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.NflogAction{
					Group:  1,
					Prefix: "DPI0|default.foo",
				},
			}))
			Expect(rules[2]).To(Equal(generictables.Rule{
				Match:  iptables.Match().MarkSingleBitSet(0x800),
				Action: iptables.RejectAction{},
			}))
		},
		ruleTestData...,
	)

	const (
		clearBothMarksRule       = "-A test --jump MARK --set-mark 0x0/0x600"
		preSetAllBlocksMarkRule  = "-A test --jump MARK --set-mark 0x200/0x600"
		allowIfAllMarkRule       = "-A test -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80"
		nflogAllowRule           = "-A test -m mark --mark 0x80/0x80 --jump NFLOG --nflog-group 1 --nflog-prefix API0|default.foo --nflog-range 80"
		allowIfAllMarkAndTCPRule = "-A test -p tcp -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80"
		allowIfAllMarkAndUDPRule = "-A test -p udp -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80"
		returnRule               = "-A test -m mark --mark 0x80/0x80 --jump RETURN"
		// allBlocksPassAndEqThisBlockPassRule is seen at the end of every positive match block
		// after the first one.  It clears the all-blocks-pass bit if the this-block-passes bit
		// is not set.
		allBlocksPassAndEqThisBlockPassRule = "-A test -m mark --mark 0/0x400 --jump MARK --set-mark 0/0x200"
	)

	DescribeTable(
		"CIDR split tests",
		func(numSrc, numNotSrc, numDst, numNotDst int, expected []string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			pRule := proto.Rule{
				SrcNet:    []string{"10.0.0.0/24", "10.0.1.0/24"}[:numSrc],
				NotSrcNet: []string{"11.0.0.0/24", "11.0.1.0/24"}[:numNotSrc],
				DstNet:    []string{"12.0.0.0/24", "12.0.1.0/24"}[:numDst],
				NotDstNet: []string{"13.0.0.0/24", "13.0.1.0/24"}[:numNotDst],
				Action:    "allow",
			}
			iptRules := renderer.ProtoRuleToIptablesRules(&pRule, 4,
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			rendered := []string{}
			for _, ir := range iptRules {
				s := iptables.NewIptablesRenderer("").RenderAppend(&ir, "test", "", &environment.Features{})
				rendered = append(rendered, s)
			}
			Expect(rendered).To(Equal(expected))
		},
		// Simple overflow of each match criteria...

		Entry("2 src, 0 !src, 0 dst, 0 !dst", 2, 0, 0, 0, []string{
			clearBothMarksRule,
			"-A test --source 10.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --source 10.0.1.0/24 --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkRule,
			returnRule,
		}),
		Entry("0 src, 2 !src, 0 dst, 0 !dst", 0, 2, 0, 0, []string{
			preSetAllBlocksMarkRule,
			"-A test --source 11.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --source 11.0.1.0/24 --jump MARK --set-mark 0/0x200",
			allowIfAllMarkRule,
			returnRule,
		}),
		Entry("0 src, 0 !src, 2 dst, 0 !dst", 0, 0, 2, 0, []string{
			clearBothMarksRule,
			"-A test --destination 12.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --destination 12.0.1.0/24 --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkRule,
			returnRule,
		}),
		Entry("0 src, 0 !src, 0 dst, 2 !dst", 0, 0, 0, 2, []string{
			preSetAllBlocksMarkRule,
			"-A test --destination 13.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 13.0.1.0/24 --jump MARK --set-mark 0/0x200",
			allowIfAllMarkRule,
			returnRule,
		}),

		// Overflow of source even though each type would fit.
		Entry("1 src, 1 !src, 0 dst, 0 !dst", 1, 1, 0, 0, []string{
			preSetAllBlocksMarkRule,
			"-A test --source 11.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --source 10.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			returnRule,
		}),
		Entry("2 src, 1 !src, 0 dst, 0 !dst", 2, 1, 0, 0, []string{
			clearBothMarksRule,
			"-A test --source 10.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --source 10.0.1.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test ! --source 11.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			returnRule,
		}),

		// Ditto for dest.
		Entry("0 src, 0 !src, 1 dst, 1 !dst", 0, 0, 1, 1, []string{
			preSetAllBlocksMarkRule,
			"-A test --destination 13.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 12.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			returnRule,
		}),
		Entry("0 src, 0 !src, 2 dst, 1 !dst", 0, 0, 2, 1, []string{
			clearBothMarksRule,
			"-A test --destination 12.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --destination 12.0.1.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test ! --destination 13.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			returnRule,
		}),

		// One of everything; only !src and !dst should overflow
		Entry("1 src, 1 !src, 1 dst, 1 !dst", 1, 1, 1, 1, []string{
			preSetAllBlocksMarkRule,
			"-A test --source 11.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 13.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --source 10.0.0.0/24 --destination 12.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			returnRule,
		}),

		// Two of everything; everything overflows.
		Entry("2 src, 2 !src, 2 dst, 2 !dst", 2, 2, 2, 2, []string{
			// Both marks start as 0.
			clearBothMarksRule,

			// Source match directly sets the AllBlocks bit.
			"-A test --source 10.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --source 10.0.1.0/24 --jump MARK --set-mark 0x200/0x200",

			// Then the Dest match sets a scratch bit.
			"-A test --destination 12.0.0.0/24 --jump MARK --set-mark 0x400/0x400",
			"-A test --destination 12.0.1.0/24 --jump MARK --set-mark 0x400/0x400",
			// If the scratch bit isn't set then we clear the AllBlocks bit.
			allBlocksPassAndEqThisBlockPassRule,

			// The negated matches clear the AllBlocks bit directly if they match.
			"-A test --source 11.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --source 11.0.1.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 13.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 13.0.1.0/24 --jump MARK --set-mark 0/0x200",

			allowIfAllMarkRule,
			returnRule,
		}),
	)

	DescribeTable(
		"CIDR split tests with flowlogs enabled",
		func(numSrc, numNotSrc, numDst, numNotDst int, expected []string) {
			rrConfigNormal.FlowLogsEnabled = true
			renderer := NewRenderer(rrConfigNormal)
			pRule := proto.Rule{
				SrcNet:    []string{"10.0.0.0/24", "10.0.1.0/24"}[:numSrc],
				NotSrcNet: []string{"11.0.0.0/24", "11.0.1.0/24"}[:numNotSrc],
				DstNet:    []string{"12.0.0.0/24", "12.0.1.0/24"}[:numDst],
				NotDstNet: []string{"13.0.0.0/24", "13.0.1.0/24"}[:numNotDst],
				Action:    "allow",
			}
			iptRules := renderer.ProtoRuleToIptablesRules(&pRule, 4,
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			rendered := []string{}
			for _, ir := range iptRules {
				s := iptables.NewIptablesRenderer("").RenderAppend(&ir, "test", "", &environment.Features{})
				rendered = append(rendered, s)
			}
			Expect(rendered).To(Equal(expected))
		},
		// Simple overflow of each match criteria...

		Entry("2 src, 0 !src, 0 dst, 0 !dst", 2, 0, 0, 0, []string{
			clearBothMarksRule,
			"-A test --source 10.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --source 10.0.1.0/24 --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkRule,
			nflogAllowRule,
			returnRule,
		}),
		Entry("0 src, 2 !src, 0 dst, 0 !dst", 0, 2, 0, 0, []string{
			preSetAllBlocksMarkRule,
			"-A test --source 11.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --source 11.0.1.0/24 --jump MARK --set-mark 0/0x200",
			allowIfAllMarkRule,
			nflogAllowRule,
			returnRule,
		}),
		Entry("0 src, 0 !src, 2 dst, 0 !dst", 0, 0, 2, 0, []string{
			clearBothMarksRule,
			"-A test --destination 12.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --destination 12.0.1.0/24 --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkRule,
			nflogAllowRule,
			returnRule,
		}),
		Entry("0 src, 0 !src, 0 dst, 2 !dst", 0, 0, 0, 2, []string{
			preSetAllBlocksMarkRule,
			"-A test --destination 13.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 13.0.1.0/24 --jump MARK --set-mark 0/0x200",
			allowIfAllMarkRule,
			nflogAllowRule,
			returnRule,
		}),

		// Overflow of source even though each type would fit.
		Entry("1 src, 1 !src, 0 dst, 0 !dst", 1, 1, 0, 0, []string{
			preSetAllBlocksMarkRule,
			"-A test --source 11.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --source 10.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		}),
		Entry("2 src, 1 !src, 0 dst, 0 !dst", 2, 1, 0, 0, []string{
			clearBothMarksRule,
			"-A test --source 10.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --source 10.0.1.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test ! --source 11.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		}),

		// Ditto for dest.
		Entry("0 src, 0 !src, 1 dst, 1 !dst", 0, 0, 1, 1, []string{
			preSetAllBlocksMarkRule,
			"-A test --destination 13.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 12.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		}),
		Entry("0 src, 0 !src, 2 dst, 1 !dst", 0, 0, 2, 1, []string{
			clearBothMarksRule,
			"-A test --destination 12.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --destination 12.0.1.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test ! --destination 13.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		}),

		// One of everything; only !src and !dst should overflow
		Entry("1 src, 1 !src, 1 dst, 1 !dst", 1, 1, 1, 1, []string{
			preSetAllBlocksMarkRule,
			"-A test --source 11.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 13.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --source 10.0.0.0/24 --destination 12.0.0.0/24 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		}),

		// Two of everything; everything overflows.
		Entry("2 src, 2 !src, 2 dst, 2 !dst", 2, 2, 2, 2, []string{
			// Both marks start as 0.
			clearBothMarksRule,

			// Source match directly sets the AllBlocks bit.
			"-A test --source 10.0.0.0/24 --jump MARK --set-mark 0x200/0x200",
			"-A test --source 10.0.1.0/24 --jump MARK --set-mark 0x200/0x200",

			// Then the Dest match sets a scratch bit.
			"-A test --destination 12.0.0.0/24 --jump MARK --set-mark 0x400/0x400",
			"-A test --destination 12.0.1.0/24 --jump MARK --set-mark 0x400/0x400",
			// If the scratch bit isn't set then we clear the AllBlocks bit.
			allBlocksPassAndEqThisBlockPassRule,

			// The negated matches clear the AllBlocks bit directly if they match.
			"-A test --source 11.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --source 11.0.1.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 13.0.0.0/24 --jump MARK --set-mark 0/0x200",
			"-A test --destination 13.0.1.0/24 --jump MARK --set-mark 0/0x200",

			allowIfAllMarkRule,
			nflogAllowRule,
			returnRule,
		}),
	)

	namedPortEntry := func(description string, pRule *proto.Rule, expected ...string) TableEntry {
		return Entry(
			description+", input = "+pRule.String(),
			pRule,
			expected,
		)
	}

	DescribeTable(
		"Named port tests",
		func(pRule *proto.Rule, expected []string) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			iptRules := renderer.ProtoRuleToIptablesRules(pRule, 4,
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			rendered := []string{}
			for _, ir := range iptRules {
				s := iptables.NewIptablesRenderer("").RenderAppend(&ir, "test", "", &environment.Features{})
				rendered = append(rendered, s)
			}
			Expect(rendered).To(Equal(expected))
		},

		// Positive source matches only.
		namedPortEntry(
			"Named port on its own rendered as single rule",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
			},
			"-A test -p tcp -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x80/0x80",
			returnRule,
		),
		namedPortEntry(
			"Two named ports need a block",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				SrcNamedPortIpSetIds: []string{"ipset-1", "ipset-2"},
			},
			clearBothMarksRule,
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-2 src,src --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndUDPRule,
			returnRule,
		),
		namedPortEntry(
			"Multiple named + numeric ports",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				SrcNamedPortIpSetIds: []string{"ipset-1", "ipset-2", "ipset-3"},
			},
			clearBothMarksRule,
			"-A test -p tcp -m multiport --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p tcp -m multiport --source-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-2 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-3 src,src --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndTCPRule,
			returnRule,
		),
		namedPortEntry(
			"Overflow of numeric ports",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				SrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
			},
			clearBothMarksRule,
			"-A test -p udp -m multiport --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p udp -m multiport --source-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndUDPRule,
			returnRule,
		),

		// Positive dest matches only.
		namedPortEntry(
			"Named + numeric ports need a block",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p tcp -m multiport --source-ports 1:2 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndTCPRule,
			returnRule,
		),
		namedPortEntry(
			"Single named port fits in rule",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				DstNamedPortIpSetIds: []string{"ipset-1"},
			},
			"-A test -p tcp -m set --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x80/0x80",
			returnRule,
		),
		namedPortEntry(
			"Two named ports need a block",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				DstNamedPortIpSetIds: []string{"ipset-1", "ipset-2"},
			},
			clearBothMarksRule,
			"-A test -m set --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-2 dst,dst --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndTCPRule,
			returnRule,
		),
		namedPortEntry(
			"Named + numeric ports need a block",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				DstPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				DstNamedPortIpSetIds: []string{"ipset-1"},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p tcp -m multiport --destination-ports 1:2 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndTCPRule,
			returnRule,
		),

		// Positive source and dest matches together.
		namedPortEntry(
			"Positive source needs block only",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				SrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
				DstPorts:             []*proto.PortRange{{First: 3, Last: 4}},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p udp -m multiport --source-ports 1:2 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -p udp -m multiport --destination-ports 3:4 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			returnRule,
		),
		namedPortEntry(
			"Positive dest needs block only",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				DstPorts:             []*proto.PortRange{{First: 3, Last: 4}},
				DstNamedPortIpSetIds: []string{"ipset-1"},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p tcp -m multiport --destination-ports 3:4 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x200/0x200",
			// Source port rendered directly into the main rule.
			"-A test -p tcp -m multiport --source-ports 1:2 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			returnRule,
		),
		namedPortEntry(
			"Positive source and dest need blocks",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
				DstPorts:             []*proto.PortRange{{First: 3, Last: 4}},
				DstNamedPortIpSetIds: []string{"ipset-2"},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p tcp -m multiport --source-ports 1:2 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			// Second block uses per-block bit.
			"-A test -p tcp -m multiport --destination-ports 3:4 --jump MARK --set-mark 0x400/0x400",
			"-A test -m set --match-set cali40ipset-2 dst,dst --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,
			allowIfAllMarkAndTCPRule,
			returnRule,
		),
		namedPortEntry(
			"Overflow of numeric ports",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				DstPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
			},
			clearBothMarksRule,
			"-A test -p udp -m multiport --destination-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p udp -m multiport --destination-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndUDPRule,
			returnRule,
		),

		// Negative src matches.
		namedPortEntry(
			"Negated named + numeric ports rendered in single rule",
			&proto.Rule{
				Protocol:                &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				NotSrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				NotSrcNamedPortIpSetIds: []string{"ipset-1"},
			},
			"-A test -p tcp -m multiport ! --source-ports 1:2 "+
				"-m set ! --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x80/0x80",
			returnRule,
		),
		namedPortEntry(
			"Multiple negated named + numeric ports rendered in single rule",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				NotSrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				NotSrcNamedPortIpSetIds: []string{"ipset-1", "ipset-2", "ipset-3"},
			},
			"-A test -p tcp "+
				"-m multiport ! --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 "+
				"-m multiport ! --source-ports 15:16 "+ // Overflow to new multiport.
				"-m set ! --match-set cali40ipset-1 src,src "+
				"-m set ! --match-set cali40ipset-2 src,src "+
				"-m set ! --match-set cali40ipset-3 src,src "+
				"--jump MARK --set-mark 0x80/0x80",
			returnRule,
		),

		// Negative dst matches.
		namedPortEntry(
			"Negated named + numeric ports rendered in single rule",
			&proto.Rule{
				Protocol:                &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				NotDstPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				NotDstNamedPortIpSetIds: []string{"ipset-1"},
			},
			"-A test -p tcp -m multiport ! --destination-ports 1:2 "+
				"-m set ! --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x80/0x80",
			returnRule,
		),
		namedPortEntry(
			"Multiple negated named + numeric ports rendered in single rule",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				NotDstPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				NotDstNamedPortIpSetIds: []string{"ipset-1", "ipset-2", "ipset-3"},
			},
			"-A test -p udp "+
				"-m multiport ! --destination-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 "+
				"-m multiport ! --destination-ports 15:16 "+ // Overflow to new multiport.
				"-m set ! --match-set cali40ipset-1 dst,dst "+
				"-m set ! --match-set cali40ipset-2 dst,dst "+
				"-m set ! --match-set cali40ipset-3 dst,dst "+
				"--jump MARK --set-mark 0x80/0x80",
			returnRule,
		),

		// CIDRs + named ports.
		namedPortEntry(
			"numeric, named ports and CIDRs",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				SrcNamedPortIpSetIds: []string{"ipset-1", "ipset-2", "ipset-3"},
				SrcNet:               []string{"10.1.0.0/16", "11.0.0.0/8"},
			},
			clearBothMarksRule,
			"-A test -p tcp -m multiport --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p tcp -m multiport --source-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-2 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-3 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test --source 10.1.0.0/16 --jump MARK --set-mark 0x400/0x400",
			"-A test --source 11.0.0.0/8 --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,
			allowIfAllMarkAndTCPRule,
			returnRule,
		),

		// CIDRs + positive and negated named ports.
		namedPortEntry(
			"positive and negatednumeric, named ports and CIDRs",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
				SrcNet:               []string{"10.1.0.0/16", "11.0.0.0/8"},
				NotSrcPorts: []*proto.PortRange{
					{First: 101, Last: 101},
				},
				NotSrcNamedPortIpSetIds: []string{"ipset-3"},
				NotSrcNet:               []string{"14.1.0.0/16", "15.0.0.0/8"},

				DstPorts: []*proto.PortRange{
					{First: 2, Last: 3},
				},
				DstNamedPortIpSetIds: []string{"ipset-2"},
				DstNet:               []string{"12.1.0.0/16", "13.0.0.0/8"},
				NotDstPorts: []*proto.PortRange{
					{First: 201, Last: 201},
				},
				NotDstNamedPortIpSetIds: []string{"ipset-4"},
				NotDstNet:               []string{"16.1.0.0/16", "17.0.0.0/8"},
			},
			clearBothMarksRule,
			// Positive source port match block.
			"-A test -p tcp -m multiport --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p tcp -m multiport --source-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",

			// Positive destination port match block..
			"-A test -p tcp -m multiport --destination-ports 2:3 --jump MARK --set-mark 0x400/0x400",
			"-A test -m set --match-set cali40ipset-2 dst,dst --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,

			// Positive source CIDRs.
			"-A test --source 10.1.0.0/16 --jump MARK --set-mark 0x400/0x400",
			"-A test --source 11.0.0.0/8 --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,

			// Positive dest CIDRs.
			"-A test --destination 12.1.0.0/16 --jump MARK --set-mark 0x400/0x400",
			"-A test --destination 13.0.0.0/8 --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,

			// Negative source CIDRs.
			"-A test --source 14.1.0.0/16 --jump MARK --set-mark 0/0x200",
			"-A test --source 15.0.0.0/8 --jump MARK --set-mark 0/0x200",

			// Negative dest CIDRs.
			"-A test --destination 16.1.0.0/16 --jump MARK --set-mark 0/0x200",
			"-A test --destination 17.0.0.0/8 --jump MARK --set-mark 0/0x200",

			// Negative port matches can be inlined into the main rule.
			"-A test -p tcp "+
				"-m multiport ! --source-ports 101 "+
				"-m set ! --match-set cali40ipset-3 src,src "+
				"-m multiport ! --destination-ports 201 "+
				"-m set ! --match-set cali40ipset-4 dst,dst "+
				"-m mark --mark 0x200/0x200 "+
				"--jump MARK --set-mark 0x80/0x80",
			returnRule,
		),
	)

	DescribeTable(
		"Named port tests with flowlogs enabled",
		func(pRule *proto.Rule, expected []string) {
			rrConfigNormal.FlowLogsEnabled = true
			renderer := NewRenderer(rrConfigNormal)
			iptRules := renderer.ProtoRuleToIptablesRules(pRule, 4,
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			rendered := []string{}
			for _, ir := range iptRules {
				s := iptables.NewIptablesRenderer("").RenderAppend(&ir, "test", "", &environment.Features{})
				rendered = append(rendered, s)
			}
			Expect(rendered).To(Equal(expected))
		},

		// Positive source matches only.
		namedPortEntry(
			"Named port on its own rendered as single rule",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
			},
			"-A test -p tcp -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Two named ports need a block",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				SrcNamedPortIpSetIds: []string{"ipset-1", "ipset-2"},
			},
			clearBothMarksRule,
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-2 src,src --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndUDPRule,
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Multiple named + numeric ports",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				SrcNamedPortIpSetIds: []string{"ipset-1", "ipset-2", "ipset-3"},
			},
			clearBothMarksRule,
			"-A test -p tcp -m multiport --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p tcp -m multiport --source-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-2 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-3 src,src --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndTCPRule,
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Overflow of numeric ports",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				SrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
			},
			clearBothMarksRule,
			"-A test -p udp -m multiport --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p udp -m multiport --source-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndUDPRule,
			nflogAllowRule,
			returnRule,
		),

		// Positive dest matches only.
		namedPortEntry(
			"Named + numeric ports need a block",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p tcp -m multiport --source-ports 1:2 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndTCPRule,
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Single named port fits in rule",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				DstNamedPortIpSetIds: []string{"ipset-1"},
			},
			"-A test -p tcp -m set --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Two named ports need a block",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				DstNamedPortIpSetIds: []string{"ipset-1", "ipset-2"},
			},
			clearBothMarksRule,
			"-A test -m set --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-2 dst,dst --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndTCPRule,
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Named + numeric ports need a block",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				DstPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				DstNamedPortIpSetIds: []string{"ipset-1"},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p tcp -m multiport --destination-ports 1:2 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndTCPRule,
			nflogAllowRule,
			returnRule,
		),

		// Positive source and dest matches together.
		namedPortEntry(
			"Positive source needs block only",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				SrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
				DstPorts:             []*proto.PortRange{{First: 3, Last: 4}},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p udp -m multiport --source-ports 1:2 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -p udp -m multiport --destination-ports 3:4 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Positive dest needs block only",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				DstPorts:             []*proto.PortRange{{First: 3, Last: 4}},
				DstNamedPortIpSetIds: []string{"ipset-1"},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p tcp -m multiport --destination-ports 3:4 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x200/0x200",
			// Source port rendered directly into the main rule.
			"-A test -p tcp -m multiport --source-ports 1:2 -m mark --mark 0x200/0x200 --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Positive source and dest need blocks",
			&proto.Rule{
				Protocol:             &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
				DstPorts:             []*proto.PortRange{{First: 3, Last: 4}},
				DstNamedPortIpSetIds: []string{"ipset-2"},
			},
			clearBothMarksRule,
			// Need to "OR" the named port and multiport matches together.
			// First positive block so it sets the all bit directly.
			"-A test -p tcp -m multiport --source-ports 1:2 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			// Second block uses per-block bit.
			"-A test -p tcp -m multiport --destination-ports 3:4 --jump MARK --set-mark 0x400/0x400",
			"-A test -m set --match-set cali40ipset-2 dst,dst --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,
			allowIfAllMarkAndTCPRule,
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Overflow of numeric ports",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				DstPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
			},
			clearBothMarksRule,
			"-A test -p udp -m multiport --destination-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p udp -m multiport --destination-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			allowIfAllMarkAndUDPRule,
			nflogAllowRule,
			returnRule,
		),

		// Negative src matches.
		namedPortEntry(
			"Negated named + numeric ports rendered in single rule",
			&proto.Rule{
				Protocol:                &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				NotSrcPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				NotSrcNamedPortIpSetIds: []string{"ipset-1"},
			},
			"-A test -p tcp -m multiport ! --source-ports 1:2 "+
				"-m set ! --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Multiple negated named + numeric ports rendered in single rule",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				NotSrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				NotSrcNamedPortIpSetIds: []string{"ipset-1", "ipset-2", "ipset-3"},
			},
			"-A test -p tcp "+
				"-m multiport ! --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 "+
				"-m multiport ! --source-ports 15:16 "+ // Overflow to new multiport.
				"-m set ! --match-set cali40ipset-1 src,src "+
				"-m set ! --match-set cali40ipset-2 src,src "+
				"-m set ! --match-set cali40ipset-3 src,src "+
				"--jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		),

		// Negative dst matches.
		namedPortEntry(
			"Negated named + numeric ports rendered in single rule",
			&proto.Rule{
				Protocol:                &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				NotDstPorts:             []*proto.PortRange{{First: 1, Last: 2}},
				NotDstNamedPortIpSetIds: []string{"ipset-1"},
			},
			"-A test -p tcp -m multiport ! --destination-ports 1:2 "+
				"-m set ! --match-set cali40ipset-1 dst,dst --jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		),
		namedPortEntry(
			"Multiple negated named + numeric ports rendered in single rule",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "udp"}},
				NotDstPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				NotDstNamedPortIpSetIds: []string{"ipset-1", "ipset-2", "ipset-3"},
			},
			"-A test -p udp "+
				"-m multiport ! --destination-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 "+
				"-m multiport ! --destination-ports 15:16 "+ // Overflow to new multiport.
				"-m set ! --match-set cali40ipset-1 dst,dst "+
				"-m set ! --match-set cali40ipset-2 dst,dst "+
				"-m set ! --match-set cali40ipset-3 dst,dst "+
				"--jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		),

		// CIDRs + named ports.
		namedPortEntry(
			"numeric, named ports and CIDRs",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				SrcNamedPortIpSetIds: []string{"ipset-1", "ipset-2", "ipset-3"},
				SrcNet:               []string{"10.1.0.0/16", "11.0.0.0/8"},
			},
			clearBothMarksRule,
			"-A test -p tcp -m multiport --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p tcp -m multiport --source-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-2 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-3 src,src --jump MARK --set-mark 0x200/0x200",
			"-A test --source 10.1.0.0/16 --jump MARK --set-mark 0x400/0x400",
			"-A test --source 11.0.0.0/8 --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,
			allowIfAllMarkAndTCPRule,
			nflogAllowRule,
			returnRule,
		),

		// CIDRs + positive and negated named ports.
		namedPortEntry(
			"positive and negatednumeric, named ports and CIDRs",
			&proto.Rule{
				Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
				SrcPorts: []*proto.PortRange{
					{First: 1, Last: 2},
					{First: 3, Last: 4},
					{First: 5, Last: 6},
					{First: 7, Last: 8},
					{First: 9, Last: 10},
					{First: 11, Last: 12},
					{First: 13, Last: 14},
					{First: 15, Last: 16},
				},
				SrcNamedPortIpSetIds: []string{"ipset-1"},
				SrcNet:               []string{"10.1.0.0/16", "11.0.0.0/8"},
				NotSrcPorts: []*proto.PortRange{
					{First: 101, Last: 101},
				},
				NotSrcNamedPortIpSetIds: []string{"ipset-3"},
				NotSrcNet:               []string{"14.1.0.0/16", "15.0.0.0/8"},

				DstPorts: []*proto.PortRange{
					{First: 2, Last: 3},
				},
				DstNamedPortIpSetIds: []string{"ipset-2"},
				DstNet:               []string{"12.1.0.0/16", "13.0.0.0/8"},
				NotDstPorts: []*proto.PortRange{
					{First: 201, Last: 201},
				},
				NotDstNamedPortIpSetIds: []string{"ipset-4"},
				NotDstNet:               []string{"16.1.0.0/16", "17.0.0.0/8"},
			},
			clearBothMarksRule,
			// Positive source port match block.
			"-A test -p tcp -m multiport --source-ports 1:2,3:4,5:6,7:8,9:10,11:12,13:14 --jump MARK --set-mark 0x200/0x200",
			"-A test -p tcp -m multiport --source-ports 15:16 --jump MARK --set-mark 0x200/0x200",
			"-A test -m set --match-set cali40ipset-1 src,src --jump MARK --set-mark 0x200/0x200",

			// Positive destination port match block..
			"-A test -p tcp -m multiport --destination-ports 2:3 --jump MARK --set-mark 0x400/0x400",
			"-A test -m set --match-set cali40ipset-2 dst,dst --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,

			// Positive source CIDRs.
			"-A test --source 10.1.0.0/16 --jump MARK --set-mark 0x400/0x400",
			"-A test --source 11.0.0.0/8 --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,

			// Positive dest CIDRs.
			"-A test --destination 12.1.0.0/16 --jump MARK --set-mark 0x400/0x400",
			"-A test --destination 13.0.0.0/8 --jump MARK --set-mark 0x400/0x400",
			allBlocksPassAndEqThisBlockPassRule,

			// Negative source CIDRs.
			"-A test --source 14.1.0.0/16 --jump MARK --set-mark 0/0x200",
			"-A test --source 15.0.0.0/8 --jump MARK --set-mark 0/0x200",

			// Negative dest CIDRs.
			"-A test --destination 16.1.0.0/16 --jump MARK --set-mark 0/0x200",
			"-A test --destination 17.0.0.0/8 --jump MARK --set-mark 0/0x200",

			// Negative port matches can be inlined into the main rule.
			"-A test -p tcp "+
				"-m multiport ! --source-ports 101 "+
				"-m set ! --match-set cali40ipset-3 src,src "+
				"-m multiport ! --destination-ports 201 "+
				"-m set ! --match-set cali40ipset-4 dst,dst "+
				"-m mark --mark 0x200/0x200 "+
				"--jump MARK --set-mark 0x80/0x80",
			nflogAllowRule,
			returnRule,
		),
	)

	Context("with default renderer", func() {
		var renderer *DefaultRuleRenderer
		BeforeEach(func() {
			rrConfigNormal.FlowLogsEnabled = false
			renderer = NewRenderer(rrConfigNormal).(*DefaultRuleRenderer)
		})

		It("should skip rules of incorrect IP version", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{IpVersion: 4}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed source CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{SrcNet: []string{"10.0.0.1"}}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed source CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{SrcNet: []string{"feed::beef"}}}, 4,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed dest CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{DstNet: []string{"10.0.0.1"}}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed dest CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{DstNet: []string{"feed::beef"}}}, 4,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed negated source CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{NotSrcNet: []string{"10.0.0.1"}}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed negated source CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{NotSrcNet: []string{"feed::beef"}}}, 4,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed negated dest CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{NotDstNet: []string{"10.0.0.1"}}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed negated dest CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{NotDstNet: []string{"feed::beef"}}}, 4,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})
	})

	Context("with default renderer and flow logs enabled", func() {
		var renderer *DefaultRuleRenderer
		BeforeEach(func() {
			rrConfigNormal.FlowLogsEnabled = true
			renderer = NewRenderer(rrConfigNormal).(*DefaultRuleRenderer)
		})

		It("should skip rules of incorrect IP version", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{IpVersion: 4}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed source CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{SrcNet: []string{"10.0.0.1"}}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed source CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{SrcNet: []string{"feed::beef"}}}, 4,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed dest CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{DstNet: []string{"10.0.0.1"}}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed dest CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{DstNet: []string{"feed::beef"}}}, 4,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed negated source CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{NotSrcNet: []string{"10.0.0.1"}}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed negated source CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{NotSrcNet: []string{"feed::beef"}}}, 4,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed negated dest CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{NotDstNet: []string{"10.0.0.1"}}}, 6,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})

		It("should skip with mixed negated dest CIDR matches", func() {
			rules := renderer.ProtoRulesToIptablesRules([]*proto.Rule{{NotDstNet: []string{"feed::beef"}}}, 4,
				RuleOwnerTypePolicy, RuleDirIngress, "default.foo", false)
			Expect(rules).To(BeEmpty())
		})
	})
})

var _ = Describe("Filtered rules (negated catch-all CIDR validation)", func() {
	rrConfigNormal := Config{
		IPIPEnabled:       true,
		IPIPTunnelAddress: nil,
		IPSetConfigV4:     ipsets.NewIPVersionConfig(ipsets.IPFamilyV4, "cali", nil, nil),
		IPSetConfigV6:     ipsets.NewIPVersionConfig(ipsets.IPFamilyV6, "cali", nil, nil),
		MarkAccept:        0x80,
		MarkPass:          0x100,
		MarkScratch0:      0x200,
		MarkScratch1:      0x400,
		MarkDrop:          0x800,
		MarkEndpoint:      0xff000,
		LogPrefix:         "calico-packet",
	}

	DescribeTable(
		"Rules with catch-all negated CIDRs should be filtered out",
		func(ipVer int, in *proto.Rule) {
			rrConfigNormal.FlowLogsEnabled = false
			renderer := NewRenderer(rrConfigNormal)
			rules := renderer.ProtoRuleToIptablesRules(in, uint8(ipVer),
				RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
			// Rules with catch-all negated CIDRs should be completely filtered out
			Expect(len(rules)).To(Equal(0))
		},
		filteredRuleTestData...,
	)
})

var _ = DescribeTable("Port split tests",
	func(in []*proto.PortRange, expected [][]*proto.PortRange) {
		Expect(SplitPortList(in)).To(Equal(expected))
	},
	Entry("nil input", ([]*proto.PortRange)(nil), ([][]*proto.PortRange)(nil)),
	Entry("empty input", []*proto.PortRange{}, ([][]*proto.PortRange)(nil)),
	Entry("single input", []*proto.PortRange{{First: 1, Last: 1}}, [][]*proto.PortRange{{{First: 1, Last: 1}}}),
	Entry("range input", []*proto.PortRange{{First: 1, Last: 10}}, [][]*proto.PortRange{{{First: 1, Last: 10}}}),
	Entry("exactly 15 single ports should give exactly one split", []*proto.PortRange{
		{First: 1, Last: 1},
		{First: 2, Last: 2},
		{First: 3, Last: 3},
		{First: 4, Last: 4},
		{First: 5, Last: 5},
		{First: 6, Last: 6},
		{First: 7, Last: 7},
		{First: 8, Last: 8},
		{First: 9, Last: 9},
		{First: 10, Last: 10},
		{First: 11, Last: 11},
		{First: 12, Last: 12},
		{First: 13, Last: 13},
		{First: 14, Last: 14},
		{First: 15, Last: 15},
	}, [][]*proto.PortRange{{
		{First: 1, Last: 1},
		{First: 2, Last: 2},
		{First: 3, Last: 3},
		{First: 4, Last: 4},
		{First: 5, Last: 5},
		{First: 6, Last: 6},
		{First: 7, Last: 7},
		{First: 8, Last: 8},
		{First: 9, Last: 9},
		{First: 10, Last: 10},
		{First: 11, Last: 11},
		{First: 12, Last: 12},
		{First: 13, Last: 13},
		{First: 14, Last: 14},
		{First: 15, Last: 15},
	}}),
	Entry("exactly 16 single ports should give exactly two splits", []*proto.PortRange{
		{First: 1, Last: 1},
		{First: 2, Last: 2},
		{First: 3, Last: 3},
		{First: 4, Last: 4},
		{First: 5, Last: 5},
		{First: 6, Last: 6},
		{First: 7, Last: 7},
		{First: 8, Last: 8},
		{First: 9, Last: 9},
		{First: 10, Last: 10},
		{First: 11, Last: 11},
		{First: 12, Last: 12},
		{First: 13, Last: 13},
		{First: 14, Last: 14},
		{First: 15, Last: 15},
		{First: 16, Last: 16},
	}, [][]*proto.PortRange{{
		{First: 1, Last: 1},
		{First: 2, Last: 2},
		{First: 3, Last: 3},
		{First: 4, Last: 4},
		{First: 5, Last: 5},
		{First: 6, Last: 6},
		{First: 7, Last: 7},
		{First: 8, Last: 8},
		{First: 9, Last: 9},
		{First: 10, Last: 10},
		{First: 11, Last: 11},
		{First: 12, Last: 12},
		{First: 13, Last: 13},
		{First: 14, Last: 14},
		{First: 15, Last: 15},
	}, {
		{First: 16, Last: 16},
	}}),
	Entry("port ranges should count for 2 single ports", []*proto.PortRange{
		{First: 1, Last: 2},
		{First: 3, Last: 4},
		{First: 5, Last: 6},
		{First: 7, Last: 8},
		{First: 9, Last: 10},
		{First: 11, Last: 12},
		{First: 13, Last: 14},
		{First: 15, Last: 15},
	}, [][]*proto.PortRange{{
		{First: 1, Last: 2},
		{First: 3, Last: 4},
		{First: 5, Last: 6},
		{First: 7, Last: 8},
		{First: 9, Last: 10},
		{First: 11, Last: 12},
		{First: 13, Last: 14},
		{First: 15, Last: 15},
	}}),
	Entry("port range straggling 15-16 should be put in second group", []*proto.PortRange{
		{First: 1, Last: 2},
		{First: 3, Last: 4},
		{First: 5, Last: 6},
		{First: 7, Last: 8},
		{First: 9, Last: 10},
		{First: 11, Last: 12},
		{First: 13, Last: 14},
		{First: 15, Last: 16},
	}, [][]*proto.PortRange{{
		{First: 1, Last: 2},
		{First: 3, Last: 4},
		{First: 5, Last: 6},
		{First: 7, Last: 8},
		{First: 9, Last: 10},
		{First: 11, Last: 12},
		{First: 13, Last: 14},
	}, {
		{First: 15, Last: 16},
	}}),
	Entry("further splits should be made in correct place", []*proto.PortRange{
		{First: 1, Last: 2},
		{First: 3, Last: 4},
		{First: 5, Last: 6},
		{First: 7, Last: 8},
		{First: 9, Last: 10},
		{First: 11, Last: 12},
		{First: 13, Last: 14},
		{First: 15, Last: 16},
		{First: 21, Last: 22},
		{First: 23, Last: 24},
		{First: 23, Last: 26},
		{First: 27, Last: 28},
		{First: 29, Last: 210},
		{First: 211, Last: 212},
		{First: 213, Last: 214},
		{First: 215, Last: 216},
	}, [][]*proto.PortRange{{
		{First: 1, Last: 2},
		{First: 3, Last: 4},
		{First: 5, Last: 6},
		{First: 7, Last: 8},
		{First: 9, Last: 10},
		{First: 11, Last: 12},
		{First: 13, Last: 14},
	}, {
		{First: 15, Last: 16},
		{First: 21, Last: 22},
		{First: 23, Last: 24},
		{First: 23, Last: 26},
		{First: 27, Last: 28},
		{First: 29, Last: 210},
		{First: 211, Last: 212},
	}, {
		{First: 213, Last: 214},
		{First: 215, Last: 216},
	}}),
)

var _ = Describe("rule metadata tests", func() {
	rule := &proto.Rule{
		Metadata: &proto.RuleMetadata{Annotations: map[string]string{
			"testkey00": "testvalue00",
			"testkey01": "testvalue01",
		}},
		Protocol: &proto.Protocol{NumberOrName: &proto.Protocol_Name{Name: "tcp"}},
	}
	rrConfigNormal := Config{
		IPIPEnabled:       true,
		IPIPTunnelAddress: nil,
		IPSetConfigV4:     ipsets.NewIPVersionConfig(ipsets.IPFamilyV4, "cali", nil, nil),
		IPSetConfigV6:     ipsets.NewIPVersionConfig(ipsets.IPFamilyV6, "cali", nil, nil),
		MarkAccept:        0x80,
		MarkPass:          0x100,
		MarkScratch0:      0x200,
		MarkScratch1:      0x400,
		MarkDrop:          0x800,
		MarkEndpoint:      0xff000,
		LogPrefix:         "calico-packet",
	}

	It("IPv4 should include annotations in comments", func() {
		rrConfigNormal.FlowLogsEnabled = false
		renderer := NewRenderer(rrConfigNormal)
		rs := renderer.ProtoRuleToIptablesRules(rule, uint8(4),
			RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
		for _, r := range rs {
			Expect(r.Comment).To(ContainElement("testkey00=testvalue00"))
			Expect(r.Comment).To(ContainElement("testkey01=testvalue01"))
		}

		// It should be the same with flowlogs enabled
		rrConfigNormal.FlowLogsEnabled = true
		renderer = NewRenderer(rrConfigNormal)
		rs1 := renderer.ProtoRuleToIptablesRules(rule, uint8(4),
			RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
		for _, r := range rs1 {
			Expect(r.Comment).To(ContainElement("testkey00=testvalue00"))
			Expect(r.Comment).To(ContainElement("testkey01=testvalue01"))
		}
	})

	It("IPv6 should include annotations in comments", func() {
		rrConfigNormal.FlowLogsEnabled = false
		renderer := NewRenderer(rrConfigNormal)
		rs := renderer.ProtoRuleToIptablesRules(rule, uint8(6),
			RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
		for _, r := range rs {
			Expect(r.Comment).To(ContainElement("testkey00=testvalue00"))
			Expect(r.Comment).To(ContainElement("testkey01=testvalue01"))
		}

		// It should be the same with flowlogs enabled
		rrConfigNormal.FlowLogsEnabled = true
		renderer = NewRenderer(rrConfigNormal)
		rs1 := renderer.ProtoRuleToIptablesRules(rule, uint8(6),
			RuleOwnerTypePolicy, RuleDirIngress, 0, "default.foo", false)
		for _, r := range rs1 {
			Expect(r.Comment).To(ContainElement("testkey00=testvalue00"))
			Expect(r.Comment).To(ContainElement("testkey01=testvalue01"))
		}
	})

	It("should include a chain name comment with a policy", func() {
		rrConfigNormal.FlowLogsEnabled = false
		renderer := NewRenderer(rrConfigNormal)
		chains := renderer.PolicyToIptablesChains(
			&types.PolicyID{
				Name: "long-policy-name-that-gets-hashed",
			},
			&proto.Policy{
				InboundRules: []*proto.Rule{{Action: "allow"}},
			},
			4,
		)
		Expect(chains).To(ConsistOf(
			&generictables.Chain{
				Name: "cali-pi-_FJ9yUkNpzshVDh2n7mg",
				Rules: []generictables.Rule{
					{
						Match:  iptables.Match(),
						Action: iptables.SetMarkAction{Mark: 0x80},
						Comment: []string{
							"Policy long-policy-name-that-gets-hashed ingress",
						},
					},
				},
			},
			&generictables.Chain{
				Name: "cali-po-_FJ9yUkNpzshVDh2n7mg",
				Rules: []generictables.Rule{
					{
						Comment: []string{
							"Policy long-policy-name-that-gets-hashed egress",
						},
					},
				},
			},
		))
	})

	It("should include a chain name comment with a policy and flowlogs enabled", func() {
		rrConfigNormal.FlowLogsEnabled = true
		renderer := NewRenderer(rrConfigNormal)
		chains := renderer.PolicyToIptablesChains(
			&types.PolicyID{
				Name: "long-policy-name-that-gets-hashed",
			},
			&proto.Policy{
				InboundRules: []*proto.Rule{{Action: "allow"}},
			},
			4,
		)
		Expect(chains).To(ConsistOf(
			&generictables.Chain{
				Name: "cali-pi-_FJ9yUkNpzshVDh2n7mg",
				Rules: []generictables.Rule{
					{
						Match:  iptables.Match(),
						Action: iptables.SetMarkAction{Mark: 0x80},
						Comment: []string{
							"Policy long-policy-name-that-gets-hashed ingress",
						},
					},
					{
						Match: iptables.Match().MarkSingleBitSet(0x80),
						Action: iptables.NflogAction{
							Group:  1,
							Prefix: "API0|long-policy-name-that-gets-hashed",
						},
						Comment: nil,
					},
				},
			},
			&generictables.Chain{
				Name: "cali-po-_FJ9yUkNpzshVDh2n7mg",
				Rules: []generictables.Rule{
					{
						Comment: []string{
							"Policy long-policy-name-that-gets-hashed egress",
						},
					},
				},
			},
		))
	})

	It("should include a chain name comment with a profile", func() {
		rrConfigNormal.FlowLogsEnabled = false
		renderer := NewRenderer(rrConfigNormal)
		inbound, outbound := renderer.ProfileToIptablesChains(
			&types.ProfileID{
				Name: "long-policy-name-that-gets-hashed",
			},
			&proto.Profile{
				InboundRules: []*proto.Rule{{Action: "allow"}},
			},
			4,
		)
		Expect([]*generictables.Chain{inbound, outbound}).To(ConsistOf(
			&generictables.Chain{
				Name: "cali-pri-_ffOMcf6pikpiZ6hgKc",
				Rules: []generictables.Rule{
					{
						Match:  iptables.Match(),
						Action: iptables.SetMarkAction{Mark: 0x80},
						Comment: []string{
							"Profile long-policy-name-that-gets-hashed ingress",
						},
					},
				},
			},
			&generictables.Chain{
				Name: "cali-pro-_ffOMcf6pikpiZ6hgKc",
				Rules: []generictables.Rule{
					{
						Comment: []string{
							"Profile long-policy-name-that-gets-hashed egress",
						},
					},
				},
			},
		))
	})

	It("should include a chain name comment with a profile and flowlogs enabled", func() {
		rrConfigNormal.FlowLogsEnabled = true
		renderer := NewRenderer(rrConfigNormal)
		inbound, outbound := renderer.ProfileToIptablesChains(
			&types.ProfileID{
				Name: "long-policy-name-that-gets-hashed",
			},
			&proto.Profile{
				InboundRules: []*proto.Rule{{Action: "allow"}},
			},
			4,
		)
		Expect([]*generictables.Chain{inbound, outbound}).To(ConsistOf(
			&generictables.Chain{
				Name: "cali-pri-_ffOMcf6pikpiZ6hgKc",
				Rules: []generictables.Rule{
					{
						Match:  iptables.Match(),
						Action: iptables.SetMarkAction{Mark: 0x80},
						Comment: []string{
							"Profile long-policy-name-that-gets-hashed ingress",
						},
					},
					{
						Match: iptables.Match().MarkSingleBitSet(0x80),
						Action: iptables.NflogAction{
							Group:  1,
							Prefix: "ARI0|long-policy-name-that-gets-hashed",
						},
						Comment: nil,
					},
				},
			},
			&generictables.Chain{
				Name: "cali-pro-_ffOMcf6pikpiZ6hgKc",
				Rules: []generictables.Rule{
					{
						Comment: []string{
							"Profile long-policy-name-that-gets-hashed egress",
						},
					},
				},
			},
		))
	})
})
