package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"tf-cloud-opa-server/lib"

	"github.com/k0kubun/pp"
	"github.com/open-policy-agent/opa/rego"
)

func (rh requestHandler) downloadPlanJSON(ctx context.Context, url, token string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := rh.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[ERROR] downloadPlanJSON status: %d, msg: %s", resp.StatusCode, body)
	}
	return body, nil
}

func (rh requestHandler) sendCallback(ctx context.Context, url, token, message string, passed bool) ([]byte, error) {
	status := "failed"
	if passed {
		status = "passed"
	}
	jsonData, _ := json.Marshal(&lib.TFCloudRunTasksResponse{
		Data: &lib.TFCloudRunTasksResponseData{
			Type: "task-results",
			Attributes: &lib.TFCloudRunTasksResponseAttributes{
				Status:  status,
				Message: message,
			},
		},
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := rh.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[ERROR] sendCallback status: %d, msg: %s", resp.StatusCode, body)
	}
	return body, nil
}

func (rh requestHandler) createComment(ctx context.Context, runID, token, message string) ([]byte, error) {
	jsonData, _ := json.Marshal(&lib.TFCloudComment{
		Data: &lib.TFCloudCommentData{
			Type: "task-results",
			Attributes: &lib.TFCloudCommentAttributes{
				Body: message,
			},
		},
	})
	url := "https://app.terraform.io/api/v2/runs/" + runID + "/comments"
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := rh.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("[ERROR] status: %d, msg: %s", resp.StatusCode, body)
	}
	return body, nil
}

func (rh requestHandler) evalRegoPolicies(ctx context.Context, req *lib.TFCloudRunTasksRequest) {
	var planJSON interface{}
	jsonByte, err := rh.downloadPlanJSON(ctx, req.PlanJSONAPIURL, req.AccessToken)
	if err != nil {
		log.Println(err)
		return
	}
	_ = json.Unmarshal(jsonByte, &planJSON)

	workspace := strings.ReplaceAll(req.WorkspaceName, "-", "_")
	q := rego.New(
		rego.Query("data.workspace."+workspace),
		rego.Load([]string{"./rego"}, nil),
		rego.Input(planJSON),
	)

	rs, err := q.Eval(ctx)
	if err != nil {
		log.Printf("[ERROR] failed to eval policies: %s", err)
		return
	}
	// debug
	pp.Println(rs)

	allowed := false
	message := "N/A"
	if len(rs) == 0 {
		allowed = true
		message = "ResultSet was empty. Make sure 'data.workspace." + workspace + "' package exists."
	} else if v, ok := rs[0].Expressions[0].Value.(map[string]interface{})["allow"]; ok && v == true {
		allowed = true
	} else {
		// due to a limitation in TFC, we cannot have more than 512 characters
		message = "Failed. Please check the Run comment below."
	}

	{
		// NOTE: One workaround to the above limit would be to create a comment on the current Run context.
		// But since the received 'req.AccessToken' doesn't have permission to create a comment,
		// you will need to provide your own TFC API token instead.
		if token := os.Getenv("TFC_API_KEY"); token != "" {
			pp.ColoringEnabled = false
			_, err = rh.createComment(
				ctx,
				req.RunID,
				token,
				"```go"+pp.Sprintln(rs)+"```",
			)
			if err != nil {
				log.Printf("[ERROR] failed to create a comment: %s", err)
			}
		}
	}

	_, err = rh.sendCallback(
		ctx,
		req.TaskResultCallbackURL,
		req.AccessToken,
		message, // this will be available on the UI
		allowed,
	)
	if err != nil {
		log.Println(err)
		return
	}
}

type requestHandler struct {
	client *http.Client
}

func (rh requestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	TFCloudRunTasksRequest := new(lib.TFCloudRunTasksRequest)
	err = json.Unmarshal(body, &TFCloudRunTasksRequest)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	// TF Cloud sends a test POST request upon initial registration
	// we just need to return empty 200 for verification
	if TFCloudRunTasksRequest.AccessToken == "test-token" {
		return
	}

	ctx := context.Background()
	go rh.evalRegoPolicies(ctx, TFCloudRunTasksRequest)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}
	c := lib.NewRetryHttpClient()
	rh := requestHandler{
		client: c,
	}

	mux := http.NewServeMux()
	mux.Handle("/run-opa", lib.ValidateSignature(rh))
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatalf("error starting server: %e\n", err)
	}
}
