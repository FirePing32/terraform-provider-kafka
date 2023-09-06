package provider

import (
	"fmt"
	"os/exec"
	"net"
	"strconv"
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

func clusterCreateItem(d *schema.ResourceData, m interface{}) error {
	for _,port := range d.Get("ports").([]int) {
		conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", strconv.Itoa(port)), 5000)
		if conn != nil {
			conn.Close()
			return fmt.Errorf("port: %d is aleady in use. Please use a different port", port)
		}
	}
	_, err := exec.Command("/bin/bash", "./../scripts/installJava.sh").Output()

    if err != nil {
    	return fmt.Errorf("error: %s", err)
    }



    return nil
}

