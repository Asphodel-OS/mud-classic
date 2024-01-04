package snapshot

import (
	"bytes"
	"latticexyz/mud/packages/services/pkg/logger"
	pb "latticexyz/mud/packages/services/protobuf/go/ecs-snapshot"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func encodeState(state ECSState, startBlockNumber uint64, endBlockNumber uint64) []byte {
	stateSnapshot := fromState(state, startBlockNumber, endBlockNumber)
	encoding, err := proto.Marshal(stateSnapshot)
	if err != nil {
		logger.GetLogger().Error("failed to encode ECSState", zap.Error(err))
	}
	return encoding
}

func decodeState(encoding []byte) ECSState {
	stateSnapshot := decode(encoding)
	state := toState(stateSnapshot)
	return state
}

// converts a state to a protobuf ECSStateSnapshot
func fromState(state ECSState, startBlockNumber uint64, endBlockNumber uint64) *pb.ECSStateSnapshot {
	stateSnapshot := &pb.ECSStateSnapshot{}

	var rawStateBuffer bytes.Buffer
	tsStart := time.Now()

	// List of components and entities strings. We pad by one to avoid protobufs omitting
	// our data.
	components := []string{"0x0"}
	entities := []string{"0x0"}

	// Map of components / entities to their position in an array. This helps us
	// assign the correct values to the ECSState slices as we build the snapshot.
	componentToIdx := map[string]uint32{}
	entitiyToIdx := map[string]uint32{}

	// Indexes tracking the position of every component and entity in the array.
	componentIdx := uint32(1)
	entityIdx := uint32(1)

	componentKeys := []string{}
	for k := range state {
		componentKeys = append(componentKeys, k)
	}
	sort.Strings(componentKeys)

	for _, componentId := range componentKeys {
		_state := state[componentId]

		if _, ok := componentToIdx[componentId]; !ok {
			components = append(components, componentId)
			componentToIdx[componentId] = componentIdx
			componentIdx++
		}

		entityKeys := []string{}
		_state.Range(func(key, value interface{}) bool {
			keyString, ok := key.(string)
			if ok {
				entityKeys = append(entityKeys, keyString)
			}
			return true
		})

		sort.Strings(entityKeys)

		for _, entityId := range entityKeys {
			value, ok := _state.Load(entityId)
			if !ok {
				logger.GetLogger().Error("did not find value in map for key", zap.String("key", entityId))
				continue
			}
			valueBytes, ok := value.([]byte)
			if !ok {
				logger.GetLogger().Fatal("value data type expected to be []byte", zap.Any("value", value))
			}

			if _, ok := entitiyToIdx[entityId]; !ok {
				entities = append(entities, entityId)
				entitiyToIdx[entityId] = entityIdx
				entityIdx++
			}

			stateSlice := &pb.ECSState{
				ComponentIdIdx: componentToIdx[componentId],
				EntityIdIdx:    entitiyToIdx[entityId],
				Value:          valueBytes,
			}
			stateSnapshot.State = append(stateSnapshot.State, stateSlice)

			rawStateBuffer.WriteString(componentId)
			rawStateBuffer.WriteString(entityId)
			rawStateBuffer.Write(valueBytes)
		}
	}

	stateSnapshot.StateComponents = components
	stateSnapshot.StateEntities = entities
	stateSnapshot.StateHash = crypto.Keccak256Hash(rawStateBuffer.Bytes()).String()

	stateSnapshot.StartBlockNumber = uint32(startBlockNumber)
	stateSnapshot.EndBlockNumber = uint32(endBlockNumber)

	tsElapsed := time.Since(tsStart)
	logger.GetLogger().Info("computed hash of snapshot", zap.String("category", "Snapshot"), zap.String("keccak256Hash", stateSnapshot.StateHash), zap.String("timeTaken", tsElapsed.String()))

	return stateSnapshot
}

// converts a protobuf ECSStateSnapshot to a state
func toState(stateSnapshot *pb.ECSStateSnapshot) ECSState {
	state := getEmptyState()

	components := stateSnapshot.StateComponents
	entities := stateSnapshot.StateEntities

	for _, stateSlice := range stateSnapshot.State {
		// First read the indexes from the snapshot, then lookup the actual
		// component / entity id values from the array.
		componentIdIdx := stateSlice.ComponentIdIdx
		entityIdIdx := stateSlice.EntityIdIdx
		value := stateSlice.Value

		componentId := components[componentIdIdx]
		entityId := entities[entityIdIdx]

		if _, ok := state[componentId]; !ok {
			state[componentId] = &sync.Map{}
		}
		state[componentId].Store(entityId, value)
	}
	return state
}

// decodes an encoded state file to a protobuf ECSStateSnapshot
func decode(encoding []byte) *pb.ECSStateSnapshot {
	stateSnapshot := &pb.ECSStateSnapshot{}
	if err := proto.Unmarshal(encoding, stateSnapshot); err != nil {
		logger.GetLogger().Error("failed to decode ECSState", zap.Error(err))
	}
	return stateSnapshot
}

func encodeWorldAddresses(worldAddresses []string) []byte {
	worlds := worldAddressesToPB(worldAddresses)
	encoding, err := proto.Marshal(worlds)
	if err != nil {
		logger.GetLogger().Error("failed to encode World addresses", zap.Error(err))
	}
	return encoding
}

func decodeWorldAddresses(encoding []byte) *pb.Worlds {
	worldsSnapshot := &pb.Worlds{}
	if err := proto.Unmarshal(encoding, worldsSnapshot); err != nil {
		logger.GetLogger().Error("failed to decode World addresses snapshot", zap.Error(err))
	}
	return worldsSnapshot
}
