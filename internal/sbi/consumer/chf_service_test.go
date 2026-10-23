package consumer

import (
	"encoding/json"
	"testing"

	"github.com/free5gc/openapi/models"
	smf_context "github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/pkg/factory"
)

type chargingSequenceTestApp struct {
	ctx *smf_context.SMFContext
}

func (a *chargingSequenceTestApp) SetLogEnable(bool)    {}
func (a *chargingSequenceTestApp) SetLogLevel(string)   {}
func (a *chargingSequenceTestApp) SetReportCaller(bool) {}
func (a *chargingSequenceTestApp) Start()               {}
func (a *chargingSequenceTestApp) Terminate()           {}
func (a *chargingSequenceTestApp) Context() *smf_context.SMFContext {
	return a.ctx
}
func (a *chargingSequenceTestApp) Config() *factory.Config { return nil }

func TestBuildConvergedChargingRequestInvocationSequence(t *testing.T) {
	app := &chargingSequenceTestApp{
		ctx: &smf_context.SMFContext{
			NfInstanceID: "smf-1",
			URIScheme:    models.UriScheme_HTTP,
			RegisterIPv4: "10.0.0.1",
			SBIPort:      8000,
		},
	}
	c, err := NewConsumer(app)
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}

	smContext := &smf_context.SMContext{
		Smf_PDUSess_SmContextCreateData: &models.Smf_PDUSess_SmContextCreateData{
			Supi: "imsi-001010000000001",
			Dnn:  "internet",
		},
		ChargingID:    7,
		PDUSessionID:  1,
		Ref:           "pdu-1",
		RequestedUnit: 1,
	}

	first := c.buildConvergedChargingRequest(smContext, nil)
	if first.InvocationSequenceNumber != 1 {
		t.Fatalf("first sequence = %d", first.InvocationSequenceNumber)
	}
	body, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	if !json.Valid(body) {
		t.Fatalf("invalid json: %s", body)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal first: %v", err)
	}
	seq, ok := decoded["invocationSequenceNumber"]
	if !ok {
		t.Fatalf("invocationSequenceNumber missing from JSON: %s", body)
	}
	if seq != float64(1) {
		t.Fatalf("json sequence = %v", seq)
	}

	second := c.buildConvergedChargingRequest(smContext, []models.Chf_ConvCharging_MultipleUnitUsage{{
		RatingGroup: 1,
	}})
	if second.InvocationSequenceNumber != 2 || smContext.InvocationSequenceNumber != 2 {
		t.Fatalf("second sequence request=%d session=%d", second.InvocationSequenceNumber, smContext.InvocationSequenceNumber)
	}
}
