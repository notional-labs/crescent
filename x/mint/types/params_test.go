package types_test

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/stretchr/testify/require"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmosquad-labs/squad/x/mint/types"
)

func TestParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

	defaultParams := types.DefaultParams()

	paramsStr := `mintdenom: stake
block_time_threshold: 10s
inflation_schedules:
- start_time: 2022-01-01T00:00:00Z
  end_time: 2023-01-01T00:00:00Z
  amount: "300000000000000"
- start_time: 2023-01-01T00:00:00Z
  end_time: 2024-01-01T00:00:00Z
  amount: "200000000000000"
`
	require.Equal(t, paramsStr, defaultParams.String())
}

func TestParamsValidate(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())

	testCases := []struct {
		name        string
		configure   func(*types.Params)
		expectedErr string
	}{
		{
			"empty mint denom",
			func(params *types.Params) {
				params.MintDenom = ""
			},
			"mint denom cannot be blank",
		},
		{
			"invalid mint denom",
			func(params *types.Params) {
				params.MintDenom = "a"
			},
			"invalid denom: a",
		},
		{
			"negative block time threshold",
			func(params *types.Params) {
				params.BlockTimeThreshold = -1
			},
			"block time threshold must be positive: -1",
		},
		{
			"nil inflation schedules",
			func(params *types.Params) {
				params.InflationSchedules = nil
			},
			"",
		},
		{
			"empty inflation schedules",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{}
			},
			"",
		},
		{
			"invalid inflation schedule start, end time",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: squadtypes.ParseTime("2023-01-01T00:00:00Z"),
						EndTime:   squadtypes.ParseTime("2022-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(300000000000000),
					},
				}
			},
			"inflation end time 2022-01-01T00:00:00Z must be greater than start time 2023-01-01T00:00:00Z",
		},
		{
			"negative inflation Amount",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: squadtypes.ParseTime("2022-01-01T00:00:00Z"),
						EndTime:   squadtypes.ParseTime("2023-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(-1),
					},
				}
			},
			"inflation schedule amount must be positive: -1",
		},
		{
			"too small inflation Amount",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: squadtypes.ParseTime("2022-01-01T00:00:00Z"),
						EndTime:   squadtypes.ParseTime("2023-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31535999),
					},
				}
			},
			"inflation amount too small, it should be over period duration seconds: 31535999",
		},
		{
			"overlapped inflation schedules",
			func(params *types.Params) {
				params.InflationSchedules = []types.InflationSchedule{
					{
						StartTime: squadtypes.ParseTime("2022-01-01T00:00:00Z"),
						EndTime:   squadtypes.ParseTime("2023-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31536000),
					},
					{
						StartTime: squadtypes.ParseTime("2022-12-01T00:00:00Z"),
						EndTime:   squadtypes.ParseTime("2024-01-01T00:00:00Z"),
						Amount:    sdk.NewInt(31536000),
					},
				}
			},
			"inflation periods cannot be overlapped 2022-01-01T00:00:00Z ~ 2023-01-01T00:00:00Z with 2022-12-01T00:00:00Z ~ 2024-01-01T00:00:00Z",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.configure(&params)
			err := params.Validate()

			var err2 error
			for _, p := range params.ParamSetPairs() {
				err := p.ValidatorFn(reflect.ValueOf(p.Value).Elem().Interface())
				if err != nil {
					err2 = err
					break
				}
			}
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				require.EqualError(t, err2, tc.expectedErr)
			} else {
				require.Nil(t, err)
				require.Nil(t, err2)
			}
		})
	}
}
