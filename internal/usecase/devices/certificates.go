package devices

import (
	"context"
	"reflect"

	"github.com/open-amt-cloud-toolkit/console/internal/usecase/devices/wsman"
	"github.com/open-amt-cloud-toolkit/console/internal/usecase/utils"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/cim/credential"
)

const (
	TypeWireless string = "Wireless"
	TypeTLS      string = "TLS"
	TypeWired    string = "Wired"
)

type SecuritySettings struct {
	AssociatedCertificates []AssociatedCertificates `json:"AssociatedCertificates"`
	Certificates           interface{}              `json:"Certificates"`
	Keys                   interface{}              `json:"PublicKeys"`
}

type AssociatedCertificates struct {
	Type              string      `json:"Type"`
	ProfileID         string      `json:"ProfileID"`
	RootCertificate   interface{} `json:"RootCertificate"`
	ClientCertificate interface{} `json:"ClientCertificate"`
	Key               interface{} `json:"PublicKey"`
}

func processCertificates(contextItems []credential.CredentialContext, response wsman.Certificates, profileType string, securitySettings *SecuritySettings) {
	for _, cert := range contextItems {
		var associatedCertificate AssociatedCertificates
		var isNewCertificate bool = false
		associatedCertificate.Type = profileType
		associatedCertificate.ProfileID = cert.ElementProvidingContext.ReferenceParameters.SelectorSet.Selectors[0].Text
		certificateHandle := cert.ElementInContext.ReferenceParameters.SelectorSet.Selectors[0].Text

		for _, publicKeyCert := range response.PublicKeyCertificateResponse.PublicKeyCertificateItems {
				if publicKeyCert.InstanceID == certificateHandle {
				if publicKeyCert.TrustedRootCertificate {
					associatedCertificate.RootCertificate = publicKeyCert
				} else {
					associatedCertificate.ClientCertificate = publicKeyCert
					for _, privateKeyPair := range response.ConcreteDependencyResponse.Items {
						if privateKeyPair.Antecedent.ReferenceParameters.SelectorSet.Selectors[0].Text == certificateHandle {
							keyHandle := privateKeyPair.Dependent.ReferenceParameters.SelectorSet.Selectors[0].Text
							for _, key := range response.PublicPrivateKeyPairResponse.PublicPrivateKeyPairItems {
								if key.InstanceID == keyHandle {
									associatedCertificate.Key = key
								}
							}
						}
					}
				}
			}
		}

		// Check if the certificate is already in the list
		for i, existingCertificate := range securitySettings.AssociatedCertificates {
			if existingCertificate.ProfileID == associatedCertificate.ProfileID {
				if associatedCertificate.RootCertificate != nil {
					securitySettings.AssociatedCertificates[i].RootCertificate = associatedCertificate.RootCertificate
				}
				if associatedCertificate.ClientCertificate != nil {
					securitySettings.AssociatedCertificates[i].ClientCertificate = associatedCertificate.ClientCertificate
				}
				if associatedCertificate.Key != nil {
					securitySettings.AssociatedCertificates[i].Key = associatedCertificate.Key
				}
				isNewCertificate = true
				break
			}
		}

		// If the certificate is not in the list, add it
		if !isNewCertificate {
			securitySettings.AssociatedCertificates = append(securitySettings.AssociatedCertificates, associatedCertificate)
		}
	}
}
func (uc *UseCase) GetCertificates(c context.Context, guid string) (interface{}, error) {
	item, err := uc.repo.GetByID(c, guid, "")
	if err != nil || item.GUID == "" {
		return nil, utils.ErrNotFound
	}

	uc.device.SetupWsmanClient(item, false, true)

	response, err := uc.device.GetCertificates()
	if err != nil {
		return nil, err
	}

	securitySettings := SecuritySettings{
		Certificates: response.PublicKeyCertificateResponse,
		Keys:         response.PublicPrivateKeyPairResponse,
	}

	if !reflect.DeepEqual(response.CIMCredentialContextResponse, credential.PullResponse{}) {
		processCertificates(response.CIMCredentialContextResponse.Items.CredentialContextTLS, response, TypeTLS, &securitySettings)
		processCertificates(response.CIMCredentialContextResponse.Items.CredentialContext, response, TypeWireless, &securitySettings)
		processCertificates(response.CIMCredentialContextResponse.Items.CredentialContext8021x, response, TypeWired, &securitySettings)
	}

	return securitySettings, nil
}
