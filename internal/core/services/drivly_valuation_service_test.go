package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_drivlyValuationService_getDeviceMileage_nil_udd(t *testing.T) {
	mileage := getDeviceMileage(nil, 2020, 2023)

	assert.Equal(t, 36000.0, mileage)
}
