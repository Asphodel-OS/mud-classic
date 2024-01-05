package snapshot

import (
	"fmt"
	"latticexyz/mud/packages/services/pkg/logger"
	"latticexyz/mud/packages/services/pkg/utils"
	pb "latticexyz/mud/packages/services/protobuf/go/ecs-snapshot"
	"os"

	"go.uber.org/zap"
)

const (
	StateFilename  string = "./snapshots/SerializedECSState" // prefix of the state snapshot filename
	WorldsFilename string = "./snapshots/SerializedWorlds"   // name for the snapshot binary of Worlds
)

/////////////////
// State

func getStateFilenameAtBlock(endBlockNumber uint64) string {
	return fmt.Sprintf("%s-%d", StateFilename, endBlockNumber)
}

// Lookup snapshot with a checksummed address, as written to disk.
func getWorldStateFilenameLatest(worldAddress string) string {
	return fmt.Sprintf("%s-latest-%s", StateFilename, utils.ChecksumAddressString(worldAddress))
}

func ReadStateAtBlock(blockNumber uint64) ECSState {
	logger.GetLogger().Info("reading snapshot",
		zap.String("category", "Snapshot"),
		zap.Uint64("blockNumber", blockNumber),
	)
	return decodeState(readRawStateAtBlock(blockNumber))
}

func ReadStateLatest(worldAddress string) ECSState {
	logger.GetLogger().Info("reading latest snapshot", zap.String("category", "Snapshot"))
	return decodeState(readRawStateLatest(worldAddress))
}

// ReadPbStateLatest returns the latest ECS state snapshot in protobuf format.
func ReadPbStateLatest(worldAddress string) *pb.ECSStateSnapshot {
	logger.GetLogger().Info("reading latest raw snapshot", zap.String("category", "Snapshot"), zap.String("worldAddress", worldAddress))
	return decode(readRawStateLatest(worldAddress))
}

// read state from disk
func readRawState(fileName string) []byte {
	encoding, err := os.ReadFile(fileName)
	if err != nil {
		logger.GetLogger().Fatal("failed to read encoded state", zap.String("fileName", fileName), zap.Error(err))
	}
	return encoding
}

func readRawStateAtBlock(blockNumber uint64) []byte {
	return readRawState(getStateFilenameAtBlock(blockNumber))
}

func readRawStateLatest(worldAddress string) []byte {
	return readRawState(getWorldStateFilenameLatest(worldAddress))
}

// write state to disk
func writeRawState(encoding []byte, fileName string) {
	if err := os.WriteFile(fileName, encoding, 0644); err != nil {
		logger.GetLogger().Fatal("failed to write ECSState", zap.String("fileName", fileName), zap.Error(err))
	}
}

func writeRawStateAtBlock(encoding []byte, endBlockNumber uint64) {
	writeRawState(encoding, getStateFilenameAtBlock(endBlockNumber))
}

func writeRawStateInitialSync(encoding []byte) {
	filename := fmt.Sprintf("%s-initial-sync", StateFilename)
	writeRawState(encoding, filename)
}

func writeRawStateLatest(encoding []byte, worldAddress string) {
	writeRawState(encoding, getWorldStateFilenameLatest(worldAddress))
}

/////////////////
// Worlds (addresses)

func writeWorlds(worldAddresses []string) {
	logger.GetLogger().Info("taking world addresses snapshot",
		zap.String("category", "Snapshot"),
		zap.Int("countAdresses", len(worldAddresses)),
	)
	encoding := encodeWorldAddresses(worldAddresses)

	if err := os.WriteFile(WorldsFilename, encoding, 0644); err != nil {
		logger.GetLogger().Fatal("failed to write World addresses state", zap.String("fileName", WorldsFilename), zap.Error(err))
	}
}

func readWorlds() []string {
	if !IsWorldAddressSnapshotAvailable() {
		return []string{}
	}
	encoding := readRawState(WorldsFilename)
	worlds := decodeWorldAddresses(encoding)
	worldAddressList := pbToWorldAddresses(worlds)
	return worldAddressList
}
