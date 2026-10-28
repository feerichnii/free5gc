package util

import (
	"fmt"

	"github.com/free5gc/openapi/models"
)

func SearchNFServiceUri(nfProfile *models.Nrf_NFDisc_NFProfile, serviceName models.Nrf_NFMgmt_ServiceName,
	nfServiceStatus models.Nrf_NFMgmt_NFServiceStatus,
) (nfUri string) {
	services := nfProfile.NfServices
	// Rel-16+: NRF may return services only in nfServiceList (nfServices is deprecated)
	if len(services) == 0 && len(nfProfile.NfServiceList) != 0 {
		services = make([]models.Nrf_NFDisc_NFService, 0, len(nfProfile.NfServiceList))
		for _, service := range nfProfile.NfServiceList {
			services = append(services, service)
		}
	}

	for _, service := range services {
		if service.ServiceName == serviceName && service.NfServiceStatus == nfServiceStatus {
			nfUri = serviceBaseURI(nfProfile, service)
		}
		if nfUri != "" {
			break
		}
	}
	return
}

// serviceBaseURI is the HTTP client base path for one NF service.
// free5GC stores scheme://ip:port in apiPrefix. A foreign profile can omit it
// and publish scheme plus ipEndPoints instead.
func serviceBaseURI(nfProfile *models.Nrf_NFDisc_NFProfile, service models.Nrf_NFDisc_NFService) string {
	if service.ApiPrefix != "" {
		return service.ApiPrefix
	}

	scheme := service.Scheme
	if scheme == "" {
		scheme = models.UriScheme_HTTP
	}

	var addr string
	var port int32
	if len(service.IpEndPoints) > 0 {
		point := service.IpEndPoints[0]
		port = point.Port
		addr = point.Ipv4Address
	}
	if addr == "" && len(nfProfile.Ipv4Addresses) > 0 {
		addr = nfProfile.Ipv4Addresses[0]
	}
	if addr == "" {
		if service.Fqdn != "" {
			addr = service.Fqdn
		} else {
			addr = nfProfile.Fqdn
		}
	}
	if addr == "" {
		return ""
	}
	return getSbiUri(scheme, addr, port)
}

func getSbiUri(scheme models.UriScheme, ipv4Address string, port int32) (uri string) {
	if port != 0 {
		uri = fmt.Sprintf("%s://%s:%d", scheme, ipv4Address, port)
	} else {
		switch scheme {
		case models.UriScheme_HTTP:
			uri = fmt.Sprintf("%s://%s:80", scheme, ipv4Address)
		case models.UriScheme_HTTPS:
			uri = fmt.Sprintf("%s://%s:443", scheme, ipv4Address)
		}
	}
	return
}
