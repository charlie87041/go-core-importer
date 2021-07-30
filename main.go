package main

import (
	"context"
	"encoding/json"
	"fmt"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"sync"
	"time"
)

var serverChannel <-chan *redis2.Message

var cache *commons
var debugger Debugger
var assetChunk = 800
var redisClient *redis2.Client
var config Config

func init() {
	//os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/b")
	debugger = &defaultdebugger{isDebugEnabled: true, timer: time.Now(), status: 0, errors: make([]interface{}, 0)}

	config.loadConfig()

	redisClient = startRedis()
	server := redisClient.Subscribe(context.Background(), config.RedisChannel)
	fmt.Println("subscribing to ", fmt.Sprint(config.RedisChannel), " on redis url ", config.RedisUrl, " with debug ", config.Debug)
	//self:= redisClient.Subscribe(context.Background(), fmt.Sprint(channelName,"_go"))
	serverChannel = server.Channel()
}

// structure {"event":{"name":"asset:import"},"data":{"0":"b","1":"a","2":"{}","socket":null},"socket":null}
func main() {
	for message := range serverChannel {
		var msj Message
		if err := json.Unmarshal([]byte(message.Payload), &msj); err != nil {
			log.Printf("Error on unmarshal JSON message %s", err)
			return
		}
		switch msj.Event["name"] {
		case "asset:import":
			db := msj.Data["0"]
			file := msj.Data["1"]
			preload(db)
			doLoad(file)
		}
	}

}

func postLoad() {
	status := []string{fmt.Sprint("status", cache.debugger.getStatus())}
	errors := cache.debugger.getErrors()
	message := make(map[string]interface{})
	message["status"] = status
	message["errors"] = errors
	response, _ := json.MarshalIndent(message, "", "\n")
	redisClient.Publish(context.Background(), fmt.Sprint(config.RedisChannel, "_response"), response)
}

func preload(db string) {
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", fmt.Sprint("postgres://postgres:postgres@localhost:5432/", db))
	}
	conn, error := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if error != nil {
		debugger.error("error conectando a ", os.Getenv("DATABASE_URL"))
		os.Exit(1)
	}
	debugger.setTimer(time.Now())
	query := "Select code from generic_assets order by id desc limit 1;"
	lastRegisteredCode := conn.QueryRow(context.Background(), query)
	lastRegisteredCode.Scan(&lastAssetCodeInDatabase)

	tagquery := "Select code from tags  where tag_type_id = 1 order by id desc limit 1;"
	vc := conn.QueryRow(context.Background(), tagquery)
	vc.Scan(&lastVirtualTagCodeInDatabase)

	optypequery := "Select id from operation_types  where code like '01';"
	ot := conn.QueryRow(context.Background(), optypequery)
	ot.Scan(&subscriptionOperationType)

	common := &commons{Conn: conn, debugger: debugger}
	cache = common
	common.init()

}

func doLoad(fileDir string) {
	defer cache.Conn.Close()

	assets := loadData(fileDir)
	LoadNomenclators(&assets)
	debugger.chrono("Finished handling nomenclation data at ")
	chunks := 5
	assetsPerCycle := chunks * assetChunk
	assetLen := len(assets)
	debugger.debug(" found assets ", assetLen)
	var (
		aspectsForInsert  [][][]interface{}
		initializedAssets []*GenericAsset
		excelRows         []*ExcelAsset
	)
	if assetsPerCycle > assetLen {
		aspectsForInsert, initializedAssets, excelRows = importAsset(assets)
	} else {
		aspectsForInsert, initializedAssets, excelRows = assetsInBatch(assets, assetsPerCycle, chunks)
	}
	if len(aspectsForInsert) > 0 {
		var wg sync.WaitGroup
		wg.Add(1)
		cache.batchInsert(&wg)
		wg.Wait()
		makePatrimonialOperations(initializedAssets)
		doAspectsBatch(aspectsForInsert)
		cache.nomenclators.Store("assets_ref", initializedAssets)
		cache.nomenclators.Store("assets_data", excelRows)
		postScripts()
		//release
		cache.nomenclators = sync.Map{}
	}
	debugger.chrono("Done with assets, aspects and Jesus as well at")
}

func assetsInBatch(assets []*ExcelAsset, assetsPerCycle int, chunks int) ([][][]interface{}, []*GenericAsset, []*ExcelAsset) {
	assetLen := len(assets)
	rest := assetLen % assetsPerCycle
	cycles := assetLen / assetsPerCycle
	var mu sync.Mutex
	var (
		aspectsForInsert  [][][]interface{}
		initializedAssets []*GenericAsset
		excelRows         []*ExcelAsset
	)
	debugger.chrono(fmt.Sprintf("Required %v cycles of %v assets and %v goroutines in order to process %v assets", cycles, assetsPerCycle, chunks, assetLen))
	for i := 0; i < assetLen; i += assetsPerCycle {
		cIndex := i
		var wg sync.WaitGroup
		wg.Add(chunks)
		for j := 0; j < chunks; j++ {
			lastIndex := cIndex + assetChunk
			if lastIndex > assetLen {
				wg.Done()
				continue
			}
			assetsForInsert := assets[cIndex:lastIndex]
			go func(wg *sync.WaitGroup, a []*ExcelAsset) {
				defer wg.Done()
				tmpInsert, tmpA, tmE := importAsset(a)
				if tmpInsert != nil {
					mu.Lock()
					aspectsForInsert = append(aspectsForInsert, tmpInsert...)
					initializedAssets = append(initializedAssets, tmpA...)
					excelRows = append(excelRows, tmE...)
					mu.Unlock()
				}
			}(&wg, assetsForInsert)
			cIndex = lastIndex
		}
		wg.Wait()
	}
	if rest > 0 {
		restAsset := assets[(assetLen - rest):assetLen]
		tmpInsert, tmpA, tmE := importAsset(restAsset)
		aspectsForInsert = append(aspectsForInsert, tmpInsert...)
		initializedAssets = append(initializedAssets, tmpA...)
		excelRows = append(excelRows, tmE...)
	}
	return aspectsForInsert, initializedAssets, excelRows
}
