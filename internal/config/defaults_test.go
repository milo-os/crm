package config_test

import (
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"go.miloapis.com/crm/internal/config"
)

func TestDecodingEmptyOperatorConfigAppliesDefaults(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := config.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := config.RegisterDefaults(scheme); err != nil {
		t.Fatal(err)
	}
	codecs := serializer.NewCodecFactory(scheme, serializer.EnableStrict)

	data := []byte("apiVersion: apiserver.config.miloapis.com/v1alpha1\nkind: CRMOperator\n")
	var cfg config.CRMOperator
	if err := runtime.DecodeInto(codecs.UniversalDecoder(), data, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.MetricsServer.SecureServing == nil || !*cfg.MetricsServer.SecureServing {
		t.Errorf("MetricsServer.SecureServing = %v, want true", cfg.MetricsServer.SecureServing)
	}
	if cfg.MetricsServer.BindAddress != ":8443" {
		t.Errorf("MetricsServer.BindAddress = %q, want %q", cfg.MetricsServer.BindAddress, ":8443")
	}
	if filepath.Base(cfg.MetricsServer.TLS.CertDir) != "serving-certs" {
		t.Errorf("MetricsServer.TLS.CertDir = %q, want a serving-certs directory", cfg.MetricsServer.TLS.CertDir)
	}
	if cfg.MetricsServer.TLS.CertName != "tls.crt" || cfg.MetricsServer.TLS.KeyName != "tls.key" {
		t.Errorf("MetricsServer.TLS cert/key = %q/%q, want tls.crt/tls.key", cfg.MetricsServer.TLS.CertName, cfg.MetricsServer.TLS.KeyName)
	}
}
