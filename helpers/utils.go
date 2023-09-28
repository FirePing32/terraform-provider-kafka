package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func ValidateName(v interface{}, k string) (ws []string, es []error) {
	var errs []error
	var warns []string
	value, ok := v.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("expected name to be string"))
		return warns, errs
	}
	whiteSpace := regexp.MustCompile(`\s+`)
	if whiteSpace.Match([]byte(value)) {
		errs = append(errs, fmt.Errorf("name cannot contain whitespace. Got %s", value))
		return warns, errs
	}
	return warns, errs
}

func ValidateReplicas(v interface{}, k string) (ws []string, es []error) {
	var errs []error
	var warns []string
	value, ok := v.(int)
	if !ok {
		errs = append(errs, fmt.Errorf("expected replica count to be integer"))
		return warns, errs
	}
	if value < 1 || value > 50 {
		errs = append(errs, fmt.Errorf("replica count should be between 1 and 50"))
		return warns, errs
	}
	return warns, errs
}

func ValidatePorts(v interface{}, k string) (ws []string, es []error) {
	var errs []error
	var warns []string
	value, ok := v.([]int)
	for _, v := range value {
		if !ok {
			errs = append(errs, fmt.Errorf("expected port number to be integer"))
			return warns, errs
		}
		if v < 1024 || v > 49151 {
			errs = append(errs, fmt.Errorf("port number should be between 1024 and 49151"))
			return warns, errs
		}
	}
	return warns, errs
}

func createHomeDir(dirname string) (string, error) {
	dir, err := os.UserHomeDir()

	if err != nil {
		return "error", fmt.Errorf("%s", err)
	}

	dirPath := fmt.Sprint(dir, "/", dirname)
	if err := os.Mkdir(dirPath, os.ModePerm); err != nil {
		return "error", fmt.Errorf("%s", err)
	}

	return dirPath, nil
}

func downloadBinary(dirPath string) error {
	out, err := os.Create(fmt.Sprint(dirPath, "kafka.tgz"))
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	defer out.Close()

	resp, err := http.Get(KafkaDownloadUri)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	defer resp.Body.Close()

	_, error := io.Copy(out, resp.Body)
	if error != nil {
		return fmt.Errorf("%s", err)
	}

	return nil
}

func DownloadKafka() error {

	if _, err := os.Stat(KafkaDir); os.IsNotExist(err) {
		dirPath, err := createHomeDir(".kafka")
		if err != nil {
			return fmt.Errorf("%s", err)
		}

		downloadbinary := downloadBinary(dirPath)
		if downloadbinary != nil {
			return fmt.Errorf("%s", downloadbinary)
		}

	} else {
		if _, err := os.Stat(fmt.Sprint(KafkaDir, "/kafka.tgz")); err != nil {
			downloadbinary := downloadBinary(KafkaDir)
			if downloadbinary != nil {
				return fmt.Errorf("%s", downloadbinary)
			}
		}
	}

	return nil
}

func SetupKafka(d *schema.ResourceData) error {
	for _, port := range d.Get("ports").([]int) {
		conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", strconv.Itoa(port)), 5000)
		if conn != nil {
			conn.Close()
			return fmt.Errorf("port: %d is aleady in use. Please use a different port", port)
		}
	}
	_, javaerr := exec.Command("/bin/bash", "./../scripts/installJava.sh").Output()

	if javaerr != nil {
		return fmt.Errorf("error: %s", javaerr)
	}

	downloadkafka := DownloadKafka()
	if downloadkafka != nil {
		return fmt.Errorf("error: %s", downloadkafka)
	}

	if _, err := os.Stat(fmt.Sprint(KafkaDir, "/kafka")); err != nil {
		_, kafkaerr := exec.Command("/bin/bash", "./../scripts/installKafka.sh").Output()
		if kafkaerr != nil {
			return fmt.Errorf("error: %s", kafkaerr)
		}
	}

	return nil
}

func StartKafka(d *schema.ResourceData) error {
	replicas := d.Get("replicas").(int)
	ports := d.Get("ports").([]int)

	zk, createerror := os.Create(fmt.Sprint(KafkaDir, "/kafka/config/zookeeper.properties"))
	check(createerror)
	defer zk.Close()
	_, err := zk.WriteString(zkprop)
	check(err)

	clusterdata, err := os.Create(fmt.Sprint(KafkaDir, "/clusterdata.json"))
	check(err)
	_, err = clusterdata.WriteString("[]")
	check(err)

	_, zooerr := exec.Command(fmt.Sprint(KafkaDir, "/kafka/bin/zookeeper-server-start.sh ", KafkaDir, "/kafka/config/zookeeper.properties")).Output()
	if zooerr != nil {
		return fmt.Errorf("error: %s", zooerr)
	}

	for i := 0; i < replicas; i++ {
		server, createerror := os.Create(fmt.Sprintf("%s/kafka/config/server-%d.properties", KafkaDir, i))
		check(createerror)
		defer server.Close()
		_, serverwriteerr := server.WriteString(fmt.Sprintf(serverProp, i, ports[i], i))
		check(serverwriteerr)

		_, kafkaerr := exec.Command(fmt.Sprintf("%s/kafka/bin/kafka-server-start.sh %s/kafka/config/server-%d.properties", KafkaDir, KafkaDir, i)).Output()
		if kafkaerr != nil {
			return fmt.Errorf("error: %s", kafkaerr)
		}
	}

	storeClusterData := storeClusterMetadata(d)
	if storeClusterData != nil {
		return fmt.Errorf("error storing cluster metadata: %s", storeClusterData)
	}

	return nil
}

func storeClusterMetadata(d *schema.ResourceData) error {

	id := d.Id()
	name := d.Get("name").(string)
	replicas := d.Get("replicas").(int)
	ports := d.Get("ports").([]int)

	f, err := os.OpenFile(fmt.Sprint(KafkaDir, "/clusterdata.json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("could not create config file: %s", err)
	}
	defer f.Close()

	var metaData []Cluster
	byteValue, _ := io.ReadAll(f)
	err = json.Unmarshal(byteValue, &metaData)
	if err != nil {
		return fmt.Errorf("error marshaling data: %s", err)
	}

	metaData = append(metaData, Cluster{Id: id, Name: name, Replicas: replicas, Ports: ports})

	marshalData, err := json.Marshal(metaData)
	if err != nil {
		return fmt.Errorf("error marshaling data: %s", err)
	}
	_, err = f.Write(marshalData)
	if err != nil {
		return fmt.Errorf("could not write config file: %s", err)
	}

	return nil
}

func UpdateCluster(d *schema.ResourceData, metadata Cluster) error {

	replicas := d.Get("replicas").(int)
	ports := d.Get("ports").([]int)

	if len(ports) != replicas {
		return fmt.Errorf("number of ports does not match the number of replicas")
	}

	if metadata.Replicas > replicas {
		return nil
	} else if metadata.Replicas < replicas {
		return nil
	}

	return nil
}
