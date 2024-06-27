package keeper

import (
	"fmt"
	"path/filepath"

	"github.com/cometbft/cometbft/libs/os"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetSchema(ctx sdk.Context, codeID uint64) (string, error) {
	filePath := filepath.Join(k.dataRoot, fmt.Sprintf("%d", codeID))
	contents, err := os.ReadFile(filePath)
	return string(contents), err
}

func (k Keeper) SetSchema(ctx sdk.Context, codeID uint64, schema string) error {
	filePath := filepath.Join(k.dataRoot, fmt.Sprintf("%d", codeID))
	return os.WriteFile(filePath, []byte(schema), 0644)
}
