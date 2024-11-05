package main

import (
	"AinedIndexItemCLI/databases"
	"AinedIndexItemCLI/db_models"
	"AinedIndexItemCLI/db_models/objects"
	"AinedIndexItemCLI/helpers"
	"AinedIndexItemCLI/s3"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

func main() {
	runtime.GOMAXPROCS(4)

	start := time.Now()

	config := databases.Config{
		Addresses: []string{"http://localhost:9200"},
	}

	db := databases.GetCon("developer_master")

	s3client, err := s3.NewS3Client()

	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}
	var tenants []db_models.Tenant
	if err := db.Table("tenants").Find(&tenants).Error; err != nil {
		panic(err)
	}

	// Создание клиента Elasticsearch
	client, err := databases.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}
	// Создание экземпляра Indexer
	indx := helpers.NewIndexer(client)

	fmt.Println("Starting delete old documents in items index")
	if err := indx.DeleteOldDocuments("items"); err != nil {
		log.Fatalf("Error deleting old documents: %v", err)
	}
	fmt.Println("Finished delete old documents in items index")

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, tenant := range tenants {
		wg.Add(1)

		go func(tenant db_models.Tenant) {
			defer wg.Done()

			mu.Lock()
			fmt.Println("Starting collect data in tenant: " + tenant.ID.String())
			mu.Unlock()

			statuses := map[string]db_models.Status{}
			var dbStatuses []db_models.Status
			var items []objects.Item
			tenantDB := databases.GetCon(tenant.ID.String())

			tenantDB.Table("status").Find(&dbStatuses)

			for _, status := range dbStatuses {
				statuses[status.Slug] = status
			}

			if err := tenantDB.Table("items").Where("model_type = ?", "Room").Find(&items).Error; err != nil {
				panic(err)
			}
			var elasticItems []map[string]interface{}
			itemChunks := helpers.ItemChunkSlice(items, 2000)

			var itemWg sync.WaitGroup

			for _, itemChunk := range itemChunks {
				itemWg.Add(1)
				go func(items []objects.Item) {
					defer itemWg.Done()
					for _, item := range items {
						var room objects.Room
						tenantDB.Table("rooms").First(&room, item.ModelID)
						if room.Empty() {
							continue
						}
						var formatedImages []string
						if len(room.Images) > 0 {
							for _, image := range room.Images {
								if image != "" {
									formatedImages = append(formatedImages, s3client.GetFilePath(tenant.ID.String(), "upload", image))
								}
							}
						}

						elasticItem := map[string]interface{}{
							"id":             item.ID.String(),
							"tenant_id":      tenant.ID.String(),
							"node_id":        item.NodeID,
							"room_id":        room.ID,
							"title":          helpers.GetTitle(item, room),
							"subtitle":       helpers.GetSubtitle(item, tenantDB),
							"category":       item.Category,
							"plan_type":      room.PlanType,
							"image":          formatedImages,
							"description":    room.Description,
							"rooms":          room.Rooms,
							"status":         statuses[room.Status].Slug,
							"default_status": statuses[room.Status].Default,
							"number_object":  room.NumberObject,
							"floor":          room.Floor,
							"area":           room.AreaFull,
							"price_full":     room.PriceFull,
							"price_unit":     room.PriceUnit,
							"updated_at":     time.Now().UTC().Format("2006-01-02T03:04:05"),
							"created_at":     time.Now().UTC().Format("2006-01-02T03:04:05"),
						}

						mu.Lock()
						elasticItems = append(elasticItems, elasticItem)
						mu.Unlock()
					}
				}(itemChunk)

			}
			itemWg.Wait()
			//for _, item := range items {
			//	var room objects.Room
			//	tenantDB.Table("rooms").First(&room, item.ModelID)
			//	if room.Empty() {
			//		continue
			//	}
			//	var formatedImages []string
			//	if len(room.Images) > 0 {
			//		for _, image := range room.Images {
			//			if image != "" {
			//				formatedImages = append(formatedImages, s3client.GetFilePath(tenant.ID.String(), "upload", image))
			//			}
			//		}
			//	}
			//
			//	elasticItem := map[string]interface{}{
			//		"id":             item.ID.String(),
			//		"tenant_id":      tenant.ID.String(),
			//		"node_id":        item.NodeID,
			//		"title":          helpers.GetTitle(item, room),
			//		"subtitle":       helpers.GetSubtitle(item, tenantDB),
			//		"image":          formatedImages,
			//		"description":    room.Description,
			//		"rooms":          room.Rooms,
			//		"status":         statuses[item.ModelType].Slug,
			//		"default_status": statuses[item.ModelType].Default,
			//		"number_object":  room.NumberObject,
			//		"floor":          room.Floor,
			//		"area":           room.AreaFull,
			//		"price_full":     room.PriceFull,
			//		"price_unit":     room.PriceUnit,
			//		"updated_at":     time.Now().UTC().Format("2006-01-02T03:04:05"),
			//		"created_at":     time.Now().UTC().Format("2006-01-02T03:04:05"),
			//	}

			//	elasticItems = append(elasticItems, elasticItem)
			//}
			//query := map[string]interface{}{
			//	"query": map[string]interface{}{
			//		"bool": map[string]interface{}{
			//			"must": []map[string]interface{}{
			//				{
			//					"term": map[string]interface{}{
			//						"id": item.ID.String(),
			//					},
			//				},
			//				{
			//					"term": map[string]interface{}{
			//						"tenant_id": tenant.ID.String(),
			//					},
			//				},
			//			},
			//		},
			//	},
			//}
			//
			//result, err := indx.SearchData("items", query)
			//if err != nil {
			//	log.Fatalf("Error searching data: %v", err)
			//}
			//
			//mu.Lock()
			//// Вывод результатов поиска
			//fmt.Println(result["hits"].(map[string]interface{})["hits"])
			//mu.Unlock()
			//if len(formatedImages) > 0 {
			//	mu.Lock()
			//	fmt.Println(formatedImages)
			//	mu.Unlock()
			//}

			mu.Lock()
			fmt.Println("Successfully collect data in tenant: " + tenant.ID.String())
			fmt.Println("Starting  indexing items for tenant: " + tenant.ID.String())
			mu.Unlock()

			chunks := helpers.MapChunkSlice(elasticItems, 50)
			chunksWg := sync.WaitGroup{}
			//chunksMu := sync.Mutex{}
			counter := 0

			//for _, item := range elasticItems {
			//	chunksWg.Add(1)
			//	go func(items map[string]interface{}) {
			//		defer chunksWg.Done()
			//		if err := indx.IndexData("items", item); err != nil {
			//			mu.Lock()
			//			log.Fatalf("Error indexing data: %v", err)
			//			mu.Unlock()
			//		}
			//		counter++
			//
			//	}(item)
			//}
			//chunksWg.Wait()

			//p := mpb.New(mpb.WithOutput(color.Output),
			//	mpb.WithAutoRefresh())
			//red, green, blue := color.New(color.FgRed), color.New(color.FgHiGreen), color.New(color.FgBlue)

			chunksWg.Add(len(chunks))
			for _, chunk := range chunks {

				//total := len(chunk)
				//bar := p.AddBar(int64(total),
				//	mpb.PrependDecorators(
				//		decor.Name("Тенант: "+tenant.ID.String()+" Чанк: "+strconv.Itoa(key)+" ", decor.WC{C: decor.DindentRight | decor.DextraSpace}),
				//		decor.OnCompleteMeta(
				//			decor.OnComplete(
				//				decor.Meta(decor.Name("indexing", decor.WCSyncSpaceR), toMetaFunc(red)),
				//				"done",
				//			),
				//			toMetaFunc(green),
				//		),
				//	),
				//	mpb.AppendDecorators(
				//		decor.OnCompleteMeta(
				//			decor.OnComplete(decor.Meta(decor.Percentage(decor.WC{W: 5}, decor.WCSyncSpace), toMetaFunc(blue)), ""),
				//			toMetaFunc(green),
				//		),
				//	),
				//)

				go func(items []map[string]interface{}) {
					defer chunksWg.Done()
					for _, item := range items {
						if err := indx.IndexData("items", item); err != nil {
							mu.Lock()
							log.Fatalf("Error indexing data: %v", err)
							mu.Unlock()
						}
						//bar.Increment()
						counter++
					}
				}(chunk)

			}
			//p.Wait()
			chunksWg.Wait()
			mu.Lock()
			fmt.Println("Indexed ", counter, "documents in ", time.Since(start))
			mu.Unlock()
		}(tenant)
	}
	wg.Wait()

	end := time.Now()
	fmt.Printf("start: %s; end: %s; result: %s;", start.Format("15:04:05"), end.Format("15:04:05"), end.Sub(start).String())
}

//func toMetaFunc(c *color.Color) func(string) string {
//	return func(s string) string {
//		return c.Sprint(s)
//	}
//}
