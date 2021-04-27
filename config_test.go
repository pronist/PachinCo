package upbit

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var defaultConfig = config{
	Bot: struct {
		AccessKey string
		SecretKey string
	} {
		"UyGWYAEVN3PRDDo3Y3pJnV6DWn69k17gVs1X47p4",
		"2FjMz4yBOuHqzpwGUdkEu0WJF5g30Z8Wx71cJbxn",
	},
}

func TestNewConfig(t *testing.T) {
	if jsonStr, err := json.Marshal(defaultConfig); err == nil {
		if file, err := ioutil.TempFile("", "configor"); err == nil {
			defer os.Remove(file.Name())

			if _, err = file.Write(jsonStr); err != nil {
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}

			conf := NewConfig(file.Name())

			if !reflect.DeepEqual(*conf, defaultConfig) {
				t.Fail()
			}
		}
	} else {
		t.Fatal(err)
	}
}
