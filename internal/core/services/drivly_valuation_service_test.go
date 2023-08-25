package services

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_drivlyValuationService_getDeviceMileage_nil_udd(t *testing.T) {
	mileage, err := getDeviceMileage(nil, 2020, 2023)
	require.NoError(t, err)

	assert.Equal(t, 36000.0, *mileage)
}
