package context

import (
	"testing"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/smf/pkg/factory"
	"gopkg.in/yaml.v2"
)

func TestSupplementStubPccRuleFromConfig(t *testing.T) {
	old := factory.SmfConfig
	factory.SmfConfig = &factory.Config{
		Configuration: &factory.Configuration{
			PccRuleTemplates: map[string]*models.Pcf_SMPolCtrl_PccRule{
				"GENERIC": {
					FlowInfos: []models.Pcf_SMPolCtrl_FlowInformation{{
						FlowDescription:   "permit out ip from any to assigned",
						FlowDirection:     models.Pcf_SMPolCtrl_FlowDirectionRm_BIDIRECTIONAL,
						PacketFilterUsage: true,
					}},
					Precedence: 255,
					RefChgData: []string{"ChgId-1"},
				},
				"SERVICE_1": {
					FlowInfos: []models.Pcf_SMPolCtrl_FlowInformation{{
						FlowDescription: "permit out ip from 1.1.1.1/32 to assigned",
						FlowDirection:   models.Pcf_SMPolCtrl_FlowDirectionRm_BIDIRECTIONAL,
					}},
					Precedence: 128,
					RefQosData: []string{"1"},
					RefChgData: []string{"ChgId-2"},
				},
			},
			QosDecs: map[string]*models.Pcf_SMPolCtrl_QosData{
				"1": {QosId: "1", Var5qi: 9, MaxbrUl: "108 Mbps", MaxbrDl: "108 Mbps"},
			},
			ChgDecs: map[string]*models.Pcf_SMPolCtrl_ChargingData{
				"ChgId-1": {ChgId: "ChgId-1", RatingGroup: 100, Online: true},
				"ChgId-2": {ChgId: "ChgId-2", RatingGroup: 200, Online: true},
			},
		},
	}
	t.Cleanup(func() { factory.SmfConfig = old })

	decision := &models.Pcf_SMPolCtrl_SmPolicyDecision{
		PccRules: map[string]*models.Pcf_SMPolCtrl_PccRule{
			"GENERIC":   {PccRuleId: "GENERIC"},
			"SERVICE_1": {PccRuleId: "SERVICE_1"},
			"PccRuleId-1": {
				PccRuleId: "PccRuleId-1",
				FlowInfos: []models.Pcf_SMPolCtrl_FlowInformation{{
					FlowDescription: "permit out ip from any to assigned",
				}},
				Precedence: 255,
				RefChgData: []string{"ChgId-1"},
			},
		},
		ChgDecs: map[string]*models.Pcf_SMPolCtrl_ChargingData{
			"ChgId-1": {ChgId: "ChgId-1", RatingGroup: 1},
		},
	}

	if err := SupplementPccRulesFromConfig(decision); err != nil {
		t.Fatalf("supplement: %v", err)
	}

	generic := decision.PccRules["GENERIC"]
	if generic.Precedence != 255 || len(generic.FlowInfos) != 1 ||
		generic.FlowInfos[0].FlowDescription != "permit out ip from any to assigned" ||
		len(generic.RefChgData) != 1 || generic.RefChgData[0] != "ChgId-1" {
		t.Fatalf("GENERIC = %+v", generic)
	}
	if decision.ChgDecs["ChgId-1"].RatingGroup != 1 {
		t.Fatalf("existing chgDec overwritten: %+v", decision.ChgDecs["ChgId-1"])
	}
	if decision.ChgDecs["ChgId-2"] == nil || decision.ChgDecs["ChgId-2"].RatingGroup != 200 {
		t.Fatalf("ChgId-2 = %+v", decision.ChgDecs["ChgId-2"])
	}
	service := decision.PccRules["SERVICE_1"]
	if len(service.RefQosData) != 1 || decision.QosDecs["1"] == nil || decision.QosDecs["1"].Var5qi != 9 {
		t.Fatalf("SERVICE_1 qos = %+v qosDecs=%+v", service, decision.QosDecs["1"])
	}
	full := decision.PccRules["PccRuleId-1"]
	if len(full.FlowInfos) != 1 || full.FlowInfos[0].PacketFilterUsage {
		t.Fatalf("complete rule changed: %+v", full)
	}
}

func TestPccRuleTemplateYAML(t *testing.T) {
	const doc = `
pccRuleTemplates:
  GENERIC:
    flowInfos:
      - flowDescription: permit out ip from any to assigned
        flowDirection: BIDIRECTIONAL
        packetFilterUsage: true
    precedence: 255
    refChgData: [ChgId-1]
chgDecs:
  ChgId-1:
    chgId: ChgId-1
    meteringMethod: VOLUME
    online: true
    ratingGroup: 100
    reportingLevel: RAT_GR_LEVEL
qosDecs:
  "1":
    qosId: "1"
    5qi: 9
    maxbrUl: 108 Mbps
`
	var cfg factory.Configuration
	if err := yaml.Unmarshal([]byte(doc), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	rule := cfg.PccRuleTemplates["GENERIC"]
	if rule == nil || rule.FlowInfos[0].FlowDescription != "permit out ip from any to assigned" ||
		rule.Precedence != 255 || rule.RefChgData[0] != "ChgId-1" {
		t.Fatalf("template = %+v", rule)
	}
	if cfg.ChgDecs["ChgId-1"].RatingGroup != 100 || !cfg.ChgDecs["ChgId-1"].Online ||
		cfg.ChgDecs["ChgId-1"].ReportingLevel != models.Pcf_SMPolCtrl_ReportingLevel_RAT_GR_LEVEL {
		t.Fatalf("chg = %+v", cfg.ChgDecs["ChgId-1"])
	}
	if cfg.QosDecs["1"].Var5qi != 9 || cfg.QosDecs["1"].MaxbrUl != "108 Mbps" {
		t.Fatalf("qos = %+v", cfg.QosDecs["1"])
	}
}

func TestSupplementStubPccRuleWithoutTemplate(t *testing.T) {
	old := factory.SmfConfig
	factory.SmfConfig = &factory.Config{Configuration: &factory.Configuration{}}
	t.Cleanup(func() { factory.SmfConfig = old })

	decision := &models.Pcf_SMPolCtrl_SmPolicyDecision{
		PccRules: map[string]*models.Pcf_SMPolCtrl_PccRule{
			"GENERIC": {PccRuleId: "GENERIC"},
		},
	}
	err := SupplementPccRulesFromConfig(decision)
	if err == nil || err.Error() != "PCC rule GENERIC has no flow description and no SMF template" {
		t.Fatalf("err = %v", err)
	}
}
