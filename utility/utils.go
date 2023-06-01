package utility

import (
	"encoding/json"
	"fmt"

	"github.com/mabhi/mimic-helm-mvp/models"
)

func Deserialize(mapData map[string]interface{}, dest *models.HelmAction) {
	// Convert map to json string
	jsonStr, err := json.Marshal(mapData)
	if err != nil {
		fmt.Println(err)
	}

	// Convert json string to struct
	if err := json.Unmarshal(jsonStr, dest); err != nil {
		fmt.Println(err)
	}
}
