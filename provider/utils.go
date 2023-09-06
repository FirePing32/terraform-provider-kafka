package provider

import (
	"fmt"
	"os"
	"regexp"
	"net/http"
	"io"
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

func downloadKafka() error {

	dirPath, err := createHomeDir(".kafka")
	if err != nil {
    	return fmt.Errorf("%s", err)
    }

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
