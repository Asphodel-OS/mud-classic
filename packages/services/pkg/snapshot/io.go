package snapshot

import (
	"fmt"
	"io/ioutil"
	"latticexyz/mud/packages/services/pkg/logger"
	"latticexyz/mud/packages/services/pkg/utils"

	"go.uber.org/zap"
)

const (
	StateFilename  string = "./snapshots/SerializedECSState" // prefix of the state snapshot filename
	WorldsFilename string = "./snapshots/SerializedWorlds"   // name for the snapshot binary of Worlds
)

func getFilenameAtBlock(endBlockNumber uint64) string {
	return fmt.Sprintf("%s-%d", StateFilename, endBlockNumber)
}

func getFilenameLatest(worldAddress string) string {
	// Always lookup a snapshot with a checksummed address, since that is how they are written to
	// disk.
	return fmt.Sprintf("%s-latest-%s", StateFilename, utils.ChecksumAddressString(worldAddress))
}

func readStateAtBlock(blockNumber uint64) []byte {
	return readState(getFilenameAtBlock(blockNumber))
}

func readStateLatest(worldAddress string) []byte {
	return readState(getFilenameLatest(worldAddress))
}

func readState(fileName string) []byte {
	encoding, err := ioutil.ReadFile(fileName)
	if err != nil {
		logger.GetLogger().Fatal("failed to read encoded state", zap.String("fileName", fileName), zap.Error(err))
	}
	return encoding
}

func writeStateInitialSync(encoding []byte) {
	filename := fmt.Sprintf("%s-initial-sync", StateFilename)
	writeState(encoding, filename)
}

func writeStateAtBlock(encoding []byte, endBlockNumber uint64) {
	writeState(encoding, getFilenameAtBlock(endBlockNumber))
}

func writeStateLatest(encoding []byte, worldAddress string) {
	writeState(encoding, getFilenameLatest(worldAddress))
}

func writeState(encoding []byte, fileName string) {
	if err := ioutil.WriteFile(fileName, encoding, 0644); err != nil {
		logger.GetLogger().Fatal("failed to write ECSState", zap.String("fileName", fileName), zap.Error(err))
	}
}
