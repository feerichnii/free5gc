package util

import (
	"testing"

	"github.com/free5gc/openapi/models"
)

func registeredCHF(ip string, port int32) models.Nrf_NFDisc_NFService {
	return models.Nrf_NFDisc_NFService{
		ServiceName:     models.Nrf_NFMgmt_ServiceName_NCHF_CONVERGEDCHARGING,
		NfServiceStatus: models.Nrf_NFMgmt_NFServiceStatus_REGISTERED,
		Scheme:          models.UriScheme_HTTP,
		IpEndPoints: []models.Nrf_NFMgmt_IpEndPoint{
			{Ipv4Address: ip, Port: port},
		},
	}
}

func TestSearchNFServiceUri_ApiPrefix(t *testing.T) {
	service := registeredCHF("10.0.0.8", 8080)
	service.ApiPrefix = "http://chf.example:9090"
	profile := &models.Nrf_NFDisc_NFProfile{NfServices: []models.Nrf_NFDisc_NFService{service}}

	got := SearchNFServiceUri(profile, models.Nrf_NFMgmt_ServiceName_NCHF_CONVERGEDCHARGING,
		models.Nrf_NFMgmt_NFServiceStatus_REGISTERED)
	if got != "http://chf.example:9090" {
		t.Fatalf("apiPrefix uri = %q", got)
	}
}

func TestSearchNFServiceUri_WithoutApiPrefix(t *testing.T) {
	profile := &models.Nrf_NFDisc_NFProfile{
		Fqdn:       "chf.example",
		NfServices: []models.Nrf_NFDisc_NFService{registeredCHF("172.16.74.2", 8080)},
	}

	got := SearchNFServiceUri(profile, models.Nrf_NFMgmt_ServiceName_NCHF_CONVERGEDCHARGING,
		models.Nrf_NFMgmt_NFServiceStatus_REGISTERED)
	if got != "http://172.16.74.2:8080" {
		t.Fatalf("endpoint uri = %q", got)
	}
}

func TestSearchNFServiceUri_ServiceListWithoutApiPrefix(t *testing.T) {
	profile := &models.Nrf_NFDisc_NFProfile{
		NfServiceList: map[string]models.Nrf_NFDisc_NFService{
			"pcf": {
				ServiceName:     models.Nrf_NFMgmt_ServiceName_NPCF_SMPOLICYCONTROL,
				NfServiceStatus: models.Nrf_NFMgmt_NFServiceStatus_REGISTERED,
				Scheme:          models.UriScheme_HTTP,
				IpEndPoints: []models.Nrf_NFMgmt_IpEndPoint{
					{Ipv4Address: "172.16.74.3", Port: 8080},
				},
			},
		},
	}

	got := SearchNFServiceUri(profile, models.Nrf_NFMgmt_ServiceName_NPCF_SMPOLICYCONTROL,
		models.Nrf_NFMgmt_NFServiceStatus_REGISTERED)
	if got != "http://172.16.74.3:8080" {
		t.Fatalf("nfServiceList uri = %q", got)
	}
}
