package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
	"strings"
	"sync"
)

type codeable struct {
	id interface{}
	code interface{}
}


type commons struct {
	Conn *pgxpool.Pool
	debugger Debugger
	nomenclators sync.Map
	pendingInsert sync.Map
	tables_relation map[string]string
}

type queableCode struct {
	code string
	data *[]interface{}
}

func (ptr *commons) cacheUsefulLifeByType()  {
	conn, error := ptr.Conn.Acquire(context.Background())
	defer conn.Release()
	if error != nil {
		debugger.error(error.Error())
		os.Exit(2)
	}
	usefulLifeQuery := "select sf.useful_life, t.code as tid from sub_families sf inner join types t on sf.id = t.sub_family_id"

	rows, err :=conn.Query(context.Background(), usefulLifeQuery)
	if err != nil {
		ptr.debugger.error(err.Error())
		os.Exit(1)
	}
	dest := make([]codeable,0)
	var code codeable
	for rows.Next() {
		rows.Scan(&code.code,&code.id)
		dest = append(dest, code)
	}
	ptr.nomenclators.Store("useful_life_by_type", &dest)
}

func (ptr *commons) init() {
	cd := make(map[string]string)
	cd["acquisition_types"] = "id, code"
	cd["sub_families"] = "id, code"
	cd["types"] = "id, code"
	cd["brands"] = "id, name"
	cd["asset_models"] = "id, name"
	cd["edifices"] = "id, code, name"
	cd["centers"] = "id, code, name"
	cd["levels"] = "id, code, name"
	cd["spaces"] = "id, code, name"
	cd["cost_centers"] = "id, code"
	cd["entities"] = "id, name"
	cd["colors"] = "id, code"
	cd["materials"] = "id, code"
	cd["areas"] = "id, code, name"
	cd["operation_vars"] = "id, code"
	cd["tags"] = "id, full_code"
	ptr.tables_relation = cd
	ptr.loadCodeablesNomenclators(make([]string,0))
	ptr.cacheUsefulLifeByType()
}


func (ptr *commons ) isEnqueuedForInsert(key string, code string)  (interface{},bool){
	queue, _ := ptr.pendingInsert.Load(key)
	if  queue != nil {
		for _, c := range queue.([]queableCode) {
			if c.code == code {
				return queue,true
			}
		}
	}
	return queue,false
}


func (ptr *commons ) batchInsert(wg *sync.WaitGroup){
	defer wg.Done()

	ptr.pendingInsert.Range(func(key interface{}, value interface{} ) bool {

		conn, error := ptr.Conn.Acquire(context.Background())
		defer conn.Release()
		if error != nil {
			debugger.error(error.Error())
			os.Exit(2)
		}
		query := getQueryForBatchInsert(key.(string))
		batch := pgx.Batch{}
		for _, queable := range value.([]queableCode) {
			if key == "imageables" {
				fmt.Println(*queable.data, queable.code)
				fmt.Println()
			}
			batch.Queue(query,*queable.data...)
		}
		if key == "imageables" {
			os.Exit(1)
		}
		r := conn.SendBatch(context.Background(), &batch)
		for i := 0; i < batch.Len(); i++ {
			_, err :=r.Exec()
			if err != nil {
				cache.debugger.debug(err.Error(), "  ",key)
			}
		}
		defer ptr.pendingInsert.Delete(key)
		return true
	})
}

func (ptr *commons ) enqueueForInsert(key string, code string, params *[]interface{})  bool{
	data, enqueued := ptr.isEnqueuedForInsert(key, code)
	if  enqueued == false {
		if data == nil {
			data = make([]queableCode,0)
		}

		queable := queableCode{code: code, data: params}
		data = append(data.([]queableCode), queable)
		ptr.pendingInsert.Store(key, data)
		return false
	}
	return true
}


func (ptr *commons) loadCodeablesNomenclators(tables []string )  {
	 tablesToProcess := make(map[string]string)
	if len(tables) >0 {
		for _, table := range tables {
			tablesToProcess[table] = ptr.tables_relation[table]
		}
	}else{
		tablesToProcess = ptr.tables_relation
	}
	len := len(tablesToProcess)
	wg := &sync.WaitGroup{}
	wg.Add(len)
	for key, val := range tablesToProcess {
		go ptr.refresh(key, val, wg)
	}
	wg.Wait()

}

func(ptr *commons) refresh(table string, val string, wg *sync.WaitGroup)  {
	defer wg.Done()
	conn := ptr.Conn
	paramsCount := len(strings.Split(val,","))
	container := make([]interface{},paramsCount)
	for i := 0; i < paramsCount; i++ {
		index := i
		var a interface{}
		container[index] = &a
	}
	target := table
	query := fmt.Sprintf("select %v from %v order by id asc ",val, target)

	rows, err :=conn.Query(context.Background(), query)
	if err != nil {
		ptr.debugger.debug(err.Error())
		return
	}
	dest := make([]codeable,0)
	var code codeable
	for rows.Next() {
		if paramsCount >2 {
			rows.Scan(container...)
			code.id = *(container[0].(*interface{}))
			code.code = make([]interface{},0)
			for _, i2 := range container[1:] {
				val:= ( *(i2.(*interface{})))
				code.code= append(code.code.([]interface{}), val)
			}
			//a := container[0].(*interface{})
		}else{
			rows.Scan(&code.id,&code.code)
		}

		dest = append(dest, code)
	}

	ptr.nomenclators.Store(table, &dest)
}

func (ptr *commons) findByCode(key string, code string) interface{}{
	list , _ := ptr.nomenclators.Load(key)
	for _, c := range *list.(*[]codeable) {
		switch   (c.code).(type) {
			case []interface{}:
				if (c.code).([]interface{})[0] == code {
					return c.id
				}

			case  interface{}:
				if c.code == code {
					return int(c.id.(int64))
				}
			case string:
				if c.code == code {
					return c.id
				}
		}
	}
	return -1
}



func (ptr *commons) fetch(key string, code string) interface{}{
	list , _ := ptr.nomenclators.Load(key)
	for _, c := range *list.(*[]codeable) {
		switch  (c.code).(type) {
			case []interface{}:
				if (c.code).([]interface{})[0] == code {
					return c
				}
			case interface{}:
			case string:
				if c.code == code {
					return c
				}
		}
	}
	return -1
}


func (ptr *commons) findById(key string, id int) interface{}{
	list , _ := ptr.nomenclators.Load(key)

	for _, c := range *list.(*[]codeable) {
		realId := realint64(c.id)
		if realId == int64(id) {
			return c.code
		}
	}
	return -1
}

func realint64(tipe interface{}) int64{
	switch tipe.(type) {
	case int:
		return int64(tipe.(int))
	case int64:
		return tipe.(int64)
	}
	return tipe.(int64)
}



