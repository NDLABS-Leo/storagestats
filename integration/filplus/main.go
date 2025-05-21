package main

import (
	"context"
	"math/rand"
	"time"

	logging "github.com/ipfs/go-log/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"storagestats/integration/filplus/util"
	"storagestats/pkg/env"
	"storagestats/pkg/model"
	"storagestats/pkg/resolver"
	"storagestats/pkg/task"
)

var logger = logging.Logger("filplus-integration")

func main() {
	filplus := NewFilPlusIntegration()
	for {
		// Step 1: Group the deals by client and provider
		dealsGrouped, err2 := getDealsGroupedByClientProvider(filplus.marketDealsCollection)
		if err2 != nil {
			logger.Error(err2)
			continue
		}

		// Initialize timer
		start := time.Now()

		// Initialize random generator
		rand.Seed(time.Now().UnixNano())

		// Step 2: Take the top 30% of each group and perform random sampling
		for client, providerDeals := range dealsGrouped {
			for provider, deals := range providerDeals {
				// Take the top 30%
				top30Count := int(float64(len(deals)) * 0.30)
				if top30Count == 0 {
					continue
				}
				top30Deals := deals[:top30Count]

				// Randomly sample up to 100 deals
				sampleCount := 100
				if len(top30Deals) < sampleCount {
					sampleCount = len(top30Deals)
				}

				sampledDeals := make([]model.DealState, 0, sampleCount)
				indices := rand.Perm(len(top30Deals))[:sampleCount]
				for _, idx := range indices {
					sampledDeals = append(sampledDeals, top30Deals[idx])
				}

				err := filplus.RunOnce(context.TODO(), sampledDeals)
				if err != nil {
					logger.Error(err)
				}

				// Log sampled deal info
				logger.Infof("Client: %s, Provider: %s, Sampled Deals: %d\n", client, provider, len(sampledDeals))
				for _, deal := range sampledDeals {
					logger.Infof("DealID: %d, PieceCID: %s\n", deal.DealID, deal.PieceCID)
				}
			}
		}

		// Record end time and print duration
		elapsed := time.Since(start)
		logger.Infof("Processing dealsGrouped took: %s", elapsed)

		//time.Sleep(time.Minute * 1)
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

// getDealsGroupedByClientProvider groups deals by client and provider, sorted by deal_id
func getDealsGroupedByClientProvider(collection *mongo.Collection) (map[string]map[string][]model.DealState, error) {

	cursor, err := collection.Aggregate(context.Background(), mongo.Pipeline{
		{
			{"$sort", bson.D{{"deal_id", -1}}}, // Sort by deal_id
		}, // Sort by deal_id descending
		{{"$limit", 30000000}}, // Limit to the most recent 30 million deals
	})

	if err != nil {
		logger.Errorf("Failed to aggregate data: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var deals []model.DealState
	if err := cursor.All(context.Background(), &deals); err != nil {
		logger.Errorf("Failed to decode data: %v", err)
		return nil, err
	}

	groupedDeals := make(map[string]map[string][]model.DealState)
	for _, deal := range deals {
		if _, ok := groupedDeals[deal.Client]; !ok {
			groupedDeals[deal.Client] = make(map[string][]model.DealState)
		}
		groupedDeals[deal.Client][deal.Provider] = append(groupedDeals[deal.Client][deal.Provider], deal)
	}

	return groupedDeals, nil
}

// sampleTop30Percent randomly samples from the top 30% of deals
func randomSampleFromGroup(deals []model.DealState, sampleSize int) []model.DealState {
	// Randomize the order of the deals
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deals), func(i, j int) {
		deals[i], deals[j] = deals[j], deals[i]
	})

	// Ensure we take no more than sampleSize items
	if len(deals) > sampleSize {
		deals = deals[:sampleSize]
	}

	return deals
}

func (f *FilPlusIntegration) RunOnce(ctx context.Context, documentsOne []model.DealState) error {
	logger.Info("start running filplus integration")

	for {
		// Check the number of tasks in the queue
		count, err := f.taskCollection.CountDocuments(ctx, bson.M{"requester": f.requester})
		if err != nil {
			return errors.Wrap(err, "failed to count tasks")
		}

		logger.With("count", count).Info("Current number of tasks in the queue")

		// If task count exceeds batch size, block and wait
		if count > int64(f.batchSize) {
			logger.Infof("Task queue still has %d tasks, waiting...", count)
			time.Sleep(10 * time.Second) // Wait for 10 seconds before rechecking
			continue
		}

		// Break loop when conditions are met and proceed to insert tasks
		break
	}

	//documents = RandomObjects(documents, len(documents)/2, f.randConst, totalPerClient)
	tasks, results := util.AddTasks(ctx, f.requester, f.ipInfo, documentsOne, f.locationResolver, f.providerResolver)

	if len(tasks) > 0 {
		_, err := f.taskCollection.InsertMany(ctx, tasks)
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
		_, err := f.resultCollection.InsertMany(ctx, results)
		if err != nil {
			return errors.Wrap(err, "failed to insert results")
		}
	}

	logger.With("count", len(results)).Info("inserted results")

	return nil
}
