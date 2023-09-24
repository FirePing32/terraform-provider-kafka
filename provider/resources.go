package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/FirePing32/terraform-provider-kafka/helpers"
)

func clusterItem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of cluster",
				ValidateFunc: helpers.ValidateName,
			},
			"replicas": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Number of brokers to maintain",
				ValidateFunc: helpers.ValidateReplicas,
				Default: 1,
			},
			"ports": {
				Type:         schema.TypeList,
				Required:     true,
				Description:  "Ports to run Kafka on",
				ValidateFunc: helpers.ValidatePorts,
			},
		},
		Create: clusterCreateItem,
		Read:   clusterReadItem,
		// Update: clusterUpdateItem,
		// Delete: clusterDeleteItem,
		// Exists: clusterExistsItem,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func clusterCreateItem(resData *schema.ResourceData, m interface{}) error {

	setupkafka := helpers.SetupKafka(resData)
	if setupkafka != nil {
		return fmt.Errorf("error: %s", setupkafka)
	}

	startkafka := helpers.StartKafka(resData)
	if startkafka != nil {
		return fmt.Errorf("error: %s", startkafka)
	}

    return nil
}

func clusterReadItem(resData *schema.ResourceData, m interface{}) error {
	return nil
}

