package main

type LegalAspect struct {
	id int64
	GenericAsset int64
	Proprietary int64
	Assignee int64
	Created_At string
	Updated_At string
}

func (ptr *LegalAspect) init(asset *ExcelAsset)  {
	now := now()
	ownerId := cache.findByCode("entities", asset.OwnerEntity)
	LenderId := cache.findByCode("entities", asset.LendEntity)
	ptr.Created_At = now
	ptr.Updated_At = now
	ptr.Proprietary = int64(ownerId.(int))
	ptr.Assignee = int64(LenderId.(int))

}
func (ptr *LegalAspect) after(asset *GenericAsset)  []string{
	return make([]string,0)
}

