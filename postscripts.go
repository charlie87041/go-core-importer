package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"os"
	"sync"
)

//post insert operations



func postScripts(){
	commonBatch := func(conn *pgx.Conn, batch *pgx.Batch) {
		br := cache.Conn.SendBatch(context.Background(), batch)
		for i := 0; i < batch.Len(); i++ {
			_, error:= br.Exec()
			if error != nil {
				debugger.debug("postscripts batch ",error.Error())
			}
		}
	}
	now := now()
	conns := 5
	var wg sync.WaitGroup
	debugger.chrono("started with postscripts at ")
	connections := make([]*pgx.Conn,0)
	for i := 0; i < conns; i++ {
		conn,error := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if error != nil {
			debugger.debug("error on postscripts conn ", error.Error())
		}

		connections = append(connections, conn)
	}
	wg.Add(conns - 1)
	go func(wg *sync.WaitGroup, conn *pgx.Conn) {
		defer wg.Done()
		query := assetMakePatrimonialSituationANDStatus(now)
		_, error := conn.Exec(context.Background(), query)
		if error != nil {
			debugger.error(error.Error())
			os.Exit(1)
		}
	}(&wg, connections[0])
	go func(wg *sync.WaitGroup, conn *pgx.Conn) {
		defer wg.Done()
		batch := updateFisicalTagsAsset()
		commonBatch(conn, &batch)

	}(&wg, connections[1])
	go func(wg *sync.WaitGroup, conn *pgx.Conn) {
		defer wg.Done()
		batch := updateVirtualTagsStatus()
		commonBatch(conn, &batch)
	}(&wg, connections[2])
	go func(wg *sync.WaitGroup, conn *pgx.Conn) {
		defer wg.Done()
		batch := updateAssociatedTerrainAndBuilding()
		commonBatch(conn, &batch)
	}(&wg, connections[3])
	wg.Wait()
	clearTemporalId(connections[4]);
	for _, connection := range connections {
		connection.Close(context.Background())
	}
	debugger.chrono("ended postscripts at ")
}

func assetMakePatrimonialSituationANDStatus(now string) string{
	a:= fmt.Sprintf("with v2 as (select t.patrimonial_situation_type_id as psit, po.id as operation_id, t.id, asset.id as asset_id from acquisition_types t inner join generic_assets asset on asset.acquisition_type_id =t.id inner join patrimonial_operations po on (asset.id =po.operationable_id and po.operationable_type ilike %s)) INSERT INTO public.asset_patrimonial_situations( validated, generic_asset_id, patrimonial_situation_type_id, created_at, updated_at, patrimonial_operation_id) select true, v2.asset_id, v2.psit, '%s', '%s', v2.operation_id from v2;","'%GenericAssetOverridden%'", now, now)+
		fmt.Sprintf("with v3 as (select id from generic_assets) INSERT INTO public.statuses( name, model_type, model_id, created_at, updated_at) select 'subscribed', 'App\\Models\\Asset\\GenericAssetOverridden', v3.id, '%v', '%v' from v3;", now, now)

	return a
}

func updateFisicalTagsAsset() pgx.Batch{
	excelRows,_ := cache.nomenclators.Load("assets_data")
	batch := pgx.Batch{}
	query := "update tags set generic_asset_id =%v, center_id =%v where full_code like '%v'"
	for _, row := range excelRows.([]*ExcelAsset) {
		tagType := row.TagType[0]
		if tagType == 'F'{
			centerId := cache.findByCode("centers", row.CenterCode)
			t := fmt.Sprintf(query, row.id,centerId, row.TagCode)
			batch.Queue(t)
		}
	}
	return batch;
}

func updateVirtualTagsStatus()  pgx.Batch{

	table := []string{"tags"}
	cache.loadCodeablesNomenclators( table)
	tags,_ := cache.nomenclators.Load("tags")
	batch := pgx.Batch{}
	now := now()
	for _, tag := range *tags.(*[]codeable) {
		code := (tag.code).(string)
		query := getQueryForBatchInsert("pending_tags_status")
		if code[0] == 'V' {
			dataCreated := []interface{}{"'created'","'App\\Models\\Tag\\Tag'", tag.id, now, now}
			dataInUse := []interface{}{"'inuse'","'App\\Models\\Tag\\Tag'", tag.id, now, now}
			batch.Queue(query, dataCreated...)
			batch.Queue(query, dataInUse...)
		}else{
			continue
		}
	}
	return batch;
}


func clearTemporalId(conn *pgx.Conn){
	_, err :=conn.Exec(context.Background(),"update generic_assets set temporal_id = ''")
	if err != nil {
		debugger.debug(err.Error())
	}
}

func updateAssociatedTerrainAndBuilding()  pgx.Batch{
	assetRef,_ := cache.nomenclators.Load("assets_data")
	batch := pgx.Batch{}
	queryBuildingAssetTag := "update generic_assets set associated_building_tag = '$1' where id = $2"
	queryTerrainAssetTag := "update generic_assets set associated_surface_tag = '$1' where id = $2"
	queryBuilding := "update general_aspects set associated_building = '$1' where generic_asset_id = $2"
	queryTerrain := "update general_aspects set associated_surface = '$1' where generic_asset_id = $2"
	for _, assetRef := range assetRef.([]*ExcelAsset) {
		if assetRef.AssociatedBuilding != "" {
			validAssociated := (cache.findByCode("asset_tag_ref", assetRef.AssociatedBuilding)).(int)
			if validAssociated == -1{
					debugger.error("wrong  associatedbuilding tag", assetRef.AssociatedBuilding, "for asset with id", assetRef.id)
					continue
			}
			batch.Queue(queryBuildingAssetTag, assetRef.AssociatedBuilding,assetRef.id)
			batch.Queue(queryBuilding, validAssociated,assetRef.id)
		}
		if assetRef.AssociatedTerrain != "" {
			validAssociated := (cache.findByCode("asset_tag_ref", assetRef.AssociatedTerrain)).(int)
			if validAssociated == -1{
					debugger.error("wrong  associatedsurface tag", assetRef.AssociatedTerrain, "for asset with id", assetRef.id)
					continue
			}
			batch.Queue(queryTerrainAssetTag, assetRef.AssociatedTerrain,assetRef.id)
			batch.Queue(queryTerrain, validAssociated,assetRef.id)
		}

	}
	return batch
}

