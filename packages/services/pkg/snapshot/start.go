package snapshot

import (
	"context"
	"latticexyz/mud/packages/services/pkg/eth"
	"latticexyz/mud/packages/services/pkg/utils"
	"math/big"
	"time"

	"github.com/avast/retry-go"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type SnapshotServerConfig struct {
	// The block number interval for how often to take regular snapshots.
	SnapshotBlockInterval int64
	// The number of blocks to fetch data for when performing an initial
	// sync. This is limited by the bandwidth limit of Geth for fetching logs, which is a hardcoded
	// constant.
	InitialSyncBlockBatchSize int64
	// The time to wait between calls to fetch batched log data when performing an initial sync.
	InitialSyncBlockBatchSyncTimeout time.Duration
	// The block number interval for how often to take intermediary
	// snapshots when performing an initial sync. This is useful in case the snapshot service
	// disconnects or fails while perfoming a lengthy initial sync.
	InitialSyncSnapshotInterval int64

	// Default to use when chunking snapshot to send snapshot in chunks over the wire.
	DefaultSnapshotChunkPercentage int
}

// Start starts the process of processing data from an Ethereum client, reducing the ECS state, and
// taking intermittent snapshots.
func Start(
	state ChainECSState,
	client *ethclient.Client,
	startBlock *big.Int,
	worldAddresses []common.Address,
	config *SnapshotServerConfig,
	logger *zap.Logger,
) {
	// Get an instance of a websocket subscription to the client.
	headers := make(chan *types.Header)
	sub, err := eth.GetEthereumSubscription(client, headers)
	if err != nil {
		logger.Fatal("failed to subscribe to new blockchain head", zap.Error(err))
	}

	snapshotInterval := big.NewInt(config.SnapshotBlockInterval)

	for {
		select {
		case err := <-sub.Err():
			logger.Error("failed to receive event from subscription", zap.Error(err))

			var retrying bool = false
			resubErr := retry.Do(
				func() error {
					var err error
					sub, err = eth.GetEthereumSubscription(client, headers)
					utils.LogErrorWhileRetrying("failed to subscribe to new blockchain head", err, &retrying, logger)
					return err
				},
				utils.ServiceDelayType,
				utils.ServiceRetryAttempts,
				utils.ServiceRetryDelay,
			)

			if resubErr != nil {
				logger.Fatal("failed while retrying to subscribe")
			}

		case header := <-headers:
			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				logger.Error("failed to fetch block by hash", zap.Error(err))
				continue
			}

			// Retrieve world events from block.
			blockNumber := block.Number()
			logs := eth.GetAllEventsInBlock(client, blockNumber, worldAddresses)
			logsFiltered := eth.FilterLogs(logs)

			// Print info about block.
			logger.Info("received block",
				zap.String("category", "Block"),
				zap.Uint64("blockNumber", block.Number().Uint64()),
				zap.String("blockHash", header.Hash().Hex()),
				zap.Int("countLogs", len(logs)),
				zap.Int("countLogsAfterFilter", len(logsFiltered)),
				zap.Int("countTransactions", len(block.Transactions())),
			)

			// Reduce the logs into the state.
			state = reduceLogsIntoState(state, logsFiltered)

			// Take a snapshot every 'SnapshotBlockInterval' blocks.
			if new(big.Int).Mod(blockNumber, snapshotInterval).Cmp(big.NewInt(0)) == 0 {
				// TODO: update this to actually only take a snapshot of the selected worlds
				go createForAll(state, startBlock.Uint64(), blockNumber.Uint64(), Latest)
			}
		}
	}
}
