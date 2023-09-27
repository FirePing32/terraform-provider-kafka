package helpers

type Cluster struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Replicas int    `json:"replicas"`
	Ports    []int  `json:"ports"`
}
