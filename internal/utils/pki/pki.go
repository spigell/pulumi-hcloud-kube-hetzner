package pki

import (
	"fmt"

	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-hcloud-kube-hetzner/internal/program"
)

type Certificate struct {
	allowedUsages []string
	ipAddreses    pulumi.StringArray

	PrivateKeyPem  pulumi.StringOutput
	CertificatePem pulumi.StringOutput
}
type PKI struct {
	ctx *program.Context

	*Certificate
}

func New(ctx *program.Context, name string) (*PKI, error) {
	key, err := program.PulumiRun(ctx, tls.NewPrivateKey, fmt.Sprintf("ca-key:%s", name), &tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"),
		RsaBits:   pulumi.Int(2048),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create private key for CA: %w", err)
	}

	caRoot, err := program.PulumiRun(ctx, tls.NewSelfSignedCert, fmt.Sprintf("ca-cert:%s", name), &tls.SelfSignedCertArgs{
		PrivateKeyPem: key.PrivateKeyPem,

		ValidityPeriodHours: pulumi.Int(87600),
		IsCaCertificate:     pulumi.Bool(true),
		AllowedUses:         pulumi.ToStringArray([]string{"cert_signing"}),
		Subject: tls.SelfSignedCertSubjectArgs{
			CommonName: pulumi.String(name),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	return &PKI{
		ctx: ctx,
		Certificate: &Certificate{
			PrivateKeyPem:  key.PrivateKeyPem,
			CertificatePem: caRoot.CertPem,
		},
	}, nil
}

func (n *PKI) NewCertificate(name string, options ...func(*Certificate)) (*Certificate, error) {
	certificate := &Certificate{}

	for _, o := range options {
		o(certificate)
	}

	certKey, err := program.PulumiRun(n.ctx, tls.NewPrivateKey, name, &tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"),
		RsaBits:   pulumi.Int(2048),
	})
	if err != nil {
		return nil, err
	}

	req, err := program.PulumiRun(n.ctx, tls.NewCertRequest, name, &tls.CertRequestArgs{
		PrivateKeyPem: certKey.PrivateKeyPem,
		IpAddresses:   certificate.ipAddreses,

		Subject: tls.CertRequestSubjectArgs{
			CommonName: pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	cert, err := program.PulumiRun(n.ctx, tls.NewLocallySignedCert, name, &tls.LocallySignedCertArgs{
		CertRequestPem: req.CertRequestPem,

		CaPrivateKeyPem: n.Certificate.PrivateKeyPem,
		CaCertPem:       n.Certificate.CertificatePem,

		ValidityPeriodHours: pulumi.Int(8784),
		EarlyRenewalHours:   pulumi.Int(360),
		AllowedUses:         pulumi.ToStringArray(certificate.allowedUsages),
	})
	if err != nil {
		return nil, err
	}

	return &Certificate{
		PrivateKeyPem:  certKey.PrivateKeyPem,
		CertificatePem: cert.CertPem,
	}, nil
}

func WithIPAddesses(ips pulumi.StringArray) func(*Certificate) {
	return func(c *Certificate) {
		c.ipAddreses = ips
	}
}

func WithAllowedUsages(usages []string) func(*Certificate) {
	return func(c *Certificate) {
		c.allowedUsages = usages
	}
}
