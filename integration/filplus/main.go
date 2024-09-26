package main

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/data-preservation-programs/RetrievalBot/integration/filplus/util"
	"github.com/data-preservation-programs/RetrievalBot/pkg/env"
	"github.com/data-preservation-programs/RetrievalBot/pkg/model"
	"github.com/data-preservation-programs/RetrievalBot/pkg/resolver"
	"github.com/data-preservation-programs/RetrievalBot/pkg/task"
	logging "github.com/ipfs/go-log/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var logger = logging.Logger("filplus-integration")

type Cache struct {
	Data      []bson.M
	Timestamp time.Time
}

var cache Cache

func getCachedData(collection *mongo.Collection) []bson.M {
	// Check if the cache is more than 24 hours old
	if time.Since(cache.Timestamp).Hours() >= 12 {
		// Cache is outdated, fetch fresh data
		logger.Infof("Fetching fresh data...")

		groupStage := bson.D{
			{"$group", bson.D{
				{"_id", bson.D{
					{"client", "$client"},
					{"provider", "$provider"},
				}},
				{"documents", bson.D{{"$push", "$$ROOT"}}},
			}},
		}

		// Aggregate pipeline
		cursor, err := collection.Aggregate(context.TODO(), mongo.Pipeline{groupStage})
		if err != nil {
			logger.Errorf("failed to Aggregate %v", err)
			return nil
		}
		defer cursor.Close(context.TODO())

		var pipelineResults []bson.M
		for cursor.Next(context.TODO()) {
			var result bson.M
			if err := cursor.Decode(&result); err != nil {
				logger.Errorf("Error decoding aggregation results: %v", err)
			}
			pipelineResults = append(pipelineResults, result)

		}

		// Update cache
		cache.Data = pipelineResults
		cache.Timestamp = time.Now()
	} else {
		logger.Infof("Using cached data...")
	}

	return cache.Data
}

func main() {
	filplus := NewFilPlusIntegration()
	// Grouping by client and provider
	for {
		// Fetch or update cache
		pipelineResults := getCachedData(filplus.marketDealsCollection)

		// Process each group
		for _, group := range pipelineResults {
			sampledDeal := make([]model.DealState, 100)
			documents := group["documents"].(bson.A)
			// Sorting documents by a certain criterion (e.g., timestamp or other)
			// Assuming that the documents have a "timestamp" field
			sort.Slice(documents, func(i, j int) bool {
				return documents[i].(bson.M)["deal_id"].(int32) > documents[i].(bson.M)["deal_id"].(int32)
			})

			// Get the top 40%
			top40Percent := documents[:int(float64(len(documents))*0.4)]

			// Random sampling from the top 40%
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(top40Percent), func(i, j int) {
				top40Percent[i], top40Percent[j] = top40Percent[j], top40Percent[i]
			})

			sampledData := top40Percent
			if len(top40Percent) > 100 {
				sampledData = top40Percent[:100]
			}

			logger.With("sampledDataCount", len(top40Percent)).Info("sampledDataCount")

			// Perform operations on the sampled data
			for i, document := range sampledData {
				if bsonDoc, ok := document.(bson.M); ok {
					sampledDeal[i] = model.DealState{
						DealID:      bsonDoc["deal_id"].(int32),
						PieceCID:    bsonDoc["piece_cid"].(string),
						PieceSize:   bsonDoc["piece_size"].(int64),
						Label:       bsonDoc["label"].(string),
						Verified:    bsonDoc["verified"].(bool),
						Client:      bsonDoc["client"].(string),
						Provider:    bsonDoc["provider"].(string),
						Start:       bsonDoc["start"].(int32),
						End:         bsonDoc["end"].(int32),
						SectorStart: bsonDoc["sector_start"].(int32),
						Slashed:     bsonDoc["slashed"].(int32),
						LastUpdated: bsonDoc["last_updated"].(int32),
					}
				}
			}

			err := filplus.RunOnce(context.TODO(), sampledDeal)
			if err != nil {
				logger.Error(err)
			}
		}

		time.Sleep(time.Minute)
	}
}

type TotalPerClient struct {
	Client string `bson:"_id"`
	Total  int64  `bson:"total"`
}

type FilPlusIntegration struct {
	taskCollection        *mongo.Collection
	marketDealsCollection *mongo.Collection
	resultCollection      *mongo.Collection
	batchSize             int
	requester             string
	locationResolver      resolver.LocationResolver
	providerResolver      resolver.ProviderResolver
	ipInfo                resolver.IPInfo
	randConst             float64
}

func GetTotalPerClient(ctx context.Context, marketDealsCollection *mongo.Collection) (map[string]int64, error) {
	var result []TotalPerClient
	agg, err := marketDealsCollection.Aggregate(ctx, []bson.M{
		{"$match": bson.M{
			"sector_start": bson.M{"$gt": 0},
			"end":          bson.M{"$gt": model.TimeToEpoch(time.Now())},
			"verified":     true,
			"slashed":      bson.M{"$lt": 0},
		}},
		{
			"$group": bson.M{
				"_id": "$client",
				"total": bson.M{
					"$sum": "$piece_size",
				},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to aggregate market deals")
	}

	err = agg.All(ctx, &result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode market deals")
	}

	totalPerClient := make(map[string]int64)
	for _, r := range result {
		totalPerClient[r.Client] = r.Total
	}

	return totalPerClient, nil
}

func NewFilPlusIntegration() *FilPlusIntegration {
	ctx := context.Background()
	taskClient, err := mongo.
		Connect(ctx, options.Client().ApplyURI(env.GetRequiredString(env.QueueMongoURI)))
	if err != nil {
		panic(err)
	}
	taskCollection := taskClient.
		Database(env.GetRequiredString(env.QueueMongoDatabase)).Collection("task_queue")

	stateMarketDealsClient, err := mongo.
		Connect(ctx, options.Client().ApplyURI(env.GetRequiredString(env.StatemarketdealsMongoURI)))
	if err != nil {
		panic(err)
	}
	marketDealsCollection := stateMarketDealsClient.
		Database(env.GetRequiredString(env.StatemarketdealsMongoDatabase)).
		Collection("state_market_deals")

	resultClient, err := mongo.Connect(ctx, options.Client().ApplyURI(env.GetRequiredString(env.ResultMongoURI)))
	if err != nil {
		panic(err)
	}
	resultCollection := resultClient.
		Database(env.GetRequiredString(env.ResultMongoDatabase)).
		Collection("task_result")

	batchSize := env.GetInt(env.FilplusIntegrationBatchSize, 100)
	providerCacheTTL := env.GetDuration(env.ProviderCacheTTL, 24*time.Hour)
	locationCacheTTL := env.GetDuration(env.LocationCacheTTL, 24*time.Hour)
	locationResolver := resolver.NewLocationResolver(env.GetRequiredString(env.IPInfoToken), locationCacheTTL)
	providerResolver, err := resolver.NewProviderResolver(
		env.GetString(env.LotusAPIUrl, "https://api.node.glif.io/rpc/v0"),
		env.GetString(env.LotusAPIToken, ""),
		providerCacheTTL)
	if err != nil {
		panic(err)
	}

	// Check public IP address
	ipInfo, err := resolver.GetPublicIPInfo(ctx, "", "")
	if err != nil {
		panic(err)
	}

	logger.With("ipinfo", ipInfo).Infof("Public IP info retrieved")

	return &FilPlusIntegration{
		taskCollection:        taskCollection,
		marketDealsCollection: marketDealsCollection,
		batchSize:             batchSize,
		requester:             "filplus",
		locationResolver:      locationResolver,
		providerResolver:      *providerResolver,
		resultCollection:      resultCollection,
		ipInfo:                ipInfo,
		randConst:             env.GetFloat64(env.FilplusIntegrationRandConst, 4.0),
	}

	
}

func (f *FilPlusIntegration) RunOnce(ctx context.Context, sampledDeal []model.DealState) error {
	logger.Info("start running filplus integration")

	// If the task queue already have batch size tasks, do nothing
	count, err := f.taskCollection.CountDocuments(ctx, bson.M{"requester": f.requester})
	if err != nil {
		return errors.Wrap(err, "failed to count tasks")
	}

	logger.With("count", count).Info("Current number of tasks in the queue")

	if count > int64(f.batchSize) {
		logger.Infof("task queue still have %d tasks, do nothing", count)

		/* Remove old tasks that has stayed in the queue for too long
		_, err = f.taskCollection.DeleteMany(ctx,
			bson.M{"requester": f.requester, "created_at": bson.M{"$lt": time.Now().UTC().Add(-24 * time.Hour)}})
		if err != nil {
			return errors.Wrap(err, "failed to remove old tasks")
		}
		*/
		return nil
	}

	
	tasks, results := util.AddTasks(ctx, f.requester, f.ipInfo, sampledDeal, f.locationResolver, f.providerResolver)

	if len(tasks) > 0 {
		_, err = f.taskCollection.InsertMany(ctx, tasks)
		if err != nil {
			return errors.Wrap(err, "failed to insert tasks")
		}
	}

	logger.With("count", len(tasks)).Info("inserted tasks")

	countPerCountry := make(map[string]int)
	countPerContinent := make(map[string]int)
	countPerModule := make(map[task.ModuleName]int)
	for _, t := range tasks {
		//nolint:forcetypeassert
		tsk := t.(task.Task)
		country := tsk.Provider.Country
		continent := tsk.Provider.Continent
		module := tsk.Module
		countPerCountry[country]++
		countPerContinent[continent]++
		countPerModule[module]++
	}

	for country, count := range countPerCountry {
		logger.With("country", country, "count", count).Info("tasks per country")
	}

	for continent, count := range countPerContinent {
		logger.With("continent", continent, "count", count).Info("tasks per continent")
	}

	for module, count := range countPerModule {
		logger.With("module", module, "count", count).Info("tasks per module")
	}

	if len(results) > 0 {
		_, err = f.resultCollection.InsertMany(ctx, results)
		if err != nil {
			return errors.Wrap(err, "failed to insert results")
		}
	}

	logger.With("count", len(results)).Info("inserted results")

	return nil
}
