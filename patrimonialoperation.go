package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"os"
	"strings"
	"sync"
	"time"
)

func patrimonialOperation(asset *GenericAsset) []interface{}{
	now := now()
	created_at := now
	updated_at := now
	validated_at := now
	validated_by := 1
	user_id := 1
	acquuisitionType := asset.acquisition_type_id
	acquisitionTypeCode := cache.findById("acquisition_types", int(acquuisitionType))
	operation_var_id := 1
	if acquuisitionType >1{
		operation_var_id = (cache.findByCode("operation_vars", acquisitionTypeCode.(string))).(int)
	}
	operationable_id := asset.id
	operationable_type := "App\\Models\\Asset\\GenericAssetOverridden"
	return []interface{}{nil, true, validated_at, validated_by, subscriptionOperationType, operation_var_id, user_id,operationable_id, operationable_type, created_at, updated_at}
}

func patrimonialOperationCode(operationVar int, pos int) string{
	year := time.Now().Year()
	operationTypeCode := "01"
	operationNumberPad := strings.Repeat("0",5)
	operationNumber := fmt.Sprint(operationNumberPad, pos)
	operationVarCode := cache.findById("operation_vars", operationVar)
	code := fmt.Sprint(year, operationTypeCode, operationVarCode, operationNumber)
	return code
}
func lastOperationInCurrentYear() int{
	year := time.Now().Year()
	conn, _err := cache.Conn.Acquire(context.Background())
	if _err != nil {
		debugger.error(_err.Error())
		os.Exit(2)
	}
	defer conn.Release()
	query := fmt.Sprintf("select count(id) from patrimonial_operations where EXTRACT( year from created_at) = '%v'", year)
	row :=conn.QueryRow(context.Background(),query )
	tot := 0
	row.Scan(&tot)
	return tot
}

func makePatrimonialOperations(assets []*GenericAsset) {
	initialSize := lastOperationInCurrentYear()
	queryArgs := make([]interface{},0)
	for i, asset := range assets {
		iIndex := i
		size := iIndex+initialSize
		data := patrimonialOperation(asset)
		operationVar := data[5]
		code := patrimonialOperationCode(operationVar.(int), size)
		data[0] = code
		queryArgs = append(queryArgs, data)
	}
	batchPatrimonialOperations(queryArgs)
}


func batchPatrimonialOperations(data []interface{}){

	dataLen := len(data)
	var wg sync.WaitGroup
	batches := dataLen/5000
	rest := dataLen%5000
	routines  := batches
	if rest >0{
		routines += 1
	}
	ch := make(chan []interface{},routines)
	wg.Add(routines)
	for i := 0; i < batches; i+=routines {
		index := i * 5000
		assets := data[index : index + 5000]
		go processoperatins(assets,&wg, ch)
	}
	if rest > 0 {
		rstAsset := data[dataLen-rest : dataLen]
		go processoperatins(rstAsset,&wg, ch)
	}
	wg.Wait()
	operationsResult := make(map[int]int)
	for i:= 0; i< routines;i++{
		r := <-ch
		for i := 0; i < len(r); i+=2 {
			iIndex := i
			operationsResult[r[iIndex].(int)] =r[iIndex+1].(int)
		}
	}
	close(ch)
	cache.nomenclators.Store("operations_ref", operationsResult)

}

func processoperatins(operations []interface{}, wg *sync.WaitGroup,channel  chan []interface{}){

	dataLen := len(operations)
	defer wg.Done()
	conn, err := cache.Conn.Acquire(context.Background())
	if err != nil {
		debugger.error(err.Error())
		os.Exit(1)
	}
	defer conn.Release()
	batch := &pgx.Batch{}
	query := getQueryForBatchInsert("patrimonial_operations")
	for _, operation := range operations {
		batch.Queue(query, operation.([]interface{})...)
	}
	r := conn.SendBatch(context.Background(),batch)

	mp := make([]interface{},0)
	for i := 0; i < dataLen; i++ {
		assetId:= 0
		opId := 0
		row := r.QueryRow()
		row.Scan( &opId, &assetId)
		mp = append(mp,  opId, assetId)
	}
	channel <- mp
}