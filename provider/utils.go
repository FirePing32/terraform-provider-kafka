package provider

import (
	"fmt"
	"os"
	"regexp"
	"net/http"
	"io"
	"os/exec"
	"net"
	"strconv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func validateName(v interface{}, k string) (ws []string, es []error) {
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

func validateReplicas(v interface{}, k string) (ws []string, es []error) {
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

func validatePorts(v interface{}, k string) (ws []string, es []error) {
	var errs []error
	var warns []string
	value, ok := v.(int)
	if !ok {
		errs = append(errs, fmt.Errorf("expected replica count to be integer"))
		return warns, errs
	}
	if value < 1024 || value > 49151 {
		errs = append(errs, fmt.Errorf("port number should be between 1024 and 49151"))
		return warns, errs
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

func downloadKafka() error {

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

func setupKafka(d *schema.ResourceData) error {
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

	downloadkafka := downloadKafka()
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
