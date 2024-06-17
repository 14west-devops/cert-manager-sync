package main

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/sts"
)

// CreateAWSSession will connect to AWS with the account's credentials from vault
func CreateAWSSession() (*session.Session, error) {
	l := log.WithFields(
		log.Fields{
			"action": "CreateAWSSession",
		},
	)
	l.Print("CreateAWSSession")
	sess, err := session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region:                        aws.String(os.Getenv("AWS_REGION"))},
	)
	if err != nil {
		l.Printf("%+v", err)
	}

	// Create a STS client of passed a role to assume
	roleToAssumeArn := os.Getenv("AWS_STS_ROLE_NAME")
	if len(strings.TrimSpace(roleToAssumeArn)) > 0 {
		svc := sts.New(sess)
		sessionName := os.Getenv("AWS_STS_SESSION_NAME")
		arresult, err := svc.AssumeRole(&sts.AssumeRoleInput{
			RoleArn:         &roleToAssumeArn,
			RoleSessionName: &sessionName,
		})

		if err != nil {
			l.Printf("%+v", err)
		}

		l.Println(arresult.AssumedRoleUser)
	}

	return sess, nil
}

// separateCerts ensures that certificates are configured appropriately
func separateCerts(name string, ca, crt, key []byte) *Certificate {

	l := log.WithFields(
		log.Fields{
			"action": "CreateAWSSession",
		},
	)

	b := "-----BEGIN CERTIFICATE-----\n"
	str := strings.Split(string(crt), b)

	//print each certificate in the chain
	for i, s := range str {
		l.Printf("Certificate %d: %s\n", i, s)
	}

	nc := b + str[1]
	l.Println("cert: ", nc)

	//ch := strings.Join(str[:len(str)-1], b)
	ch := b + strings.Join(str[2:], b)
	l.Println("chain: ", ch)
	cert := &Certificate{
		SecretName:  name,
		Chain:       []byte(ch),
		Certificate: []byte(nc),
		Key:         key,
	}
	return cert
}

// separateCertsACM wraps separateCerts and returns an acm ImportCertificateInput Object
func separateCertsACM(name string, ca, crt, key []byte) *acm.ImportCertificateInput {
	cert := separateCerts(name, ca, crt, key)
	im := &acm.ImportCertificateInput{
		CertificateChain: cert.Chain,
		Certificate:      cert.Certificate,
		PrivateKey:       cert.Key,
	}
	return im
}

// ImportCertificate imports a cert into ACM
func ImportCertificate(s *session.Session, im *acm.ImportCertificateInput, arn string) (string, error) {
	l := log.WithFields(
		log.Fields{
			"action": "ImportCertificate",
		},
	)
	l.Print("ImportCertificate")
	svc := acm.New(s)
	if arn != "" {
		im.CertificateArn = &arn
	}
	cert, err := svc.ImportCertificate(im)
	if err != nil {
		l.Printf("awsacm.ImportCertificate svc.ImportCertificate error: %v\n", err)
		return "", err
	}
	return *cert.CertificateArn, nil
}
