package cmd

import (
	"testing"

	"go.miloapis.com/instrumentation/instrumenttest"
)

func TestInstrumentationAdopted(t *testing.T) {
	instrumenttest.AssertAdopted(t)
}
