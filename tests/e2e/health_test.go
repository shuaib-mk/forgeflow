//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestRunningStackIsHealthy(t *testing.T){baseURL:=os.Getenv("FORGEFLOW_E2E_URL");if baseURL==""{baseURL="http://localhost:8080"};response,err:=http.Get(baseURL+"/health");if err!=nil{t.Fatal(err)};defer response.Body.Close();if response.StatusCode!=http.StatusOK{t.Fatalf("status=%s",response.Status)};var payload map[string]string;if err:=json.NewDecoder(response.Body).Decode(&payload);err!=nil{t.Fatal(err)};if payload["status"]!="ok"{t.Fatalf("payload=%v",payload)}}

