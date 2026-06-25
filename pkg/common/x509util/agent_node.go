package x509util

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
)

// AgentNodeNameOID identifies the SPIRE workload X509-SVID extension carrying
// the Kubernetes node name of the agent that requested the SVID. This preview
// OID should be replaced with an assigned SPIRE OID before the extension is
// standardized.
var AgentNodeNameOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 1, 1}

// NewAgentNodeNameExtension returns a non-critical X.509 extension containing
// the agent node name as an ASN.1 UTF8String.
func NewAgentNodeNameExtension(nodeName string) (pkix.Extension, error) {
	if nodeName == "" {
		return pkix.Extension{}, errors.New("agent node name is empty")
	}
	value, err := asn1.MarshalWithParams(nodeName, "utf8")
	if err != nil {
		return pkix.Extension{}, fmt.Errorf("marshal agent node name extension: %w", err)
	}
	return pkix.Extension{
		Id:    AgentNodeNameOID,
		Value: value,
	}, nil
}

// AgentNodeNameFromCertificate extracts the agent node name extension from a
// certificate. The boolean return value reports whether the extension was
// present.
func AgentNodeNameFromCertificate(cert *x509.Certificate) (string, bool, error) {
	if cert == nil {
		return "", false, nil
	}
	return AgentNodeNameFromExtensions(cert.Extensions)
}

// AgentNodeNameFromExtensions extracts the agent node name extension from a
// list of X.509 extensions. The boolean return value reports whether the
// extension was present.
func AgentNodeNameFromExtensions(extensions []pkix.Extension) (string, bool, error) {
	var (
		nodeName string
		found    bool
	)
	for _, extension := range extensions {
		if !extension.Id.Equal(AgentNodeNameOID) {
			continue
		}
		if found {
			return "", true, errors.New("multiple agent node name extensions found")
		}
		found = true

		rest, err := asn1.Unmarshal(extension.Value, &nodeName)
		if err != nil {
			return "", true, fmt.Errorf("unmarshal agent node name extension: %w", err)
		}
		if len(rest) > 0 {
			return "", true, errors.New("agent node name extension has trailing data")
		}
		if nodeName == "" {
			return "", true, errors.New("agent node name extension is empty")
		}
	}
	return nodeName, found, nil
}
