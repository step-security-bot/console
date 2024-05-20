package devices

import (
	"context"
	"reflect"

	"github.com/open-amt-cloud-toolkit/console/internal/usecase/utils"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/cim/credential"
)

type SecuritySettings struct {
	AssociatedCertificates []AssociatedCertificates `json:"associated_certificates"`
	Certificates           interface{}              `json:"certificates"`
	Keys                   interface{}              `json:"keys"`
}

type AssociatedCertificates struct {
	Type              string      `json:"type"`
	ProfileID         string      `json:"profile_id"`
	RootCertificate   interface{} `json:"root_certificate"`
	ClientCertificate interface{} `json:"client_certificate"`
	Key               interface{} `json:"key"`
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
		Keys: 	   response.PublicPrivateKeyPairResponse,
	}

	if !reflect.DeepEqual(response.CIMCredentialContextResponse, credential.PullResponse{}) {
		for _, cert := range response.CIMCredentialContextResponse.Items.CredentialContext {
			var associatedCertificate AssociatedCertificates
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
								for _, key := range response.PublicPrivateKeyPairResponse.PublicPrivateKeyPairItems{
									if key.InstanceID == keyHandle {
										associatedCertificate.Key = key
									}
								}
							}
						}
					}
				}
			}
			securitySettings.AssociatedCertificates = append(securitySettings.AssociatedCertificates, associatedCertificate)
		}
		
	}

	return securitySettings, nil
}
