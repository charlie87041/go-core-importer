package main

type Firstinvestment struct {
	is_main bool
	validated bool
	genericassetid int64
	id int64
	purchase_order_id int64
	acquisition_type_id int64
	start_assigment_date string
	end_assigment_date string
	created_at string
	updated_at string

}

func (ptr *Firstinvestment) init(excelAsset *ExcelAsset) {
	now := now()
	acquisitionCode := excelAsset.AcquisitionTypeCode
	acquisitionId := (cache.findByCode("acquisition_types", acquisitionCode)).(int)
	if acquisitionId < 1 {
		acquisitionId = 1
	}
	ptr.created_at = now
	ptr.updated_at = now
	ptr.is_main = true
	ptr.validated = true
	ptr.genericassetid = excelAsset.id
	ptr.purchase_order_id = 1
	ptr.start_assigment_date = excelAsset.BorrowStart
	ptr.end_assigment_date = excelAsset.BorrowEnd
	ptr.acquisition_type_id = int64(acquisitionId)
}
func (ptr *Firstinvestment) after(asset *GenericAsset)  []string {
	return make([]string,0)
}
