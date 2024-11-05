package helpers

import (
	"AinedIndexItemCLI/db_models"
	"AinedIndexItemCLI/db_models/interfaces"
	"AinedIndexItemCLI/db_models/objects"
	"encoding/json"
	"gorm.io/gorm"
	"strconv"
	"sync"
)

var (
	once     = sync.Once{}
	BuildCat map[string]string
)

func GetBuildingCategories() map[string]string {
	once.Do(func() {
		var categories = make(map[string]string)
		categories["mkd"] = "МКД"
		categories["stock"] = "Склад"
		categories["private_house"] = "Частный дом"
		categories["cottage"] = "Коттедж"
		categories["townhouse"] = "Таунхаус"
		categories["business"] = "Бизнес-центр"
		categories["mall"] = "Торговый центр"
		categories["parking"] = "Паркинг"
		BuildCat = categories
	})
	return BuildCat
}

func GetTitle(item objects.Item, room objects.Room) string {
	title := ""
	switch item.Category {
	case "flat":
		if room.Rooms != 0 {
			title = strconv.Itoa(room.Rooms) + "-к квартира"
		} else {
			title = "Квартира"
		}
		break
	case "office":
		title = "Офис"
		break
	case "parking_space":
		title = "Машиноместо"
		break
	}

	if room.AreaFull != 0 {
		title += " " + strconv.FormatFloat(room.AreaFull, 'f', 2, 64) + "м²"
	}
	if room.NumberObject != "" {
		title += " №" + room.NumberObject
	}
	return title
}

func GetSubtitle(item objects.Item, DB *gorm.DB) string {
	subtitle := ""
	buildCat := GetBuildingCategories()
	var parent db_models.Tree
	DB.Table("trees").First(&parent, item.NodeID)
	if (parent != db_models.Tree{}) {
		if parent.ModelType == "Item" {

			jk := FindFirstComplex(parent, DB)
			if (jk != objects.Complex{}) {
				subtitle = "ЖК " + jk.Name
			}
			if parent.ModelID.UUIDValue.String() != "" {
				var itemDB objects.Item
				DB.Table("items").First(&itemDB, parent.ModelID.UUIDValue)
				if (itemDB != objects.Item{}) {
					category := buildCat[itemDB.Category]
					if category != "" {
						if subtitle != "" {
							subtitle += " / "
						}
						subtitle += category
					}
					var build objects.Building
					DB.Table("buildings").First(&build, itemDB.ModelID)
					if (build != objects.Building{}) {
						if build.NumberObject != "" {
							subtitle += " №" + build.NumberObject
						}
					}
				}

			}

		}

	}

	return subtitle
}

func GetParents(item objects.Item) {

}

func GetModelsByNode(nodes []db_models.Tree, DB *gorm.DB) []interfaces.Object {
	var models []interfaces.Object
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for _, node := range nodes {
		wg.Add(1)
		go func(node db_models.Tree) {
			defer wg.Done()
			model := GetModelByNode(node, DB)
			if model != nil {
				mu.Lock()
				models = append(models, model)
				mu.Unlock()
			}
		}(node)
	}
	wg.Wait()
	return models
}

func GetModelByNode(node db_models.Tree, DB *gorm.DB) interfaces.Object {
	switch node.ModelType {
	case "Item":
		var item objects.Item
		DB.Table("items").First(&item, node.ModelID.UUIDValue)
		if (item != objects.Item{}) {
			var build objects.Building
			DB.Table("buildings").First(&build, item.ModelID)
			if (build != objects.Building{}) {
				return build
			}
		}
		break
	case "BaseGroup":
		var item objects.BaseGroup
		DB.Table("base_groups").First(&item, node.ModelID.IntValue)
		if (item != objects.BaseGroup{}) {
			return item
		}
		break
	case "ComplexGroup":
		var item objects.Complex
		DB.Table("complex_groups").First(&item, node.ModelID.IntValue)
		if (item != objects.Complex{}) {
			return item
		}
		break
	}
	return nil
}

func FindFirstComplex(node db_models.Tree, DB *gorm.DB) objects.Complex {
	var complexGroup interfaces.Object
	if node.ModelType != "ComplexGroup" {
		var nodes []db_models.Tree
		if node.ParentID != 0 {
			if err := DB.Find(&nodes, node.ParentID).Error; err != nil {
				panic(err)
			}
			var wg sync.WaitGroup
			for _, item := range nodes {
				wg.Add(1)
				go func(item db_models.Tree) {
					defer wg.Done()
					if item.ModelType != "ComplexGroup" {
						FindFirstComplex(item, DB)
					} else {
						complexGroup = GetModelByNode(item, DB)
						return
					}
				}(item)
			}
			wg.Wait()
		}

	} else {
		complexGroup = GetModelByNode(node, DB)
	}
	return ConvertToComplex(complexGroup)
}

func ConvertToComplex(jk interface{}) objects.Complex {
	complexByte, err := json.Marshal(jk)
	if err != nil {
		panic(err)
	}
	var comGroup objects.Complex
	if err := json.Unmarshal(complexByte, &comGroup); err != nil {
		panic(err)
	}
	return comGroup
}
