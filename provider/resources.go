package provider

import (
	"fmt"
	"os"
	"io"
	"encoding/json"
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
		Update: clusterUpdateItem,
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

	f, err := os.OpenFile(fmt.Sprint(helpers.KafkaDir, "/clusterdata.json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("could not create config file: %s", err)
	}
	defer f.Close()

	var metaData []helpers.Cluster
	byteValue, _ := io.ReadAll(f)
	err = json.Unmarshal(byteValue, &metaData)
	if err != nil {
		return fmt.Errorf("error marshaling data: %s", err)
	}

	id := resData.Get("id").(string)
	for _,v := range metaData {
		if v.Id == id {
			resData.Set("name", v.Name)
			resData.Set("replicas", v.Replicas)
			resData.Set("ports", v.Ports)
		}
	}

	return nil
}


func clusterUpdateItem(resData *schema.ResourceData, m interface{}) error {

	id := resData.Get("id").(string)
	name := resData.Get("name").(string)
	replicas := resData.Get("replicas").(int)
	ports := resData.Get("ports").([]int)

	f, err := os.OpenFile(fmt.Sprint(helpers.KafkaDir, "/clusterdata.json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("could not create config file: %s", err)
	}
	defer f.Close()

	var metaData []helpers.Cluster
	byteValue, _ := io.ReadAll(f)
	err = json.Unmarshal(byteValue, &metaData)
	if err != nil {
		return fmt.Errorf("error marshaling data: %s", err)
	}

	for _,v := range metaData {
		if v.Id == id {

			err = helpers.UpdateCluster(resData, v)
			if err != nil {
				return fmt.Errorf("cannot update cluster: %s", err)
			}

			v.Name = name
			v.Replicas = replicas
			v.Ports = ports
		}
	}

	return nil
}
