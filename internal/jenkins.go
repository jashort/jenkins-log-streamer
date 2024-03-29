package jenkins

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func jobStatusUrl(url string) string {
	return url + "/lastBuild/api/json"
}

func jobLogUrl(url string, start int64) string {
	return fmt.Sprintf("%s/lastBuild/logText/progressiveText?start=%d", url, start)
}

func FetchJobStatus(server ServerInfo) (*JobStatus, error) {
	jobStatus := new(JobStatus)
	err := getJson(server, jobStatus)
	if err != nil {
		return nil, err
	}
	return jobStatus, nil
}

func fetchLogChunk(server ServerInfo, position int64) *http.Response {
	logUrl := jobLogUrl(server.JobBaseUrl, position)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", logUrl, nil)
	req.Header.Add("Authorization", "Basic "+basicAuth(server.User, server.Token))
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		fmt.Printf("%s %s\n", req.Method, req.URL)
		log.Fatalf("Jenkins responded with status %d, expecting 200. Bad credentials?", resp.StatusCode)
	}
	return resp
}

type LogChunk struct {
	Body        string
	BuildNumber int
	Start       int64
	MoreData    bool
	NewPosition int64
}

func FetchLog(server ServerInfo, start int64) LogChunk {
	resp := fetchLogChunk(server, start)
	buf := processLogChunk(resp)

	moreData, err := strconv.ParseBool(resp.Header.Get("X-More-Data"))
	if err != nil {
		moreData = false
	}

	newPosition, err := strconv.ParseInt(resp.Header.Get("X-Text-Size"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return LogChunk{
		Body:        buf,
		BuildNumber: 0,
		Start:       start,
		MoreData:    moreData,
		NewPosition: newPosition,
	}
}

func processLogChunk(resp *http.Response) string {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanRunes)
	var buf bytes.Buffer
	for scanner.Scan() {
		buf.WriteString(scanner.Text())
	}
	return buf.String()
}

func getJson(server ServerInfo, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", jobStatusUrl(server.JobBaseUrl), nil)
	req.Header.Add("Authorization", "Basic "+basicAuth(server.User, server.Token))
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Printf("%s %s\n", req.Method, req.URL)
		log.Fatalf("Jenkins responded with status %d, expecting 200. Bad credentials?", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

type ServerInfo struct {
	JobBaseUrl string
	User       string
	Token      string
}

type JobStatus struct {
	Class   string `json:"_class"`
	Actions []struct {
		Class  string `json:"_class,omitempty"`
		Causes []struct {
			Class            string `json:"_class"`
			ShortDescription string `json:"shortDescription"`
		} `json:"causes,omitempty"`
		BuildsByBranchName struct {
			Master struct {
				Class       string      `json:"_class"`
				BuildNumber int         `json:"buildNumber"`
				BuildResult interface{} `json:"buildResult"`
				Marked      struct {
					SHA1   string `json:"SHA1"`
					Branch []struct {
						SHA1 string `json:"SHA1"`
						Name string `json:"name"`
					} `json:"branch"`
				} `json:"marked"`
				Revision struct {
					SHA1   string `json:"SHA1"`
					Branch []struct {
						SHA1 string `json:"SHA1"`
						Name string `json:"name"`
					} `json:"branch"`
				} `json:"revision"`
			} `json:"master"`
		} `json:"buildsByBranchName,omitempty"`
		LastBuiltRevision struct {
			SHA1   string `json:"SHA1"`
			Branch []struct {
				SHA1 string `json:"SHA1"`
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"lastBuiltRevision,omitempty"`
		RemoteUrls []string `json:"remoteUrls,omitempty"`
		ScmName    string   `json:"scmName,omitempty"`
	} `json:"actions"`
	Artifacts         []interface{} `json:"artifacts"`
	Building          bool          `json:"building"`
	Description       interface{}   `json:"description"`
	DisplayName       string        `json:"displayName"`
	Duration          int           `json:"duration"`
	EstimatedDuration int           `json:"estimatedDuration"`
	Executor          interface{}   `json:"executor"`
	FullDisplayName   string        `json:"fullDisplayName"`
	Id                string        `json:"id"`
	KeepLog           bool          `json:"keepLog"`
	Number            int           `json:"number"`
	QueueId           int           `json:"queueId"`
	Result            string        `json:"result"`
	Timestamp         int64         `json:"timestamp"`
	Url               string        `json:"url"`
	ChangeSets        []struct {
		Class string `json:"_class"`
		Items []struct {
			Class         string   `json:"_class"`
			AffectedPaths []string `json:"affectedPaths"`
			CommitId      string   `json:"commitId"`
			Timestamp     int64    `json:"timestamp"`
			Author        struct {
				AbsoluteUrl string `json:"absoluteUrl"`
				FullName    string `json:"fullName"`
			} `json:"author"`
			AuthorEmail string `json:"authorEmail"`
			Comment     string `json:"comment"`
			Date        string `json:"date"`
			Id          string `json:"id"`
			Msg         string `json:"msg"`
			Paths       []struct {
				EditType string `json:"editType"`
				File     string `json:"file"`
			} `json:"paths"`
		} `json:"items"`
		Kind string `json:"kind"`
	} `json:"changeSets"`
	Culprits []struct {
		AbsoluteUrl string `json:"absoluteUrl"`
		FullName    string `json:"fullName"`
	} `json:"culprits"`
	InProgress    bool        `json:"inProgress"`
	NextBuild     interface{} `json:"nextBuild"`
	PreviousBuild struct {
		Number int    `json:"number"`
		Url    string `json:"url"`
	} `json:"previousBuild"`
}
