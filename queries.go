package main

import (
	"os"
)

func getQueryForBatchInsert(key string) string{
	switch key {

	case "materials" :
		return "INSERT INTO public.materials(name, code, description, created_at, updated_at)  VALUES ($1, $2, $3, $4, $5)"

	case "colors" :
		return "INSERT INTO public.colors(name, code, description, created_at, updated_at)  VALUES ($1, $2, $3, $4, $5)"

	case "areas" :
		return "INSERT INTO public.areas(name, code, description, created_at, updated_at)  VALUES ($1, $2, $3, $4, $5)"

	case "brands" :
		return "INSERT INTO public.brands(name, code, description, manufacturer_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)"

	case "asset_models" :
		return "INSERT INTO public.asset_models(name, code, description, brand_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)"

	case "edifices" :
		return "INSERT INTO public.edifices(name, code, scode, description, center_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"

	case "levels" :
		return "INSERT INTO public.levels(name, code, scode, description, edifice_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"

	case "spaces" :
		return "INSERT INTO public.spaces(name, code, scode, description, level_id, validated,  created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"

	case "generic_assets" :
		return "INSERT INTO public.generic_assets( verified, subscription_date, code, internal_item_code,  acquisition_type_id, center_id, associated_surface_tag, associated_building_tag, validated,type_id,created_at, updated_at,  operational_state, temporal_id, code, useful_life) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14, $15) returning id,temporal_id"

	case "general_aspects" :
		return "INSERT INTO public.general_aspects(description, denomination, address, postal_code, surface_field, surface_build, length, characteristics, width, height, depth, serial_number, plate, frame_number, old_tag_code, validated, risk_level, catastro_reference, generic_asset_id, asset_model_id, created_at, updated_at, color_id, material_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)"

	case "localization_aspects" :
		return "INSERT INTO public.localization_aspects(validated, active,  generic_asset_id, space_id, created_at, updated_at, area_id, patrimonial_operation_id,str_value) VALUES ( $1, $2, $3, $4, $5, $6, $7,$8,$9)"

	case "virtual_tags" :
		return "INSERT INTO public.tags(code, full_code, tag_type_id,  created_at, updated_at,generic_asset_id,center_id) VALUES ( $1, $2, $3, $4,$5,$6,$7)"


	case "legal_aspects" :
		return "INSERT INTO public.legal_aspects(generic_asset_id, proprietary_id, assignee_id, created_at, updated_at)VALUES ( $1, $2, $3, $4, $5)"

	case "first_investment" :
		return "INSERT INTO public.investments(start_assigment_date, end_assigment_date, validated, is_main, generic_asset_id, acquisition_type_id, purchase_order_id, created_at, updated_at, patrimonial_operation_id) VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9,$10)"


	case "patrimonial_operations" :
		return "INSERT INTO public.patrimonial_operations(code, validated, validated_at, validated_by_id, operation_type_id, operation_var_id, user_id,  operationable_id, operationable_type, created_at, updated_at) VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning operationable_id,id"

	case "pending_tags_status" :
		return "INSERT INTO public.statuses( name, model_type, model_id, created_at, updated_at) values ($1, $2, $3, $4, $5)"

	case "update_tag" :
		return "update tags set generic_asset_id = $1,center_id=$2 where id = $3"

	case "images" :
		return `INSERT INTO "public"."images" ("name", "title",  "url", "image_type_id", "album_id", "created_at", "updated_at") VALUES ($1, $2,  '/images/assets', 1, 1, $3, $4)`

	case "imageables" :
		return `INSERT INTO "public"."imageables" ("image_id",  "imageable_type", "imageable_id", "created_at", "updated_at") VALUES ( $1, 'AppModelsAssetGenericAssetOverridden',$2, $3, $4)`

	}
	cache.debugger.error(key, " no updateable")
	os.Exit(1)
	return ""
}
