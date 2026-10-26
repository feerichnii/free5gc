package consumer

import (
	"encoding/json"
	"testing"

	"github.com/free5gc/openapi/models"
)

func TestSmPolicySubscribedDefaultQosFillsMissingArpPreemption(t *testing.T) {
	src := &models.SubscribedDefaultQos{
		Var5qi: 9,
		Arp: &models.Arp{
			PriorityLevel: 8,
		},
	}

	got := smPolicySubscribedDefaultQos(src)
	if src.Arp.PreemptCap != "" || src.Arp.PreemptVuln != "" {
		t.Fatalf("source ARP was modified: %+v", src.Arp)
	}
	if got.Arp.PreemptCap != models.PreemptionCapability_NOT_PREEMPT {
		t.Fatalf("preemptCap = %q", got.Arp.PreemptCap)
	}
	if got.Arp.PreemptVuln != models.PreemptionVulnerability_NOT_PREEMPTABLE {
		t.Fatalf("preemptVuln = %q", got.Arp.PreemptVuln)
	}
	if got.Arp.PriorityLevel != 8 || got.Var5qi != 9 {
		t.Fatalf("qos changed: %+v", got)
	}

	body, err := json.Marshal(models.Pcf_SMPolCtrl_SmPolicyContextData{SubsDefQos: got})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	subs, _ := decoded["subsDefQos"].(map[string]any)
	arp, _ := subs["arp"].(map[string]any)
	if arp["preemptCap"] != "NOT_PREEMPT" || arp["preemptVuln"] != "NOT_PREEMPTABLE" {
		t.Fatalf("json arp = %#v body=%s", arp, body)
	}
}

func TestSmPolicySubscribedDefaultQosKeepsProvidedPreemption(t *testing.T) {
	src := &models.SubscribedDefaultQos{
		Var5qi: 9,
		Arp: &models.Arp{
			PriorityLevel: 1,
			PreemptCap:    models.PreemptionCapability_MAY_PREEMPT,
			PreemptVuln:   models.PreemptionVulnerability_PREEMPTABLE,
		},
	}

	got := smPolicySubscribedDefaultQos(src)
	if got != src {
		t.Fatal("complete ARP should be forwarded unchanged")
	}
}
