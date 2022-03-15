package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const gasTrackingKey = "__gt_key__"

type SessionRecord struct {
	ActualSDKGas    sdk.Gas
	OriginalSDKGas  sdk.Gas
	OriginalVMGas   sdk.Gas
	ActualVMGas     sdk.Gas
	ContractAddress string
}

type VMRecord struct {
	OriginalVMGas       sdk.Gas
	ActualVMGas         sdk.Gas
	ActualStoreSDKGas   sdk.Gas
	OriginalStoreSDKGas sdk.Gas
}

type activeSession struct {
	meter               *ContractSDKGasMeter
	gasFilledIn         bool
	originalVMGas       sdk.Gas
	actualVMGas         sdk.Gas
	originalStoreSDKGas sdk.Gas
	actualStoreSDKGas   sdk.Gas
}

type gasTracking struct {
	depth                    uint64
	limitForUpcomingGasMeter uint64
	mainGasMeter             sdk.GasMeter
	activeSessions           []*activeSession
	sessionRecords           []*SessionRecord
}

func getGasTrackingData(ctx sdk.Context) (*gasTracking, error) {
	queryTracking, ok := ctx.Value(gasTrackingKey).(*gasTracking)
	if queryTracking == nil || !ok {
		return nil, fmt.Errorf("unable to read query tracking value")
	}

	return queryTracking, nil
}

func doDestroyCurrentSession(ctx *sdk.Context, queryTracking *gasTracking) error {
	currentSession := queryTracking.activeSessions[len(queryTracking.activeSessions)-1]
	if !currentSession.gasFilledIn {
		return fmt.Errorf("vm gas is not recorded in query tracking")
	}

	queryTracking.mainGasMeter.ConsumeGas(currentSession.meter.GasConsumed(), "contract sub-query")

	queryTracking.sessionRecords = append(queryTracking.sessionRecords, &SessionRecord{
		ActualSDKGas:    currentSession.meter.GetActualGas() + currentSession.actualStoreSDKGas,
		OriginalSDKGas:  currentSession.meter.GetOriginalGas() + currentSession.originalStoreSDKGas,
		ContractAddress: currentSession.meter.GetContractAddress(),
		OriginalVMGas:   currentSession.originalVMGas,
		ActualVMGas:     currentSession.actualVMGas,
	})
	queryTracking.activeSessions = queryTracking.activeSessions[:len(queryTracking.activeSessions)-1]

	// Revert to previous gas meter
	if len(queryTracking.activeSessions) != 0 {
		*ctx = ctx.WithGasMeter(queryTracking.activeSessions[len(queryTracking.activeSessions)-1].meter)
	} else {
		*ctx = ctx.WithGasMeter(queryTracking.mainGasMeter)
	}

	return nil
}

func IsGasTrackingInitialized(ctx sdk.Context) bool {
	_, err := getGasTrackingData(ctx)
	return err == nil
}

func InitializeGasTracking(ctx *sdk.Context, initialContractGasMeter *ContractSDKGasMeter) error {
	data := ctx.Value(gasTrackingKey)
	if data != nil {
		return fmt.Errorf("query gas tracking is already initialized")
	}

	queryTracking := gasTracking{
		depth:        0,
		mainGasMeter: ctx.GasMeter(),
		activeSessions: []*activeSession{
			{
				meter: initialContractGasMeter,
			},
		},
		sessionRecords: nil,
	}

	*ctx = ctx.WithValue(gasTrackingKey, &queryTracking)
	*ctx = ctx.WithGasMeter(initialContractGasMeter)
	return nil
}

func TerminateGasTracking(ctx *sdk.Context) ([]*SessionRecord, *SessionRecord, error) {
	queryTracking, err := getGasTrackingData(*ctx)
	if err != nil {
		return nil, nil, err
	}

	if queryTracking.depth != 0 {
		return nil, nil, fmt.Errorf("cannot terminate gas tracking as there are sessions in progress")
	}

	if len(queryTracking.activeSessions) != 1 {
		if len(queryTracking.activeSessions) == 0 {
			return nil, nil, fmt.Errorf("internal error: the initial contract gas meter not found")
		} else {
			return nil, nil, fmt.Errorf("internal error: multiple active gas trackers in session")
		}
	}

	if err := doDestroyCurrentSession(ctx, queryTracking); err != nil {
		return nil, nil, err
	}

	*ctx = ctx.WithValue(gasTrackingKey, nil)
	*ctx = ctx.WithGasMeter(queryTracking.mainGasMeter)

	querySessionRecords := queryTracking.sessionRecords[:len(queryTracking.sessionRecords)-1]
	txSessionRecord := queryTracking.sessionRecords[len(queryTracking.sessionRecords)-1]

	return querySessionRecords, txSessionRecord, nil
}

func AddVMRecord(ctx sdk.Context, vmRecord *VMRecord) error {
	queryTracking, err := getGasTrackingData(ctx)
	if err != nil {
		return err
	}

	if len(queryTracking.activeSessions) == 0 {
		return fmt.Errorf("internal error: no active sessions")
	}

	lastSession := queryTracking.activeSessions[len(queryTracking.activeSessions)-1]
	if lastSession.gasFilledIn {
		return fmt.Errorf("gas information already present for current session")
	}

	lastSession.gasFilledIn = true
	lastSession.originalVMGas = vmRecord.OriginalVMGas
	lastSession.actualVMGas = vmRecord.ActualVMGas
	lastSession.originalStoreSDKGas = vmRecord.OriginalStoreSDKGas
	lastSession.actualStoreSDKGas = vmRecord.ActualStoreSDKGas

	return nil
}

func AssociateMeterWithCurrentSession(ctx *sdk.Context, gasMeterFn func(gasLimit uint64) *ContractSDKGasMeter) error {
	queryTracking, err := getGasTrackingData(*ctx)
	if err != nil {
		return err
	}

	contractGasMeter := gasMeterFn(queryTracking.limitForUpcomingGasMeter)
	queryTracking.activeSessions = append(queryTracking.activeSessions, &activeSession{
		meter:       contractGasMeter,
		gasFilledIn: false,
	})

	*ctx = ctx.WithGasMeter(contractGasMeter)
	return nil
}

func CreateNewSession(ctx sdk.Context, gasLimitForSession uint64) error {
	queryTracking, err := getGasTrackingData(ctx)
	if err != nil {
		return err
	}

	queryTracking.depth += 1
	queryTracking.limitForUpcomingGasMeter = gasLimitForSession
	return nil
}

func DestroySession(ctx *sdk.Context) error {
	queryTracking, err := getGasTrackingData(*ctx)
	if err != nil {
		return err
	}

	if queryTracking.depth == 0 {
		return fmt.Errorf("trying to destroy last session which does not exists")
	}

	if queryTracking.depth < uint64(len(queryTracking.activeSessions))-1 {
		return fmt.Errorf("internal data corruption: mismatch between number of sessions and active gas meters")
	}

	if queryTracking.depth == uint64(len(queryTracking.activeSessions))-1 {
		if err := doDestroyCurrentSession(ctx, queryTracking); err != nil {
			return err
		}
	}
	queryTracking.depth -= 1

	return nil
}
