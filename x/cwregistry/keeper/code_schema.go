package keeper

import (
	"fmt"
	"path/filepath"

	"github.com/cometbft/cometbft/libs/os"
)

func (k Keeper) GetSchema(codeID uint64) (string, error) {
	filePath := filepath.Join(k.dataRoot, fmt.Sprintf("%d", codeID))
	contents, err := os.ReadFile(filePath)
	return string(contents), err
}

func (k Keeper) SetSchema(codeID uint64, schema string) error {
	filePath := filepath.Join(k.dataRoot, fmt.Sprintf("%d", codeID))
	return os.WriteFile(filePath, []byte(schema), 0644)
}
