package helpers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"encoding/json"
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
	if value < 1 || value > 50  {
		errs = append(errs, fmt.Errorf("replica count should be between 1 and 50"))
		return warns, errs
	}
	return warns, errs
}

func ValidatePorts(v interface{}, k string) (ws []string, es []error) {
	var errs []error
	var warns []string
	value, ok := v.([]int)
	for _,v := range value {
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

	resp, err := http.Get(kafkaDownloadUri)
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

	if _, err := os.Stat(kafkaDir); os.IsNotExist(err) {
      	dirPath, err := createHomeDir(".kafka")
		if err != nil {
			return fmt.Errorf("%s", err)
		}

		downloadbinary := downloadBinary(dirPath)
		if downloadbinary != nil {
			return fmt.Errorf("%s", downloadbinary)
		}

    } else {
		if _, err := os.Stat(fmt.Sprint(kafkaDir, "/kafka.tgz")); err != nil {
			downloadbinary := downloadBinary(kafkaDir)
			if downloadbinary != nil {
				return fmt.Errorf("%s", downloadbinary)
			}
		}
	}

	return nil
}

func SetupKafka(d *schema.ResourceData) error {
	for _,port := range d.Get("ports").([]int) {
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

	if _, err := os.Stat(fmt.Sprint(kafkaDir, "/kafka")); err != nil {
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

	zk, createerror := os.Create("$HOME/.kafka/kafka/config/zookeeper.properties")
	check(createerror)
	defer zk.Close()
	_, zoowriteerr := zk.WriteString(zkprop)
	check(zoowriteerr)

	_, zooerr := exec.Command("$HOME/.kafka/kafka/bin/zookeeper-server-start.sh $HOME/.kafka/kafka/config/zookeeper.properties").Output()
	if zooerr != nil {
		return fmt.Errorf("error: %s", zooerr)
	}

	for i := 0; i < replicas; i++ {
		server, createerror := os.Create(fmt.Sprintf("$HOME/.kafka/kafka/config/server-%d.properties", i))
		check(createerror)
		defer server.Close()
		_, serverwriteerr := server.WriteString(fmt.Sprintf(serverProp, i, ports[i], i))
		check(serverwriteerr)

		_, kafkaerr := exec.Command(fmt.Sprintf("$HOME/.kafka/kafka/bin/kafka-server-start.sh $HOME/.kafka/kafka/config/server-%d.properties", i)).Output()
		if kafkaerr != nil {
			return fmt.Errorf("error: %s", kafkaerr)
		}
	}

	return nil
}

func storeClusterMetadata(d *schema.ResourceData) error {
	clusterData := map[string]interface{} {
		"name": d.Get("name"),
		"replicas": d.Get("replicas"),
		"ports": d.Get("ports"),
	}
	jsonMarshal, err := json.Marshal(clusterData)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	_, err = os.OpenFile("$HOME/.kafka/clusterdata.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		_, err = os.Create("$HOME/.kafka/clusterdata.json")
		if err != nil {
			return fmt.Errorf("could not create config file: %s", err)
		}
	}
	writeerr := os.WriteFile("$HOME/.kafka/clusterdata.json", jsonMarshal, 0644)

	if writeerr != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	return nil
}
