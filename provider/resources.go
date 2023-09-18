package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func clusterItem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"replicas": {
				Type:         schema.TypeSet,
				Required:     true,
				Description:  "Number of replicas to maintain",
				ValidateFunc: validateReplicas,
				Default: 1,
			},
			"ports": {
				Type:         schema.TypeSet,
				Required:     true,
				Description:  "Ports to run Kafka on",
				ValidateFunc: validatePorts,
			},
		},
		Create: clusterCreateItem,
		// Read:   clusterReadItem,
		// Update: clusterUpdateItem,
		// Delete: clusterDeleteItem,
		// Exists: clusterExistsItem,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func clusterCreateItem(resData *schema.ResourceData, m interface{}) error {

	setupkafka := setupKafka(resData)
	if setupkafka != nil {
		return fmt.Errorf("error: %s", setupkafka)
	}

    return nil
}

