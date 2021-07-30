package main

import (
	"context"
	"github.com/jackc/pgx/v4"
	"os"
	"sync"
)

func doAspectsBatch(data [][][]interface{}) {

	batchQueryG := &pgx.Batch{}
	batchQueryLoc := &pgx.Batch{}
	batchQueryLeg := &pgx.Batch{}
	batchQueryInv := &pgx.Batch{}
	operationsCache, _ := cache.nomenclators.Load("operations_ref")
	op := operationsCache.(map[int]int)
	for _, datum := range data{
		assetId := int(datum[2][1].(int64))
		operationId := op[assetId]
		if operationId <= 0{
			debugger.debug("Invalid operation id ", operationId, " assetId ", assetId)
		}
		datum[1][8] = operationId
		datum[3][10] = operationId

		makeAspectBatch(datum, batchQueryG,batchQueryLoc,batchQueryLeg,batchQueryInv)
	}
	wg := &sync.WaitGroup{}
	wg.Add(4)
	go runAspectBatch(batchQueryG, wg)
	go runAspectBatch(batchQueryLoc, wg)
	go runAspectBatch(batchQueryLeg, wg)
	go runAspectBatch(batchQueryInv, wg)
	wg.Wait()
}

func makeAspectBatch(data [][]interface{}, batch ...*pgx.Batch) {
	for i := 0; i < len(batch); i++ {
		iIndex := i
		currentBatch := batch[iIndex]
		params := data[iIndex]
		query := params[0].(string)
		rst := params[1:]
		currentBatch.Queue(query, rst...)
	}
}

func runAspectBatch(batchQuery *pgx.Batch, wg *sync.WaitGroup) {
	defer wg.Done()
	conn , err := cache.Conn.Acquire(context.Background())
	if err != nil {
		debugger.error(err.Error())
		os.Exit(2)
	}
	defer conn.Release()
	batchLen := batchQuery.Len()
	r := conn.SendBatch(context.Background(), batchQuery)
	for i := 0; i < batchLen; i++ {
		_, _err := r.Exec()
		if _err != nil {
			debugger.error("Error inserting aspects ", _err.Error())
			os.Exit(1)
		}
	}

}
