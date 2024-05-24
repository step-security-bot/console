package devices

import (
	"context"
	"reflect"
	"strings"

	"github.com/open-amt-cloud-toolkit/console/internal/usecase/devices/wsman"
	"github.com/open-amt-cloud-toolkit/console/internal/usecase/utils"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/amt/publickey"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/amt/publicprivate"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/cim/credential"
)

const (
	TypeWireless string = "Wireless"
	TypeTLS      string = "TLS"
	TypeWired    string = "Wired"
)

type SecuritySettings struct {
	ProfileAssociation []ProfileAssociation `json:"ProfileAssociation"`
	Certificates       interface{}          `json:"Certificates"`
	Keys               interface{}          `json:"PublicKeys"`
}

type ProfileAssociation struct {
	Type              string      `json:"Type"`
	ProfileID         string      `json:"ProfileID"`
	RootCertificate   interface{} `json:"RootCertificate,omitempty"`
	ClientCertificate interface{} `json:"ClientCertificate,omitempty"`
	Key               interface{} `json:"PublicKey,omitempty"`
}

func processCertificates(contextItems []credential.CredentialContext, response wsman.Certificates, profileType string, securitySettings *SecuritySettings) {
	for _, cert := range contextItems {
		var profileAssociation ProfileAssociation
		// var isNewCertificate bool = true
		var isNewProfileAssociation bool = true
		profileAssociation.Type = profileType
		profileAssociation.ProfileID = strings.TrimPrefix(cert.ElementProvidingContext.ReferenceParameters.SelectorSet.Selectors[0].Text, "Intel(r) AMT:IEEE 802.1x Settings ")
		certificateHandle := cert.ElementInContext.ReferenceParameters.SelectorSet.Selectors[0].Text

		for _, publicKeyCert := range response.PublicKeyCertificateResponse.PublicKeyCertificateItems {
			if publicKeyCert.InstanceID == certificateHandle {
				if publicKeyCert.TrustedRootCertificate {
					profileAssociation.RootCertificate = publicKeyCert
				} else {
					profileAssociation.ClientCertificate = publicKeyCert
					for _, privateKeyPair := range response.ConcreteDependencyResponse.Items {
						if privateKeyPair.Antecedent.ReferenceParameters.SelectorSet.Selectors[0].Text == certificateHandle {
							keyHandle := privateKeyPair.Dependent.ReferenceParameters.SelectorSet.Selectors[0].Text
							for _, key := range response.PublicPrivateKeyPairResponse.PublicPrivateKeyPairItems {
								if key.InstanceID == keyHandle {
									profileAssociation.Key = key
								}
							}
						}
					}
				}
			}
		}

		// Check if the certificate is already in the list
		for i, existingCertificate := range securitySettings.ProfileAssociation {
			if existingCertificate.ProfileID == profileAssociation.ProfileID {
				if profileAssociation.RootCertificate != nil {
					securitySettings.ProfileAssociation[i].RootCertificate = profileAssociation.RootCertificate
				}
				if profileAssociation.ClientCertificate != nil {
					securitySettings.ProfileAssociation[i].ClientCertificate = profileAssociation.ClientCertificate
				}
				if profileAssociation.Key != nil {
					securitySettings.ProfileAssociation[i].Key = profileAssociation.Key
				}
				isNewProfileAssociation = false
				break
			}
		}

		// If the profile is not in the list, add it
		if isNewProfileAssociation {
			securitySettings.ProfileAssociation = append(securitySettings.ProfileAssociation, profileAssociation)
		}

		// If a client cert, update the associated public key w/ the cert's handle
		if profileAssociation.ClientCertificate != nil {
			var publicKeyHandle string
			// Loop thru public keys looking for the one that matches the current profileAssociation's key
			for i, existingKeyPair := range securitySettings.Keys.(publicprivate.RefinedPullResponse).PublicPrivateKeyPairItems {
				// If found update that key with the profileAssociation's certificate handle
				if existingKeyPair.InstanceID == profileAssociation.Key.(publicprivate.RefinedPublicPrivateKeyPair).InstanceID {
					securitySettings.Keys.(publicprivate.RefinedPullResponse).PublicPrivateKeyPairItems[i].CertificateHandle = profileAssociation.ClientCertificate.(publickey.RefinedPublicKeyCertificateResponse).InstanceID
					// save this public key handle since we know it pairs with the profileAssociation's certificate
					publicKeyHandle = securitySettings.Keys.(publicprivate.RefinedPullResponse).PublicPrivateKeyPairItems[i].InstanceID
					break
				}
			}

			// Loop thru certificates looking for the one that matches the current profileAssociation's certificate
			for i, existingCert := range securitySettings.Certificates.(publickey.RefinedPullResponse).PublicKeyCertificateItems {
				// if found associate the previously found key handle with it
				if existingCert.InstanceID == profileAssociation.ClientCertificate.(publickey.RefinedPublicKeyCertificateResponse).InstanceID {
					securitySettings.Certificates.(publickey.RefinedPullResponse).PublicKeyCertificateItems[i].PublicKeyHandle = publicKeyHandle
					break
				}
			}
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
