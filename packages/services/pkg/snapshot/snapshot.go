package snapshot

import (
	"fmt"
	"io/ioutil"
	"latticexyz/mud/packages/services/pkg/logger"
	"latticexyz/mud/packages/services/pkg/utils"
	pb "latticexyz/mud/packages/services/protobuf/go/ecs-snapshot"

	"math"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

// A SnapshotType distinguishes between snapshot types if those are required.
type SnapshotType int

const (
	Latest        SnapshotType = iota // latest available snapshot
	BlockSpecific                     // snapshot at a specific block
	InitialSync                       // snapshot taken right after the service has performed a sync
)

// checks whether a state snapshot exists for a given world
func IsAvailableLatest(worldAddress string) bool {
	_, err := os.Stat(getFilenameLatest(worldAddress))
	return err == nil
}

func typeToName(snapshotType SnapshotType) (string, error) {
	if snapshotType == Latest {
		return "latest", nil
	} else if snapshotType == BlockSpecific {
		return "block range", nil
	} else if snapshotType == InitialSync {
		return "initial sync", nil
	} else {
		return "", fmt.Errorf("received unsupported SnapshotType")
	}
}

func takeStateSnapshotChain(chainState ChainECSState, startBlockNumber uint64, endBlockNumber uint64, snapshotType SnapshotType) {
	logger.GetLogger().Info("taking full chain state snapshot",
		zap.String("category", "Snapshot"),
	)
	worldAddresses := []string{}
	for worldAddress, ecsStateForWorld := range chainState {
		// Collect the list of worlds.
		worldAddresses = append(worldAddresses, worldAddress)

		// Serialize the state for a given world and take a snapshot.
		takeStateSnapshot(ecsStateForWorld, worldAddress, startBlockNumber, endBlockNumber, snapshotType)
	}
	// Take a snapshot of the world addresses.
	takeWorldAddressesSnapshot(worldAddresses)
}

func takeStateSnapshot(state ECSState, worldAddress string, startBlockNumber uint64, endBlockNumber uint64, snapshotType SnapshotType) {
	encoding := encodeState(state, startBlockNumber, endBlockNumber)

	snapshotTypeName, err := typeToName(snapshotType)
	if err != nil {
		logger.GetLogger().Fatal("received an unsupported SnapshotType", zap.Int("snapshotType", int(snapshotType)))
	}
	logger.GetLogger().Info("taking snapshot",
		zap.String("category", "Snapshot"),
		zap.String("type", snapshotTypeName),
		zap.Uint64("startBlockNumber", startBlockNumber),
		zap.Uint64("endBlockNumber", endBlockNumber),
	)

	if snapshotType == Latest {
		writeStateLatest(encoding, worldAddress)
	} else if snapshotType == BlockSpecific {
		writeStateAtBlock(encoding, endBlockNumber)
	} else if snapshotType == InitialSync {
		writeStateInitialSync(encoding)
	}
}

func readStateSnapshotAtBlock(blockNumber uint64) ECSState {
	logger.GetLogger().Info("reading snapshot",
		zap.String("category", "Snapshot"),
		zap.Uint64("blockNumber", blockNumber),
	)
	return decodeState(readStateAtBlock(blockNumber))
}

func readStateSnapshotLatest(worldAddress string) ECSState {
	logger.GetLogger().Info("reading latest snapshot", zap.String("category", "Snapshot"))
	return decodeState(readStateLatest(worldAddress))
}

// RawReadStateSnapshotLatest returns the latest ECS state snapshot in protobuf format.
func RawReadStateSnapshotLatest(worldAddress string) *pb.ECSStateSnapshot {
	logger.GetLogger().Info("reading latest raw snapshot", zap.String("category", "Snapshot"), zap.String("worldAddress", worldAddress))
	return decode(readStateLatest(worldAddress))
}

// ChunkRawStateSnapshot splits a rawStateSnapshot ECSStateSnapshot in protobuf format into a list
// of ECSStateSnapshot's also in protobuf format. Each ECSStateSnapshot after chunking is
// chunkPercentage fraction size of the original snapshot.
func ChunkRawStateSnapshot(rawStateSnapshot *pb.ECSStateSnapshot, chunkPercentage int) []*pb.ECSStateSnapshot {
	chunked := []*pb.ECSStateSnapshot{}
	chunkIdx := 0
	chunkSize := int(math.Ceil(float64(len(rawStateSnapshot.State))/float64(100))) * chunkPercentage

	logger.GetLogger().Info("start chunking raw state snapshot", zap.String("category", "Snapshot"), zap.Int("fullStateLength", len(rawStateSnapshot.State)), zap.Int("chunkSize", chunkSize), zap.String("chunkPercentage", fmt.Sprintf("%d%%", chunkPercentage)))
	tsStart := time.Now()

	for chunkIdx < len(rawStateSnapshot.State) {
		chunkUpperBound := func(a, b int) int {
			if a < b {
				return a
			}
			return b
		}

		stateSlice := rawStateSnapshot.State[chunkIdx:chunkUpperBound(chunkIdx+chunkSize, len(rawStateSnapshot.State))]

		// List of ECSState, components, and entities re-indexed since here we are working with a slice of
		// the complete state.
		reIndexedStateSlice := []*pb.ECSState{}
		reIndexedComponents := []string{"0x0"}
		reIndexedEntities := []string{"0x0"}

		// Map of components / entities to their position in an array. This helps us
		// assign the correct values to the ECSState slices as we build the snapshot.
		componentToIdx := map[string]uint32{}
		entitiyToIdx := map[string]uint32{}

		// Indexes tracking the position of every component and entity in the array.
		componentIdx := uint32(1)
		entityIdx := uint32(1)

		for _, state := range stateSlice {
			// Get the actual string values for the component and entity.
			componentId := rawStateSnapshot.StateComponents[state.ComponentIdIdx]
			entityId := rawStateSnapshot.StateEntities[state.EntityIdIdx]

			// Add the component to the list and to the mapping.
			if _, ok := componentToIdx[componentId]; !ok {
				reIndexedComponents = append(reIndexedComponents, componentId)
				componentToIdx[componentId] = componentIdx
				componentIdx++
			}

			// Add the entity to the list and to the mapping.
			if _, ok := entitiyToIdx[entityId]; !ok {
				reIndexedEntities = append(reIndexedEntities, entityId)
				entitiyToIdx[entityId] = entityIdx
				entityIdx++
			}

			// Extend the ECSState by creating a new re-mapped element. The value stays the
			// same, but the component and entity indexes are now pointing to the components
			// and entities that are specific to this slice.
			reIndexedState := &pb.ECSState{
				ComponentIdIdx: componentToIdx[componentId],
				EntityIdIdx:    entitiyToIdx[entityId],
				Value:          state.Value,
			}
			reIndexedStateSlice = append(reIndexedStateSlice, reIndexedState)
		}

		chunk := &pb.ECSStateSnapshot{
			State:            reIndexedStateSlice,
			StateComponents:  reIndexedComponents,
			StateEntities:    reIndexedEntities,
			StateHash:        rawStateSnapshot.StateHash,
			StartBlockNumber: rawStateSnapshot.StartBlockNumber,
			EndBlockNumber:   rawStateSnapshot.EndBlockNumber,
		}
		chunked = append(chunked, chunk)
		chunkIdx += chunkSize
	}

	tsElapsed := time.Since(tsStart)
	logger.GetLogger().Info("done chunking raw state snapshot", zap.String("category", "Snapshot"), zap.Int("numChunks", len(chunked)), zap.String("timeTaken", tsElapsed.String()))

	return chunked
}

///
/// World list state snapshots.
///

func worldSnapshotToWorldAddressList(snapshot *pb.Worlds) []string {
	worldAddresses := []string{}
	for _, worldAddress := range snapshot.WorldAddress {
		worldAddresses = append(worldAddresses, worldAddress)
	}
	return worldAddresses
}

func worldAddressListToWorldsSnapshot(worldAddresses []string) *pb.Worlds {
	worlds := &pb.Worlds{}

	for _, worldAddress := range worldAddresses {
		worlds.WorldAddress = append(worlds.WorldAddress, worldAddress)
	}
	return worlds
}

func takeWorldAddressesSnapshot(worldAddresses []string) {
	logger.GetLogger().Info("taking world addresses snapshot",
		zap.String("category", "Snapshot"),
		zap.Int("countAdresses", len(worldAddresses)),
	)
	encoding := encodeWorldAddresses(worldAddresses)

	if err := ioutil.WriteFile(WorldsFilename, encoding, 0644); err != nil {
		logger.GetLogger().Fatal("failed to write World addresses state", zap.String("fileName", WorldsFilename), zap.Error(err))
	}
}

func readWorldAddressesSnapshot() []string {
	if !IsWorldAddressSnapshotAvailable() {
		return []string{}
	}
	encoding := readState(WorldsFilename)
	worlds := decodeWorldAddresses(encoding)
	worldAddressList := worldSnapshotToWorldAddressList(worlds)
	return worldAddressList
}

// RawReadWorldAddressesSnapshot returns a snapshot of all indexed World addresses in protobuf
// format.
func RawReadWorldAddressesSnapshot() *pb.Worlds {
	return decodeWorldAddresses(readState(WorldsFilename))
}

// IsWorldAddressSnapshotAvailable returns if a snapshot of all indexed World addresses is
// available.
func IsWorldAddressSnapshotAvailable() bool {
	_, err := os.Stat(WorldsFilename)
	return err == nil
}

// PruneSnapshotOwnedByComponent prunes a given ECSStateSnapshot, given an address.
// This helps get rid of unnecessary state that a given address does not depend on in order
// to perform actions.
func PruneSnapshotOwnedByComponent(snapshot *pb.ECSStateSnapshot, pruneForAddress string) *pb.ECSStateSnapshot {
	// Default to 'OwnedBy' componentId, since that's the component that stores information
	// about entities owned by specific addresses, and we can discard those that are not
	// the ones for the given address.
	pruneComponentId := "0xaf90be6cd7aa92d6569a9ae629178b74e1b0fbdd1097c27ec1dfffd2dc4c7540"
	prunedState := []*pb.ECSState{}

	// Iterate all state and lookup the component for each.
	for _, stateEntry := range snapshot.State {
		componentId := snapshot.StateComponents[stateEntry.ComponentIdIdx]
		if componentId == pruneComponentId {
			// Extract the address that is the 'value' of OwnedBy.
			ownedByValue := hexutil.Encode(stateEntry.Value[12:])
			// Discard this state entry if the value is not for the specified address.
			if utils.ChecksumAddressString(ownedByValue) != utils.ChecksumAddressString(pruneForAddress) {
				continue
			}
		}
		prunedState = append(prunedState, stateEntry)
	}

	percentSizeAfterPrune := float64(len(prunedState)) / float64(len(snapshot.State))
	logger.GetLogger().Info("pruned snapshot", zap.String("pruneForAddress", pruneForAddress), zap.Float64("percentSizeAfterPrune", percentSizeAfterPrune))

	return &pb.ECSStateSnapshot{
		State:            prunedState,
		StateComponents:  snapshot.StateComponents,
		StateEntities:    snapshot.StateEntities,
		StateHash:        snapshot.StateHash,
		StartBlockNumber: snapshot.StartBlockNumber,
		EndBlockNumber:   snapshot.EndBlockNumber,
		WorldAddress:     snapshot.WorldAddress,
	}
}
