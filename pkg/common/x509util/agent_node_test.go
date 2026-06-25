package x509util_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"testing"

	"github.com/spiffe/spire/pkg/common/x509util"
	"github.com/stretchr/testify/require"
)

func TestAgentNodeNameExtension(t *testing.T) {
	extension, err := x509util.NewAgentNodeNameExtension("k8s-node-1")
	require.NoError(t, err)
	require.True(t, extension.Id.Equal(x509util.AgentNodeNameOID))
	require.False(t, extension.Critical)

	nodeName, found, err := x509util.AgentNodeNameFromExtensions([]pkix.Extension{extension})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "k8s-node-1", nodeName)

	nodeName, found, err = x509util.AgentNodeNameFromCertificate(&x509.Certificate{Extensions: []pkix.Extension{extension}})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "k8s-node-1", nodeName)
}

func TestNewAgentNodeNameExtensionRequiresNodeName(t *testing.T) {
	_, err := x509util.NewAgentNodeNameExtension("")
	require.ErrorContains(t, err, "agent node name is empty")
}

func TestAgentNodeNameFromCertificateAbsent(t *testing.T) {
	nodeName, found, err := x509util.AgentNodeNameFromCertificate(&x509.Certificate{})
	require.NoError(t, err)
	require.False(t, found)
	require.Empty(t, nodeName)

	nodeName, found, err = x509util.AgentNodeNameFromCertificate(nil)
	require.NoError(t, err)
	require.False(t, found)
	require.Empty(t, nodeName)
}

func TestAgentNodeNameFromExtensionsRejectsInvalidExtension(t *testing.T) {
	validValue, err := asn1.MarshalWithParams("k8s-node-1", "utf8")
	require.NoError(t, err)
	emptyValue, err := asn1.MarshalWithParams("", "utf8")
	require.NoError(t, err)

	for _, tt := range []struct {
		name       string
		extensions []pkix.Extension
		expectErr  string
	}{
		{
			name:       "invalid ASN.1",
			extensions: []pkix.Extension{{Id: x509util.AgentNodeNameOID, Value: []byte{1, 2, 3}}},
			expectErr:  "unmarshal agent node name extension",
		},
		{
			name:       "trailing data",
			extensions: []pkix.Extension{{Id: x509util.AgentNodeNameOID, Value: append(append([]byte(nil), validValue...), 0)}},
			expectErr:  "agent node name extension has trailing data",
		},
		{
			name:       "empty node name",
			extensions: []pkix.Extension{{Id: x509util.AgentNodeNameOID, Value: emptyValue}},
			expectErr:  "agent node name extension is empty",
		},
		{
			name: "multiple extensions",
			extensions: []pkix.Extension{
				{Id: x509util.AgentNodeNameOID, Value: validValue},
				{Id: x509util.AgentNodeNameOID, Value: validValue},
			},
			expectErr: "multiple agent node name extensions found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			nodeName, found, err := x509util.AgentNodeNameFromExtensions(tt.extensions)
			require.ErrorContains(t, err, tt.expectErr)
			require.True(t, found)
			require.Empty(t, nodeName)
		})
	}
}
