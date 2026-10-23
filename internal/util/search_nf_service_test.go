package util_test

import (
	"testing"

	"github.com/free5gc/amf/internal/util"
	"github.com/free5gc/openapi/models"
)

func pcfService(ipv4 string, port int32) models.Nrf_NFDisc_NFService {
	return models.Nrf_NFDisc_NFService{
		ServiceName:     models.Nrf_NFMgmt_ServiceName_NPCF_AM_POLICY_CONTROL,
		NfServiceStatus: models.Nrf_NFMgmt_NFServiceStatus_REGISTERED,
		Scheme:          models.UriScheme_HTTP,
		IpEndPoints: []models.Nrf_NFMgmt_IpEndPoint{
			{Ipv4Address: ipv4, Port: port},
		},
	}
}

func TestSearchNFServiceUri_NfServices(t *testing.T) {
	profile := &models.Nrf_NFDisc_NFProfile{
		NfServices: []models.Nrf_NFDisc_NFService{
			pcfService("10.0.0.1", 8080),
		},
	}

	got := util.SearchNFServiceUri(profile,
		models.Nrf_NFMgmt_ServiceName_NPCF_AM_POLICY_CONTROL,
		models.Nrf_NFMgmt_NFServiceStatus_REGISTERED)
	if got != "http://10.0.0.1:8080" {
		t.Fatalf("NfServices uri = %q", got)
	}
}

func TestSearchNFServiceUri_NfServiceListWhenServicesEmpty(t *testing.T) {
	profile := &models.Nrf_NFDisc_NFProfile{
		NfServiceList: map[string]models.Nrf_NFDisc_NFService{
			"pcf-am-policy": pcfService("172.16.74.1", 8082),
		},
	}

	got := util.SearchNFServiceUri(profile,
		models.Nrf_NFMgmt_ServiceName_NPCF_AM_POLICY_CONTROL,
		models.Nrf_NFMgmt_NFServiceStatus_REGISTERED)
	if got != "http://172.16.74.1:8082" {
		t.Fatalf("NfServiceList uri = %q", got)
	}
}

func TestSearchNFServiceUri_PrefersNfServices(t *testing.T) {
	profile := &models.Nrf_NFDisc_NFProfile{
		NfServices: []models.Nrf_NFDisc_NFService{
			pcfService("10.0.0.1", 8080),
		},
		NfServiceList: map[string]models.Nrf_NFDisc_NFService{
			"pcf-am-policy": pcfService("172.16.74.1", 8082),
		},
	}

	got := util.SearchNFServiceUri(profile,
		models.Nrf_NFMgmt_ServiceName_NPCF_AM_POLICY_CONTROL,
		models.Nrf_NFMgmt_NFServiceStatus_REGISTERED)
	if got != "http://10.0.0.1:8080" {
		t.Fatalf("preferred uri = %q", got)
	}
}

func TestSearchNFServiceUri_Empty(t *testing.T) {
	profile := &models.Nrf_NFDisc_NFProfile{}
	got := util.SearchNFServiceUri(profile,
		models.Nrf_NFMgmt_ServiceName_NPCF_AM_POLICY_CONTROL,
		models.Nrf_NFMgmt_NFServiceStatus_REGISTERED)
	if got != "" {
		t.Fatalf("empty uri = %q", got)
	}
}
